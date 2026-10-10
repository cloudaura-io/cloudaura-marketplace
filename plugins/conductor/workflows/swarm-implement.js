export const meta = {
  name: 'swarm-implement',
  description: 'Conductor swarm: implement one plan phase with parallel task workers in isolated worktrees (default 5 at once)',
  whenToUse: 'Run only by /conductor:swarm implement, which passes the phase and its pending tasks in args. If invoked with no args (a user typed the bare slash command), do not call Workflow: tell the user to run /conductor:swarm implement.',
  phases: [
    { title: 'Plan', detail: 'group pending tasks into dependency-free waves' },
    { title: 'Build', detail: 'one task-worker per task in its own worktree, at most maxParallel at once' },
    { title: 'Integrate', detail: 'serial cherry-pick, git note, plan.md update, full test run after each wave' },
  ],
}

let A = args
if (typeof A === 'string') { try { A = JSON.parse(A) } catch (e) { A = null } }
if (!A || !Array.isArray(A.tasks) || !A.tasks.length) {
  log('started without a phase to implement -- nothing dispatched')
  return { started: false, next: 'Tell the user to run /conductor:swarm implement; this workflow needs the phase and tasks that skill supplies.' }
}

// Swarm size from args.maxParallel (default 5). ponytail: capped at 16, the runtime's default concurrency ceiling.
const MAX = Math.min(16, Math.max(1, Math.floor(Number(A.maxParallel)) || 5))
log(`swarm size: ${MAX} agent(s) at once`)

// Run fn over items with at most MAX agents in flight.
async function pool(items, fn) {
  const out = new Array(items.length).fill(null)
  let next = 0
  await parallel(Array.from({ length: Math.min(MAX, items.length) }, () => async () => {
    while (next < items.length) {
      const k = next++
      try { out[k] = await fn(items[k], k) } catch (e) { out[k] = null }
    }
  }))
  return out
}

const ctx = `Track files (paths relative to the repository root, which is your working directory): workflow ${A.workflowPath}, spec ${A.specPath}, plan ${A.planPath}. Phase: "${A.phase}".`
const byId = Object.fromEntries(A.tasks.map(t => [t.id, t]))
const order = id => A.tasks.findIndex(t => t.id === id)

phase('Plan')
// No planner answer = fully sequential waves, never "everything at once".
let waves = A.tasks.map(t => [t.id])
if (A.tasks.length > 1) {
  const plan = await agent(
    `${ctx}\nGroup these pending tasks into ordered waves. Tasks in one wave are built in parallel in separate worktrees, so they must not depend on each other's code and should not edit the same files. A task that needs another task's output goes in a later wave. Read the plan, spec and relevant code to judge. Do not edit anything. Use the task ids exactly.\n${JSON.stringify(A.tasks, null, 1)}`,
    {
      label: 'plan-waves', phase: 'Plan',
      schema: {
        type: 'object',
        properties: { waves: { type: 'array', items: { type: 'array', items: { type: 'string' } } }, reasoning: { type: 'string' } },
        required: ['waves'],
      },
    })
  if (plan) waves = plan.waves
  else log('planner failed -- running tasks one per wave')
}
// Drop unknown/duplicate ids, append anything the planner skipped as a final wave.
const seen = new Set()
waves = waves.map(w => w.filter(id => byId[id] && !seen.has(id) && seen.add(id)).sort((a, b) => order(a) - order(b))).filter(w => w.length)
const skipped = A.tasks.map(t => t.id).filter(id => !seen.has(id))
if (skipped.length) { log(`planner skipped ${skipped.join(', ')} -- added as a final wave`); waves.push(skipped) }
log(`${waves.length} wave(s): ${waves.map(w => w.length).join(', ')} task(s)`)

const BUILT = {
  type: 'object',
  properties: {
    taskId: { type: 'string' },
    status: { type: 'string', enum: ['done', 'failed'] },
    sha: { type: 'string' },
    worktreePath: { type: 'string' },
    branch: { type: 'string' },
    summary: { type: 'string' },
    files: { type: 'array', items: { type: 'string' } },
    reason: { type: 'string' },
    question: { type: 'string', description: 'question for the user, or empty' },
  },
  required: ['taskId', 'status', 'sha', 'worktreePath', 'branch', 'summary', 'files', 'reason', 'question'],
}
const MERGED = {
  type: 'object',
  properties: {
    merged: { type: 'array', items: { type: 'object', properties: { taskId: { type: 'string' }, sha: { type: 'string' } }, required: ['taskId', 'sha'] } },
    failed: { type: 'array', items: { type: 'object', properties: { taskId: { type: 'string' }, reason: { type: 'string' } }, required: ['taskId', 'reason'] } },
    testCommand: { type: 'string' },
    testsPassed: { type: 'boolean' },
    testOutput: { type: 'string' },
  },
  required: ['merged', 'failed', 'testCommand', 'testsPassed', 'testOutput'],
}

const out = { started: true, merged: [], failed: [], notStarted: [], worktrees: [], testsPassed: true, testCommand: '', testOutput: '' }

for (let w = 0; w < waves.length; w++) {
  const ids = waves[w]
  const built = await pool(ids, id => agent(
    `${ctx}\nYour task (id ${id}):\n${byId[id].text}`,
    { label: `task:${id}`, phase: 'Build', agentType: 'conductor:task-worker', isolation: 'worktree', schema: BUILT }))

  const ok = []
  built.forEach((b, i) => {
    if (b && b.worktreePath) out.worktrees.push({ path: b.worktreePath, branch: b.branch })
    if (b && b.status === 'done' && b.sha) ok.push({ taskId: ids[i], task: byId[ids[i]].text, sha: b.sha, summary: b.summary, files: b.files })
    else out.failed.push({ taskId: ids[i], reason: b ? (b.reason || 'no commit') : 'worker died', question: (b && b.question) || '' })
  })
  if (!ok.length) continue

  const res = await agent(
    `${ctx}\nIntegrate these built tasks into the main checkout, in this order:\n${JSON.stringify(ok, null, 1)}`,
    { label: `integrate:wave-${w + 1}`, phase: 'Integrate', agentType: 'conductor:integrator', schema: MERGED })
  const rest = waves.slice(w + 1).flat()
  if (!res) {
    out.failed.push(...ok.map(b => ({ taskId: b.taskId, reason: 'integrator died; check git log and plan.md' })))
    out.notStarted = rest
    break
  }
  out.merged.push(...res.merged)
  out.failed.push(...res.failed)
  out.testCommand = res.testCommand
  if (!res.testsPassed) {
    out.testsPassed = false
    out.testOutput = res.testOutput
    out.notStarted = rest
    if (rest.length) log(`tests failed after wave ${w + 1} -- stopping, ${rest.length} task(s) not started`)
    break
  }
}

return out
