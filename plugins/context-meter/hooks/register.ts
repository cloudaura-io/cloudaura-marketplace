import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register, SessionCompacted, SessionContextUsage } from 'claude-code'

import type { ContextDelta, ContextFill } from '../types'
import { ALERT_AT, contextSpans, tokens } from './format'
import { withPart } from './hint'

const REARM_BELOW = 75

const fill = atom({ plugin: 'context-meter', key: 'fill' } as const, null)
const lastTurnEnd = atom({ plugin: 'context-meter', key: 'lastTurnEnd' } as const, null)
const lastDelta = atom({ plugin: 'context-meter', key: 'lastDelta' } as const, null)
const compacting = atom({ plugin: 'context-meter', key: 'compacting' } as const, false)
const alerted = atom({ plugin: 'context-meter', key: 'alerted' } as const, false)

// Writing it redraws the prompt's hint line, which reads it.
async function record($: EngineInterface, ctx: SessionContextUsage) {
  const { tokens, window, percent } = ctx
  const next: ContextFill = { window }

  if (tokens !== undefined) {
    next.tokens = tokens
  }
  if (percent !== undefined) {
    next.percent = percent
  }

  await update($, fill, () => next)
}

async function alertPastThreshold($: EngineInterface, ctx: SessionContextUsage) {
  if (ctx.percent === undefined) {
    return
  }

  const wasAlerted = await read($, alerted)

  if (ctx.percent >= ALERT_AT && !wasAlerted) {
    await update($, alerted, () => true)
    $.ui.toast(
      `Context ${ctx.percent}% full (${tokens(ctx.tokens ?? 0)}/${tokens(ctx.window)}). Consider /compact`,
      { timeoutMs: 8000 },
    )
  } else if (ctx.percent < REARM_BELOW && wasAlerted) {
    await update($, alerted, () => false)
  }
}

// The live reading is empty from a compaction until the next response, so the
// result's own count stands in; without one, the next turn's delta is marked.
async function compacted($: EngineInterface, { tokensBefore, tokensAfter }: SessionCompacted) {
  if (tokensAfter === undefined) {
    await update($, compacting, () => true)

    return
  }

  const { window } = (await $.session.usage()).context
  const ctx: SessionContextUsage = { tokens: tokensAfter, window }

  if (window > 0) {
    ctx.percent = Math.round((tokensAfter / window) * 100)
  }

  const before = tokensBefore ?? (await read($, lastTurnEnd))

  await record($, ctx)
  await alertPastThreshold($, ctx)
  await update($, lastDelta, () =>
    before === null ? null : { tokens: tokensAfter - before, compacted: true },
  )
  await update($, lastTurnEnd, () => tokensAfter)
  await update($, compacting, () => false)
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    // 0.1.0 pinned its own status line; the meter now rides the hint line.
    $.ui.status(undefined)
    await record($, (await $.session.usage()).context)

    return next(e)
  })

  on('session.measure', async ($, e, next) => {
    // An empty reading follows a compaction; it would wipe the count it left.
    if (e.changed.includes('context') && e.context.tokens !== undefined) {
      await record($, e.context)
      await alertPastThreshold($, e.context)
    }

    return next(e)
  })

  on('session.compact', async ($, e, next) => {
    const result = await next(e)

    if (e.agentId === undefined && e.trigger !== 'precompute' && result.skip === undefined) {
      await compacted($, result)
    }

    return result
  }).catch(($, e, next) => next(e))

  on('turn.complete', async ($, e, next) => {
    const result = await next(e)

    if (e.agentId !== undefined) {
      return result
    }

    const { context } = await $.session.usage()
    const now = context.tokens

    if (now !== undefined) {
      const before = await read($, lastTurnEnd)
      const delta: ContextDelta = {
        tokens: now - (before ?? 0),
        compacted: await read($, compacting),
      }

      await update($, lastTurnEnd, () => now)
      await update($, lastDelta, () => delta)
      await update($, compacting, () => false)
      await record($, context)
    }

    return result
  })

  // A /clear goes on in a new session with its $.state empty and no
  // session.start, so the meter is set up again here, in the new one.
  on('classic.SessionStart', async ($, e, next) => {
    if (e.source === 'clear') {
      const { context } = await $.session.usage()

      await update($, lastTurnEnd, () => null)
      await update($, lastDelta, () => null)
      await update($, compacting, () => false)
      await update($, alerted, () => false)
      await update($, fill, () => ({ window: context.window }))
    }

    return next(e)
  }).catch(($, e, next) => next(e))

  // Drawn in the engine's line's place rather than as its `tail`, which is dim
  // text alone. First of the mods' parts on the row (hint.ts's PARTS).
  on('ui.render', { component: 'PromptHint' }, async ($, e, next) => {
    const ctx = await read($, fill)

    if (ctx === null) {
      return next(e)
    }

    const spans = contextSpans(ctx, await read($, lastDelta))

    return withPart(await next(e), e.props.hint, spans, 'context-meter')
  })
}
