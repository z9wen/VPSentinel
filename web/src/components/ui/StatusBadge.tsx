import type { NodeStatus } from '../../types'

const tones: Record<NodeStatus, string> = {
  online: 'badge-success',
  degraded: 'badge-warning',
  offline: 'badge-error',
  pending: 'badge-ghost',
}

export function StatusBadge({ status, label }: { status: NodeStatus; label?: string }) {
  return <span className={`badge badge-sm gap-2 ${tones[status]}`}><span className="size-1.5 rounded-full bg-current" />{label ?? status}</span>
}
