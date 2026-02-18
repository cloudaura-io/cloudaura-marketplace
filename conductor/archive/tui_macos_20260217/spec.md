# Spec: Add macOS Build Targets for conductor-tui

## Overview

Add macOS build support to `conductor-tui`, covering both Apple Silicon (`arm64`) and Intel (`x64`) architectures. This involves updating the build script, the CI release workflow, the install script, and the README.

## Background

`conductor-tui` currently produces binaries for Linux (x64, arm64) and Windows (x64). macOS is not supported despite being a primary development platform for many users. Bun supports cross-compilation to `bun-darwin-x64` and `bun-darwin-arm64` targets.

## Functional Requirements

1. **Build Script (`build.sh`)**: Add `bun-darwin-x64` and `bun-darwin-arm64` to the build targets, producing artifacts named `conductor-tui-darwin-x64` and `conductor-tui-darwin-arm64`.

2. **CI Release Workflow (`.github/workflows/release-conductor-tui.yml`)**: Add two new matrix entries for macOS builds:
   - `{ target: bun-darwin-x64, artifact: conductor-tui-darwin-x64 }`
   - `{ target: bun-darwin-arm64, artifact: conductor-tui-darwin-arm64 }`

3. **Install Script (`install.sh`)**: Update to detect macOS (`Darwin`) via `uname -s` and download the appropriate `darwin-<arch>` binary instead of a Linux binary.

4. **README (`README.md`)**: Update platform support table/section to include macOS, and update installation instructions to show macOS install command.

## Non-Functional Requirements

- The macOS binaries must be produced via Bun's cross-compilation (no macOS runner required in CI).
- The install script must remain POSIX-compatible (`/bin/sh`).

## Acceptance Criteria

- [ ] `build.sh` produces `conductor-tui-darwin-x64` and `conductor-tui-darwin-arm64` in `dist/` when run locally.
- [ ] The CI release workflow uploads macOS artifacts to GitHub Releases on new `conductor-tui-v*` tags.
- [ ] Running `install.sh` on macOS downloads and installs the correct darwin binary.
- [ ] README documents macOS as a supported platform with installation instructions.

## Out of Scope

- macOS-specific packaging (`.dmg`, Homebrew formula) — plain binary only.
- Code signing / notarization of the macOS binary.
