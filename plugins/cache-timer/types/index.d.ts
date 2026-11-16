export type CacheTtl = '5m' | '1h'

declare module 'claude-code' {
  interface PluginState {
    'cache-timer': {
      /** When the main conversation's last request reached the API (ms since epoch); null while nothing is warm. */
      lastRequestAt: number | null
      /** The TTL the engine reported on a model switch; null until it has. */
      learnedTtl: CacheTtl | null
      /** The `lastRequestAt` already warned about, so each warm period toasts once. */
      warnedFor: number | null
    }
  }
}
