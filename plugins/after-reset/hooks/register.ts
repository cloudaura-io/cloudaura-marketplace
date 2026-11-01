import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register, SessionRateLimit, Timer } from 'claude-code'

import type { Job, Limit, Reading } from '../types'
import { asReading, BUFFER_MS, listing, name, parse, pickReset, preview, timing, when } from './when'

const TICK_MS = 30_000
const RETRY_MS = 5 * 60_000
const LIMITS_KEY = 'limits'

const jobs = atom({ plugin: 'after-reset', key: 'jobs' } as const, [])
const seenAt = atom({ plugin: 'after-reset', key: 'seenAt' } as const, null)

// The status line last pinned, so a quiet tick pins nothing.
let pinned: string | undefined

// The tick's interval. session.start runs once per load, never on /clear, and
// a reload drops the old load's timers; cancelling first keeps it one.
let ticker: Timer | undefined

// Whether a tick is under way, so the next one waits its turn.
let isTicking = false

// The queue as a /clear left it, for the session it goes on in: $.state is the
// session's, this module's variables the process's.
let carried: Job[] | undefined

function toLimits(windows: readonly SessionRateLimit[]): Limit[] {
  return windows.map(({ kind, percentUsed, resetsAt }) =>
    resetsAt === undefined ? { kind, percentUsed } : { kind, percentUsed, resetsAt },
  )
}

function byDue(a: Job, b: Job): number {
  return a.dueAt - b.dueAt
}

// The windows are the account's: this session's reading, or the one any
// session saved last, whichever was taken later. A new session has none of its
// own before its first reply.
async function knownLimits($: EngineInterface): Promise<Limit[]> {
  const live = toLimits((await $.session.usage()).rateLimits)
  const mine = { at: (await read($, seenAt)) ?? 0, limits: live }
  const saved = asReading(await $.store.get(LIMITS_KEY))
  const newest = [mine, saved]
    .filter((one): one is Reading => one !== undefined && one.limits.length > 0)
    .reduce<Reading | undefined>((a, b) => (a === undefined || b.at > a.at ? b : a), undefined)

  return newest?.limits ?? []
}

function pin($: EngineInterface, list: readonly Job[], now: number) {
  const next = list[0]
  const text =
    next === undefined
      ? undefined
      : `${list.length === 1 ? 'prompt' : `${list.length} prompts, next`} at ${when(next.dueAt, now)}`

  if (text !== pinned) {
    pinned = text
    $.ui.status(text)
  }
}

// Why the prompt did not enter, or undefined once it has: its turn started, or
// it is queued behind the running one.
async function send($: EngineInterface, job: Job): Promise<string | undefined> {
  try {
    return (await $.prompt.submit({ text: job.text, asUser: true })).drop
  } catch (error) {
    return error instanceof Error ? error.message : String(error)
  }
}

// Sends what is due, one prompt at a time and each taken off the queue only
// once it entered, unless a window is still full: then waits for its reset.
async function sendDue($: EngineInterface) {
  const now = await $.clock.now()

  if ((await read($, jobs)).some(job => job.dueAt <= now)) {
    const reset = pickReset(await knownLimits($), now)

    if (reset?.isFull) {
      const dueAt = reset.at + BUFFER_MS
      const reason = `after the ${name(reset.kind)} resets`

      await update($, jobs, list =>
        list.map(job => (job.dueAt <= now ? { ...job, dueAt, reason } : job)).sort(byDue),
      )
      $.ui.toast(`the ${name(reset.kind)} is still reached; waiting until ${when(dueAt, now)}`, {
        timeoutMs: 8000,
      })
    } else {
      for (;;) {
        const job = (await read($, jobs)).find(one => one.dueAt <= now)

        if (job === undefined) {
          break
        }

        const refused = await send($, job)

        if (refused === undefined) {
          await update($, jobs, list => list.filter(one => one.id !== job.id))
          continue
        }

        // Kept, and tried again later; the ones behind it wait for the next tick.
        const dueAt = now + RETRY_MS

        await update($, jobs, list =>
          list
            .map(one => (one.id === job.id ? { ...one, dueAt, reason: 'retrying after a failed send' } : one))
            .sort(byDue),
        )
        $.ui.toast(`could not send "${preview(job.text)}": ${preview(refused)}; trying again at ${when(dueAt, now)}`, {
          timeoutMs: 15000,
        })
        break
      }
    }
  }

  pin($, await read($, jobs), now)
}

