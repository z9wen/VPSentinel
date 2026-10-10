import { useState } from 'react'
import type { Alert } from '../../types'
import { EmptyState } from '../../components/ui/EmptyState'
import { Panel } from '../../components/ui/Panel'
import { AlertRow } from './AlertRow'

type Filter = 'all' | 'firing' | 'resolved'

const filters: { value: Filter; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'firing', label: 'Firing' },
  { value: 'resolved', label: 'Resolved' },
]

type Props = { alerts: Alert[]; activeAlerts: Alert[]; nodeName: (id: string | null) => string }

export function AlertsView({ alerts, activeAlerts, nodeName }: Props) {
  const [filter, setFilter] = useState<Filter>('all')
  const activeIDs = new Set(activeAlerts.map((alert) => alert.id))
  const count = (value: Filter) => (value === 'all' ? alerts.length : alerts.filter((alert) => alert.state === value).length)
  const visible = filter === 'all' ? alerts : alerts.filter((alert) => alert.state === filter)

  const controls = (
    <div className="flex rounded-field border border-base-300 bg-base-200 p-0.5" role="tablist" aria-label="Filter events">
      {filters.map(({ value, label }) => (
        <button
          key={value}
          type="button"
          role="tab"
          aria-selected={filter === value}
          onClick={() => setFilter(value)}
          className={`rounded-[calc(var(--radius-field)-2px)] px-2.5 py-1 text-xs font-medium transition-colors ${
            filter === value ? 'bg-base-100 text-base-content shadow-sm' : 'text-base-content/60 hover:text-base-content'
          }`}
        >
          {label} <span className="tabular-nums text-base-content/45">{count(value)}</span>
        </button>
      ))}
    </div>
  )

  return (
    <Panel title="Event history" description="The latest 100 webhook deliveries, newest first." action={controls}>
      {visible.length === 0 ? (
        <EmptyState title="No matching events" text="Grafana firing and resolved notifications are stored here for investigation." />
      ) : (
        <ul className="divide-y divide-base-300">
          {visible.map((alert) => <AlertRow key={alert.id} alert={alert} nodeName={nodeName} active={activeIDs.has(alert.id)} />)}
        </ul>
      )}
    </Panel>
  )
}
