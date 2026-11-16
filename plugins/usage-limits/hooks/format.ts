import type { Limit } from '../types'

const CELLS = 5
const WARN_AT = 70

export const ALERT_AT = 90

// How full a window reads; the hint line draws each in a theme status color.
export type Tone = 'ok' | 'warn' | 'high'

// A run of the line's text; one without a tone is drawn dim, as the line is.
export type Span = { text: string; tone?: Tone }

const LABELS: Record<string, string> = {
  five_hour: '5h',
  seven_day: 'wk',
  seven_day_opus: 'wk opus',
  seven_day_sonnet: 'wk sonnet',
  spend_limit: 'spend',
}
const NAMES: Record<string, string> = {
  five_hour: '5-hour limit',
  seven_day: 'Weekly limit',
}
const ORDER = Object.keys(LABELS)

export const WAITING = 'limits: waiting'

export function label(kind: string): string {
  return LABELS[kind] ?? kind
}

export function name(kind: string): string {
  return NAMES[kind] ?? `${label(kind)} limit`
}

export function bar(percent: number): string {
  const filled = Math.max(0, Math.min(CELLS, Math.round(percent / (100 / CELLS))))

  return '▰'.repeat(filled) + '▱'.repeat(CELLS - filled)
}

export function tone(percent: number): Tone {
  if (percent >= ALERT_AT) {
    return 'high'
  }

  return percent >= WARN_AT ? 'warn' : 'ok'
}

export function countdown(ms: number): string {
  const minutes = Math.floor(ms / 60_000)

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

function rank(kind: string): number {
  const at = ORDER.indexOf(kind)

  return at === -1 ? ORDER.length : at
}

function joined(spans: readonly Span[]): string {
  return spans.map(one => one.text).join('')
}

// A window whose reset time has passed reads as 0% until a fresh reading.
export function limitSpans(limit: Limit, now: number): Span[] {
  const resetAt = limit.resetsAt === undefined ? NaN : Date.parse(limit.resetsAt)
  const isReset = !Number.isNaN(resetAt) && resetAt <= now
  const percent = isReset ? 0 : Math.round(limit.percentUsed)
  const when = Number.isNaN(resetAt)
    ? ''
    : isReset
      ? ' (reset)'
      : ` (${countdown(resetAt - now)})`

  return [
    { text: `${label(limit.kind)} ` },
    { text: `${bar(percent)} ${percent}%`, tone: tone(percent) },
    { text: when },
  ]
}

// Compact, as it shares the prompt's hint line with the mode and other mods.
export function limitsSpans(limits: readonly Limit[], now: number, isFresh: boolean): Span[] {
  if (limits.length === 0) {
    return [{ text: WAITING }]
  }

  const sorted = [...limits].sort((a, b) => rank(a.kind) - rank(b.kind))
  const spans: Span[] = isFresh ? [] : [{ text: '~' }]

  sorted.forEach((one, at) => {
    if (at > 0) {
      spans.push({ text: ' · ' })
    }
    spans.push(...limitSpans(one, now))
  })

  return spans
}

export function formatLimit(limit: Limit, now: number): string {
  return joined(limitSpans(limit, now))
}

export function formatLimits(limits: readonly Limit[], now: number, isFresh: boolean): string {
  return joined(limitsSpans(limits, now, isFresh))
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
