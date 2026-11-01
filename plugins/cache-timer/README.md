# cache-timer

Counts down the prompt cache's time-to-live, starting from the main conversation's last API request. It shows on the line under the prompt:

![cache-timer showing cache 59:56 left](docs/cache-timer.png)

- The countdown turns yellow in the last 5 minutes (the last minute on a 5-minute TTL), and a toast reminds you to send something to keep the cache warm.
- Once the cache has expired it reads `cache expired 12m ago`.
- Only main-conversation requests count, because subagents cache a prefix of their own.
- `/clear`, a compaction or a switch to a different model empties the cache, so the timer disappears until the next request.

It shares the row with [context-meter](../context-meter/) and [usage-limits](../usage-limits/) when they're installed, and sits between them.

![context-meter, cache-timer and usage-limits sharing one row](docs/shared-row.png)

## Options

| Option | Default | Meaning |
|--------|---------|---------|
| `ttl` | `1h` | The TTL to count down from (`1h` or `5m`) until Claude Code reports the real one, which it does on a model switch. Use `5m` when you're in usage overage or on a 5-minute API setup. |

Set it when you install, or later with `claude plugin configure cache-timer`.

## Install

```bash
claude plugin marketplace add cloudaura-io/cloudaura-marketplace
claude plugin install cache-timer@cloudaura-marketplace
```

## Tests

```bash
claude plugin test plugins/cache-timer
```
