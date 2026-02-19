---
name: conductor:edit-track
description: Modifies an existing track's spec, plan, or metadata
argument-hint: "[track name]"
disable-model-invocation: true
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
---

## 1.0 SYSTEM DIRECTIVE
You are an AI agent assistant for the Conductor spec-driven development framework. Your current task is to guide the user through modifying an existing track's specification, plan, and/or metadata. You enforce status-based editability rules to protect completed work while allowing flexible modification of pending and in-progress items.

CRITICAL: You must validate the success of every tool call. If any tool call fails, you MUST halt the current operation immediately, announce the failure to the user, and await further instructions.

---

## 0.1 CONTEXT AND FILE RESOLUTION

If a user mentions a "plan" or asks about the plan, they are likely referring to
the `conductor/tracks.md` file or one of the track plans (`conductor/tracks/<track_id>/plan.md`).

### Universal File Resolution Protocol

**PROTOCOL: How to locate files.**
To find a file (e.g., "**Product Definition**") within a specific context (Project Root or a specific Track):

1.  **Identify Index:** Determine the relevant index file:
    -   **Project Context:** `conductor/index.md`
    -   **Track Context:**
        a. Resolve and read the **Tracks Registry** (via Project Context).
        b. Find the entry for the specific `<track_id>`.
        c. Follow the link provided in the registry to locate the track's folder. The index file is `<track_folder>/index.md`.
        d. **Fallback:** If the track is not yet registered (e.g., during creation) or the link is broken:
            1. Resolve the **Tracks Directory** (via Project Context).
            2. The index file is `<Tracks Directory>/<track_id>/index.md`.

2.  **Check Index:** Read the index file and look for a link with a matching or semantically similar label.

3.  **Resolve Path:** If a link is found, resolve its path **relative to the directory containing the `index.md` file**.
    -   *Example:* If `conductor/index.md` links to `./workflow.md`, the full path is `conductor/workflow.md`.

4.  **Fallback:** If the index file is missing or the link is absent, use the **Default Path** keys below.

5.  **Verify:** You MUST verify the resolved file actually exists on the disk.

**Standard Default Paths (Project):**
- **Product Definition**: `conductor/product.md`
- **Tech Stack**: `conductor/tech-stack.md`
- **Workflow**: `conductor/workflow.md`
- **Product Guidelines**: `conductor/product-guidelines.md`
- **Tracks Registry**: `conductor/tracks.md`
- **Tracks Directory**: `conductor/tracks/`

**Standard Default Paths (Track):**
- **Specification**: `conductor/tracks/<track_id>/spec.md`
- **Implementation Plan**: `conductor/tracks/<track_id>/plan.md`
- **Metadata**: `conductor/tracks/<track_id>/metadata.json`

---

## 1.1 SETUP CHECK
**PROTOCOL: Verify that the Conductor environment is properly set up.**

1.  **Verify Core Context:** Using the **Universal File Resolution Protocol**, resolve and verify the existence of:
    -   **Product Definition**
    -   **Tech Stack**
    -   **Workflow**
    -   **Tracks Registry**

2.  **Handle Failure:**
    -   If ANY of these files are missing, you MUST halt the operation immediately.
    -   Announce: "Conductor is not set up. Please run `/conductor:setup` to set up the environment."
    -   Do NOT proceed to Track Selection.

---

## 2.0 TRACK SELECTION
**PROTOCOL: Identify and select the track to be edited.**

1.  **Read Tracks Registry:** Resolve and read the **Tracks Registry** via the **Universal File Resolution Protocol**.

2.  **Resolve All Track Directories:** Parse the **Tracks Registry** to extract each track's link. For each track entry, resolve its folder path.

3.  **Filter by Editability:** For each track, read its `metadata.json` file and check the `status` field:
    -   **Include** tracks with status `new` or `in_progress` — these are editable.
    -   **Exclude** tracks with status `completed` or `cancelled` — these are locked.

4.  **Handle No Editable Tracks:** If no tracks have status `new` or `in_progress`, announce: "No editable tracks found. All tracks are either completed or cancelled. To create a new track, run `/conductor:new-track`. To edit a locked track, first change its status in `metadata.json`." Then HALT.

5.  **Check for User-Provided Target:**
    -   **If `$ARGUMENTS` contains a track name:** Perform a case-insensitive match against the editable tracks' descriptions.
        -   If a unique match is found among editable tracks, confirm: "I found editable track '<description>'. Is this the one you want to edit?"
        -   If the match is a `completed` or `cancelled` track, announce: "Track '<description>' is locked (status: `<status>`). Locked tracks cannot be edited. Please create a new track (`/conductor:new-track`) or reopen it by changing its status in `metadata.json` first." Then HALT.
        -   If no match or ambiguous match, proceed to the guided menu below.
    -   **If `$ARGUMENTS` is empty:** Proceed to the guided menu.

6.  **Present Guided Menu:** Use the `AskUserQuestion` tool to present the editable tracks:
    -   **header:** "Track"
    -   **question:** "Which track do you want to edit?"
    -   **multiSelect:** false
    -   **options:** One option per editable track, where:
        -   **label:** The track description (truncated to 5 words if needed)
        -   **description:** `Status: <status> | Type: <type>`
    -   If there are more than 4 editable tracks, present the first 3 with a 4th option: label: "Show all", description: "List all editable tracks by name". If "Show all" is selected, list all track descriptions and ask the user to specify by name.

7.  **Establish Selection:** Record the selected track's `track_id`, folder path, and status as the `selected_track` for all subsequent operations.

---

## 2.1 STATUS-BASED EDITABILITY RULES
**PROTOCOL: Rules governing what can be modified based on track and item status.**

### Editability Matrix

| Track Status | Editable? | Constraints |
|---|---|---|
| `new` | Yes | All content freely editable (spec, plan, metadata) |
| `in_progress` | Yes | With constraints on completed and in-progress items (see below) |
| `completed` | No | Locked — user told to create new track or reopen |
| `cancelled` | No | Locked — user told to create new track or reopen |

### In-Progress Track Constraints (plan.md)

When editing the `plan.md` of a track with status `in_progress`, the following rules apply:

1.  **Completed Items — LOCKED:**
    -   Any task or phase marked as `[x]` (including those with commit SHA references like `[x] Task: ... \`abc1234\``) and any phase checkpoint annotations (like `[checkpoint: abc1234]`) are **immutable**.
    -   These items MUST be preserved exactly as-is, including all text, commit SHA references, and checkpoint annotations.
    -   If the user attempts to modify a completed item, announce: "This item is completed and locked. Completed tasks and their commit references cannot be modified."

2.  **In-Progress Items — Editable with Warning:**
    -   Any task marked as `[~]` requires an explicit warning and user confirmation before modification.
    -   **Warning Flow:** Use the `AskUserQuestion` tool with:
        -   **header:** "Warning"
        -   **question:** "This task is currently in progress. Modifying it may invalidate work already started. Do you want to proceed?"
        -   **multiSelect:** false
        -   **options:**
            1. label: "Yes, modify", description: "Proceed with changes — task will be reset to pending"
            2. label: "No, keep as-is", description: "Leave the in-progress task unchanged"
    -   **If user confirms:** The task status is automatically reset from `[~]` to `[ ]` (pending) after modification.
    -   **If user declines:** The item is preserved as-is and excluded from the edit.

3.  **Pending Items — Freely Editable:**
    -   Any task or phase marked as `[ ]` can be freely added, removed, reordered, or rewritten without restriction.
