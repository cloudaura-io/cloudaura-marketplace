---
name: conductor:implement
description: Executes tasks from the specified track's plan
argument-hint: "[track name]"
disable-model-invocation: true
model: opus
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Agent
  - SendMessage
  - Workflow
  - Workflow(conductor:swarm-implement)
---

## 1.0 SYSTEM DIRECTIVE
You are an AI agent assistant for the Conductor spec-driven development framework. Your current task is to implement a track. You MUST follow this protocol precisely.

CRITICAL: You must validate the success of every tool call. If any tool call fails, you MUST halt the current operation immediately, announce the failure to the user, and await further instructions.

---

## 0.0 PRE-LOADED PROJECT CONTEXT

The following project files have been automatically injected at skill invocation time.
If any section shows "NOT FOUND", the file does not exist — Conductor may not be set up.

### Conductor Index
!`cat conductor/index.md 2>/dev/null || echo "NOT FOUND — conductor/index.md does not exist."`

### Tracks Registry (conductor/tracks.md)
!`cat conductor/tracks.md 2>/dev/null || echo "NOT FOUND — conductor/tracks.md does not exist."`

### Workflow (conductor/workflow.md)
!`cat conductor/workflow.md 2>/dev/null || echo "NOT FOUND — conductor/workflow.md does not exist."`

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
- **Repositories**: `conductor/repos.md` (multi-repo projects only)

**Standard Default Paths (Track):**
- **Specification**: `conductor/tracks/<track_id>/spec.md`
- **Implementation Plan**: `conductor/tracks/<track_id>/plan.md`
- **Metadata**: `conductor/tracks/<track_id>/metadata.json`

### Multi-Repo Projects

Check whether the **Repositories** file (`conductor/repos.md`) exists. If it does, this project is one system spread across several git repositories: read that file before running any git command, and follow its **Multi-Repo Protocol** for every git operation in this command. The protocol overrides the single-repository git instructions in this file (bare `git` commands in the project root, commits without a target repository, plain SHAs).

---

## 1.1 SETUP CHECK
**PROTOCOL: Verify that the Conductor environment is properly set up.**

1.  **Verify Core Context:** Using the **Universal File Resolution Protocol**, resolve and verify the existence of:
    -   **Product Definition**
    -   **Tech Stack**
    -   **Workflow**

2.  **Handle Failure:** If ANY of these are missing (or their resolved paths do not exist), Announce: "Conductor is not set up. Please run `/conductor:setup`." and HALT.

---

## 1.2 EXECUTION MODE
**PROTOCOL: Ask how the track should be implemented. This runs in the main session, which is the only place that talks to the user.**

**User questions stay in this main session.** Subagents and swarm agents never ask the user anything; they return questions in their results. Ask every question in this protocol — including the ones the **Workflow** phrases as plain text (manual verification, yes/no confirmations, doc sync approvals) — with the `AskUserQuestion` tool, here.

1.  **Ask for Mode:** Use the `AskUserQuestion` tool with:
    - **header:** "Mode"
    - **question:** "How should this track be implemented?"
    - **multiSelect:** false
    - **options:**
        1. label: "Swarm", description: "Parallel agents build independent tasks in separate git worktrees (dynamic workflows)"
        2. label: "Single agent", description: "One subagent implements the tasks one by one"

2.  **If "Swarm", Ask for Size:** Use the `AskUserQuestion` tool with:
    - **header:** "Agents"
    - **question:** "How many agents may run at once?"
    - **multiSelect:** false
    - **options:**
        1. label: "15 (Recommended)", description: "Default swarm size"
        2. label: "5", description: "Smaller swarm"
        3. label: "10", description: "Medium swarm"
        4. label: "25", description: "Large swarm; Claude Code runs at most 16 at once unless CLAUDE_CODE_WORKFLOW_MAX_CONCURRENT_AGENTS is raised"
    - A custom answer is any whole number of 1 or more; do not cap it. Remember it as `maxParallel`. Swarm agents never run more than `maxParallel` at once: the workflow enforces it, and in fallback mode you send at most `maxParallel` Agent calls per message.
    - Choosing "Swarm" is the user's explicit request to run the `conductor:swarm-implement` workflow.

3.  **Remember the Mode** and continue with **2.0 TRACK SELECTION**. The mode is used in section **3.0**, step 4.


---

## 2.0 TRACK SELECTION
**PROTOCOL: Identify and select the track to be implemented.**

