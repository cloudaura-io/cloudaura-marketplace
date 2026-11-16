export type Limit = { kind: string; percentUsed: number; resetsAt?: string }

declare module 'claude-code' {
  interface PluginState {
    'usage-limits': {
      limits: Limit[]
      isFresh: boolean
      hasNoWindows: boolean
      alerted: string[]
    }
  }
}
