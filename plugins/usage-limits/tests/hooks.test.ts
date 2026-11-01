import { expect, mock, test } from 'claude-code/testing'
import type { Engine } from 'claude-code/testing'
import type { On, RenderElement, RenderNode, SessionRateLimit } from 'claude-code'

import { withPart } from '../hooks/hint'

const NOW = Date.parse('2026-10-07T10:00:00Z')
const at = (ms: number) => new Date(NOW + ms).toISOString()
const MIN = 60_000
const HOUR = 60 * MIN

const FIVE_HOUR: SessionRateLimit = {
  kind: 'five_hour',
  percentUsed: 31,
  resetsAt: at(2 * HOUR + 13 * MIN),
}
const WEEK: SessionRateLimit = { kind: 'seven_day', percentUsed: 12, resetsAt: at(76 * HOUR) }

const START = { cwd: '/tmp', surface: 'terminal', isInteractive: true } as const
const TURN = {
  answer: '',
  durationMs: 1,
  isAborted: false,
  turnId: 't',
  reason: 'answer',
  usage: {
    input_tokens: 1,
    output_tokens: 1,
    cache_read_input_tokens: 0,
    cache_creation_input_tokens: 0,
    model: 'claude-opus-5-5',
  },
} as const

const HINT = '? for shortcuts'
const LINE: RenderElement = { type: 'engine', ref: 0 }

// What the mods drew in the engine's line's place, as the person reads it.
function text(node: RenderNode): string {
  if (typeof node === 'string') {
    return node
  }

  return node.type === 'Box' || node.type === 'Text' ? (node.children ?? []).map(text).join('') : ''
}

// Whether the tree keeps the engine's line, which puts it on a row under the mode.
function keepsLine(node: RenderNode): boolean {
  if (typeof node === 'string') {
    return false
  }

  return node.type === 'engine' || ((node.type === 'Box' || node.type === 'Text') && (node.children ?? []).some(keepsLine))
}

// Each colored run, with its color.
function colors(node: RenderNode): [string, unknown][] {
  if (typeof node === 'string' || (node.type !== 'Box' && node.type !== 'Text')) {
    return []
  }
  if (node.type === 'Text' && node.props?.color !== undefined) {
    return [[text(node), node.props.color]]
  }

  return (node.children ?? []).flatMap(colors)
}

// Each of the row's boxes, as its key and what it reads.
function parts(node: RenderNode): [unknown, string][] {
  if (typeof node === 'string' || node.type !== 'Box') {
    return []
  }

  return (node.children ?? []).map(one =>
    typeof one === 'string' || one.type !== 'Box' ? [undefined, text(one)] : [one.props?.key, text(one)],
  )
}

// Stands in for the engine beneath the plugin: answers usage from `live`,
// echoes the observe-only events, and records what the plugin shows.
function engine(on: On, stored: Record<string, unknown> = {}) {
  const live: { rateLimits: SessionRateLimit[] } = { rateLimits: [] }
  const statuses: (string | undefined)[] = []
  const toasts: string[] = []
  const invalidated: string[] = []
  const clock = mock.clock(on, { now: NOW })
  let beneath: string | undefined

  mock.store(on, stored)
  on('session.usage', () => ({
    value: { startedAt: 0, context: { window: 200_000 }, rateLimits: live.rateLimits },
  }))
  on('session.start', ($, e) => ({ cwd: e.cwd }))
  on('session.measure', ($, e) => ({ changed: e.changed }))
  on('session.end', ($, e) => ({ sessionId: e.sessionId }))
  on('classic.SessionStart', () => ({}))
  on('turn.complete', ($, e) => ({ text: e.answer }))
  on('ui.status', ($, e) => {
    statuses.push(e.text)

    return { value: undefined }
  })
  on('ui.toast', ($, e) => {
    toasts.push(e.text)

    return { value: undefined }
  })
  on('ui.invalidate', ($, e) => {
    invalidated.push(e.event)

    return { value: undefined }
  })
  on('ui.render', { component: 'PromptHint' }, ($, e) =>
    beneath === undefined ? LINE : withPart(LINE, e.props.hint, [{ text: beneath }], 'context-meter'),
  )

  // Draws the hint line, `other` standing for context-meter's part drawn beneath.
  async function draw($: Engine, other?: string, hint = HINT) {
    beneath = other

    return $.ui.render({
      surface: 'terminal',
      component: 'PromptHint',
      requestId: 'hint',
      props: { isDraft: false, isWorking: false, hint },
    })
  }

  async function shown($: Engine, other?: string, hint = HINT) {
    const row = await draw($, other, hint)

    return row.type === 'engine' ? undefined : text(row)
  }

  return { live, clock, statuses, toasts, invalidated, draw, shown }
}

function measure(rateLimits: SessionRateLimit[]) {
  return { context: { window: 200_000 }, rateLimits, changed: ['rateLimits' as const] }
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

test('shows fresh readings at start', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [WEEK, FIVE_HOUR]

  await $.session.start(START)

  expect(await world.shown($)).toBe('? for shortcuts · 5h ▰▰▱▱▱ 31% (2h13m) · wk ▰▱▱▱▱ 12% (3d4h)')
  expect(world.statuses).toEqual([undefined])
})

