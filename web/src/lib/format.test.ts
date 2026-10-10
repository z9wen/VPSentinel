import { describe, expect, it, vi } from 'vitest'
import { elapsedSince, formatBytes, formatDuration, formatPercent, percentOf, relativeTime } from './format'

describe('metric formatting', () => {
  it('formats missing and measured values', () => {
    expect(formatPercent(null)).toBe('—')
    expect(formatPercent(83.456)).toBe('83.5%')
    expect(formatBytes(1024 ** 3)).toBe('1.0 GB')
    expect(formatDuration(90_000)).toBe('1d 1h')
  })

  it('derives usage percentages only when both values are known', () => {
    expect(percentOf(25, 200)).toBe(12.5)
    expect(percentOf(null, 200)).toBeNull()
    expect(percentOf(25, 0)).toBeNull()
  })

  it('formats timestamps relative to now', () => {
    vi.setSystemTime(new Date('2026-07-25T12:00:00Z'))
    expect(relativeTime('2026-07-25T11:58:00Z')).toBe('2m ago')
    expect(relativeTime(null)).toBe('Never')
    expect(elapsedSince('2026-07-25T09:55:00Z')).toBe('2h 5m')
    vi.useRealTimers()
  })
})
