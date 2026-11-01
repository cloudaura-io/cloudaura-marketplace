import { expect, test } from 'claude-code/testing'

import { countdown, formatLimits, isLimitList, WAITING } from '../hooks/format'

const NOW = Date.parse('2026-10-07T10:00:00Z')
const at = (ms: number) => new Date(NOW + ms).toISOString()
const MIN = 60_000
const HOUR = 60 * MIN

test('countdown boundaries', () => {
  expect(countdown(30_000)).toBe('<1m')
  expect(countdown(5 * MIN)).toBe('5m')
  expect(countdown(2 * HOUR + 5 * MIN)).toBe('2h05m')
  expect(countdown(23 * HOUR + 59 * MIN)).toBe('23h59m')
  expect(countdown(76 * HOUR)).toBe('3d4h')
})

test('5h first, then weekly, unknown kinds last', () => {
  const line = formatLimits(
    [
      { kind: 'mystery', percentUsed: 50 },
      { kind: 'seven_day', percentUsed: 12, resetsAt: at(76 * HOUR) },
      { kind: 'five_hour', percentUsed: 31.4, resetsAt: at(2 * HOUR + 13 * MIN) },
    ],
    NOW,
    true,
  )

  expect(line).toBe(
    '5h ▰▰▱▱▱ 31% (2h13m) · wk ▰▱▱▱▱ 12% (3d4h) · mystery ▰▰▰▱▱ 50%',
  )
})

test('stale readings carry a tilde', () => {
  expect(formatLimits([{ kind: 'five_hour', percentUsed: 40 }], NOW, false)).toBe(
    '~5h ▰▰▱▱▱ 40%',
  )
})

test('a window past its reset reads 0%', () => {
  expect(formatLimits([{ kind: 'five_hour', percentUsed: 97, resetsAt: at(-MIN) }], NOW, false)).toBe(
    '~5h ▱▱▱▱▱ 0% (reset)',
  )
})

test('no readings', () => {
  expect(formatLimits([], NOW, true)).toBe(WAITING)
})

test('isLimitList guards the store', () => {
  expect(isLimitList([{ kind: 'five_hour', percentUsed: 3 }])).toBe(true)
  expect(isLimitList([{ kind: 'five_hour' }])).toBe(false)
  expect(isLimitList('nope')).toBe(false)
})