async function tick($: EngineInterface) {
  if (isTicking) {
    return
  }

  isTicking = true
  try {
    await sendDue($)
  } finally {
    isTicking = false
  }
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    await $.command.register({
      name: 'after-reset',
      description: 'Send a prompt once your usage limit resets (or in 2h, at 14:30)',
      argumentHint: '[in 2h | at 14:30] <prompt> | cancel [n]',
      immediate: true,
    })

    // Readings the engine already holds, as when the plugin is turned on
    // mid-session, count from now; a reload keeps the time it had.
    if ((await $.session.usage()).rateLimits.length > 0 && (await read($, seenAt)) === null) {
      const now = await $.clock.now()
      await update($, seenAt, () => now)
    }

    // A reload drops the timer, so this also picks up what came due meanwhile.
    pinned = undefined
    ticker?.cancel()
    ticker = $.clock.every(TICK_MS, () => void tick($))
    void tick($)

    return next(e)
  })

  on('session.measure', async ($, e, next) => {
    if (e.changed.includes('rateLimits') && e.rateLimits.length > 0) {
      const reading: Reading = { at: await $.clock.now(), limits: toLimits(e.rateLimits) }

      await update($, seenAt, () => reading.at)
      await $.store.set(LIMITS_KEY, reading)
    }

    return next(e)
  })

  // A /clear goes on in a new session with its $.state empty and no
  // session.start: the queue is carried over to it, and the person told.
  on('session.end', async ($, e, next) => {
    if (e.reason === 'clear') {
      carried = await read($, jobs)
    }

    return next(e)
  })

  on('classic.SessionStart', async ($, e, next) => {
    const kept = carried

    if (e.source === 'clear' && kept !== undefined) {
      carried = undefined

      if (kept.length > 0) {
        const ids = kept.map(job => job.id)

        await update($, jobs, list => [...list.filter(job => !ids.includes(job.id)), ...kept].sort(byDue))
        $.ui.toast(
          kept.length === 1
            ? '1 scheduled prompt kept; /after-reset cancel drops it'
            : `${kept.length} scheduled prompts kept; /after-reset cancel drops them`,
          { timeoutMs: 8000 },
        )
        pinned = undefined
        pin($, await read($, jobs), await $.clock.now())
      }
    }

    return next(e)
  }).catch(($, e, next) => next(e))

  on('command.run', { command: 'after-reset' }, async ($, e) => {
    const now = await $.clock.now()
    const parsed = parse(e.args, now)

    if (parsed.kind === 'error') {
      return { text: parsed.message }
    }
    if (parsed.kind === 'list') {
      return { text: listing(await read($, jobs), pickReset(await knownLimits($), now), now) }
    }
    if (parsed.kind === 'cancel') {
      const list = await read($, jobs)
      const gone = parsed.which === 'all' ? list : list.slice(parsed.which - 1, parsed.which)

      if (gone.length === 0) {
        return { text: list.length === 0 ? 'Nothing is scheduled.' : `There is no prompt ${parsed.which}; see /after-reset.` }
      }

      const ids = gone.map(job => job.id)
      await update($, jobs, all => all.filter(job => !ids.includes(job.id)))
      pin($, await read($, jobs), now)

      return { text: gone.length === 1 ? `Cancelled: ${preview(gone[0]?.text ?? '')}` : `Cancelled ${gone.length} prompts.` }
    }

    let dueAt = parsed.at
    let reason = ''

    if (dueAt === undefined) {
      const reset = pickReset(await knownLimits($), now)

      if (reset === undefined) {
        return {
          text: 'No reset time is known yet (Claude Code learns it from the first reply). Give one instead: /after-reset in 2h <prompt> or /after-reset at 14:30 <prompt>.',
        }
      }

      dueAt = reset.at + BUFFER_MS
      reason = `after the ${name(reset.kind)} resets`
    }

    const job: Job = { id: crypto.randomUUID(), text: parsed.text, dueAt, reason }
    await update($, jobs, list => [...list, job].sort(byDue))
    pin($, await read($, jobs), now)

    return {
      text: `Will send at ${timing(job, now)}: ${preview(parsed.text)}\nKeep this session open. /after-reset lists, /after-reset cancel undoes.`,
    }
  })

  // A turn cut off by the limit: say how to pick it up once the window resets.
  on('classic.StopFailure', async ($, e, next) => {
    const result = await next(e)

    if (e.error === 'rate_limit' && (await read($, jobs)).length === 0) {
      const now = await $.clock.now()
      const reset = pickReset(await knownLimits($), now)
      const at = reset === undefined ? ' once it resets' : ` at ${when(reset.at + BUFFER_MS, now)}`

      $.ui.toast(`Usage limit reached. /after-reset continue sends "continue"${at}.`, { timeoutMs: 15000 })
    }

    return result
  }).catch(($, e, next) => next(e))
}
