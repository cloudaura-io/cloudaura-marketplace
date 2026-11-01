import type { ContextDelta, ContextFill } from '../types'

const CELLS = 5
const WARN_AT = 50

export const ALERT_AT = 80

// How full the gauge reads; the hint line draws each in a theme status color.
export type Tone = 'ok' | 'warn' | 'high'

// A run of the meter's text; one without a tone is drawn dim, as the line is.
export type Span = { text: string; tone?: Tone }

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

export function tokens(n: number): string {
  const abs = Math.abs(n)

  if (abs < 1000) {
    return String(n)
  }
  if (abs < 999_950) {
    return `${Number((n / 1000).toFixed(1))}k`
  }

  return `${Number((n / 1_000_000).toFixed(2))}M`
}

export function signed(n: number): string {
  return n < 0 ? `−${tokens(-n)}` : `+${tokens(n)}`
}

// Compact, as it shares the prompt's hint line with the mode and other mods.
export function contextSpans(ctx: ContextFill, delta: ContextDelta | null): Span[] {
  if (ctx.tokens === undefined) {
    return [{ text: `ctx ${bar(0)} —/${tokens(ctx.window)}` }]
  }

  const percent =
    ctx.percent ?? (ctx.window > 0 ? Math.round((ctx.tokens / ctx.window) * 100) : 0)
  const spans: Span[] = [
    { text: 'ctx ' },
    { text: `${bar(percent)} ${percent}%`, tone: tone(percent) },
    { text: ` ${tokens(ctx.tokens)}/${tokens(ctx.window)}` },
  ]

  if (delta !== null) {
    spans.push({ text: ` Δ${signed(delta.tokens)}${delta.compacted ? ' compacted' : ''}` })
  }

  return spans
}

export function formatContext(ctx: ContextFill, delta: ContextDelta | null): string {
  return contextSpans(ctx, delta)
    .map(one => one.text)
    .join('')
}
