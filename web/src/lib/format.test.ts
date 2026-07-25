import { describe, expect, it, vi } from 'vitest'
import { formatBytes, formatDuration, formatPercent, relativeTime } from './format'

describe('metric formatting', () => {
  it('formats missing and measured values', () => {
    expect(formatPercent(null)).toBe('—')
    expect(formatPercent(83.456)).toBe('83.5%')
    expect(formatBytes(1024 ** 3)).toBe('1.0 GB')
    expect(formatDuration(90_000)).toBe('1d 1h')
  })

  it('formats last-seen timestamps relative to now', () => {
    vi.setSystemTime(new Date('2026-07-25T12:00:00Z'))
    expect(relativeTime('2026-07-25T11:58:00Z')).toBe('2m ago')
    expect(relativeTime(null)).toBe('Never')
    vi.useRealTimers()
  })
})
