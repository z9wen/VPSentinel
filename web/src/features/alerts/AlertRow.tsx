import { CheckCircle2, TriangleAlert } from 'lucide-react'
import type { Alert } from '../../types'
import { StatusBadge } from '../../components/ui/StatusBadge'

export function AlertRow({ alert }: { alert: Alert }) {
  const Icon = alert.state === 'firing' ? TriangleAlert : CheckCircle2
  return <article className="grid min-h-24 grid-cols-[36px_1fr_auto] items-center gap-4 border-b border-base-300 px-5 py-4 last:border-0"><span className={`grid size-9 place-items-center rounded-full ${alert.state === 'firing' ? 'bg-error/10 text-error' : 'bg-success/10 text-success'}`}><Icon size={17} /></span><div className="min-w-0"><div className="flex flex-wrap items-center gap-2"><strong className="text-sm">{alert.rule_name}</strong><span className={`badge badge-xs ${alert.severity === 'critical' ? 'badge-error' : 'badge-warning'}`}>{alert.severity}</span></div><p className="mt-1 truncate text-xs text-base-content/60">{alert.summary}</p><small className="font-mono text-[9px] text-base-content/35">{alert.node_id ?? 'Control plane'} · {new Date(alert.received_at).toLocaleString()}</small></div><StatusBadge status={alert.state === 'firing' ? 'offline' : 'online'} label={alert.state} /></article>
}
