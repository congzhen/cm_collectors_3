import { describe, expect, it } from 'vitest'
import { canAutoCheck, normalizePolicy } from '../softwareUpdate'

describe('automatic update reminder policy', () => {
  const now = new Date(2026, 9, 27, 12)
  it('caps all attempts including the first at five and resets tomorrow', () => {
    const p = normalizePolicy({}, now)
    expect(canAutoCheck({ ...p, attempts: 4 }, now.getTime())).toBe(true)
    expect(canAutoCheck({ ...p, attempts: 5 }, now.getTime())).toBe(false)
    const tomorrow = new Date(2026, 9, 28, 12)
    expect(canAutoCheck(normalizePolicy({ ...p, attempts: 5 }, tomorrow), tomorrow.getTime())).toBe(
      true,
    )
  })
  it('does not retry immediately or check again after success', () => {
    const p = normalizePolicy({}, now)
    expect(canAutoCheck({ ...p, lastAttempt: now.getTime() - 30_000 }, now.getTime())).toBe(false)
    expect(canAutoCheck({ ...p, checked: true }, now.getTime())).toBe(false)
    expect(canAutoCheck({ ...p, prompted: true }, now.getTime())).toBe(false)
  })
  it('preserves a seven-day mute across date resets', () => {
    const mutedUntil = now.getTime() + 7 * 86400_000
    const p = normalizePolicy(
      { ...normalizePolicy({}, now), mutedUntil },
      new Date(2026, 9, 28, 12),
    )
    expect(canAutoCheck(p, mutedUntil - 1)).toBe(false)
    expect(canAutoCheck(p, mutedUntil)).toBe(true)
  })
  it('normalizes invalid stored values', () => {
    const p = normalizePolicy({ attempts: NaN, mutedUntil: Infinity, lastAttempt: NaN }, now)
    expect(p.attempts).toBe(0)
    expect(p.mutedUntil).toBe(0)
  })
})
