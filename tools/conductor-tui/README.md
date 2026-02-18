# Conductor TUI

Terminal UI for browsing [Conductor](../../plugins/conductor/) tracks, phases, and tasks. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Go).

## Install

**Pre-built binary** (Linux/macOS/Windows):

```bash
# Linux/Unix
curl -fsSL https://raw.githubusercontent.com/cloudaura-io/cloudaura-marketplace/main/tools/conductor-tui/install.sh | sh
```
```powershell
# Windows (PowerShell)
irm https://raw.githubusercontent.com/cloudaura-io/cloudaura-marketplace/main/tools/conductor-tui/install.ps1 | iex
```

**From source** (requires [Go](https://go.dev/) 1.25+):

```bash
cd tools/conductor-tui
go build -o conductor-tui ./cmd/conductor-tui
```

## Usage

Run `conductor-tui` in a repo with a `conductor/` directory. Navigate with arrow keys, Enter to drill down, Esc to go back, `q` to quit. Press `a` to toggle archived tracks. Data auto-refreshes every 2s.

## Project Structure

```
tools/conductor-tui/
├── cmd/
│   └── conductor-tui/
│       └── main.go              # entrypoint
├── internal/
│   ├── data/                    # types, metadata, plan parsing, track discovery
│   ├── tui/                     # Bubble Tea model, views, keys, styles
│   └── util/                    # string helpers, status colors
├── testdata/                    # test fixtures
├── build.sh                     # cross-compilation script
├── install.sh / install.ps1     # install scripts
├── go.mod / go.sum
└── README.md
```

## Development

```bash
# Run tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Build for current platform
go build -o conductor-tui ./cmd/conductor-tui

# Cross-compile all targets
bash build.sh
```
