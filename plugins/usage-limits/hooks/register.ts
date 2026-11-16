import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register, SessionRateLimit, Timer } from 'claude-code'

import type { Limit } from '../types'
import { ALERT_AT, countdown, isLimitList, limitsSpans, name } from './format'
import { withPart } from './hint'

const REFRESH_MS = 60_000
const STORE_KEY = 'last'

const limits = atom({ plugin: 'usage-limits', key: 'limits' } as const, [])
const isFresh = atom({ plugin: 'usage-limits', key: 'isFresh' } as const, false)
const hasNoWindows = atom({ plugin: 'usage-limits', key: 'hasNoWindows' } as const, false)
const alerted = atom({ plugin: 'usage-limits', key: 'alerted' } as const, [])

// The redraw's interval. session.start runs once per load, never on /clear,
// and a reload drops the old load's timers; cancelling first keeps it one.
let ticker: Timer | undefined

function toLimits(windows: readonly SessionRateLimit[]): Limit[] {
  return windows.map(({ kind, percentUsed, resetsAt }) =>
    resetsAt === undefined ? { kind, percentUsed } : { kind, percentUsed, resetsAt },
  )
}

async function alertPastThreshold($: EngineInterface, list: readonly Limit[]) {
  const seen = await read($, alerted)
  const now = await $.clock.now()
  const due = list.filter(
    one => one.percentUsed >= ALERT_AT && !seen.includes(`${one.kind}@${one.resetsAt ?? ''}`),
  )

  if (due.length === 0) {
    return
  }

  for (const one of due) {
    const resetAt = one.resetsAt === undefined ? NaN : Date.parse(one.resetsAt)
    const when = Number.isNaN(resetAt) ? '' : ` (resets in ${countdown(resetAt - now)})`
    $.ui.toast(`${name(one.kind)} at ${Math.round(one.percentUsed)}%${when}`, { timeoutMs: 8000 })
  }

  const keys = due.map(one => `${one.kind}@${one.resetsAt ?? ''}`)
  await update($, alerted, list => [...list, ...keys].slice(-20))
}

// This session's readings, else the last any session saw: the windows are the
// account's, so a new session (a /clear's too) need not wait for its own.
async function seed($: EngineInterface) {
  const { rateLimits } = await $.session.usage()

  if (rateLimits.length > 0) {
    const list = toLimits(rateLimits)
    await update($, limits, () => list)
    await update($, isFresh, () => true)
  } else if ((await read($, limits)).length === 0) {
    // Nothing yet this session (a hot reload keeps $.state): last session's.
    const cached = await $.store.get(STORE_KEY)

    if (isLimitList(cached) && cached.length > 0) {
      await update($, limits, () => cached)
      await update($, isFresh, () => false)
    }
  }
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    // 0.1.0 pinned its own status line; the limits now ride the hint line.
    $.ui.status(undefined)
    await seed($)

    // Keeps the countdowns current between readings.
    ticker?.cancel()
    ticker = $.clock.every(REFRESH_MS, () => $.ui.invalidate('ui.render'))

    return next(e)
  })

  // A /clear goes on in a new session with its $.state empty and no
  // session.start, so the readings are put back here, in the new one.
  on('classic.SessionStart', async ($, e, next) => {
    if (e.source === 'clear') {
      await seed($)
    }

    return next(e)
  }).catch(($, e, next) => next(e))

  on('session.measure', async ($, e, next) => {
    if (e.changed.includes('rateLimits') && e.rateLimits.length > 0) {
      const list = toLimits(e.rateLimits)

      await update($, limits, () => list)
      await update($, isFresh, () => true)
      await update($, hasNoWindows, () => false)
      await $.store.set(STORE_KEY, list)
      await alertPastThreshold($, list)
    }

    return next(e)
  })

  // A finished main turn that still reports no windows means this account has
  // none (an API key, Bedrock, Vertex): leave the line to the others.
  on('turn.complete', async ($, e, next) => {
    const result = await next(e)

    if (e.agentId === undefined && e.usage !== undefined) {
      const { rateLimits } = await $.session.usage()

      if (rateLimits.length === 0) {
        await update($, limits, () => [])
        await update($, hasNoWindows, () => true)
      }
    }

    return result
  })

  // Drawn in the engine's line's place rather than as its `tail`, which is dim
  // text alone. Last of the mods' parts on the row (hint.ts's PARTS).
  on('ui.render', { component: 'PromptHint' }, async ($, e, next) => {
    if (await read($, hasNoWindows)) {
      return next(e)
    }

    const spans = limitsSpans(await read($, limits), await $.clock.now(), await read($, isFresh))

    return withPart(await next(e), e.props.hint, spans, 'usage-limits')
  })
}
