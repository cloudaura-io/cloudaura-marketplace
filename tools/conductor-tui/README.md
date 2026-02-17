# Conductor TUI

Terminal UI for browsing [Conductor](../../plugins/conductor/) tracks, phases, and tasks. Built with [Ink](https://github.com/vadimdemedes/ink) (React for CLI).

## Install

**Pre-built binary** (Linux/Windows):

```bash
# Linux/Unix
curl -fsSL https://raw.githubusercontent.com/cloudaura-io/cloudaura-marketplace/main/tools/conductor-tui/install.sh | sh
```
```powershell
# Windows (PowerShell)
irm https://raw.githubusercontent.com/cloudaura-io/cloudaura-marketplace/main/tools/conductor-tui/install.ps1 | iex
```

**From source** (requires [Bun](https://bun.sh)):

```bash
bun install && bun run tui.tsx
```

## Usage

Run `conductor-tui` in a repo with a `conductor/` directory. Navigate with arrow keys, Enter to drill down, Esc to go back, `q` to quit. Data auto-refreshes every 2s.
