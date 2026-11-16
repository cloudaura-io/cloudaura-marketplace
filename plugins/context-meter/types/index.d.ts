export type ContextDelta = { tokens: number; compacted: boolean }
export type ContextFill = { tokens?: number; window: number; percent?: number }

declare module 'claude-code' {
  interface PluginState {
    'context-meter': {
      fill: ContextFill | null
      lastTurnEnd: number | null
      lastDelta: ContextDelta | null
      compacting: boolean
      alerted: boolean
    }
  }
}
