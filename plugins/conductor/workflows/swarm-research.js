export const meta = {
  name: 'swarm-research',
  description: 'Conductor swarm: research a track from the codebase and the web, verify every answer, write research.md (default 15 at once)',
  whenToUse: 'Run only by /conductor:research in Swarm mode, which passes the track files and the research questions in args. If invoked with no args (a user typed the bare slash command), do not call Workflow: tell the user to run /conductor:research and choose Swarm.',
  phases: [
    { title: 'Research', detail: 'one researcher per question (code + web), then a skeptic re-checks its evidence and sources' },
    { title: 'Write', detail: 'one writer merges the verified answers into research.md' },
  ],
}

let A = args
if (typeof A === 'string') { try { A = JSON.parse(A) } catch (e) { A = null } }
if (!A || !Array.isArray(A.questions) || !A.questions.length || !A.researchPath) {
  log('started without research questions -- nothing dispatched')
  return { started: false, next: 'Tell the user to run /conductor:research and choose Swarm; this workflow needs the questions that skill supplies.' }
}

// Swarm size from args.maxParallel (default 15). No upper cap here: the Claude Code runtime
// enforces its own concurrency limit (CLAUDE_CODE_WORKFLOW_MAX_CONCURRENT_AGENTS).
const MAX = Math.max(1, Math.floor(Number(A.maxParallel)) || 15)
log(`swarm size: ${MAX} agent(s) at once, ${A.questions.length} question(s)`)

// Run fn over items with at most MAX chains in flight.
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

const MULTI = Array.isArray(A.repos) && A.repos.length > 0
const ctx = `Track ${A.trackId}. Spec: ${A.specPath}. Plan: ${A.planPath}. Tech stack: ${A.techStackPath}. Paths are relative to ${MULTI ? 'the project root (your working directory), which is not a git repository; repositories: ' + A.repos.map(r => `${r.name} at ${r.path}`).join(', ') + '. Follow conductor/repos.md' : 'the repository root, your working directory'}.${A.focus ? ` The user asked to focus on: ${A.focus}.` : ''}`

const SOURCE = {
  type: 'object',
  properties: { title: { type: 'string' }, url: { type: 'string' }, version: { type: 'string' }, quote: { type: 'string' } },
  required: ['title', 'url', 'version', 'quote'],
}
const OPEN = {
  type: 'object',
  properties: {
    question: { type: 'string' },
    options: { type: 'array', items: { type: 'string' }, description: '2 to 4 concrete answers' },
    why: { type: 'string' },
  },
  required: ['question', 'options', 'why'],
}
const ANSWER = {
  type: 'object',
  properties: {
    answer: { type: 'string' },
    evidence: { type: 'array', items: { type: 'string' }, description: 'file:line references' },
    sources: { type: 'array', items: SOURCE },
    confidence: { type: 'string', enum: ['high', 'medium', 'low'] },
    planImpact: { type: 'string', description: 'what this changes in the plan, or "none"' },
    openQuestions: { type: 'array', items: OPEN },
  },
  required: ['answer', 'evidence', 'sources', 'confidence', 'planImpact', 'openQuestions'],
}
const VERDICT = {
  type: 'object',
  properties: {
    supported: { type: 'boolean' },
    corrections: { type: 'string', description: 'what does not hold, or empty' },
    confidence: { type: 'string', enum: ['high', 'medium', 'low'] },
  },
  required: ['supported', 'corrections', 'confidence'],
}

phase('Research')
// Each question runs researcher -> skeptic as one chain; chains do not wait for each other.
const results = await pool(A.questions, async q => {
  const where = `${q.ref ? ` It serves: ${q.ref}.` : ''}${q.repos && q.repos.length ? ` Repositories: ${q.repos.join(', ')}.` : ''}`
  const found = await agent(
    `${ctx}\nAct as a researcher. Question ${q.id} (${q.kind || 'code + web'}): ${q.text}${where}`,
    { label: `research:${q.id}`, phase: 'Research', agentType: 'conductor:researcher', schema: ANSWER })
  if (!found) return null
  const verdict = await agent(
    `${ctx}\nAct as a skeptic. Try to refute this answer to question ${q.id} ("${q.text}"):\n${JSON.stringify(found, null, 1)}`,
    { label: `verify:${q.id}`, phase: 'Research', agentType: 'conductor:researcher', schema: VERDICT })
  return { id: q.id, question: q.text, ref: q.ref || '', ...found, verdict: verdict || { supported: false, corrections: 'skeptic returned nothing; unverified', confidence: 'low' } }
})

const answered = results.filter(Boolean)
const lost = A.questions.filter((q, i) => !results[i]).map(q => q.id)
if (lost.length) log(`no answer for ${lost.join(', ')} -- listed as unanswered in research.md`)
const refuted = answered.filter(r => !r.verdict.supported).length
log(`${answered.length} answered, ${refuted} not supported by their skeptic`)

phase('Write')
const WRITTEN = {
  type: 'object',
  properties: {
    path: { type: 'string' },
    openQuestions: { type: 'array', items: { type: 'object', properties: { id: { type: 'string' }, ...OPEN.properties }, required: ['id', 'question', 'options', 'why'] } },
    planChanges: {
      type: 'array',
      items: {
        type: 'object',
        properties: { id: { type: 'string' }, phase: { type: 'string' }, change: { type: 'string' }, reason: { type: 'string' } },
        required: ['id', 'phase', 'change', 'reason'],
      },
    },
  },
  required: ['path', 'openQuestions', 'planChanges'],
}
const written = await agent(
  `${ctx}\nWrite the research file ${A.researchPath} for this track, following the "research.md format" section of ${A.skillPath}. Date: ${A.date}. Mode: swarm (${MAX}).
If the file already exists, keep its sections for questions not in this run and its "Decisions" section untouched; replace the sections of re-run questions.
Use the skeptic's verdict: when supported is false, apply its corrections, lower the confidence to the skeptic's, and mark the answer "(unverified)".
Merge duplicate open questions and plan changes across answers; give them ids o1, o2, ... and c1, c2, ... continuing after the ones already in the file.
Do not edit any other file. Questions without an answer: ${lost.length ? lost.join(', ') : 'none'}.
Verified answers:\n${JSON.stringify(answered, null, 1)}`,
  { label: 'write:research.md', phase: 'Write', schema: WRITTEN })

return { started: true, answered: answered.length, unanswered: lost, refuted, written }
