import { useState } from 'react'
import type { Alert } from '../../types'
import { EmptyState } from '../../components/ui/EmptyState'
import { Panel } from '../../components/ui/Panel'
import { AlertRow } from './AlertRow'

export function AlertsView({ alerts }: { alerts: Alert[] }) {
  const [filter, setFilter] = useState<'all' | 'firing' | 'resolved'>('all')
  const visible = filter === 'all' ? alerts : alerts.filter((alert) => alert.state === filter)
  const controls = <div className="join">{(['all', 'firing', 'resolved'] as const).map((value) => <button className={`btn btn-xs join-item ${filter === value ? 'btn-active' : 'btn-ghost'}`} onClick={() => setFilter(value)} key={value}>{value}</button>)}</div>
  return <Panel eyebrow="GRAFANA WEBHOOK EVENTS" title={`${visible.length} events`} action={controls}>{visible.length === 0 ? <EmptyState title="No matching incidents" text="Alert events will be retained here for investigation." /> : visible.map((alert) => <AlertRow key={alert.id} alert={alert} />)}</Panel>
}
