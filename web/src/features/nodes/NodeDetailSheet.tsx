import { ExternalLink, KeyRound, TriangleAlert } from 'lucide-react'
import { useState } from 'react'
import type { ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import type { AgentToken, AuthedRequest, Metric, Node } from '../../types'
import { APIError } from '../../lib/api'
import { formatBytes, formatDuration, formatTimestamp, percentOf, relativeTime } from '../../lib/format'
import { thresholds } from '../../lib/thresholds'
import { Button } from '../../components/ui/Button'
import { ConfirmDialog } from '../../components/ui/ConfirmDialog'
import { EmptyState } from '../../components/ui/EmptyState'
import { Meter } from '../../components/ui/Meter'
import { Modal } from '../../components/ui/Modal'
import { Pill, StatusBadge } from '../../components/ui/StatusBadge'
import { TokenReveal } from './TokenReveal'

type Props = { node: Node | null; authedRequest: AuthedRequest; onClose: () => void }

export function NodeDetailSheet({ node, authedRequest, onClose }: Props) {
  return (
    <Modal
      variant="sheet"
      open={node !== null}
      title={node?.name ?? ''}
      description={node ? <span className="font-mono text-xs">{node.hostname || node.id}</span> : undefined}
      onClose={onClose}
      footer={
        node && (
          <a className="btn btn-primary btn-sm gap-2" href={node.grafana_url} target="_blank" rel="noreferrer">
            Open in Grafana
            <ExternalLink size={14} />
          </a>
        )
      }
    >
      {node && <NodeDetail key={node.id} node={node} authedRequest={authedRequest} />}
    </Modal>
  )
}

function Section({ title, aside, children }: { title: string; aside?: ReactNode; children: ReactNode }) {
  return (
    <section className="border-t border-base-300 pt-5 first:border-t-0 first:pt-0">
      <div className="mb-4 flex items-center justify-between gap-3">
        <h3 className="text-sm font-semibold">{title}</h3>
        {aside}
      </div>
      {children}
    </section>
  )
}

function Fact({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="min-w-0">
      <dt className="text-xs text-base-content/50">{label}</dt>
      <dd className="mt-0.5 truncate text-sm tabular-nums">{children}</dd>
    </div>
  )
}

function NodeDetail({ node, authedRequest }: { node: Node; authedRequest: AuthedRequest }) {
  const queryClient = useQueryClient()
  const [issuedToken, setIssuedToken] = useState('')
  const [issuing, setIssuing] = useState(false)
  const [tokenError, setTokenError] = useState('')
  const [revoking, setRevoking] = useState<AgentToken | null>(null)

  const latest = useQuery({
    queryKey: ['node', node.id, 'latest-metric'],
    queryFn: () => authedRequest<Metric>(`/api/v1/nodes/${node.id}/metrics/latest`),
    refetchInterval: 15_000,
    retry: false,
  })
  const tokens = useQuery({
    queryKey: ['node', node.id, 'tokens'],
    queryFn: async () => (await authedRequest<{ tokens: AgentToken[] }>(`/api/v1/nodes/${node.id}/tokens`)).tokens,
  })

  async function issueToken() {
    setIssuing(true)
    setTokenError('')
    try {
      const label = `console-${new Date().toISOString().slice(0, 10)}`
      const result = await authedRequest<{ token: string }>(`/api/v1/nodes/${node.id}/tokens`, { method: 'POST', body: JSON.stringify({ label }) })
      setIssuedToken(result.token)
      await queryClient.invalidateQueries({ queryKey: ['node', node.id, 'tokens'] })
    } catch (caught) {
      setTokenError(caught instanceof Error ? caught.message : 'Could not issue a token')
    } finally {
      setIssuing(false)
    }
  }

  async function revoke(token: AgentToken) {
    await authedRequest(`/api/v1/nodes/${node.id}/tokens/${token.id}`, { method: 'DELETE' })
    await queryClient.invalidateQueries({ queryKey: ['node', node.id, 'tokens'] })
  }

  const metric = latest.data
  const waiting = latest.error instanceof APIError && latest.error.status === 404

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center gap-2 text-sm text-base-content/60">
        <StatusBadge status={node.status} />
        <span>Last seen {relativeTime(node.last_seen_at)}</span>
        <span aria-hidden="true">·</span>
        <span>{node.environment}</span>
      </div>

      <Section title="Latest sample" aside={metric && <span className="text-xs text-base-content/50">recorded {relativeTime(metric.recorded_at)}</span>}>
        {latest.isLoading ? (
          <div className="grid gap-4 sm:grid-cols-3">
            {[1, 2, 3].map((value) => <div key={value} className="skeleton h-16" />)}
          </div>
        ) : waiting || !metric ? (
          <EmptyState
            title={waiting ? 'Waiting for the first sample' : 'Metrics unavailable'}
            text={waiting ? 'Install the agent with a token below; samples arrive every 15 seconds.' : latest.error?.message ?? 'Try refreshing.'}
          />
        ) : (
          <div className="space-y-6">
            <div className="grid gap-5 sm:grid-cols-3">
              <Meter label="CPU" percent={metric.cpu_percent} detail={`Alerts above ${thresholds.cpu.warning}% for 10 min`} warning={thresholds.cpu.warning} />
              <Meter
                label="Memory"
                percent={percentOf(metric.memory_used_bytes, metric.memory_total_bytes)}
                detail={`${formatBytes(metric.memory_used_bytes)} of ${formatBytes(metric.memory_total_bytes)}`}
                warning={thresholds.memory.warning}
              />
              <Meter
                label="Disk"
                percent={percentOf(metric.disk_used_bytes, metric.disk_total_bytes)}
                detail={`${formatBytes(metric.disk_used_bytes)} of ${formatBytes(metric.disk_total_bytes)}`}
                warning={thresholds.disk.warning}
                critical={thresholds.disk.critical}
              />
            </div>
            <dl className="grid grid-cols-2 gap-x-6 gap-y-4 sm:grid-cols-3">
              <Fact label="Load 1 / 5 / 15">
                {[metric.load_1, metric.load_5, metric.load_15].map((value) => value?.toFixed(2) ?? '—').join(' / ')}
              </Fact>
              <Fact label="Uptime">{formatDuration(metric.uptime_seconds)}</Fact>
              <Fact label="Network RX / TX">{`${formatBytes(metric.network_rx_bytes)} / ${formatBytes(metric.network_tx_bytes)}`}</Fact>
            </dl>
            {metric.failed_collectors.length > 0 && (
              <p className="flex items-start gap-2 rounded-field border border-warning/30 bg-warning/10 px-3 py-2 text-sm">
                <TriangleAlert size={15} className="mt-0.5 shrink-0" />
                Partial sample: {metric.failed_collectors.join(', ')} could not be collected.
              </p>
            )}
          </div>
        )}
      </Section>

      <Section title="Host">
        <dl className="grid grid-cols-2 gap-x-6 gap-y-4">
          <Fact label="Agent version"><span className="font-mono text-xs">{node.agent_version}</span></Fact>
          <Fact label="Operating system">{node.os_name ?? '—'}</Fact>
          <Fact label="Kernel">{node.kernel_version ?? '—'}</Fact>
          <Fact label="Registered">{formatTimestamp(node.created_at)}</Fact>
          <Fact label="Node ID"><span className="font-mono text-xs">{node.id}</span></Fact>
        </dl>
      </Section>

      <Section
        title="Agent tokens"
        aside={
          !issuedToken && (
            <Button size="sm" onClick={() => void issueToken()} disabled={issuing}>
              {issuing ? <span className="loading loading-spinner loading-xs" /> : <KeyRound size={14} />}
              Issue token
            </Button>
          )
        }
      >
        {issuedToken && (
          <div className="mb-4 space-y-3">
            <TokenReveal token={issuedToken} />
            <div className="flex justify-end">
              <Button size="sm" onClick={() => setIssuedToken('')}>I saved it</Button>
            </div>
          </div>
        )}
        {tokenError && <p className="mb-3 text-sm text-error" role="alert">{tokenError}</p>}
        {tokens.data && tokens.data.length > 0 ? (
          <ul className="divide-y divide-base-300 rounded-field border border-base-300">
            {tokens.data.map((token) => (
              <li key={token.id} className="flex items-center gap-3 px-3 py-2.5">
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{token.label}</p>
                  <p className="text-xs text-base-content/50">
                    Created {formatTimestamp(token.created_at)} · {token.last_used_at ? `last used ${relativeTime(token.last_used_at)}` : 'never used'}
                  </p>
                </div>
                <Pill tone={token.revoked_at ? 'neutral' : 'success'}>{token.revoked_at ? 'Revoked' : 'Active'}</Pill>
                {!token.revoked_at && (
                  <Button size="sm" variant="ghost" className="text-error" onClick={() => setRevoking(token)}>
                    Revoke
                  </Button>
                )}
              </li>
            ))}
          </ul>
        ) : (
          !tokens.isLoading && <p className="text-sm text-base-content/55">No tokens have been issued for this node.</p>
        )}
        <p className="mt-3 text-xs leading-5 text-base-content/50">
          To rotate without a monitoring gap: issue a new token, update <code className="font-mono">/etc/vpsentinel/agent.env</code>, confirm a new
          sample arrives, then revoke the old token.
        </p>
      </Section>

      <ConfirmDialog
        open={revoking !== null}
        title="Revoke token?"
        description={
          <>
            Agents using <strong>{revoking?.label}</strong> will be rejected immediately and <strong>{node.name}</strong> will go offline unless another
            active token is installed.
          </>
        }
        confirmLabel="Revoke token"
        onConfirm={() => (revoking ? revoke(revoking) : Promise.resolve())}
        onClose={() => setRevoking(null)}
      />
    </div>
  )
}
