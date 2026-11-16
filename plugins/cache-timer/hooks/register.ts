import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register, Timer } from 'claude-code'

import type { CacheTtl } from '../types'
import { TTL_MS, WARN_MS, cacheSpans, formatCache } from './format'
import { withPart } from './hint'

const lastRequestAt = atom({ plugin: 'cache-timer', key: 'lastRequestAt' } as const, null)
const learnedTtl = atom({ plugin: 'cache-timer', key: 'learnedTtl' } as const, null)
const warnedFor = atom({ plugin: 'cache-timer', key: 'warnedFor' } as const, null)

type Left = { at: number; ttl: CacheTtl; ms: number }

/** What the main conversation's cache has left; null while nothing is warm. */
async function left($: EngineInterface, configuredTtl: CacheTtl): Promise<Left | null> {
  const at = await read($, lastRequestAt)
  if (at === null) return null

  const ttl = (await read($, learnedTtl)) ?? configuredTtl
  return { at, ttl, ms: at + TTL_MS[ttl] - (await $.clock.now()) }
}

let drawn: string | undefined

// The countdown's interval. session.start runs once per load, never on /clear,
// and a reload drops the old load's timers; cancelling first keeps it one.
let ticker: Timer | undefined

async function tick($: EngineInterface, configuredTtl: CacheTtl) {
  const now = await left($, configuredTtl)

  // The countdown reads the clock, which no state write redraws; redraw only
  // when its text moves (each second while warm, each minute once expired).
  const text = now === null ? undefined : formatCache(now.ms, now.ttl)
  if (text !== drawn) {
    drawn = text
    $.ui.invalidate('ui.render')
  }

  if (now === null || now.ms <= 0 || now.ms > WARN_MS[now.ttl]) return
  if ((await read($, warnedFor)) === now.at) return

  await update($, warnedFor, () => now.at)
  $.ui.toast(
    `Prompt cache expires in ~${WARN_MS[now.ttl] / 60_000} min: send something to keep it warm.`,
    { timeoutMs: 15_000 },
  )
}

export const register: Register = (on, options) => {
  const configuredTtl: CacheTtl = options.ttl === '5m' ? '5m' : '1h'

  on('session.start', async ($, e, next) => {
    // 0.1.0 pinned its own status line; the timer now rides the hint line.
    $.ui.status(undefined)
    ticker?.cancel()
    ticker = $.clock.every(1000, () => void tick($, configuredTtl))
    return next(e)
  })

  // Every main-loop request reads (and so refreshes) the cached prefix as it is
  // sent; a subagent's requests carry a prefix of their own.
  on('turn.step', async function* ($, e, next) {
    if (e.agentId !== undefined) return yield* next(e)

    const sentAt = await $.clock.now()
    const response = yield* next(e)
    if (response.usage !== null) await update($, lastRequestAt, () => sentAt)

    return response
  })

  on('classic.SessionStart', async ($, e, next) => {
    if (e.source === 'clear' || e.source === 'compact') {
      await update($, lastRequestAt, () => null)
    } else if (e.seconds_since_last_response !== undefined) {
      const since = e.seconds_since_last_response * 1000
      const now = await $.clock.now()
      await update($, lastRequestAt, () => now - since)
    }
    return next(e)
  })

  // The engine says the real TTL here; a new model starts with a cold cache.
  on('classic.PostModelSwitch', async ($, e, next) => {
    await update($, learnedTtl, () => e.cache_ttl)
    if (e.from_model !== e.to_model) await update($, lastRequestAt, () => null)
    return next(e)
  })

  // Drawn in the engine's line's place, between context-meter's part and
  // usage-limits', which share the row.
  on('ui.render', { component: 'PromptHint' }, async ($, e, next) => {
    const now = await left($, configuredTtl)
    if (now === null) return next(e)

    return withPart(await next(e), e.props.hint, cacheSpans(now.ms, now.ttl), 'cache-timer')
  })
}
