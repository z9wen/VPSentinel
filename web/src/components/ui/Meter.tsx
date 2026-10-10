import { formatPercent } from '../../lib/format'

type Props = {
  label: string
  percent: number | null
  detail: string
  warning: number
  critical?: number
}

export function Meter({ label, percent, detail, warning, critical }: Props) {
  const fill =
    percent == null ? 'bg-base-300'
      : critical != null && percent >= critical ? 'bg-error'
        : percent >= warning ? 'bg-warning'
          : 'bg-success'
  const width = percent == null ? 0 : Math.min(100, Math.max(0, percent))

  return (
    <div>
      <div className="flex items-baseline justify-between gap-3">
        <span className="text-sm text-base-content/70">{label}</span>
        <span className="text-sm font-semibold tabular-nums">{formatPercent(percent)}</span>
      </div>
      <div
        className="mt-2 h-1.5 overflow-hidden rounded-full bg-base-200"
        role="meter"
        aria-label={label}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={percent == null ? undefined : Math.round(width)}
      >
        <div className={`h-full rounded-full transition-[width] duration-500 ${fill}`} style={{ width: `${width}%` }} />
      </div>
      <p className="mt-1.5 text-xs tabular-nums text-base-content/50">{detail}</p>
    </div>
  )
}
