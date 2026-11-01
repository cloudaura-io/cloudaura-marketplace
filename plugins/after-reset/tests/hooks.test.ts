import { expect, mock, test } from 'claude-code/testing'
import type { Engine } from 'claude-code/testing'
import type { On, PromptSubmitInput, SessionRateLimit } from 'claude-code'

// Local wall-clock times, so the tests read the same in any time zone.
const NOW = new Date(2026, 9, 8, 13, 0).getTime()
const iso = (ms: number) => new Date(ms).toISOString()
const MIN = 60_000
const HOUR = 60 * MIN

const FULL_5H: SessionRateLimit = { kind: 'five_hour', percentUsed: 100, resetsAt: iso(NOW + 2 * HOUR + 5 * MIN) }
const WEEK: SessionRateLimit = { kind: 'seven_day', percentUsed: 64, resetsAt: iso(NOW + 50 * HOUR) }

const START = { cwd: '/tmp', surface: 'terminal', isInteractive: true } as const

// Stands in for the engine beneath the plugin: answers usage from `live` and
// the store from `store` (another session may write it meanwhile), refuses a
// prompt for each reason in `refusals` (undefined lets one in), and records
// the prompts the plugin sends and what it shows.
function engine(on: On, stored: Record<string, unknown> = {}) {
  const live: { rateLimits: SessionRateLimit[] } = { rateLimits: [] }
  const store = new Map(Object.entries(stored))
  const refusals: (string | undefined)[] = []
  const sent: PromptSubmitInput[] = []
  const statuses: (string | undefined)[] = []
  const toasts: string[] = []
  const clock = mock.clock(on, { now: NOW })

  on('store.get', ($, e) => ({ value: store.get(e.key) }))
  on('store.set', ($, e) => {
    store.set(e.key, e.value)

    return { value: undefined }
  })
  on('session.usage', () => ({
    value: { startedAt: 0, context: { window: 200_000 }, rateLimits: live.rateLimits },
  }))
  on('session.start', ($, e) => ({ cwd: e.cwd }))
  on('session.measure', ($, e) => ({ changed: e.changed }))
  on('session.end', ($, e) => ({ sessionId: e.sessionId }))
  on('classic.SessionStart', () => ({}))
  on('command.register', ($, e) => ({ value: { command: e.name } }))
  on('prompt.submit', ($, e) => {
    const refused = refusals.shift()

    if (refused !== undefined) {
      return { drop: refused }
    }

    sent.push(e)

    return { text: e.text }
  })
  on('ui.status', ($, e) => {
    statuses.push(e.text)

    return { value: undefined }
  })
  on('ui.toast', ($, e) => {
    toasts.push(e.text)

    return { value: undefined }
  })
  on('classic.StopFailure', () => ({}))

  return { live, store, refusals, sent, statuses, toasts, clock }
}

// Stands in for the new session a /clear starts: after `wipe()` every value
// reads as never written, as in the engine's fresh session, until written
// again. Versions stay as they stood, so the plugin's writes still land.
function clearable(on: On) {
  const rewritten = new Set<string>()
  let isWiped = false
  const name = (e: { key: string; id?: string }) => `${e.key}#${e.id ?? ''}`

  on('state.get', async ($, e, next) => {
    const held = await next(e)

    return isWiped && !rewritten.has(name(e)) && 'value' in held
      ? { value: { ...held.value, value: undefined } }
      : held
  })
  on('state.set', async ($, e, next) => {
    const result = await next(e)

    if ('value' in result && result.value.isSet) {
      rewritten.add(name(e))
    }

    return result
  })

  return {
    wipe() {
      isWiped = true
      rewritten.clear()
    },
  }
}

async function clear($: Engine, session: { wipe: () => void }) {
  await $.session.end({ reason: 'clear', sessionId: 's', resume: { id: 's' } })
  session.wipe()
  await $.classic.SessionStart({ source: 'clear' })
}

