# CloudAura Marketplace

A collection of plugins for Claude Code by CloudAura.

## Installation

```bash
claude plugin marketplace add cloudaura-io/cloudaura-marketplace
```

## Available Plugins

> [!WARNING]
> **Skills may not appear in autocomplete** due to a [known Claude Code limitation](https://github.com/anthropics/claude-code/issues/18949). Type commands manually or run the plugin's `scripts/setup-autocomplete.sh` script.

| Plugin | Description |
|--------|-------------|
| [conductor](plugins/conductor/) | Context-driven development framework. Measure twice, code once.<br>Ported from [gemini-cli-extensions/conductor](https://github.com/gemini-cli-extensions/conductor). |
| [context-meter](plugins/context-meter/) | Context window fill (tokens and %) and how much the last turn added, on the line under the prompt. |
| [cache-timer](plugins/cache-timer/) | Counts down the prompt cache's TTL on the line under the prompt, with a toast before it expires. |
| [usage-limits](plugins/usage-limits/) | The 5-hour and weekly rate-limit windows with reset countdowns, on the line under the prompt. |
| [after-reset](plugins/after-reset/) | `/after-reset <prompt>` sends a prompt once your usage limit resets, or at a time you name. |

context-meter, cache-timer and usage-limits share one row under the prompt, in that order; install any of them on their own or together.

## Tools

| Tool | Description |
|------|-------------|
| [conductor-tui](tools/conductor-tui/) | Terminal UI for browsing Conductor tracks, phases, and tasks. |

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for local testing instructions.

## License

Apache License 2.0
