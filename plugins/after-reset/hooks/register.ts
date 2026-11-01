import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register, SessionRateLimit } from 'claude-code'

import type { Job, Limit } from '../types'
import { BUFFER_MS, isLimitList, listing, name, parse, pickReset, preview, timing, when } from './when'

const TICK_MS = 30_000
const LIMITS_KEY = 'limits'

const jobs = atom({ plugin: 'after-reset', key: 'jobs' } as const, [])

// The status line last pinned, so a quiet tick pins nothing.
let pinned: string | undefined

function toLimits(windows: readonly SessionRateLimit[]): Limit[] {
  return windows.map(({ kind, percentUsed, resetsAt }) =>
    resetsAt === undefined ? { kind, percentUsed } : { kind, percentUsed, resetsAt },
  )
}

function byDue(a: Job, b: Job): number {
  return a.dueAt - b.dueAt
}

// This session's readings, else the last any session saw: the windows are the
// account's, and a new session has none before its first reply.
async function knownLimits($: EngineInterface): Promise<Limit[]> {
  const { rateLimits } = await $.session.usage()

  if (rateLimits.length > 0) {
    return toLimits(rateLimits)
  }

  const cached = await $.store.get(LIMITS_KEY)

  return isLimitList(cached) ? cached : []
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

// Sends what is due, unless a window is still full: then waits for its reset.
async function tick($: EngineInterface) {
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
      let due: Job[] = []

      await update($, jobs, list => {
        due = list.filter(job => job.dueAt <= now)

        return list.filter(job => job.dueAt > now)
      })
      for (const job of due) {
        $.prompt.submit({ text: job.text, asUser: true }).catch(() => {
          $.ui.toast(`could not send "${preview(job.text)}"`, { timeoutMs: 15000 })
        })
      }
    }
  }

  pin($, await read($, jobs), now)
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    await $.command.register({
      name: 'after-reset',
      description: 'Send a prompt once your usage limit resets (or in 2h, at 14:30)',
      argumentHint: '[in 2h | at 14:30] <prompt> | cancel [n]',
      immediate: true,
    })

    // A reload drops the timer, so this also picks up what came due meanwhile.
    pinned = undefined
    $.clock.every(TICK_MS, () => void tick($))
    void tick($)

    return next(e)
  })

  on('session.measure', async ($, e, next) => {
    if (e.changed.includes('rateLimits') && e.rateLimits.length > 0) {
      await $.store.set(LIMITS_KEY, toLimits(e.rateLimits))
    }

    return next(e)
  })

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