function measure(rateLimits: SessionRateLimit[]) {
  return { context: { window: 200_000 }, rateLimits, changed: ['rateLimits' as const] }
}

async function run($: Engine, args: string) {
  const presentation = { isFullscreen: true, columns: 120 }

  return (await $.command.run({ command: 'after-reset', args, origin: { kind: 'composer' }, presentation })).text ?? ''
}

test('sends the prompt as the person’s, just after the full window resets', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [WEEK, FULL_5H]

  await $.session.start(START)

  expect(await run($, 'continue where you left off')).toBe(
    'Will send at 15:07 (in 2h07m), after the 5-hour limit resets: continue where you left off\n' +
      'Keep this session open. /after-reset lists, /after-reset cancel undoes.',
  )
  expect(world.statuses.at(-1)).toBe('prompt at 15:07 (in 2h07m)')

  await world.clock.advance(2 * HOUR + 6 * MIN)
  expect(world.sent).toEqual([])

  await world.clock.advance(MIN)
  expect(world.sent.map(e => e.text)).toEqual(['continue where you left off'])
  expect(world.sent[0]?.origin).toEqual(expect.objectContaining({ kind: 'plugin', asUser: true }))
  expect(world.statuses.at(-1)).toBeUndefined()

  await world.clock.advance(HOUR)
  expect(world.sent.length).toBe(1)
})

test('waits for the weekly reset when both windows are full', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [{ ...WEEK, percentUsed: 100 }, FULL_5H]

  await $.session.start(START)

  expect(await run($, 'go on')).toContain('Sat 10 Oct 15:02 (in 2d2h), after the weekly limit resets')
})

test('postpones when another window is full by the time it is due', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [WEEK, FULL_5H]

  await $.session.start(START)
  await run($, 'go on')
  world.live.rateLimits = [{ ...WEEK, percentUsed: 100 }, FULL_5H]
  await world.clock.advance(2 * HOUR + 7 * MIN)

  expect(world.sent).toEqual([])
  expect(world.toasts).toEqual([
    'the weekly limit is still reached; waiting until Sat 10 Oct 15:02 (in 1d23h)',
  ])
})

test('in and at need no known reset', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)

  expect(await run($, 'go on')).toContain('No reset time is known yet')
  expect(await run($, 'in 90m first')).toContain('Will send at 14:30 (in 1h30m): first')
  expect(await run($, 'at 14:00 second')).toContain('Will send at 14:00 (in 1h00m): second')
  expect(world.statuses.at(-1)).toBe('2 prompts, next at 14:00 (in 1h00m)')

  await world.clock.advance(90 * MIN)
  expect(world.sent.map(e => e.text)).toEqual(['second', 'first'])
})

test('falls back to the readings another session saved', async ($, on) => {
  const world = engine(on, { limits: [{ kind: 'five_hour', percentUsed: 100, resetsAt: iso(NOW + HOUR) }] })

  await $.session.start(START)

  expect(await run($, 'go on')).toContain('Will send at 14:02 (in 1h02m), after the 5-hour limit resets')
})

test('saves readings for later sessions', async ($, on) => {
  engine(on)

  await $.session.start(START)
  await $.session.measure({ context: { window: 200_000 }, rateLimits: [FULL_5H], changed: ['rateLimits'] })

  // Live readings stay empty here, so the time comes from what was saved.
  expect(await run($, 'go on')).toContain('Will send at 15:07 (in 2h07m), after the 5-hour limit resets')
})

test('a newer reading from another session wins over this one’s', async ($, on) => {
  const world = engine(on)
  const nearlyFull = { ...FULL_5H, percentUsed: 97 }
  world.live.rateLimits = [nearlyFull]

  await $.session.start(START)
  await $.session.measure(measure([nearlyFull]))
  await run($, 'in 10m go on')

  // Another session fills the window and saves what it read.
  await world.clock.advance(MIN)
  world.store.set('limits', { at: world.clock.now(), limits: [FULL_5H] })
  await world.clock.advance(9 * MIN)

  expect(world.sent).toEqual([])
  expect(world.toasts).toEqual(['the 5-hour limit is still reached; waiting until 15:07 (in 1h57m)'])
})

