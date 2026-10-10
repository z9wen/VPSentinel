import { toneDot, toneSurface, type Tone } from '../../lib/tone'
import type { Alert, NodeStatus } from '../../types'

export function Pill({ tone, children }: { tone: Tone; children: string }) {
  return (
    <span className={`inline-flex items-center gap-1.5 whitespace-nowrap rounded-full border px-2 py-0.5 text-xs font-medium ${toneSurface[tone]}`}>
      <span className={`size-1.5 rounded-full ${toneDot[tone]}`} />
      {children}
    </span>
  )
}

const nodeTones: Record<NodeStatus, Tone> = {
  online: 'success',
  degraded: 'warning',
  offline: 'error',
  pending: 'neutral',
}

const nodeLabels: Record<NodeStatus, string> = {
  online: 'Online',
  degraded: 'Degraded',
  offline: 'Offline',
  pending: 'Pending',
}

export function StatusBadge({ status }: { status: NodeStatus }) {
  return <Pill tone={nodeTones[status]}>{nodeLabels[status]}</Pill>
}

// A firing event that a later recovery superseded is history, not an open incident.
export function AlertStateBadge({ state, active }: { state: Alert['state']; active: boolean }) {
  if (state === 'resolved') return <Pill tone="success">Resolved</Pill>
  return active ? <Pill tone="error">Firing</Pill> : <Pill tone="neutral">Fired</Pill>
}
