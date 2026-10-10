---
name: conductor:swarm
description: Implements or reviews a track with a swarm of parallel agents (default 5, configurable), using Claude Code dynamic workflows
argument-hint: "implement|review [track name] [--max N]"
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
  - Workflow
  - Workflow(conductor:swarm-implement)
  - Workflow(conductor:swarm-review)
  - Agent(conductor:task-worker, conductor:integrator, conductor:reviewer)
---

## 1.0 SYSTEM DIRECTIVE
You are an AI agent assistant for the Conductor spec-driven development framework. Your task is to implement or review a track with a swarm of parallel agents. You orchestrate from the main session: you own every user question, every phase checkpoint, and the final state of `plan.md`. You MUST follow this protocol precisely.

The user typing `/conductor:swarm` is their explicit request to run the Conductor swarm workflows below.

**User questions stay in this main session.** Only you talk to the user. Swarm agents (workflow agents and fallback subagents) never ask the user anything; they return questions in their results. Ask every question — your own, the ones agents return, and the ones the sibling protocols and the **Workflow** phrase as plain text (yes/no confirmations, manual verification, doc sync approvals) — with the `AskUserQuestion` tool, here, never inside a workflow or subagent. Never tell an agent to ask the user.

CRITICAL: You must validate the success of every tool call. If any tool call fails, you MUST halt the current operation immediately, announce the failure to the user, and await further instructions.

**Sibling protocols** (read them when a step below says so; do not re-implement them):
- Implement: `${CLAUDE_SKILL_DIR}/../implement/SKILL.md`
- Review: `${CLAUDE_SKILL_DIR}/../review/SKILL.md`

**Swarm limit (`maxParallel`):** never more than `maxParallel` agents at once. The workflows enforce it; in fallback mode you enforce it (at most `maxParallel` Agent calls per message). `maxParallel` defaults to 5 and is set in section 1.1, step 2.

---

## 0.0 PRE-LOADED PROJECT CONTEXT

### Tracks Registry (conductor/tracks.md)
!`cat conductor/tracks.md 2>/dev/null || echo "NOT FOUND — conductor/tracks.md does not exist."`

### Workflow (conductor/workflow.md)
!`cat conductor/workflow.md 2>/dev/null || echo "NOT FOUND — conductor/workflow.md does not exist."`

---

## 1.1 MODE AND PREFLIGHT

1.  **Mode:** The user provided: `$ARGUMENTS`. If the first word is `implement` or `review`, use it and treat the rest as the track name (after removing the swarm size, step 2). Otherwise use the `AskUserQuestion` tool with:
    - **header:** "Swarm"
    - **question:** "What should the swarm do?"
    - **multiSelect:** false
    - **options:**
        1. label: "Implement (Recommended)", description: "Build the pending tasks of a track in parallel worktrees"
        2. label: "Review", description: "Review a track's commits with parallel reviewers and verification"

2.  **Swarm Size:** Set `maxParallel` from the arguments or the user's request, in any form: a flag (`--max 10`, `--max=3`) or plain words ("max agent 10", "10 agents", "use 3 agents", "maks 4 agenty"). Remove it from the track name. Default to 5 when none is given. Any whole number of 1 or more is valid; do not cap it — Claude Code's workflow runtime enforces its own concurrency limit, which the user can raise with `CLAUDE_CODE_WORKFLOW_MAX_CONCURRENT_AGENTS`. Announce the swarm size in your start message.

3.  **Setup Check:** Execute section **1.1 SETUP CHECK** of the Implement protocol.

4.  **Clean Tree:** Run `git status --porcelain`. Workers build in git worktrees, which only contain committed files. If the output is not empty, use the `AskUserQuestion` tool with:
    - **header:** "Changes"
    - **question:** "There are uncommitted changes. Worktrees won't see them. How do you want to proceed?"
    - **multiSelect:** false
    - **options:**
        1. label: "Commit them", description: "Commit all current changes, then continue"
        2. label: "Stop", description: "Stop so you can handle them yourself"
    - On "Commit them": stage everything and commit with `chore(conductor): Save work before swarm`. On "Stop": halt.

