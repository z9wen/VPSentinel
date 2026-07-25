import { ExternalLink, KeyRound, Trash2 } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import type { AgentToken, AuthedRequest, Metric, Node } from '../../types'
import { formatBytes, formatDuration, formatPercent } from '../../lib/format'
import { EmptyState } from '../../components/ui/EmptyState'
import { Modal } from '../../components/ui/Modal'
import { StatusBadge } from '../../components/ui/StatusBadge'
import { Button } from '../../components/ui/Button'

function MetricCard({ label, value }: { label: string; value: string }) {
  return <div className="flex min-h-28 flex-col justify-between rounded-box border border-base-300 bg-base-200 p-4"><span className="text-xs text-base-content/45">{label}</span><strong className="font-mono text-lg font-medium">{value}</strong></div>
}

export function NodeDetailDialog({ node, authedRequest, onClose }: { node: Node | null; authedRequest: AuthedRequest; onClose: () => void }) {
  const [metric, setMetric] = useState<Metric | null>(null)
	const [tokens, setTokens] = useState<AgentToken[]>([])
  const [error, setError] = useState('')
	const loadTokens = useCallback(async () => {
		if (!node) return
		const result = await authedRequest<{ tokens: AgentToken[] }>(`/api/v1/nodes/${node.id}/tokens`)
		setTokens(result.tokens)
	}, [authedRequest, node])
  useEffect(() => {
    if (!node) return
    setMetric(null)
    setError('')
    authedRequest<Metric>(`/api/v1/nodes/${node.id}/metrics/latest`).then(setMetric).catch((caught) => setError(caught instanceof Error ? caught.message : 'No metrics available'))
		void loadTokens()
  }, [authedRequest, loadTokens, node])
  if (!node) return null
  const nodeID = node.id
  const percent = (used: number | null, total: number | null) => used != null && total ? (used / total) * 100 : null
  async function revoke(tokenID: number) {
		await authedRequest(`/api/v1/nodes/${nodeID}/tokens/${tokenID}`, { method: 'DELETE' })
		await loadTokens()
	}
	return <Modal open title={node.name} onClose={onClose} wide><div className="mb-5 flex flex-wrap items-center gap-3 text-xs text-base-content/50"><StatusBadge status={node.status} /><span>{node.hostname}</span><span>{node.environment}</span></div>{error && !metric ? <EmptyState title="Waiting for first sample" text={error} /> : <div className="mb-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-3"><MetricCard label="CPU" value={formatPercent(metric?.cpu_percent)} /><MetricCard label="Memory" value={formatPercent(percent(metric?.memory_used_bytes ?? null, metric?.memory_total_bytes ?? null))} /><MetricCard label="Disk" value={formatPercent(percent(metric?.disk_used_bytes ?? null, metric?.disk_total_bytes ?? null))} /><MetricCard label="Load 1 / 5 / 15" value={metric ? [metric.load_1, metric.load_5, metric.load_15].map((value) => value?.toFixed(2) ?? '—').join(' / ') : '—'} /><MetricCard label="Uptime" value={formatDuration(metric?.uptime_seconds)} /><MetricCard label="Network RX / TX" value={metric ? `${formatBytes(metric.network_rx_bytes)} / ${formatBytes(metric.network_tx_bytes)}` : '—'} /></div>}<section className="mb-5 rounded-box border border-base-300"><header className="flex items-center gap-2 border-b border-base-300 px-4 py-3 text-sm font-semibold"><KeyRound size={15} className="text-primary" />Agent tokens</header>{tokens.length === 0 ? <p className="p-4 text-xs text-base-content/45">No tokens have been created.</p> : tokens.map((token) => <div className="flex items-center gap-3 border-b border-base-300 px-4 py-3 text-xs last:border-0" key={token.id}><div><strong>{token.label}</strong><p className="mt-1 text-base-content/40">Last used: {token.last_used_at ? new Date(token.last_used_at).toLocaleString() : 'Never'}</p></div><span className={`badge badge-xs ml-auto ${token.revoked_at ? 'badge-ghost' : 'badge-success'}`}>{token.revoked_at ? 'revoked' : 'active'}</span>{!token.revoked_at && <Button variant="danger" size="sm" className="btn-square" title="Revoke token" onClick={() => void revoke(token.id)}><Trash2 size={14} /></Button>}</div>)}</section><a className="btn btn-primary w-full" href={node.grafana_url} target="_blank">View historical dashboard <ExternalLink size={15} /></a></Modal>
}
