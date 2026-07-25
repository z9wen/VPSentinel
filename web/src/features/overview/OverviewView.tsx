import { Activity, CircleAlert, ExternalLink, Server, Wifi } from 'lucide-react'
import type { Alert, Node, NodeStatus } from '../../types'
import { relativeTime } from '../../lib/format'
import { EmptyState } from '../../components/ui/EmptyState'
import { Panel } from '../../components/ui/Panel'
import { StatusBadge } from '../../components/ui/StatusBadge'
import { AlertRow } from '../alerts/AlertRow'

function StatCard({ label, value, detail, tone, Icon }: { label: string; value: number; detail: string; tone: 'primary' | 'success' | 'warning' | 'error'; Icon: typeof Server }) {
  const colors = { primary: 'text-primary bg-primary/10', success: 'text-success bg-success/10', warning: 'text-warning bg-warning/10', error: 'text-error bg-error/10' }
  return <article className="card border border-base-300 bg-base-100"><div className="card-body gap-1 p-5"><div className="flex items-center justify-between text-xs text-base-content/55"><span>{label}</span><span className={`grid size-8 place-items-center rounded-lg ${colors[tone]}`}><Icon size={15} /></span></div><strong className="mt-2 font-mono text-4xl font-medium tracking-tight">{value}</strong><small className="text-base-content/35">{detail}</small></div></article>
}

export function OverviewView({ nodes, activeAlerts, alerts, onNode }: { nodes: Node[]; activeAlerts: Alert[]; alerts: Alert[]; onNode: (node: Node) => void }) {
  const count = (status: NodeStatus) => nodes.filter((node) => node.status === status).length
  return <div className="space-y-5"><section className="grid grid-cols-2 gap-3 xl:grid-cols-4"><StatCard label="Total nodes" value={nodes.length} detail="Registered fleet" tone="primary" Icon={Server} /><StatCard label="Online" value={count('online')} detail="Reporting within 2m" tone="success" Icon={Wifi} /><StatCard label="Needs attention" value={count('degraded') + count('offline')} detail="Delayed or offline" tone="warning" Icon={CircleAlert} /><StatCard label="Active incidents" value={activeAlerts.length} detail={`${alerts.length} retained events`} tone={activeAlerts.length ? 'error' : 'success'} Icon={Activity} /></section>
    <Panel eyebrow="LIVE INVENTORY" title="Fleet state" action={<a className="btn btn-ghost btn-sm text-primary" href="/grafana/d/vpsentinel-fleet/fleet-overview" target="_blank">Open Grafana <ExternalLink size={14} /></a>}>{nodes.length === 0 ? <EmptyState title="No nodes registered" text="Add a node to generate its first agent token." /> : <div className="overflow-x-auto"><table className="table"><thead><tr><th>Node</th><th>Environment</th><th>Status</th><th>Last seen</th><th>Agent</th></tr></thead><tbody>{nodes.map((node) => <tr className="cursor-pointer hover" key={node.id} onClick={() => onNode(node)}><td><strong className="block text-sm">{node.name}</strong><small className="text-base-content/40">{node.hostname || node.id}</small></td><td>{node.environment}</td><td><StatusBadge status={node.status} /></td><td>{relativeTime(node.last_seen_at)}</td><td className="font-mono text-xs">{node.agent_version}</td></tr>)}</tbody></table></div>}</Panel>
    <Panel eyebrow="INCIDENT SIGNAL" title="Active alerts">{activeAlerts.length === 0 ? <EmptyState title="No active incidents" text="Firing alerts and their recovery events will appear here." good /> : activeAlerts.map((alert) => <AlertRow key={alert.id} alert={alert} />)}</Panel>
  </div>
}
