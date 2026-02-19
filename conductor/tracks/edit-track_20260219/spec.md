# Specification: Conductor Edit Track Skill

## Overview

Implement a new `/conductor:edit-track` skill that allows users to modify an existing track's specification, plan, and metadata. The skill enforces status-based editability rules to protect completed work while allowing flexible modification of pending and in-progress items.

## Functional Requirements

### FR-1: Track Selection (Guided Menu)

- On invocation, the skill reads the **Tracks Registry** and lists all tracks with status `new` or `in_progress` from their `metadata.json`.
- Tracks are presented via `AskUserQuestion` as selectable options.
- Tracks with status `completed` or `cancelled` are excluded from the menu. If a user attempts to edit one directly, the skill informs them: "This track is locked. Please create a new track (`/conductor:new-track`) or reopen it by changing its status first."
- If no editable tracks exist, the skill halts with an informative message.

### FR-2: Status-Based Editability Rules

| Track Status | Editable? | Constraints |
|---|---|---|
| `new` | Yes | All content freely editable |
| `in_progress` | Yes | With constraints (see below) |
| `completed` | No | Locked - user told to create new track or reopen |
| `cancelled` | No | Locked - user told to create new track or reopen |

**In-Progress Track Constraints (plan.md):**
- **Completed items** (`[x]` with commit SHAs or phase checkpoints): Locked. Cannot be modified. The skill must preserve these exactly as-is, including all commit SHA references and checkpoint annotations.
- **In-progress items** (`[~]`): Editable only after an explicit warning and user confirmation. Warning text: "This task is currently in progress. Modifying it may invalidate work already started. Do you want to proceed?" If modified, the task is automatically reset from `[~]` to `[ ]` (pending).
- **Pending items** (`[ ]`): Freely editable - can be added, removed, reordered, or rewritten.

### FR-3: Edit Mode Selection

After selecting a track, the user chooses an edit mode via `AskUserQuestion`:

#### Mode 1: Edit Spec
- Opens the track's `spec.md` for interactive modification.
- The skill presents the current spec content and asks the user what changes they want to make.
- After modifications, a dated `## Changes` log entry is appended to the bottom of `spec.md` for traceability. Format:
  ```
  ## Changes

  ### YYYY-MM-DD
  - <description of change>
  ```
  If a `## Changes` section already exists, the new dated entry is prepended within it (most recent first).

#### Mode 2: Edit Plan
- Opens the track's `plan.md` for interactive modification of pending tasks and phases.
- The skill displays the current plan, clearly marking which items are locked (completed/in-progress) and which are editable (pending).
- Supports: adding new tasks/phases, removing pending tasks/phases, reordering pending items, and rewriting pending task descriptions.
- All completed work, commit references, and checkpoint annotations are preserved exactly.

#### Mode 3: Rescope
- A combined operation: update the spec AND regenerate the remaining pending portions of the plan.
- **Stage 1 (Spec):** Follow the Edit Spec flow. User approves the updated spec first.
- **Stage 2 (Plan):** Based on the approved updated spec, regenerate only the pending (`[ ]`) portions of `plan.md`. All completed (`[x]`) and in-progress (`[~]`) items are preserved. The regenerated plan follows the **Workflow** methodology (TDD structure, Phase Completion Verification tasks, etc.).
- Each stage requires separate user approval (two-stage approval flow).

#### Mode 4: Edit Metadata
- Allows modification of the track's `description` and/or `type` fields in `metadata.json`.
- Updates the `updated_at` timestamp.
- If the description changes, the corresponding entry in the **Tracks Registry** (`tracks.md`) is also updated to reflect the new description.

### FR-4: Change Preview

- Before writing any changes to disk, the skill presents an inline summary of proposed changes in a markdown code block.
- The summary clearly shows what will change: sections/items added, removed, or modified.
- The user must approve the changes via `AskUserQuestion` before files are written.

### FR-5: Commit Protocol

- All file modifications are committed together in a single commit.
- Commit message format: `conductor(edit): <description of what was edited>`.
- Examples:
  - `conductor(edit): Update spec for edit-track skill`
  - `conductor(edit): Rescope plan for auth feature track`
  - `conductor(edit): Update metadata type for bugfix track`
- The `updated_at` field in `metadata.json` is always updated when any track artifact is modified.

### FR-6: Universal File Resolution Protocol

- The skill follows the Universal File Resolution Protocol (as defined in other Conductor skills) for locating all project and track files.
- Includes the standard Setup Check to verify core Conductor files exist.

### FR-7: AskUserQuestion Format

- All user interactions use the `AskUserQuestion` tool with proper `header`, `question`, `options`, and `multiSelect` fields.
- Questions are asked sequentially (one at a time).

## Non-Functional Requirements

- **Consistency:** The skill's SKILL.md follows the same structural patterns as existing Conductor skills (setup check, file resolution protocol, phased execution).
- **Safety:** No data loss - completed work is never modified or deleted.
- **Traceability:** All changes are logged (spec changes log, commit messages) for audit purposes.

## Acceptance Criteria

1. The skill file exists at `plugins/conductor/skills/edit-track/SKILL.md` with correct YAML frontmatter.
2. Invoking `/conductor:edit-track` presents a guided menu of editable tracks (status `new` or `in_progress`).
3. Selecting a completed/cancelled track results in a clear rejection message.
4. All four edit modes (Edit Spec, Edit Plan, Rescope, Edit Metadata) function as specified.
5. Completed tasks/phases in `plan.md` are never modified regardless of edit mode.
6. In-progress tasks require explicit warning and confirmation before modification, and are auto-reset to pending after changes.
7. A dated Changes log entry is appended to `spec.md` when it is modified.
8. The Rescope mode uses two-stage approval (spec then plan).
9. An inline summary preview is shown before any file writes.
10. All modifications are committed with `conductor(edit): ...` prefixed messages.
11. The `metadata.json` `updated_at` field is refreshed on any edit.
12. The Tracks Registry entry is updated when metadata description changes.

## Out of Scope

- Reopening completed/cancelled tracks (changing status back to `in_progress`) - this would be a separate skill.
- Editing the Workflow, Product Definition, or Tech Stack files.
- Merging or splitting tracks.
- Editing multiple tracks simultaneously.
