import type { Job, Limit } from '../types'

const MIN = 60_000

// Sent a little after the reset, so the window has surely rolled over.
export const BUFFER_MS = 2 * MIN

const NAMES: Record<string, string> = {
  five_hour: '5-hour limit',
  seven_day: 'weekly limit',
  seven_day_opus: 'weekly Opus limit',
  seven_day_sonnet: 'weekly Sonnet limit',
  spend_limit: 'spend limit',
}
const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

export const USAGE =
  'Usage: /after-reset <prompt> · /after-reset in 2h30m <prompt> · /after-reset at 14:30 <prompt> · /after-reset cancel [n]'

export function name(kind: string): string {
  return NAMES[kind] ?? `${kind} limit`
}

export type Reset = { kind: string; at: number; isFull: boolean }

// The reset worth waiting for. Every full window must clear, so the latest of
// those; with none full, the nearest.
export function pickReset(limits: readonly Limit[], now: number): Reset | undefined {
  const ahead = limits.flatMap(one => {
    const at = one.resetsAt === undefined ? NaN : Date.parse(one.resetsAt)

    return Number.isNaN(at) || at <= now ? [] : [{ kind: one.kind, at, isFull: one.percentUsed >= 100 }]
  })
  const full = ahead.filter(one => one.isFull)

  if (full.length > 0) {
    return full.reduce((a, b) => (b.at > a.at ? b : a))
  }

  return ahead.reduce<Reset | undefined>((a, b) => (a === undefined || b.at < a.at ? b : a), undefined)
}

export type Parsed =
  | { kind: 'list' }
  | { kind: 'cancel'; which: 'all' | number }
  | { kind: 'schedule'; text: string; at?: number }
  | { kind: 'error'; message: string }

const UNIT = String.raw`(?:h(?:ours?|rs?)?|m(?:in(?:ute)?s?)?)`
const DURATION = new RegExp(String.raw`^in\s+((?:\d+(?:\.\d+)?\s*${UNIT}\s*)+?)(?:\s+([\s\S]*))?$`, 'i')
const AT = /^at\s+(\d{1,2}):(\d{2})(?:\s+([\s\S]*))?$/i
const CANCEL = /^cancel(?:\s+(all|\d+))?$/i

// What reads as a mistyped command rather than a prompt: `cancel` and one more
// word, `in` or `at` and a number. "cancel the deploy", "in the README" pass.
const CANCEL_TYPO = /^cancel\s+\S+$/i
const IN_TYPO = /^in\s+\d/i
const AT_TYPO = /^at\s+\d/i

function durationMs(spec: string): number {
  let total = 0

  for (const [, amount, unit] of spec.matchAll(new RegExp(String.raw`(\d+(?:\.\d+)?)\s*(${UNIT})`, 'gi'))) {
    total += Number(amount) * (/^h/i.test(unit ?? '') ? 60 : 1) * MIN
  }

  return Math.round(total)
}

// The next HH:MM on the local clock: today, or tomorrow once it has passed.
function nextClock(hours: number, minutes: number, now: number): number {
  const at = new Date(now)
  at.setHours(hours, minutes, 0, 0)

  if (at.getTime() <= now) {
    at.setDate(at.getDate() + 1)
  }

  return at.getTime()
}