test('this session’s reading wins over an older saved one', async ($, on) => {
  const world = engine(on, { limits: { at: NOW - HOUR, limits: [FULL_5H] } })
  world.live.rateLimits = [{ ...FULL_5H, percentUsed: 40 }]

  await $.session.start(START)
  await run($, 'in 10m go on')
  await world.clock.advance(10 * MIN)

  expect(world.sent.map(e => e.text)).toEqual(['go on'])
})

test('due prompts go one at a time, and a refusal holds back only its own', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [WEEK, FULL_5H]
  world.refusals.push(undefined, 'the session is busy')

  await $.session.start(START)
  await run($, 'first')
  await run($, 'second')
  await run($, 'third')
  await world.clock.advance(2 * HOUR + 7 * MIN)

  expect(world.sent.map(e => e.text)).toEqual(['first'])
  expect(world.toasts).toEqual(['could not send "second": the session is busy; trying again at 15:12 (in 5m)'])
  expect(await run($, '')).toContain('  2. 15:12 (in 5m), retrying after a failed send: second')

  await world.clock.advance(30_000)
  expect(world.sent.map(e => e.text)).toEqual(['first', 'third'])

  await world.clock.advance(5 * MIN)
  expect(world.sent.map(e => e.text)).toEqual(['first', 'third', 'second'])
  expect(world.statuses.at(-1)).toBeUndefined()
})

test('/clear keeps the queue and says so', async ($, on) => {
  const world = engine(on)
  const session = clearable(on)

  await $.session.start(START)
  await run($, 'in 1h one')
  await run($, 'in 2h two')
  await clear($, session)

  expect(world.toasts).toEqual(['2 scheduled prompts kept; /after-reset cancel drops them'])
  expect(world.statuses.at(-1)).toBe('2 prompts, next at 14:00 (in 1h00m)')
  expect(await run($, '')).toContain('  2. 15:00 (in 2h00m): two')

  await world.clock.advance(2 * HOUR)
  expect(world.sent.map(e => e.text)).toEqual(['one', 'two'])
})

test('/clear with nothing queued says nothing', async ($, on) => {
  const world = engine(on)
  const session = clearable(on)

  await $.session.start(START)
  await clear($, session)

  expect(world.toasts).toEqual([])
})

test('lists and cancels', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [WEEK, FULL_5H]

  await $.session.start(START)
  expect(await run($, '')).toContain('Nothing is scheduled.')

  await run($, 'in 1h one')
  await run($, 'in 2h two')
  await run($, 'three')

  const listed = await run($, '')
  expect(listed).toContain('  1. 14:00 (in 1h00m): one')
  expect(listed).toContain('  3. 15:07 (in 2h07m), after the 5-hour limit resets: three')
  expect(listed).toContain('Next reset that matters: the 5-hour limit (full) at 15:05 (in 2h05m)')

  expect(await run($, 'cancel 2')).toBe('Cancelled: two')
  expect(await run($, 'cancel 9')).toBe('There is no prompt 9; see /after-reset.')
  expect(await run($, 'cancel')).toBe('Cancelled 2 prompts.')
  expect(world.statuses.at(-1)).toBeUndefined()

  await world.clock.advance(3 * HOUR)
  expect(world.sent).toEqual([])
})

test('a turn cut off by the limit says how to pick it up', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [WEEK, FULL_5H]

  await $.session.start(START)
  await $.classic.StopFailure({ error: 'overloaded' })
  await $.classic.StopFailure({ error: 'rate_limit' })

  expect(world.toasts).toEqual(['Usage limit reached. /after-reset continue sends "continue" at 15:07 (in 2h07m).'])
})
