import type { CacheTtl } from '../types'

export const TTL_MS: Record<CacheTtl, number> = { '1h': 60 * 60_000, '5m': 5 * 60_000 }
export const WARN_MS: Record<CacheTtl, number> = { '1h': 5 * 60_000, '5m': 60_000 }

// How close to expiry the cache reads; the hint line draws each in a theme status color.
export type Tone = 'ok' | 'warn' | 'high'

// A run of the timer's text; one without a tone is drawn dim, as the line is.
export type Span = { text: string; tone?: Tone }

const pad = (n: number) => String(n).padStart(2, '0')

/** Time left as MM:SS, rounded up so a fresh 1h reads 60:00. */
export function clock(ms: number): string {
  const seconds = Math.ceil(ms / 1000)

  return `${pad(Math.floor(seconds / 60))}:${pad(seconds % 60)}`
}

/** Time since expiry, coarse: <1m, 12m, 2h 5m. */
export function ago(ms: number): string {
  const minutes = Math.floor(ms / 60_000)

  if (minutes < 1) {
    return '<1m'
  }
  if (minutes < 60) {
    return `${minutes}m`
  }

  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`
}

// Compact, as it shares the prompt's hint line with the mode and other mods.
export function cacheSpans(remaining: number, ttl: CacheTtl): Span[] {
  if (remaining <= 0) {
    return [{ text: `cache expired ${ago(-remaining)} ago` }]
  }

  return [
    { text: 'cache ' },
    { text: clock(remaining), tone: remaining <= WARN_MS[ttl] ? 'warn' : 'ok' },
    { text: ' left' },
  ]
}

export function formatCache(remaining: number, ttl: CacheTtl): string {
  return cacheSpans(remaining, ttl)
    .map(one => one.text)
    .join('')
}
