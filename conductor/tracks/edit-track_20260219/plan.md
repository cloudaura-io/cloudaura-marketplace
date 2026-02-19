# Implementation Plan: Conductor Edit Track Skill

## Phase 1: Skill Scaffold and Foundation [checkpoint: 0f8d52f]

- [x] Task: Create SKILL.md with YAML frontmatter and system directive `7009cce`
    - [x] Create directory `plugins/conductor/skills/edit-track/`
    - [x] Write YAML frontmatter (name, description, argument-hint, disable-model-invocation, allowed-tools)
    - [x] Write Section 1.0 SYSTEM DIRECTIVE matching existing skill patterns
    - [x] Write Section 0.1 CONTEXT AND FILE RESOLUTION with Universal File Resolution Protocol (copied from existing skills)
    - [x] Write Section 1.1 SETUP CHECK protocol (verify Product Definition, Tech Stack, Workflow, Tracks Registry)

- [x] Task: Verify scaffold structure against existing skills `7009cce`
    - [x] Compare frontmatter fields with `new-track/SKILL.md`, `revert/SKILL.md`, and `implement/SKILL.md`
    - [x] Verify Universal File Resolution Protocol section matches verbatim across skills
    - [x] Verify Setup Check section follows established pattern

- [x] Task: Conductor - User Manual Verification 'Skill Scaffold and Foundation' (Protocol in workflow.md) `0f8d52f`

## Phase 2: Track Selection and Editability Rules [checkpoint: 259bf28]

- [x] Task: Write Section 2.0 TRACK SELECTION protocol `83d2e27`
    - [x] Write step to read Tracks Registry and resolve all track directories
    - [x] Write step to read each track's `metadata.json` and filter by status (`new` or `in_progress`)
    - [x] Write step to present editable tracks via AskUserQuestion (guided menu)
    - [x] Write handling for completed/cancelled tracks: inform user to create new track or reopen
    - [x] Write handling for no editable tracks found: halt with informative message

- [x] Task: Write Section 2.1 STATUS-BASED EDITABILITY RULES `83d2e27`
    - [x] Document the editability matrix (new: fully editable, in_progress: constrained, completed/cancelled: locked)
    - [x] Write in-progress constraint rules: completed items `[x]` locked, in-progress `[~]` requires warning + confirmation, pending `[ ]` freely editable
    - [x] Write the warning text and AskUserQuestion confirmation flow for `[~]` task modification
    - [x] Write auto-reset rule: modified `[~]` tasks revert to `[ ]`

- [x] Task: Verify track selection and editability logic `83d2e27`
    - [x] Verify AskUserQuestion format follows spec (header max 12 chars, options with label/description)
    - [x] Verify all status combinations are handled (new, in_progress, completed, cancelled)
    - [x] Verify locked item detection regex patterns match plan.md format (`[x]` with SHAs, checkpoints)

- [x] Task: Conductor - User Manual Verification 'Track Selection and Editability Rules' (Protocol in workflow.md) `259bf28`

## Phase 3: Edit Mode Selection and Mode Implementations

- [x] Task: Write Section 3.0 EDIT MODE SELECTION `7fbf390`
    - [x] Write AskUserQuestion presenting four modes: Edit Spec, Edit Plan, Rescope, Edit Metadata
    - [x] Write routing logic to the correct mode section based on user selection

- [x] Task: Write Section 3.1 MODE 1 - EDIT SPEC `7fbf390`
    - [x] Write step to read and display current spec.md content
    - [x] Write interactive questioning flow for gathering desired changes
    - [x] Write step to generate updated spec content
    - [x] Write Changes log entry logic: append dated `## Changes` section (or prepend within existing section)
    - [x] Write change preview presentation (inline summary format)
    - [x] Write AskUserQuestion confirmation before writing

- [x] Task: Write Section 3.2 MODE 2 - EDIT PLAN `7fbf390`
    - [x] Write step to read and display current plan.md with locked/editable annotations
    - [x] Write interactive flow for plan modifications (add, remove, reorder, rewrite pending items)
    - [x] Write preservation logic for completed items, commit SHAs, and checkpoint annotations
    - [x] Write change preview presentation (inline summary format)
    - [x] Write AskUserQuestion confirmation before writing

- [x] Task: Write Section 3.3 MODE 3 - RESCOPE `7fbf390`
    - [x] Write Stage 1: Edit Spec flow (reuse Mode 1 logic, user approves updated spec)
    - [x] Write Stage 2: Regenerate pending plan from updated spec
    - [x] Write step to read Workflow file and apply TDD methodology to regenerated plan
    - [x] Write preservation logic for completed/in-progress items during plan regeneration
    - [x] Write injection of Phase Completion Verification tasks for new phases
    - [x] Write two-stage approval flow (spec approval, then plan approval)

- [x] Task: Write Section 3.4 MODE 4 - EDIT METADATA `7fbf390`
    - [x] Write step to read and display current metadata.json
    - [x] Write AskUserQuestion for selecting fields to edit (description, type)
    - [x] Write step to update `updated_at` timestamp
    - [x] Write step to update Tracks Registry entry if description changes
    - [x] Write change preview and confirmation flow

- [x] Task: Verify all four edit modes `7fbf390`
    - [x] Verify Edit Spec includes dated Changes log format
    - [x] Verify Edit Plan preserves all `[x]` items and commit SHA references
    - [x] Verify Rescope uses two-stage approval and follows Workflow methodology
    - [x] Verify Edit Metadata updates both metadata.json and tracks.md when description changes
    - [x] Verify all modes use AskUserQuestion with correct format

- [~] Task: Conductor - User Manual Verification 'Edit Mode Selection and Mode Implementations' (Protocol in workflow.md)

## Phase 4: Change Preview, Commit Protocol, and Finalization

- [ ] Task: Write Section 4.0 CHANGE PREVIEW AND WRITE PROTOCOL
    - [ ] Write inline summary preview generation logic (before/after for each modified file)
    - [ ] Write AskUserQuestion approval step before any file writes
    - [ ] Write file write execution steps (spec.md, plan.md, metadata.json as applicable)
    - [ ] Write `updated_at` timestamp update in metadata.json for all edit modes

- [ ] Task: Write Section 5.0 COMMIT PROTOCOL
    - [ ] Write step to stage all modified track files
    - [ ] Write commit message format: `conductor(edit): <description>`
    - [ ] Write examples for each edit mode's commit message
    - [ ] Write completion announcement to user

- [ ] Task: Final structural review of complete SKILL.md
    - [ ] Verify all sections are numbered consistently
    - [ ] Verify all AskUserQuestion calls have proper header, question, options, multiSelect fields
    - [ ] Verify section ordering and flow matches existing skill patterns
    - [ ] Verify no references to undefined sections or protocols
    - [ ] Verify the skill handles edge cases (empty tracks, single-task plans, plans with no pending items)

- [ ] Task: Conductor - User Manual Verification 'Change Preview, Commit Protocol, and Finalization' (Protocol in workflow.md)
