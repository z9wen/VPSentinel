import type { Alert } from '../../types'
import { elapsedSince, formatTimestamp, relativeTime } from '../../lib/format'
import { severityTone, toneDot } from '../../lib/tone'
import { AlertStateBadge, Pill } from '../../components/ui/StatusBadge'

type Props = {
  alert: Alert
  nodeName: (id: string | null) => string
  // Active means this firing event has not been followed by a recovery yet.
  active: boolean
  compact?: boolean
}

function timing(alert: Alert, active: boolean) {
  if (alert.state === 'resolved') return `resolved ${relativeTime(alert.received_at)}`
  return active ? `for ${elapsedSince(alert.starts_at)}` : `fired ${relativeTime(alert.received_at)}`
}

export function AlertRow({ alert, nodeName, active, compact = false }: Props) {
  const firing = alert.state === 'firing'
  const tone = !firing ? 'success' : active ? severityTone(alert.severity) : 'neutral'

  return (
    <li className="flex gap-3 px-4 py-3 sm:px-5">
      <span className={`mt-1.5 size-2 shrink-0 rounded-full ${toneDot[tone]}`} aria-hidden="true" />
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
          <span className="text-sm font-medium">{alert.rule_name}</span>
          <Pill tone={severityTone(alert.severity)}>{alert.severity}</Pill>
        </div>
        {!compact && <p className="mt-1 truncate text-sm text-base-content/65" title={alert.summary}>{alert.summary}</p>}
        <p className="mt-1 text-xs text-base-content/50">
          {nodeName(alert.node_id)} · started {formatTimestamp(alert.starts_at)}
        </p>
      </div>
      <div className="flex shrink-0 flex-col items-end gap-1.5 text-xs tabular-nums text-base-content/55">
        {!compact && <AlertStateBadge state={alert.state} active={active} />}
        <span>{timing(alert, active)}</span>
      </div>
    </li>
  )
}
