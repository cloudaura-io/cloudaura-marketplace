import { expect, mock, test } from 'claude-code/testing'
import type { Engine } from 'claude-code/testing'
import type { On, RenderElement, RenderNode } from 'claude-code'

import { withPart } from '../hooks/hint'
import type { Part } from '../hooks/hint'

const USAGE = {
  input_tokens: 10,
  output_tokens: 5,
  cache_read_input_tokens: 50_000,
  cache_creation_input_tokens: 200,
  model: 'claude-opus-5-5',
}

const SWITCH = {
  from_model: 'claude-opus-5-5',
  to_model: 'claude-sonnet-5-5',
  requested_model: 'sonnet',
  source: 'command',
  context_tokens: 50_000,
  prompt_cache_warm: true,
  cache_ttl: '5m',
  estimated_cache_write_usd: 0.5,
  pricing: 'catalog',
} as const

const HINT = '? for shortcuts'
const LINE: RenderElement = { type: 'engine', ref: 0 }

// What the mods drew in the engine's line's place, as the person reads it.
function text(node: RenderNode): string {
  if (typeof node === 'string') return node
  return node.type === 'Box' || node.type === 'Text' ? (node.children ?? []).map(text).join('') : ''
}

// Whether the tree keeps the engine's line, which puts it on a row under the mode.
function keepsLine(node: RenderNode): boolean {
  if (typeof node === 'string') return false
  return node.type === 'engine' || ((node.type === 'Box' || node.type === 'Text') && (node.children ?? []).some(keepsLine))
}

// Each colored run, with its color.
function colors(node: RenderNode): [string, unknown][] {
  if (typeof node === 'string' || (node.type !== 'Box' && node.type !== 'Text')) return []
  if (node.type === 'Text' && node.props?.color !== undefined) return [[text(node), node.props.color]]
  return (node.children ?? []).flatMap(colors)
}

// Each of the row's boxes, as its key and what it reads.
function parts(node: RenderNode): [unknown, string][] {
  if (typeof node === 'string' || node.type !== 'Box') return []
  return (node.children ?? []).map(one =>
    typeof one === 'string' || one.type !== 'Box' ? [undefined, text(one)] : [one.props?.key, text(one)],
  )
}

/** The engine beneath the plugin: a mocked clock, and bottoms that record. */
function world(on: On) {
  const clock = mock.clock(on, { now: 1_000_000 })
  const statuses: (string | undefined)[] = []
  const toasts: string[] = []
  const invalidated: string[] = []
  let beneath: [Part, string][] = []

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
  on('session.start', ($, e) => ({ cwd: e.cwd }))
  on('turn.step', async function* ($, e) {
    return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], stopReason: 'end_turn', usage: USAGE }
  })
  on('classic.SessionStart', () => ({}))
  on('classic.PostModelSwitch', () => ({}))
  on('ui.render', { component: 'PromptHint' }, ($, e) =>
    beneath.reduce<RenderElement>((tree, [part, said]) => withPart(tree, e.props.hint, [{ text: said }], part), LINE),
  )

  // Draws the hint line, `others` standing for the other mods' parts drawn beneath.
  async function draw($: Engine, others: [Part, string][] = [], hint = HINT) {
    beneath = others
    return $.ui.render({
      surface: 'terminal',
      component: 'PromptHint',
      requestId: 'hint',
      props: { isDraft: false, isWorking: false, hint },
    })
  }

  async function shown($: Engine, others: [Part, string][] = [], hint = HINT) {
    const row = await draw($, others, hint)
    return row.type === 'engine' ? undefined : text(row)
  }

  return { clock, statuses, toasts, invalidated, draw, shown }
}

async function start($: Engine) {
  await $.session.start({ cwd: '/work', surface: 'terminal', isInteractive: true })
}

async function step($: Engine, agentId?: string) {
  const stream = $.turn.step({ turnId: 't', index: 0, model: 'claude-opus-5-5', messageCount: 2, agentId })
  for await (const _ of stream) {
  }
  return stream.result
}

test('leaves the line alone before the first request, and pins no status', async ($, on) => {
  const w = world(on)
  await start($)
  await w.clock.advance(3000)
  expect(await w.shown($)).toBe(undefined)
  expect(w.statuses).toEqual([undefined])
})

test('counts the 1h TTL down from the last request', async ($, on) => {
  const w = world(on)
  await start($)
  await step($)
  await w.clock.advance(1000)
  expect(await w.shown($)).toBe('? for shortcuts · cache 59:59 left')
  expect(w.statuses).toEqual([undefined])
})

test('warns once near expiry, then reads expired', { options: { ttl: '5m' }, timeoutMs: 30_000 }, async ($, on) => {
  const w = world(on)
  await start($)
  await step($)
  await w.clock.advance(239_000)
  expect(w.toasts).toEqual([])
  expect(colors(await w.draw($))).toEqual([['01:01', 'success']])

  await w.clock.advance(2000)
  expect(w.toasts.length).toBe(1)
  expect(w.toasts[0]).toContain('~1 min')
  expect(colors(await w.draw($))).toEqual([['00:59', 'warning']])

  await w.clock.advance(60_000)
  expect(w.toasts.length).toBe(1)
  expect(await w.shown($)).toBe('? for shortcuts · cache expired <1m ago')

  await w.clock.advance(120_000)
  expect(await w.shown($)).toBe('? for shortcuts · cache expired 2m ago')
})

