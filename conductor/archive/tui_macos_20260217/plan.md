# Plan: Add macOS Build Targets for conductor-tui

## Phase 1: Build Script & CI Workflow [checkpoint: 586c20f]

- [x] Task: Update `build.sh` to add macOS targets `fc6e96e`
    - [x] Add `bun-darwin-x64` and `bun-darwin-arm64` to the `targets` array
    - [x] Verify output filenames follow existing convention (`conductor-tui-darwin-x64`, `conductor-tui-darwin-arm64`)

- [x] Task: Update CI release workflow for macOS `fc6e96e`
    - [x] Add matrix entry `{ target: bun-darwin-x64, artifact: conductor-tui-darwin-x64 }` to `.github/workflows/release-conductor-tui.yml`
    - [x] Add matrix entry `{ target: bun-darwin-arm64, artifact: conductor-tui-darwin-arm64 }` to `.github/workflows/release-conductor-tui.yml`

- [x] Task: Conductor - User Manual Verification 'Phase 1: Build Script & CI Workflow' (Protocol in workflow.md)

---

## Phase 2: Install Script Update [checkpoint: 46f3247]

- [x] Task: Update `install.sh` to support macOS `3c44b95`
    - [x] Detect OS via `uname -s` and set platform to `darwin` on macOS, `linux` on Linux
    - [x] Build the download URL using `${platform}-${arch}` (e.g., `darwin-arm64`, `linux-x64`)
    - [x] Ensure the script exits cleanly with an unsupported OS message for other platforms

- [x] Task: Conductor - User Manual Verification 'Phase 2: Install Script Update' (Protocol in workflow.md)

---

## Phase 3: Documentation Update [checkpoint: 30ca10d]

- [x] Task: Update `README.md` for macOS support `6c4d63a`
    - [x] Add macOS to the supported platforms list/table
    - [x] Add macOS install command to the installation section

- [x] Task: Conductor - User Manual Verification 'Phase 3: Documentation Update' (Protocol in workflow.md)