1.  **Check for User Input:** First, check if the user provided a track name as an argument (e.g., `/conductor:implement <track_description>`).

2.  **Locate and Parse Tracks Registry:**
    -   The **Tracks Registry** content has been pre-loaded in section **0.0 PRE-LOADED PROJECT CONTEXT** above. Use that content directly — do NOT re-read the file unless the pre-loaded content shows "NOT FOUND".
    -   Parse the tracks content by splitting it by the `---` separator to identify each track section. For each section, extract the status (`[ ]`, `[~]`, `[x]`), the track description (from the `##` heading), and the link to the track folder.
    -   **CRITICAL:** If no track sections are found after parsing, announce: "The tracks file is empty or malformed. No tracks to implement." and halt.

3.  **Continue:** Immediately proceed to the next step to select a track.

4.  **Select Track:**
    -   **If a track name was provided:**
        1.  Perform an exact, case-insensitive match for the provided name against the track descriptions you parsed.
        2.  If a unique match is found, confirm the selection with the user: "I found track '<track_description>'. Is this correct?"
        3.  If no match is found, or if the match is ambiguous, inform the user and ask for clarification. Suggest the next available track as below.
    -   **If no track name was provided (or if the previous step failed):**
        1.  **Identify Next Track:** Find the first track in the parsed tracks file that is NOT marked as `[x] Completed`.
        2.  **If a next track is found:**
            -   Announce: "No track name provided. Automatically selecting the next incomplete track: '<track_description>'."
            -   Proceed with this track.
        3.  **If no incomplete tracks are found:**
            -   Announce: "No incomplete tracks found in the tracks file. All tasks are completed!"
            -   Halt the process and await further user instructions.

5.  **Handle No Selection:** If no track is selected, inform the user and await further instructions.

---

## 3.0 TRACK IMPLEMENTATION
**PROTOCOL: Execute the selected track.**

1.  **Announce Action:** Announce which track you are beginning to implement.

2.  **Update Status to 'In Progress':**
    -   Before beginning any work, you MUST update the status of the selected track in the **Tracks Registry** file.
    -   This requires finding the specific heading for the track (e.g., `## [ ] Track: <Description>`) and replacing it with the updated status (e.g., `## [~] Track: <Description>`) in the **Tracks Registry** file you identified earlier.

3.  **Load Track Context:**
    a. **Identify Track Folder:** From the tracks file, identify the track's folder link to get the `<track_id>`.
    b. **Read Files:**
        -   **Track Context:** Using the **Universal File Resolution Protocol**, resolve and read the **Specification** and **Implementation Plan** for the selected track.
        -   **Workflow:** The **Workflow** has been pre-loaded in section **0.0 PRE-LOADED PROJECT CONTEXT** above. Use that content directly — do NOT re-read unless the pre-loaded content shows "NOT FOUND".
    c. **Error Handling:** If you fail to read any of these files, you MUST stop and inform the user of the error.

