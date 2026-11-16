import { expect, test } from 'claude-code/testing'

import { bar, formatContext, signed, tokens } from '../hooks/format'

test('tokens are compact', () => {
  expect(tokens(950)).toBe('950')
  expect(tokens(84_123)).toBe('84.1k')
  expect(tokens(200_000)).toBe('200k')
  expect(tokens(999_960)).toBe('1M')
  expect(tokens(1_020_000)).toBe('1.02M')
})

test('signed deltas', () => {
  expect(signed(3_210)).toBe('+3.2k')
  expect(signed(-120_400)).toBe('−120.4k')
  expect(signed(0)).toBe('+0')
})

test('bar fills by fifths and clamps', () => {
  expect(bar(0)).toBe('▱▱▱▱▱')
  expect(bar(42)).toBe('▰▰▱▱▱')
  expect(bar(50)).toBe('▰▰▰▱▱')
  expect(bar(100)).toBe('▰▰▰▰▰')
  expect(bar(130)).toBe('▰▰▰▰▰')
})

test('before the first response', () => {
  expect(formatContext({ window: 200_000 }, null)).toBe('ctx ▱▱▱▱▱ —/200k')
})

test('fill and last turn', () => {
  const ctx = { tokens: 84_100, window: 200_000, percent: 42 }

  expect(formatContext(ctx, null)).toBe('ctx ▰▰▱▱▱ 42% 84.1k/200k')
  expect(formatContext(ctx, { tokens: 3_200, compacted: false })).toBe(
    'ctx ▰▰▱▱▱ 42% 84.1k/200k Δ+3.2k',
  )
  expect(formatContext(ctx, { tokens: -120_400, compacted: true })).toBe(
    'ctx ▰▰▱▱▱ 42% 84.1k/200k Δ−120.4k compacted',
  )
})

test('percent is derived when absent', () => {
  expect(formatContext({ tokens: 50_000, window: 200_000 }, null)).toBe('ctx ▰▱▱▱▱ 25% 50k/200k')
})