test('falls back to the last session’s readings, marked stale', async ($, on) => {
  const world = engine(on, { last: [{ kind: 'five_hour', percentUsed: 55, resetsAt: at(HOUR) }] })

  await $.session.start(START)

  expect(await world.shown($)).toBe('? for shortcuts · ~5h ▰▰▰▱▱ 55% (1h00m)')
})

test('/clear keeps the windows, marked stale until the next reading', async ($, on) => {
  const world = engine(on)
  const session = clearable(on)

  await $.session.start(START)
  await $.session.measure(measure([FIVE_HOUR, WEEK]))
  await clear($, session)

  expect(await world.shown($)).toBe('? for shortcuts · ~5h ▰▰▱▱▱ 31% (2h13m) · wk ▰▱▱▱▱ 12% (3d4h)')

  await $.session.measure(measure([FIVE_HOUR, WEEK]))
  expect(await world.shown($)).toBe('? for shortcuts · 5h ▰▰▱▱▱ 31% (2h13m) · wk ▰▱▱▱▱ 12% (3d4h)')
})

test('waits when nothing is known yet', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)

  expect(await world.shown($)).toBe('? for shortcuts · limits: waiting')
})

test('a measurement replaces stale readings', async ($, on) => {
  const world = engine(on, { last: [{ kind: 'five_hour', percentUsed: 55 }] })

  await $.session.start(START)
  await $.session.measure(measure([FIVE_HOUR]))

  expect(await world.shown($)).toBe('? for shortcuts · 5h ▰▰▱▱▱ 31% (2h13m)')
})

test('toasts once per window at 90%', async ($, on) => {
  const world = engine(on)
  const high = { ...FIVE_HOUR, percentUsed: 92 }

  await $.session.start(START)
  await $.session.measure(measure([high, WEEK]))
  await $.session.measure(measure([{ ...high, percentUsed: 95 }, WEEK]))
  expect(world.toasts).toEqual(['5-hour limit at 92% (resets in 2h13m)'])

  await $.session.measure(measure([{ ...high, resetsAt: at(5 * HOUR) }, WEEK]))
  expect(world.toasts.length).toBe(2)
})

test('the countdown redraws every minute', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [FIVE_HOUR]

  await $.session.start(START)
  await world.clock.advance(MIN)

  expect(world.invalidated).toEqual(['ui.render'])
  expect(await world.shown($)).toBe('? for shortcuts · 5h ▰▰▱▱▱ 31% (2h12m)')
})

test('leaves the line alone when a turn reports no windows', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  await $.turn.complete(TURN)

  expect(await world.shown($)).toBeUndefined()
})

test('reads after another mod’s part of the line', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [FIVE_HOUR]

  await $.session.start(START)

  expect(await world.shown($, 'ctx 18%')).toBe('? for shortcuts · ctx 18% · 5h ▰▰▱▱▱ 31% (2h13m)')
})

test('wraps between the parts, never inside one', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [FIVE_HOUR]

  await $.session.start(START)

  const row = await world.draw($, 'ctx 18%')

  expect(row.type === 'Box' && row.props?.flexWrap).toBe('wrap')
  expect(parts(row)).toEqual([
    ['claudeforge-hint:hint', '? for shortcuts · '],
    ['claudeforge-hint:context-meter', 'ctx 18% · '],
    ['claudeforge-hint:usage-limits', '5h ▰▰▱▱▱ 31% (2h13m)'],
  ])
})

test('stands in for the engine’s line, so it reads beside the mode', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [FIVE_HOUR]

  await $.session.start(START)

  expect(keepsLine(await world.draw($))).toBe(false)
  expect(keepsLine(await world.draw($, 'ctx 18%'))).toBe(false)
})

test('an empty hint leaves no leading dot', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [FIVE_HOUR]

  await $.session.start(START)

  expect(await world.shown($, undefined, '')).toBe('5h ▰▰▱▱▱ 31% (2h13m)')
  expect(await world.shown($, 'ctx 18%', '')).toBe('ctx 18% · 5h ▰▰▱▱▱ 31% (2h13m)')
})

test('colors each window by how full it is', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [
    { ...FIVE_HOUR, percentUsed: 93 },
    { ...WEEK, percentUsed: 74 },
    { kind: 'seven_day_opus', percentUsed: 12 },
  ]

  await $.session.start(START)

  expect(colors(await world.draw($))).toEqual([
    ['▰▰▰▰▰ 93%', 'error'],
    ['▰▰▰▰▱ 74%', 'warning'],
    ['▰▱▱▱▱ 12%', 'success'],
  ])
})

test('the row validates where the hint line is drawn', async ($, on) => {
  const world = engine(on)
  world.live.rateLimits = [FIVE_HOUR, WEEK]

  await $.session.start(START)

  for (const surface of ['terminal', 'desktop'] as const) {
    const ui = await $.ui.mount({
      plugin: 'usage-limits',
      surface,
      component: 'PromptHint',
      props: { isDraft: false, isWorking: false, hint: '? for shortcuts' },
    })

    expect(await ui.find({ type: 'Text', text: /5h/ })).toBeDefined()
    await ui.unmount()
  }
})
