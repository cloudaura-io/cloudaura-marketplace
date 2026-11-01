export type Limit = { kind: string; percentUsed: number; resetsAt?: string }

// A prompt waiting to be sent: `dueAt` in epoch ms, `reason` how it was timed.
export type Job = { id: string; text: string; dueAt: number; reason: string }

// The windows as a session read them, and when (epoch ms).
export type Reading = { at: number; limits: Limit[] }

declare module 'claude-code' {
  interface PluginState {
    'after-reset': {
      jobs: Job[]
      // When this session last read the windows (epoch ms).
      seenAt: number | null
    }
  }
}
