export type Tone = 'success' | 'warning' | 'error' | 'neutral'

// Tone classes are spelled out in full so Tailwind can detect them at build time.
export const toneDot: Record<Tone, string> = {
  success: 'bg-success',
  warning: 'bg-warning',
  error: 'bg-error',
  neutral: 'bg-base-content/30',
}

export const toneSurface: Record<Tone, string> = {
  success: 'border-success/25 bg-success/10',
  warning: 'border-warning/30 bg-warning/12',
  error: 'border-error/25 bg-error/10',
  neutral: 'border-base-300 bg-base-200',
}

export function severityTone(severity: string): Tone {
  return severity === 'critical' ? 'error' : 'warning'
}