export function parse(args: string, now: number): Parsed {
  const input = args.trim()

  if (input === '') {
    return { kind: 'list' }
  }

  const cancel = CANCEL.exec(input)

  if (cancel) {
    return { kind: 'cancel', which: cancel[1] === undefined || cancel[1].toLowerCase() === 'all' ? 'all' : Number(cancel[1]) }
  }

  const inMatch = DURATION.exec(input)

  if (inMatch) {
    const spec = (inMatch[1] ?? '').trim()
    const ms = durationMs(spec)
    const text = (inMatch[2] ?? '').trim()

    if (ms <= 0) {
      return { kind: 'error', message: `"${spec}" is no wait. ${USAGE}` }
    }
    if (text === '') {
      return { kind: 'error', message: `Give the prompt to send after the wait. ${USAGE}` }
    }

    return { kind: 'schedule', text, at: now + ms }
  }

  const atMatch = AT.exec(input)

  if (atMatch) {
    const hours = Number(atMatch[1])
    const minutes = Number(atMatch[2])
    const text = (atMatch[3] ?? '').trim()

    if (hours > 23 || minutes > 59) {
      return { kind: 'error', message: `${atMatch[1]}:${atMatch[2]} is not a time of day. ${USAGE}` }
    }
    if (text === '') {
      return { kind: 'error', message: `Give the prompt to send at ${atMatch[1]}:${atMatch[2]}. ${USAGE}` }
    }

    return { kind: 'schedule', text, at: nextClock(hours, minutes, now) }
  }

  if (CANCEL_TYPO.test(input)) {
    return { kind: 'error', message: `"cancel" takes a prompt number or "all". ${USAGE}` }
  }
  if (IN_TYPO.test(input)) {
    return { kind: 'error', message: `"in" takes hours and minutes, like in 2h30m. ${USAGE}` }
  }
  if (AT_TYPO.test(input)) {
    return { kind: 'error', message: `"at" takes a 24-hour time, like at 21:30. ${USAGE}` }
  }

  return { kind: 'schedule', text: input }
}

export function countdown(ms: number): string {
  const minutes = Math.floor(ms / MIN)

  if (minutes < 1) {
    return '<1m'
  }
  if (minutes < 60) {
    return `${minutes}m`
  }

  const hours = Math.floor(minutes / 60)

  if (hours < 24) {
    return `${hours}h${String(minutes % 60).padStart(2, '0')}m`
  }

  return `${Math.floor(hours / 24)}d${hours % 24}h`
}

function isSameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

// A local wall-clock time, with the day when it is not today.
export function clock(at: number, now: number): string {
  const date = new Date(at)
  const hhmm = `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
  const tomorrow = new Date(now)
  tomorrow.setDate(tomorrow.getDate() + 1)

  if (isSameDay(date, new Date(now))) {
    return hhmm
  }
  if (isSameDay(date, tomorrow)) {
    return `tomorrow ${hhmm}`
  }

  return `${DAYS[date.getDay()]} ${date.getDate()} ${MONTHS[date.getMonth()]} ${hhmm}`
}

export function when(at: number, now: number): string {
  return `${clock(at, now)} (in ${countdown(at - now)})`
}

// When a prompt goes, and why then when it waits on a reset.
export function timing(job: Pick<Job, 'dueAt' | 'reason'>, now: number): string {
  return job.reason === '' ? when(job.dueAt, now) : `${when(job.dueAt, now)}, ${job.reason}`
}

export function preview(text: string): string {
  const line = (text.split('\n')[0] ?? '').trim()

  return line.length > 60 || text.includes('\n') ? `${line.slice(0, 57)}...` : line
}

export function listing(jobs: readonly Job[], reset: Reset | undefined, now: number): string {
  const lines =
    jobs.length === 0
      ? ['Nothing is scheduled.']
      : [
          'Scheduled:',
          ...jobs.map((job, at) => `  ${at + 1}. ${timing(job, now)}: ${preview(job.text)}`),
        ]

  if (reset !== undefined) {
    lines.push(`Next reset that matters: the ${name(reset.kind)}${reset.isFull ? ' (full)' : ''} at ${when(reset.at, now)}`)
  }
  lines.push(USAGE)

  return lines.join('\n')
}

export function isLimitList(value: unknown): value is Limit[] {
  return (
    Array.isArray(value) &&
    value.every(
      one =>
        typeof one === 'object' &&
        one !== null &&
        typeof one.kind === 'string' &&
        typeof one.percentUsed === 'number' &&
        (one.resetsAt === undefined || typeof one.resetsAt === 'string'),
    )
  )
}