5.  **Engine:** If the `Workflow` tool is available to you, use **workflow mode**. If not, announce once: "Dynamic workflows are not available in this session (enable them in /config). Running the swarm with parallel subagents instead." and use **fallback mode**.

6.  **Track Selection:** Execute section **2.0 TRACK SELECTION** of the Implement protocol, using the pre-loaded Tracks Registry above. Track headings may be written as `## [ ] Track: <Description>` or `- [ ] **Track: <Description>**`; accept both, and keep whichever format the file uses when you update it. For review mode, a track marked `[x]` is also selectable.

7.  **Load Track Context:** Resolve and read the track's **Specification** and **Implementation Plan**. Note their paths relative to the repository root — agents only ever get relative paths.

---

## 1.2 RESUMING A WORKFLOW RUN

Every `Workflow` call returns a run ID (`wf_...`) and a script path. Remember both, together with the exact `args` you passed, and tell the user the run ID in your one-line start message.

If the run does not complete — its notification reports it failed or was stopped, or the user interrupted it and asks to continue — resume it instead of starting over:

1.  If the run is still going (check `/workflows` or the task list), stop it first.
2.  Call the `Workflow` tool again with the same `name`, the SAME `args` (unchanged, or the cache misses — only `maxParallel` may change, e.g. when the user asks for more or fewer agents), and `resumeFromRunId: "<run ID>"`. If the tool refuses the name, pass `scriptPath: "<script path>"` instead of `name`.
3.  Agents that already finished return their cached results instantly; only the rest run live. The integrator skips tasks that are already `[x]`, so a resumed implement run does not merge twice.
4.  Handle the new result as if it were the first one.

Resume only works within the same Claude Code session. In a new session, start a fresh run: implement mode treats leftover `[~]` tasks as in-flight (section 2.0, step 2a); run `git worktree prune` first.

---

## 2.0 IMPLEMENT MODE

1.  **Mark Track In Progress:** If the track is `[ ]` in the **Tracks Registry**, change it to `[~]` (do not commit yet).

