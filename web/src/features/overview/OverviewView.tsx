import { ExternalLink } from 'lucide-react'
import type { Alert, Node, NodeStatus } from '../../types'
import { relativeTime } from '../../lib/format'
import { toneDot, type Tone } from '../../lib/tone'
import { EmptyState } from '../../components/ui/EmptyState'
import { Panel } from '../../components/ui/Panel'
import { StatusBadge } from '../../components/ui/StatusBadge'
import { AlertRow } from '../alerts/AlertRow'

type Props = {
  nodes: Node[]
  alerts: Alert[]
  activeAlerts: Alert[]
  nodeName: (id: string | null) => string
  onNode: (node: Node) => void
}

const statusOrder: { status: NodeStatus; tone: Tone }[] = [
  { status: 'online', tone: 'success' },
  { status: 'degraded', tone: 'warning' },
  { status: 'offline', tone: 'error' },
  { status: 'pending', tone: 'neutral' },
]

function Stat({ label, value, detail, tone }: { label: string; value: number; detail: string; tone?: Tone }) {
  return (
    <div className="bg-base-100 px-4 py-4 sm:px-5">
      <div className="flex items-center gap-2 text-sm text-base-content/60">
        {tone && <span className={`size-2 rounded-full ${toneDot[tone]}`} aria-hidden="true" />}
        {label}
      </div>
      <div className="mt-2 text-3xl font-semibold tabular-nums tracking-tight">{value}</div>
      <div className="mt-1 truncate text-xs text-base-content/50">{detail}</div>
    </div>
  )
}

// FleetBar shows the share of nodes in each state as one segmented bar.
function FleetBar({ counts, total }: { counts: Record<NodeStatus, number>; total: number }) {
  if (total === 0) return null
  return (
    <div className="flex items-center gap-4 px-4 py-3 sm:px-5">
      <div className="flex h-2 flex-1 gap-0.5 overflow-hidden rounded-full" role="img" aria-label="Fleet status distribution">
        {statusOrder.map(({ status, tone }) =>
          counts[status] > 0 ? <span key={status} className={toneDot[tone]} style={{ flexGrow: counts[status] }} /> : null,
        )}
      </div>
      <div className="hidden shrink-0 gap-3 text-xs text-base-content/60 sm:flex">
        {statusOrder.map(({ status, tone }) =>
          counts[status] > 0 ? (
            <span key={status} className="flex items-center gap-1.5">
              <span className={`size-1.5 rounded-full ${toneDot[tone]}`} />
              {counts[status]} {status}
            </span>
          ) : null,
        )}
      </div>
    </div>
  )
}

export function OverviewView({ nodes, alerts, activeAlerts, nodeName, onNode }: Props) {
  const counts: Record<NodeStatus, number> = { online: 0, degraded: 0, offline: 0, pending: 0 }
  nodes.forEach((node) => counts[node.status]++)
  const attention = counts.degraded + counts.offline
  const production = nodes.filter((node) => node.environment === 'production').length

  return (
    <div className="space-y-6">
      <section className="grid grid-cols-2 gap-px overflow-hidden rounded-box border border-base-300 bg-base-300 lg:grid-cols-4">
        <Stat label="Nodes" value={nodes.length} detail={`${production} production · ${nodes.length - production} other`} />
        <Stat label="Online" value={counts.online} detail="Reported in the last 2 minutes" tone="success" />
        <Stat
          label="Needs attention"
          value={attention}
          detail={`${counts.degraded} degraded · ${counts.offline} offline`}
          tone={attention > 0 ? 'warning' : undefined}
        />
        <Stat
          label="Active incidents"
          value={activeAlerts.length}
          detail={`${alerts.length} events retained`}
          tone={activeAlerts.length > 0 ? 'error' : undefined}
        />
      </section>

      <div className="grid items-start gap-6 xl:grid-cols-[minmax(0,1.7fr)_minmax(0,1fr)]">
        <Panel
          title="Fleet"
          description="Select a node to inspect its latest sample and tokens."
          action={
            <a className="btn btn-ghost btn-sm gap-1.5 font-medium" href="/grafana/d/vpsentinel-fleet/fleet-overview" target="_blank" rel="noreferrer">
              Grafana
              <ExternalLink size={14} />
            </a>
          }
        >
          {nodes.length === 0 ? (
            <EmptyState title="No nodes registered" text="Add a node on the Nodes page to generate its first agent token." />
          ) : (
            <>
              <FleetBar counts={counts} total={nodes.length} />
              <div className="overflow-x-auto border-t border-base-300">
                <table className="table">
                  <thead>
                    <tr className="text-xs text-base-content/55">
                      <th className="font-medium">Node</th>
                      <th className="font-medium">Status</th>
                      <th className="font-medium">Last seen</th>
                      <th className="hidden font-medium sm:table-cell">Environment</th>
                    </tr>
                  </thead>
                  <tbody>
                    {nodes.map((node) => (
                      <tr key={node.id} className="cursor-pointer border-base-300 transition-colors hover:bg-base-200/60" onClick={() => onNode(node)}>
                        <td>
                          <button type="button" className="text-left text-sm font-medium hover:underline">
                            {node.name}
                          </button>
                          <div className="font-mono text-xs text-base-content/50">{node.hostname || node.id}</div>
                        </td>
                        <td><StatusBadge status={node.status} /></td>
                        <td className="whitespace-nowrap text-sm tabular-nums text-base-content/70">{relativeTime(node.last_seen_at)}</td>
                        <td className="hidden text-sm text-base-content/70 sm:table-cell">{node.environment}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          )}
        </Panel>

        <Panel title="Active incidents" description="Alerts whose latest event is still firing.">
          {activeAlerts.length === 0 ? (
            <EmptyState title="No active incidents" text="Firing alerts appear here until Grafana sends the matching recovery." good />
          ) : (
            <ul className="divide-y divide-base-300">
              {activeAlerts.map((alert) => <AlertRow key={alert.id} alert={alert} nodeName={nodeName} active compact />)}
            </ul>
          )}
        </Panel>
      </div>
    </div>
  )
}
