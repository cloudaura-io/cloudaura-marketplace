import { expect, test } from 'claude-code/testing'
import type { Engine } from 'claude-code/testing'
import type { On, RenderElement, RenderNode, SessionContextUsage } from 'claude-code'

import { withPart } from '../hooks/hint'

const WINDOW = 200_000

function ctxAt(tokens: number): SessionContextUsage {
  return { tokens, window: WINDOW, percent: Math.round((tokens / WINDOW) * 100) }
}

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
// echoes the observe-only events, and records what the plugin shows. With
// `compactTo` set, a compaction reports its counts and empties the reading.
function engine(on: On) {
  const live: { context: SessionContextUsage; compactTo?: number } = {
    context: { window: WINDOW },
  }
  const statuses: (string | undefined)[] = []
  const toasts: string[] = []
  let beneath: string | undefined

  on('session.usage', () => ({
    value: { startedAt: 0, context: live.context, rateLimits: [] },
  }))
  on('session.start', ($, e) => ({ cwd: e.cwd }))
  on('session.measure', ($, e) => ({ changed: e.changed }))
  on('session.end', ($, e) => ({ sessionId: e.sessionId }))
  on('classic.SessionStart', () => ({}))
  on('turn.complete', ($, e) => ({ text: e.answer }))
  on('session.compact', ($, e) => {
    if (live.compactTo === undefined) {
      return { messages: e.messages }
    }

    const tokensBefore = live.context.tokens

    if (e.agentId === undefined && e.trigger !== 'precompute') {
      live.context = { window: WINDOW }
    }

    return { messages: e.messages, tokensBefore, tokensAfter: live.compactTo }
  })
  on('ui.status', ($, e) => {
    statuses.push(e.text)

    return { value: undefined }
  })
  on('ui.toast', ($, e) => {
    toasts.push(e.text)

    return { value: undefined }
  })
  on('ui.render', { component: 'PromptHint' }, ($, e) =>
    beneath === undefined ? LINE : withPart(LINE, e.props.hint, [{ text: beneath }], 'usage-limits'),
  )

  // Draws the hint line, `other` standing for usage-limits' part drawn beneath.
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

  return { live, statuses, toasts, draw, shown }
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

async function clear($: Engine, session?: { wipe: () => void }) {
  await $.session.end({ reason: 'clear', sessionId: 's', resume: { id: 's' } })
  session?.wipe()
  await $.classic.SessionStart({ source: 'clear' })
}

const START = { cwd: '/tmp', surface: 'terminal', isInteractive: true } as const
const TURN = { answer: '', durationMs: 1, isAborted: false, turnId: 't', reason: 'answer' } as const
const COMPACT = {
  trigger: 'manual' as const,
  messages: [{ role: 'user' as const, text: 'hello', toolUses: [] }],
}

test('shows a placeholder before the first response', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)

  expect(await world.shown($)).toBe('? for shortcuts · ctx ▱▱▱▱▱ —/200k')
  expect(world.statuses).toEqual([undefined])
})

test('measure updates the fill', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  await $.session.measure({ context: ctxAt(84_100), rateLimits: [], changed: ['context'] })

  expect(await world.shown($)).toBe('? for shortcuts · ctx ▰▰▱▱▱ 42% 84.1k/200k')
})

test('turn.complete reports what the turn added', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  world.live.context = ctxAt(20_000)
  await $.turn.complete({ ...TURN, turnId: 't1' })
  expect(await world.shown($)).toBe('? for shortcuts · ctx ▰▱▱▱▱ 10% 20k/200k Δ+20k')

  world.live.context = ctxAt(23_200)
  await $.turn.complete({ ...TURN, turnId: 't2' })
  expect(await world.shown($)).toBe('? for shortcuts · ctx ▰▱▱▱▱ 12% 23.2k/200k Δ+3.2k')
})

test('subagent turns are ignored', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  world.live.context = ctxAt(50_000)
  await $.turn.complete({ ...TURN, agentId: 'sub-1' })

  expect(await world.shown($)).toBe('? for shortcuts · ctx ▱▱▱▱▱ —/200k')
})

test('a compaction without counts marks the next delta', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  world.live.context = ctxAt(150_000)
  await $.turn.complete({ ...TURN, turnId: 't1' })

  await $.session.compact(COMPACT)
  world.live.context = ctxAt(30_000)
  await $.turn.complete({ ...TURN, turnId: 't2' })

  expect(await world.shown($)).toBe('? for shortcuts · ctx ▰▱▱▱▱ 15% 30k/200k Δ−120k compacted')
})

test('/compact updates the meter right away', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  world.live.context = ctxAt(150_000)
  await $.turn.complete({ ...TURN, turnId: 't1' })

  world.live.compactTo = 18_000
  await $.session.compact(COMPACT)
  expect(await world.shown($)).toBe('? for shortcuts · ctx ▱▱▱▱▱ 9% 18k/200k Δ−132k compacted')

  world.live.context = ctxAt(30_000)
  await $.turn.complete({ ...TURN, turnId: 't2' })
  expect(await world.shown($)).toBe('? for shortcuts · ctx ▰▱▱▱▱ 15% 30k/200k Δ+12k')
})