test('redraws each second while warm, each minute once expired', { options: { ttl: '5m' }, timeoutMs: 30_000 }, async ($, on) => {
  const w = world(on)
  await start($)
  await step($)
  await w.clock.advance(1000)
  w.invalidated.length = 0

  await w.clock.advance(5000)
  expect(w.invalidated.length).toBe(5)

  await w.clock.advance(300_000)
  w.invalidated.length = 0
  await w.clock.advance(60_000)
  expect(w.invalidated.length).toBe(1)
})

test('a new request restarts the countdown and re-arms the warning', { options: { ttl: '5m' }, timeoutMs: 30_000 }, async ($, on) => {
  const w = world(on)
  await start($)
  await step($)
  await w.clock.advance(250_000)
  expect(w.toasts.length).toBe(1)

  await step($)
  await w.clock.advance(1000)
  expect(await w.shown($)).toBe('? for shortcuts · cache 04:59 left')

  await w.clock.advance(240_000)
  expect(w.toasts.length).toBe(2)
})

test("a subagent's request does not refresh the main cache", { options: { ttl: '5m' } }, async ($, on) => {
  const w = world(on)
  await start($)
  await step($)
  await w.clock.advance(120_000)
  await step($, 'agent-1')
  await w.clock.advance(1000)
  expect(await w.shown($)).toBe('? for shortcuts · cache 02:59 left')
})

test("a model switch goes cold and adopts the engine's TTL", async ($, on) => {
  const w = world(on)
  await start($)
  await step($)
  await w.clock.advance(1000)
  expect(await w.shown($)).toBe('? for shortcuts · cache 59:59 left')

  await $.classic.PostModelSwitch(SWITCH)
  await w.clock.advance(1000)
  expect(await w.shown($)).toBe(undefined)

  await step($)
  await w.clock.advance(1000)
  expect(await w.shown($)).toBe('? for shortcuts · cache 04:59 left')
})

test('a resume picks up the time already elapsed; /clear drops it', async ($, on) => {
  const w = world(on)
  await start($)
  await $.classic.SessionStart({ source: 'resume', seconds_since_last_response: 600 })
  await w.clock.advance(1000)
  expect(await w.shown($)).toBe('? for shortcuts · cache 49:59 left')

  await $.classic.SessionStart({ source: 'clear' })
  await w.clock.advance(1000)
  expect(await w.shown($)).toBe(undefined)
})

test('reads between the other mods’ parts, whichever drew first', async ($, on) => {
  const w = world(on)
  await start($)
  await step($)
  await w.clock.advance(1000)

  const ctx: [Part, string] = ['context-meter', 'ctx 18%']
  const limits: [Part, string] = ['usage-limits', '5h 31%']
  const all = '? for shortcuts · ctx 18% · cache 59:59 left · 5h 31%'

  expect(await w.shown($, [ctx])).toBe('? for shortcuts · ctx 18% · cache 59:59 left')
  expect(await w.shown($, [limits])).toBe('? for shortcuts · cache 59:59 left · 5h 31%')
  expect(await w.shown($, [ctx, limits])).toBe(all)
  expect(await w.shown($, [limits, ctx])).toBe(all)
  expect(await w.shown($, [ctx, limits], '')).toBe('ctx 18% · cache 59:59 left · 5h 31%')
})

test('wraps between the parts, never inside one', async ($, on) => {
  const w = world(on)
  await start($)
  await step($)
  await w.clock.advance(1000)

  const row = await w.draw($, [['usage-limits', '5h 31%'], ['context-meter', 'ctx 18%']])

  expect(row.type === 'Box' && row.props?.flexWrap).toBe('wrap')
  expect(parts(row)).toEqual([
    ['claudeforge-hint:hint', '? for shortcuts · '],
    ['claudeforge-hint:context-meter', 'ctx 18% · '],
    ['claudeforge-hint:cache-timer', 'cache 59:59 left · '],
    ['claudeforge-hint:usage-limits', '5h 31%'],
  ])
})

test('stands in for the engine’s line, so it reads beside the mode', async ($, on) => {
  const w = world(on)
  await start($)
  await step($)
  await w.clock.advance(1000)

  expect(keepsLine(await w.draw($))).toBe(false)
  expect(keepsLine(await w.draw($, [['usage-limits', '5h 31%']]))).toBe(false)
})

test('the row validates where the hint line is drawn', async ($, on) => {
  const w = world(on)
  await start($)
  await step($)
  await w.clock.advance(1000)

  for (const surface of ['terminal', 'desktop'] as const) {
    const ui = await $.ui.mount({
      plugin: 'cache-timer',
      surface,
      component: 'PromptHint',
      props: { isDraft: false, isWorking: false, hint: HINT },
    })

    expect(await ui.find({ type: 'Text', text: /cache/ })).toBeDefined()
    await ui.unmount()
  }
})