4.  **Execute Tasks and Update Track Plan:**
    a. **Announce:** State that you will now execute the tasks from the track's **Implementation Plan** by following the procedures in the **Workflow**.
    b. **Swarm Mode:** Execute section **3.1 SWARM IMPLEMENTATION**, then continue with step 5 below.
    c. **Single-Agent Mode:** For each phase of the **Implementation Plan**, in order, skipping phases that already have a `[checkpoint: <sha>]`:
        i. **Dispatch:** Use the `Agent` tool to start one `general-purpose` subagent with `model: "opus"`. Give it the track id, the relative paths of the **Specification**, **Implementation Plan** and **Workflow**, and the phase name, and instruct it to:
            - Implement every `[ ]` or `[~]` task of this phase in order, following the "Task Workflow" section of the **Workflow** precisely (implementation, testing, commits, git notes, plan updates). Skip the `Conductor - User Manual Verification` meta-task.
            - Then execute steps 1 to 4 of the **Workflow**'s "Phase Completion Verification and Checkpointing Protocol" (test coverage, automated tests, manual verification plan), but NOT the user confirmation or the checkpoint commit.
            - In a multi-repo workspace (`conductor/repos.md` exists): read that file first and follow its **Multi-Repo Protocol** for every git command (track branch per repository, commits in the repository that owns the files, `<name>@<sha>` references, plan commits in `conductor/`).
            - Never ask the user anything. If it needs a decision, stop and return the question.
            - Return: the tasks completed with their short SHAs, the test command and result, the manual verification plan, and any open question.
        ii. **Questions:** If the subagent returned a question, ask it with the `AskUserQuestion` tool, send the answer back with `SendMessage` to the same subagent (or dispatch a new one with the answer if it cannot be resumed), and wait for its new result.
        iii. **Manual Verification:** Show the user the manual verification plan, then use the `AskUserQuestion` tool with:
            - **header:** "Verify"
            - **question:** "Does phase '<phase name>' meet your expectations?"
            - **multiSelect:** false
            - **options:**
                1. label: "Yes", description: "Create the phase checkpoint"
                2. label: "Needs changes", description: "Describe what to change"
            - If the user wants changes, send the feedback to the subagent with `SendMessage` (or a new subagent), wait for its new result, and ask again.
        iv. **Checkpoint:** Once confirmed, execute steps 6 to 10 of the "Phase Completion Verification and Checkpointing Protocol" yourself (checkpoint commit, git note with the verification report and the user's confirmation, `[checkpoint: <sha>]` in the plan, plan commit, announcement). Mark the `Conductor - User Manual Verification` meta-task `[x]`.

5.  **Finalize Track:**
    -   After all tasks in the track's local **Implementation Plan** are completed, you MUST update the track's status in the **Tracks Registry**.
    -   This requires finding the specific heading for the track (e.g., `## [~] Track: <Description>`) and replacing it with the completed status (e.g., `## [x] Track: <Description>`).
    -   **Commit Changes:** Stage the **Tracks Registry** file and commit with the message `chore(conductor): Mark track '<track_description>' as complete`.
    -   Announce that the track is fully complete and the tracks file has been updated.

6.  **Pull Requests (Multi-repo workspace only):**
    -   List every repository in the track's `"repos"` with its track branch, base branch and commit count (`git -C <path> rev-list --count <base>..<track branch>`).
    -   Use the `AskUserQuestion` tool with:
        - **header:** "PRs"
        - **question:** "Push the track branches and open one pull request per repository?"
        - **multiSelect:** false
        - **options:**
            1. label: "Push and open PRs", description: "Push each track branch and open a PR against its base branch"
            2. label: "Push only", description: "Push the branches; I will open the PRs myself"
            3. label: "Not now", description: "Leave the branches local"
    -   On "Push and open PRs": push each branch (`git -C <path> push -u origin <track branch>`), then open the pull requests with the hosting CLI available for that remote (`gh`, `az repos`, `glab`). Use the same title for all of them (`<type>: <track description>`); each description names the track and links the sibling pull requests. If no CLI is available for a remote, push and print the URL to create the PR manually.
    -   On "Push only": push each branch and print the compare URLs.
    -   Report every branch and pull request URL.

---

## 3.1 SWARM IMPLEMENTATION
**PROTOCOL: Build the track's tasks with parallel agents. Only used when "Swarm" was chosen in 1.2 EXECUTION MODE.**

0.  **Multi-Repo Workspace:** If `conductor/repos.md` exists, the swarm runs across repositories. Apply these changes to the steps below, and follow the Multi-Repo Protocol for everything else:
    -   **Phase repositories:** For each phase, the repositories it works in are named in its heading (e.g. `## Phase 2: Sync engine (api, portal)`). If a heading names none, ask the user which repositories the phase touches with the `AskUserQuestion` tool (`multiSelect: true`, options from the **Repositories** table).
    -   **Clean Tree (step 1)** runs per phase repository with `git -C <path> status --porcelain`, before any track branch is created. Offer "Commit them" (commit in that repository on its current branch), "Stash them" (`git -C <path> stash push -u -m "conductor: before swarm <track_id>"`, tell the user to `stash pop` later) and "Stop".
    -   **Track branches:** Before the swarm starts on a phase, prepare the track branch in every phase repository (rule 4 of the protocol) and add the repositories to `"repos"` in `metadata.json`. Workers branch their worktrees from these track branches.
    -   **Mark Batch (step 4d)** commits `plan.md`, the **Tracks Registry** and `metadata.json` in the `conductor/` repository.
    -   **Run the Swarm (step 4e):** add `repos: [{ name, path, trackBranch }, ...]` (the phase repositories, `path` relative to the project root) to the workflow `args`. In fallback mode, dispatch `conductor:repo-worker` Agents instead of `conductor:task-worker`, with the same repository list in their prompt; each returns `commits` and `worktrees` lists. The integrator gets each task's `commits` and the same repository list.
    -   **Clean Up Worktrees (step 4f):** each `worktrees` entry has a `repo`: run `git -C <path> worktree remove --force <worktree path>` and `git -C <path> branch -D <branch>`, then `git -C <path> worktree prune` in each repository.
    -   **Verify Plan State (step 4g):** task references are `<repo>@<sha7>`, space-separated when a task spans repositories.
    -   **Leftovers** are implemented in the real checkouts, on the track branches.

1.  **Clean Tree:** Run `git status --porcelain`. Workers build in git worktrees, which only contain committed files. If the output is not empty, use the `AskUserQuestion` tool with:
    - **header:** "Changes"
    - **question:** "There are uncommitted changes. Worktrees won't see them. How do you want to proceed?"
    - **multiSelect:** false
    - **options:**
        1. label: "Commit them", description: "Commit all current changes, then continue"
        2. label: "Stop", description: "Stop so you can handle them yourself"
    - On "Commit them": stage everything and commit with `chore(conductor): Save work before swarm`. On "Stop": halt.

2.  **Engine:** If the `Workflow` tool is available to you, use **workflow mode**. If not, announce once: "Dynamic workflows are not available in this session (enable them in /config). Running the swarm with parallel subagents instead." and use **fallback mode**.

3.  **Paths:** Note the relative paths (from the repository root) of the **Specification**, **Implementation Plan** and **Workflow**. Agents only ever get relative paths. If `[~]` tasks are left from an interrupted swarm in an earlier session, run `git worktree prune` first.

4.  **For each phase of the plan, in order, skip phases that already have a `[checkpoint: <sha>]`:**

    a.  **Resume In-Flight Tasks:** For every task of this phase marked `[~]`, finish it yourself in the main checkout by following the "Task Workflow" of the **Workflow**.

    b.  **Collect Pending Tasks:** List the `[ ]` top-level tasks of this phase, excluding the `Task: Conductor - User Manual Verification ...` meta-task. Give each an id `p<phase number>-t<task number>` and a text made of the task line plus its indented sub-task lines.

    c.  **Small Phase:** If there is at most one pending task, implement it yourself by following the "Task Workflow" and go to step **g**. A swarm is not worth the worktree overhead.

    d.  **Mark Batch:** In `plan.md`, change every pending task of this phase from `[ ]` to `[~]`. Stage `plan.md` and the **Tracks Registry** and commit with `conductor(plan): Start phase '<phase name>' with swarm`. Workers branch from this commit.

    e.  **Run the Swarm:**
        -   **Workflow mode:** Call the `Workflow` tool with `name: "conductor:swarm-implement"` and `args: { phase: "<phase name>", workflowPath: "<relative>", specPath: "<relative>", planPath: "<relative>", tasks: [{ id, text }, ...], maxParallel: <N> }` (a JSON object, not a string). The call returns immediately and runs in the background. Tell the user in one line that the swarm started for this phase and that `/workflows` shows progress, then END YOUR TURN and do nothing else. A task notification with the workflow result re-invokes you; continue with step **f**. If the result is truncated, read the output file the notification names.
        -   **Resuming a workflow run:** Every `Workflow` call returns a run ID (`wf_...`) and a script path. Remember both, together with the exact `args` you passed, and tell the user the run ID in your one-line start message. If the run does not complete — its notification reports it failed or was stopped, or the user interrupted it and asks to continue — resume it instead of starting over:
        1. If the run is still going (check `/workflows` or the task list), stop it first.
        2. Call the `Workflow` tool again with the same `name`, the SAME `args` (unchanged, or the cache misses — only `maxParallel` may change, e.g. when the user asks for more or fewer agents), and `resumeFromRunId: "<run ID>"`. If the tool refuses the name, pass `scriptPath: "<script path>"` instead of `name`.
        3. Agents that already finished return their cached results instantly; only the rest run live.
        4. Handle the new result as if it were the first one.
        Resume only works within the same Claude Code session. The integrator skips tasks that are already `[x]`, so a resumed run does not merge twice.
        -   **Fallback mode:** Do the same work yourself:
            1. Group the pending tasks into ordered waves: tasks in one wave must not depend on each other's code or edit the same files. When unsure, put a task in its own wave.
            2. For each wave, dispatch one `conductor:task-worker` Agent per task with the same task text and relative paths the workflow would pass. Send at most `maxParallel` Agent calls in one message; dispatch the rest after those return. Each worker ends with a JSON object (`taskId`, `status`, `sha`, `worktreePath`, `branch`, `summary`, `files`, `reason`, `question`).
            3. Dispatch one `conductor:integrator` Agent with the wave's `done` tasks in plan order (`taskId`, `task`, `sha`, `summary`, `files`). It ends with a JSON object (`merged`, `failed`, `testCommand`, `testsPassed`, `testOutput`).
            4. If `testsPassed` is false, stop starting new waves; the remaining tasks are "not started".

    f.  **Handle the Result** (`merged`, `failed`, `notStarted`, `worktrees`, `testsPassed`, `testOutput`):
        -   **Clean Up Worktrees:** For each entry in `worktrees`, run `git worktree remove --force <path>` and `git branch -D <branch>`, then `git worktree prune`. Their commits are already cherry-picked or abandoned.
        -   **Report:** Tell the user which tasks were merged (with their new short SHAs), which failed and why, and which were not started.
        -   **Failing Tests:** If `testsPassed` is false, show `testOutput` and follow the **Workflow**'s debugging rule (at most two fix attempts, then ask the user for guidance) before continuing.
        -   **Agent Questions:** For every failed task with a non-empty `question`, ask it with the `AskUserQuestion` tool (offer the sensible answers as options) and use the answer when you implement that task.
        -   **Leftovers:** Implement every failed or not-started task yourself, one by one, by following the "Task Workflow" (they are already `[~]`).

    g.  **Verify Plan State:** Every task of this phase except the meta-task must now be `[x]` with a short SHA in backticks (`- [x] Task: <name> `abc1234``). Fix any line that is not, from `git log`.

    h.  **Phase Checkpoint:** Execute the "Phase Completion Verification and Checkpointing Protocol" of the **Workflow**, asking the manual verification question with the `AskUserQuestion` tool. Then mark the `Conductor - User Manual Verification` meta-task `[x]`.

---

## 4.0 SYNCHRONIZE PROJECT DOCUMENTATION
**PROTOCOL: Update project-level documentation based on the completed track.**

1.  **Execution Trigger:** This protocol MUST only be executed when a track has reached a `[x]` status in the tracks file. DO NOT execute this protocol for any other track status changes.

2.  **Announce Synchronization:** Announce that you are now synchronizing the project-level documentation with the completed track's specifications.

3.  **Load Track Specification:** Read the track's **Specification**.

4.  **Load Project Documents:**
    -   Resolve and read:
        -   **Product Definition**
        -   **Tech Stack**
        -   **Product Guidelines**

5.  **Analyze and Update:**
    a.  **Analyze Specification:** Carefully analyze the **Specification** to identify any new features, changes in functionality, or updates to the technology stack.
    b.  **Update Product Definition:**
        i. **Condition for Update:** Based on your analysis, you MUST determine if the completed feature or bug fix significantly impacts the description of the product itself.
        ii. **Propose and Confirm Changes:** If an update is needed, generate the proposed changes. Then, present them to the user for confirmation:
            > "Based on the completed track, I propose the following updates to the **Product Definition**:"
            > ```diff
            > [Proposed changes here, ideally in a diff format]
            > ```
            > "Do you approve these changes? (yes/no)"
        iii. **Action:** Only after receiving explicit user confirmation, perform the file edits to update the **Product Definition** file. Keep a record of whether this file was changed.
    c.  **Update Tech Stack:**
        i. **Condition for Update:** Similarly, you MUST determine if significant changes in the technology stack are detected as a result of the completed track.
        ii. **Propose and Confirm Changes:** If an update is needed, generate the proposed changes. Then, present them to the user for confirmation:
            > "Based on the completed track, I propose the following updates to the **Tech Stack**:"
            > ```diff
            > [Proposed changes here, ideally in a diff format]
            > ```
            > "Do you approve these changes? (yes/no)"
        iii. **Action:** Only after receiving explicit user confirmation, perform the file edits to update the **Tech Stack** file. Keep a record of whether this file was changed.
    d. **Update Product Guidelines (Strictly Controlled):**
        i. **CRITICAL WARNING:** This file defines the core identity and communication style of the product. It should be modified with extreme caution and ONLY in cases of significant strategic shifts, such as a product rebrand or a fundamental change in user engagement philosophy. Routine feature updates or bug fixes should NOT trigger changes to this file.
        ii. **Condition for Update:** You may ONLY propose an update to this file if the track's **Specification** explicitly describes a change that directly impacts branding, voice, tone, or other core product guidelines.
        iii. **Propose and Confirm Changes:** If the conditions are met, you MUST generate the proposed changes and present them to the user with a clear warning:
            > "WARNING: The completed track suggests a change to the core **Product Guidelines**. This is an unusual step. Please review carefully:"
            > ```diff
            > [Proposed changes here, ideally in a diff format]
            > ```
            > "Do you approve these critical changes to the **Product Guidelines**? (yes/no)"
        iv. **Action:** Only after receiving explicit user confirmation, perform the file edits. Keep a record of whether this file was changed.

6.  **Final Report:** Announce the completion of the synchronization process and provide a summary of the actions taken.
    - **Construct the Message:** Based on the records of which files were changed, construct a summary message.
    - **Commit Changes:**
        - If any files were changed (**Product Definition**, **Tech Stack**, or **Product Guidelines**), you MUST stage them and commit them.
        - **Commit Message:** `docs(conductor): Synchronize docs for track '<track_description>'`
    - **Example (if Product Definition was changed, but others were not):**
        > "Documentation synchronization is complete.
        > - **Changes made to Product Definition:** The user-facing description of the product was updated to include the new feature.
        > - **No changes needed for Tech Stack:** The technology stack was not affected.
        > - **No changes needed for Product Guidelines:** Core product guidelines remain unchanged."
    - **Example (if no files were changed):**
        > "Documentation synchronization is complete. No updates were necessary for project documents based on the completed track."

---

## 5.0 TRACK CLEANUP
**PROTOCOL: Offer to archive or delete the completed track.**

1.  **Execution Trigger:** This protocol MUST only be executed after the current track has been successfully implemented and the `SYNCHRONIZE PROJECT DOCUMENTATION` step is complete.

2.  **Ask for User Choice:** Use the `AskUserQuestion` tool with:
    - **header:** "Cleanup"
    - **question:** "Track '<track_description>' is complete. What would you like to do?"
    - **multiSelect:** false
    - **options:**
        1. label: "Review (Recommended)", description: "Run /conductor:review to verify changes before finalizing"
        2. label: "Archive", description: "Move track folder to conductor/archive/ and remove from registry"
        3. label: "Delete", description: "Permanently delete track folder and remove from registry"
        4. label: "Skip", description: "Do nothing, leave track in the registry"

3.  **Handle User Response:**
    *   **If user chooses "Review":**
        *   Announce: "Please run `/conductor:review` to verify your changes. You will be able to archive or delete the track after the review."
    *   **If user chooses "Archive":**
        i.   **Create Archive Directory:** Check for the existence of `conductor/archive/`. If it does not exist, create it.
        ii.  **Archive Track Folder:** Move the track's folder from its current location (resolved via the **Tracks Directory**) to `conductor/archive/<track_id>`.
        iii. **Remove from Tracks File:** Read the content of the **Tracks Registry** file, remove the entire section for the completed track (the part that starts with `---` and contains the track description), and write the modified content back to the file.
        iv.  **Commit Changes:** Stage the **Tracks Registry** file and `conductor/archive/`. Commit with the message `chore(conductor): Archive track '<track_description>'`.
        v.   **Announce Success:** Announce: "Track '<track_description>' has been successfully archived."
    *   **If user chooses "Delete":**
        i. **CRITICAL WARNING:** Before proceeding, you MUST ask for a final confirmation due to the irreversible nature of the action.
            > "WARNING: This will permanently delete the track folder and all its contents. This action cannot be undone. Are you sure you want to proceed? (yes/no)"
        ii. **Handle Confirmation:**
            - **If 'yes'**:
                a. **Delete Track Folder:** Resolve the **Tracks Directory** and permanently delete the track's folder from `<Tracks Directory>/<track_id>`.
                b. **Remove from Tracks File:** Read the content of the **Tracks Registry** file, remove the entire section for the completed track, and write the modified content back to the file.
                c. **Commit Changes:** Stage the **Tracks Registry** file and the deletion of the track directory. Commit with the message `chore(conductor): Delete track '<track_description>'`.
                d. **Announce Success:** Announce: "Track '<track_description>' has been permanently deleted."
            - **If 'no' (or anything else)**:
                a. **Announce Cancellation:** Announce: "Deletion cancelled. The track has not been changed."
    *   **If user chooses "Skip" or provides any other input:**
        *   Announce: "Okay, the completed track will remain in your tracks file for now."
