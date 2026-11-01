import type { RenderElement, RenderNode } from 'claude-code'

import type { Span, Tone } from './format'

// Kept the same in context-meter, cache-timer and usage-limits: the three share
// the row.

const COLORS: Record<Tone, string> = { ok: 'success', warn: 'warning', high: 'error' }

// The mods' parts in reading order, after the engine's hint.
const PARTS = ['context-meter', 'cache-timer', 'usage-limits'] as const

export type Part = (typeof PARTS)[number]

// Names the mods' row, so whichever mod's hook runs later finds the others'. A
// Text takes no key, so each run sits in a Box keyed by whose it is:
// `claudeforge-hint:context-meter`.
const ROW = 'claudeforge-hint'
const HINT = 'hint'
const DOT = ' · '

// One run of the row and whose it is: the engine's hint, or a mod's part.
type Run = { owner: string; node: RenderNode }

function span({ text, tone }: Span): RenderElement {
  return {
    type: 'Text',
    props: tone === undefined ? { dimColor: true } : { color: COLORS[tone] },
    children: [text],
  }
}

function run(spans: readonly Span[]): RenderElement {
  return { type: 'Text', children: spans.filter(one => one.text !== '').map(span) }
}

function rank(owner: string): number {
  if (owner === HINT) {
    return -1
  }

  const at = (PARTS as readonly string[]).indexOf(owner)

  return at === -1 ? PARTS.length : at
}

// Dim like the engine's own line, the gauges in the theme's status colors. Each
// run is a Box of its own, so a row too narrow for them all wraps between runs,
// never inside one; the dot rides the end of the run before it, so no wrapped
// row starts with one. A run wider than the whole row is cut where it ends.
function row(runs: readonly Run[]): RenderElement {
  return {
    type: 'Box',
    props: { key: ROW, flexDirection: 'row', flexWrap: 'wrap' },
    children: runs.map((one, at) => ({
      type: 'Box',
      props: { key: `${ROW}:${one.owner}`, flexShrink: 1 },
      children: [
        {
          type: 'Text',
          props: { wrap: 'truncate-end' },
          children: at === runs.length - 1 ? [one.node] : [one.node, span({ text: DOT })],
        },
      ],
    })),
  }
}

// The run one of the row's Boxes holds; undefined for anything else.
function runOf(node: RenderNode): Run | undefined {
  if (typeof node === 'string' || node.type !== 'Box') {
    return undefined
  }

  const key = node.props?.key
  const [line] = node.children ?? []

  if (typeof key !== 'string' || !key.startsWith(`${ROW}:`) || typeof line === 'string' || line?.type !== 'Text') {
    return undefined
  }

  const [first] = line.children ?? []

  return first === undefined ? undefined : { owner: key.slice(ROW.length + 1), node: first }
}

function runsOf(node: RenderElement): Run[] | undefined {
  if (node.type !== 'Box' || node.props?.key !== ROW) {
    return undefined
  }

  const runs = (node.children ?? []).map(runOf)

  return runs.every((one): one is Run => one !== undefined) ? runs : undefined
}

// The row is drawn in the engine's line's place: its hint, then each mod's part
// in PARTS' order. A tree with no `engine` element stands in for the line, and
// the engine draws it beside the mode; one that keeps the element gets a row of
// its own under the mode. A mod that finds the row beneath it adds its part
// there, so the parts keep their order whichever mod's hook runs first.
export function withPart(
  beneath: RenderElement,
  hint: string,
  spans: readonly Span[],
  part: Part,
): RenderElement {
  const mine: Run = { owner: part, node: run(spans) }
  const runs = runsOf(beneath)

  if (runs !== undefined) {
    return row([...runs, mine].sort((a, b) => rank(a.owner) - rank(b.owner)))
  }
  if (beneath.type === 'engine') {
    return row(hint === '' ? [mine] : [{ owner: HINT, node: run([{ text: hint }]) }, mine])
  }

  // Another plugin's tree: the part goes after it.
  return { type: 'Box', children: [beneath, run([{ text: DOT }, ...spans])] }
}
