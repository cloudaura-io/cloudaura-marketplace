import { expect, test } from 'claude-code/testing'

import { clock, countdown, parse, pickReset } from '../hooks/when'

// Local wall-clock times, so the tests read the same in any time zone.
const NOW = new Date(2026, 9, 8, 13, 0).getTime()
const local = (day: number, hours: number, minutes: number) => new Date(2026, 9, day, hours, minutes).getTime()
const iso = (ms: number) => new Date(ms).toISOString()
const MIN = 60_000
const HOUR = 60 * MIN

test('bare lists, cancel takes all or a number', () => {
  expect(parse('', NOW)).toEqual({ kind: 'list' })
  expect(parse('  ', NOW)).toEqual({ kind: 'list' })
  expect(parse('cancel', NOW)).toEqual({ kind: 'cancel', which: 'all' })
  expect(parse('cancel all', NOW)).toEqual({ kind: 'cancel', which: 'all' })
  expect(parse('cancel 2', NOW)).toEqual({ kind: 'cancel', which: 2 })
})

test('a prompt alone waits for the reset', () => {
  expect(parse('continue where you left off', NOW)).toEqual({ kind: 'schedule', text: 'continue where you left off' })
  expect(parse('in the README, fix the typos', NOW)).toEqual({ kind: 'schedule', text: 'in the README, fix the typos' })
  expect(parse('at the end, summarise', NOW)).toEqual({ kind: 'schedule', text: 'at the end, summarise' })
})

test('in <duration> waits that long', () => {
  expect(parse('in 90m go on', NOW)).toEqual({ kind: 'schedule', text: 'go on', at: NOW + 90 * MIN })
  expect(parse('in 1h30m go on', NOW)).toEqual({ kind: 'schedule', text: 'go on', at: NOW + 90 * MIN })
  expect(parse('in 2 hours go on', NOW)).toEqual({ kind: 'schedule', text: 'go on', at: NOW + 2 * HOUR })
  expect(parse('in 1.5h go on', NOW)).toEqual({ kind: 'schedule', text: 'go on', at: NOW + 90 * MIN })
  expect(parse('in 3h\nline one\nline two', NOW)).toEqual({ kind: 'schedule', text: 'line one\nline two', at: NOW + 3 * HOUR })
  expect(parse('in 2h', NOW).kind).toBe('error')
  expect(parse('in 0m go', NOW).kind).toBe('error')
})

test('at HH:MM is today, or tomorrow once passed', () => {
  expect(parse('at 14:30 go on', NOW)).toEqual({ kind: 'schedule', text: 'go on', at: local(8, 14, 30) })
  expect(parse('at 9:15 go on', NOW)).toEqual({ kind: 'schedule', text: 'go on', at: local(9, 9, 15) })
  expect(parse('at 13:00 go on', NOW)).toEqual({ kind: 'schedule', text: 'go on', at: local(9, 13, 0) })
  expect(parse('at 25:00 go on', NOW).kind).toBe('error')
  expect(parse('at 14:30', NOW).kind).toBe('error')
})

test('a mistyped command is refused, not queued as a prompt', () => {
  expect(parse('cancel first', NOW)).toEqual({ kind: 'error', message: expect.stringContaining('"cancel" takes') })
  expect(parse('at 9:00pm deploy', NOW)).toEqual({ kind: 'error', message: expect.stringContaining('"at" takes') })
  expect(parse('at 9 deploy', NOW)).toEqual({ kind: 'error', message: expect.stringContaining('"at" takes') })
  expect(parse('in 2 days check', NOW)).toEqual({ kind: 'error', message: expect.stringContaining('"in" takes') })
})

test('prompts that only start like a command still queue', () => {
  expect(parse('cancel the deploy and roll back', NOW)).toEqual({ kind: 'schedule', text: 'cancel the deploy and roll back' })
  expect(parse('in the tests, fix the flake', NOW)).toEqual({ kind: 'schedule', text: 'in the tests, fix the flake' })
  expect(parse('at some point, tidy up', NOW)).toEqual({ kind: 'schedule', text: 'at some point, tidy up' })
})

test('waits for the latest full window, else the nearest reset', () => {
  const fiveHour = { kind: 'five_hour', percentUsed: 100, resetsAt: iso(NOW + 2 * HOUR) }
  const week = { kind: 'seven_day', percentUsed: 64, resetsAt: iso(NOW + 50 * HOUR) }

  expect(pickReset([week, fiveHour], NOW)).toEqual({ kind: 'five_hour', at: NOW + 2 * HOUR, isFull: true })
  expect(pickReset([{ ...week, percentUsed: 100 }, fiveHour], NOW)).toEqual({
    kind: 'seven_day',
    at: NOW + 50 * HOUR,
    isFull: true,
  })
  expect(pickReset([week, { ...fiveHour, percentUsed: 40 }], NOW)).toEqual({
    kind: 'five_hour',
    at: NOW + 2 * HOUR,
    isFull: false,
  })
})

test('ignores windows already reset or with no time', () => {
  expect(pickReset([{ kind: 'five_hour', percentUsed: 100, resetsAt: iso(NOW - MIN) }], NOW)).toBeUndefined()
  expect(pickReset([{ kind: 'five_hour', percentUsed: 100 }], NOW)).toBeUndefined()
  expect(pickReset([], NOW)).toBeUndefined()
})

test('times read as the local clock, with the day when not today', () => {
  expect(clock(local(8, 15, 7), NOW)).toBe('15:07')
  expect(clock(local(9, 9, 5), NOW)).toBe('tomorrow 09:05')
  expect(clock(local(10, 18, 0), NOW)).toBe('Sat 10 Oct 18:00')
  expect(countdown(30_000)).toBe('<1m')
  expect(countdown(2 * HOUR + 7 * MIN)).toBe('2h07m')
  expect(countdown(50 * HOUR)).toBe('2d2h')
})