2.  **For each phase of the plan, in order, skip phases that already have a `[checkpoint: <sha>]`:**

    a.  **Resume In-Flight Tasks:** For every task of this phase marked `[~]`, finish it yourself in the main checkout by following the "Task Workflow" of the **Workflow**.

    b.  **Collect Pending Tasks:** List the `[ ]` top-level tasks of this phase, excluding the `Task: Conductor - User Manual Verification ...` meta-task. Give each an id `p<phase number>-t<task number>` and a text made of the task line plus its indented sub-task lines.

    c.  **Small Phase:** If there is at most one pending task, implement it yourself by following the "Task Workflow" and go to step **g**. A swarm is not worth the worktree overhead.

    d.  **Mark Batch:** In `plan.md`, change every pending task of this phase from `[ ]` to `[~]`. Stage `plan.md` and the **Tracks Registry** and commit with `conductor(plan): Start phase '<phase name>' with swarm`. Workers branch from this commit.

    e.  **Run the Swarm:**
        -   **Workflow mode:** Call the `Workflow` tool with `name: "conductor:swarm-implement"` and `args: { phase: "<phase name>", workflowPath: "<relative>", specPath: "<relative>", planPath: "<relative>", tasks: [{ id, text }, ...], maxParallel: <N> }` (a JSON object, not a string). The call returns immediately and runs in the background. Tell the user in one line that the swarm started for this phase and that `/workflows` shows progress, then END YOUR TURN and do nothing else. A task notification with the workflow result re-invokes you; continue with step **f**. If the result is truncated, read the output file the notification names. If the run failed or was stopped, follow **1.2 RESUMING A WORKFLOW RUN**.
        -   **Fallback mode:** Do the same work yourself:
            1. Group the pending tasks into ordered waves: tasks in one wave must not depend on each other's code or edit the same files. When unsure, put a task in its own wave.
            2. For each wave, dispatch one `conductor:task-worker` Agent per task with the same task text and relative paths the workflow would pass. Send at most `maxParallel` Agent calls in one message; dispatch the rest after those return. Each worker ends with a JSON object (`taskId`, `status`, `sha`, `worktreePath`, `branch`, `summary`, `files`, `reason`).
            3. Dispatch one `conductor:integrator` Agent with the wave's `done` tasks in plan order (`taskId`, `task`, `sha`, `summary`, `files`). It ends with a JSON object (`merged`, `failed`, `testCommand`, `testsPassed`, `testOutput`).
            4. If `testsPassed` is false, stop starting new waves; the remaining tasks are "not started".

    f.  **Handle the Result** (`merged`, `failed`, `notStarted`, `worktrees`, `testsPassed`, `testOutput`):
        -   **Clean Up Worktrees:** For each entry in `worktrees`, run `git worktree remove --force <path>` and `git branch -D <branch>`, then `git worktree prune`. Their commits are already cherry-picked or abandoned.
        -   **Report:** Tell the user which tasks were merged (with their new short SHAs), which failed and why, and which were not started.
        -   **Failing Tests:** If `testsPassed` is false, show `testOutput` and follow the **Workflow**'s debugging rule (at most two fix attempts, then ask the user for guidance) before continuing.
        -   **Agent Questions:** For every failed task with a non-empty `question`, ask it with the `AskUserQuestion` tool (offer the sensible answers as options) and use the answer when you implement that task.
        -   **Leftovers:** Implement every failed or not-started task yourself, one by one, by following the "Task Workflow" (they are already `[~]`).

    g.  **Verify Plan State:** Every task of this phase except the meta-task must now be `[x]` with a short SHA in backticks (`- [x] Task: <name> `abc1234``). Fix any line that is not, from `git log`.

    h.  **Phase Checkpoint:** Execute the "Phase Completion Verification and Checkpointing Protocol" of the **Workflow**, including the manual verification with the user. Then mark the `Conductor - User Manual Verification` meta-task `[x]`.

3.  **Finalize:** When every phase is checkpointed, read the Implement protocol and execute step **5. Finalize Track** of section **3.0**, then sections **4.0 SYNCHRONIZE PROJECT DOCUMENTATION** and **5.0 TRACK CLEANUP**.

---

## 3.0 REVIEW MODE

1.  **Scope:** The swarm reviews a track's commits. If the user asked to review uncommitted changes (`current`), tell them to run `/conductor:review` instead and halt.

2.  **Revision Range:** From the track's `plan.md`, collect the recorded commit SHAs. The range is `<first sha>^..<last sha>`. Verify each end with `git rev-parse`.

3.  **Changed Files:** Run `git diff --numstat <range>`. Build `files: [{ path, lines }]` with `lines` = added + deleted. Skip binary entries (`-`), lock files and generated assets.

4.  **Run the Swarm:**
    -   **Workflow mode:** Call the `Workflow` tool with `name: "conductor:swarm-review"` and `args: { range: "<range>", planPath: "<relative>", specPath: "<relative>", files: [...], maxParallel: <N> }`. Tell the user in one line that the review swarm started and that `/workflows` shows progress, then END YOUR TURN and do nothing else. The task notification with the result re-invokes you; continue with step 5. If the run failed or was stopped, follow **1.2 RESUMING A WORKFLOW RUN**.
    -   **Fallback mode:** Dispatch `conductor:reviewer` Agents yourself: one with lens `tests` over the whole range (it runs the test suite), plus one per lens `intent`, `style`, `correctness` for each chunk of files (about 300 changed lines per chunk). At most `maxParallel` Agent calls per message. Merge duplicate findings (same file and start line), then for each Critical or High finding dispatch one `conductor:reviewer` as a skeptic and drop the finding if it is refuted.

5.  **Report:** Read the Review protocol. Write the report exactly in the format of its section **2.4 Output Findings**, using the confirmed `findings` and the `tests` result. If `unreviewedSlices` is above zero, say so in the Summary. Then execute its section **3.0 COMPLETION PHASE**.