test('an empty reading after compaction keeps its count', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  world.live.context = ctxAt(150_000)
  await $.turn.complete({ ...TURN, turnId: 't1' })

  world.live.compactTo = 18_000
  await $.session.compact(COMPACT)
  await $.session.measure({ context: { window: WINDOW }, rateLimits: [], changed: ['context'] })
  await $.turn.complete({ ...TURN, turnId: 't2' })

  expect(await world.shown($)).toBe('? for shortcuts · ctx ▱▱▱▱▱ 9% 18k/200k Δ−132k compacted')
})

test('a compaction re-arms the 80% toast', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  world.live.context = ctxAt(170_000)
  await $.session.measure({ context: world.live.context, rateLimits: [], changed: ['context'] })
  expect(world.toasts.length).toBe(1)

  world.live.compactTo = 20_000
  await $.session.compact(COMPACT)
  await $.session.measure({ context: ctxAt(165_000), rateLimits: [], changed: ['context'] })
  expect(world.toasts.length).toBe(2)
})

test('subagent and precompute compactions leave the meter alone', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  world.live.context = ctxAt(150_000)
  await $.turn.complete({ ...TURN, turnId: 't1' })

  world.live.compactTo = 18_000
  await $.session.compact({ ...COMPACT, agentId: 'sub-1' })
  await $.session.compact({ ...COMPACT, trigger: 'precompute' })

  expect(await world.shown($)).toBe('? for shortcuts · ctx ▰▰▰▰▱ 75% 150k/200k Δ+150k')
})

test('toasts once at 80% and re-arms below 75%', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  await $.session.measure({ context: ctxAt(162_000), rateLimits: [], changed: ['context'] })
  await $.session.measure({ context: ctxAt(170_000), rateLimits: [], changed: ['context'] })
  expect(world.toasts.length).toBe(1)
  expect(world.toasts[0]).toContain('Context 81% full')

  await $.session.measure({ context: ctxAt(40_000), rateLimits: [], changed: ['context'] })
  await $.session.measure({ context: ctxAt(165_000), rateLimits: [], changed: ['context'] })
  expect(world.toasts.length).toBe(2)
})

test('/clear shows the empty meter in the new session', async ($, on) => {
  const world = engine(on)
  const session = clearable(on)

  await $.session.start(START)
  world.live.context = ctxAt(60_000)
  await $.turn.complete(TURN)
  world.live.context = { window: WINDOW }
  await clear($, session)

  expect(await world.shown($)).toBe('? for shortcuts · ctx ▱▱▱▱▱ —/200k')
})

test('/clear resets the meter where the state was kept', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  world.live.context = ctxAt(170_000)
  await $.turn.complete(TURN)
  world.live.context = { window: WINDOW }
  await clear($)

  expect(await world.shown($)).toBe('? for shortcuts · ctx ▱▱▱▱▱ —/200k')
})

test('reads before another mod’s part of the line', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  await $.session.measure({ context: ctxAt(84_100), rateLimits: [], changed: ['context'] })

  expect(await world.shown($, '5h 31%')).toBe('? for shortcuts · ctx ▰▰▱▱▱ 42% 84.1k/200k · 5h 31%')
})

test('wraps between the parts, never inside one', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  await $.session.measure({ context: ctxAt(84_100), rateLimits: [], changed: ['context'] })

  const row = await world.draw($, '5h 31%')

  expect(row.type === 'Box' && row.props?.flexWrap).toBe('wrap')
  expect(parts(row)).toEqual([
    ['claudeforge-hint:hint', '? for shortcuts · '],
    ['claudeforge-hint:context-meter', 'ctx ▰▰▱▱▱ 42% 84.1k/200k · '],
    ['claudeforge-hint:usage-limits', '5h 31%'],
  ])
})

test('stands in for the engine’s line, so it reads beside the mode', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)

  expect(keepsLine(await world.draw($))).toBe(false)
  expect(keepsLine(await world.draw($, '5h 31%'))).toBe(false)
})

test('an empty hint leaves no leading dot', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)

  expect(await world.shown($, undefined, '')).toBe('ctx ▱▱▱▱▱ —/200k')
  expect(await world.shown($, '5h 31%', '')).toBe('ctx ▱▱▱▱▱ —/200k · 5h 31%')
})

test('colors the gauge by how full it is', async ($, on) => {
  const world = engine(on)

  await $.session.start(START)
  expect(colors(await world.draw($))).toEqual([])

  await $.session.measure({ context: ctxAt(84_100), rateLimits: [], changed: ['context'] })
  expect(colors(await world.draw($))).toEqual([['▰▰▱▱▱ 42%', 'success']])

  await $.session.measure({ context: ctxAt(120_000), rateLimits: [], changed: ['context'] })
  expect(colors(await world.draw($))).toEqual([['▰▰▰▱▱ 60%', 'warning']])

  await $.session.measure({ context: ctxAt(170_000), rateLimits: [], changed: ['context'] })
  expect(colors(await world.draw($))).toEqual([['▰▰▰▰▱ 85%', 'error']])
})

test('the row validates where the hint line is drawn', async ($, on) => {
  engine(on)

  await $.session.start(START)
  await $.session.measure({ context: ctxAt(170_000), rateLimits: [], changed: ['context'] })

  for (const surface of ['terminal', 'desktop'] as const) {
    const ui = await $.ui.mount({
      plugin: 'context-meter',
      surface,
      component: 'PromptHint',
      props: { isDraft: false, isWorking: false, hint: '? for shortcuts' },
    })

    expect(await ui.find({ type: 'Text', text: /ctx/ })).toBeDefined()
    await ui.unmount()
  }
})
