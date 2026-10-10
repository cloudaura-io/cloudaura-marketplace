export const meta = {
  name: 'swarm-review',
  description: 'Conductor swarm: review a track diff by lens x file chunk, then adversarially verify findings (default 15 agents at once)',
  whenToUse: 'Run only by /conductor:review in Swarm mode, which passes the revision range and changed files in args. If invoked with no args (a user typed the bare slash command), do not call Workflow: tell the user to run /conductor:review and choose Swarm.',
  phases: [
    { title: 'Find', detail: 'one reviewer per lens x file chunk, plus one test run' },
    { title: 'Verify', detail: 'skeptics try to refute each finding (3 for Critical/High, 1 otherwise)' },
  ],
}

const CHUNK = 300 // changed lines per finder, same threshold as /conductor:review

let A = args
if (typeof A === 'string') { try { A = JSON.parse(A) } catch (e) { A = null } }
if (!A || !A.range || !Array.isArray(A.files) || !A.files.length) {
  log('started without a diff to review -- nothing dispatched')
  return { started: false, next: 'Tell the user to run /conductor:review and choose Swarm; this workflow needs the range and files that skill supplies.' }
}

// Swarm size from args.maxParallel (default 15). No upper cap here: the Claude Code runtime
// enforces its own concurrency limit (CLAUDE_CODE_WORKFLOW_MAX_CONCURRENT_AGENTS).
const MAX = Math.max(1, Math.floor(Number(A.maxParallel)) || 15)
log(`swarm size: ${MAX} agent(s) at once`)

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

const ctx = `Revision range: ${A.range}. Track plan: ${A.planPath}. Track spec: ${A.specPath}. Paths are relative to the repository root, your working directory.`

// Group files into chunks of about CHUNK changed lines; a big file gets its own chunk.
const chunks = []
let cur = [], n = 0
for (const f of A.files) {
  if (cur.length && n + f.lines > CHUNK) { chunks.push(cur); cur = []; n = 0 }
  cur.push(f.path); n += f.lines
}
if (cur.length) chunks.push(cur)

const units = [{ lens: 'tests', files: null }]
for (const c of chunks) for (const lens of ['intent', 'style', 'correctness']) units.push({ lens, files: c })
log(`${A.files.length} file(s) in ${chunks.length} chunk(s) -> ${units.length} reviewers`)

const FOUND = {
  type: 'object',
  properties: {
    findings: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          severity: { type: 'string', enum: ['Critical', 'High', 'Medium', 'Low'] },
          title: { type: 'string' },
          file: { type: 'string' },
          startLine: { type: 'integer' },
          endLine: { type: 'integer' },
          context: { type: 'string' },
          suggestion: { type: 'string', description: 'diff, or empty' },
        },
        required: ['severity', 'title', 'file', 'startLine', 'endLine', 'context', 'suggestion'],
      },
    },
    tests: {
      type: 'object',
      description: 'only for the tests lens',
      properties: {
        command: { type: 'string' },
        passed: { type: 'boolean' },
        summary: { type: 'string' },
        newTests: { type: 'boolean' },
        coverage: { type: 'string', enum: ['Yes', 'No', 'Partial'] },
      },
      required: ['command', 'passed', 'summary', 'newTests', 'coverage'],
    },
  },
  required: ['findings'],
}

phase('Find')
const found = await pool(units, u => agent(
  `${ctx}\nLens: ${u.lens}.\n${u.files ? `Files: ${u.files.join(', ')}` : 'Files: the whole range. Run the test suite.'}`,
  { label: `find:${u.lens}${u.files ? ':' + u.files[0] : ''}`, phase: 'Find', agentType: 'conductor:reviewer', schema: FOUND }))
const lost = found.filter(r => !r).length
if (lost) log(`${lost} reviewer(s) returned nothing -- their slices are unreviewed`)

const tests = (found[0] && found[0].tests) || null
// Dedup across lenses: same file and start line = same finding, keep the highest severity.
const RANK = { Critical: 0, High: 1, Medium: 2, Low: 3 }
const byKey = new Map()
found.filter(Boolean).flatMap(r => r.findings).forEach(f => {
  const k = `${f.file}:${f.startLine}`
  if (!byKey.has(k) || RANK[f.severity] < RANK[byKey.get(k).severity]) byKey.set(k, f)
})
const findings = [...byKey.values()]

phase('Verify')
const VERDICT = {
  type: 'object',
  properties: { refuted: { type: 'boolean' }, why: { type: 'string' } },
  required: ['refuted', 'why'],
}
const votes = findings.flatMap((f, i) => Array.from({ length: RANK[f.severity] <= 1 ? 3 : 1 }, (_, v) => ({ i, v })))
const verdicts = await pool(votes, ({ i, v }) => agent(
  `${ctx}\nAct as a skeptic (vote ${v + 1}). Try to refute this finding:\n${JSON.stringify(findings[i], null, 1)}`,
  { label: `verify:${findings[i].file}:${findings[i].startLine}#${v + 1}`, phase: 'Verify', agentType: 'conductor:reviewer', schema: VERDICT }))

// A finding survives when a strict majority of its returned votes did not refute it.
const confirmed = findings.filter((f, i) => {
  const mine = votes.map((x, k) => (x.i === i ? verdicts[k] : null)).filter(Boolean)
  return mine.length > 0 && mine.filter(x => !x.refuted).length * 2 > mine.length
}).sort((a, b) => RANK[a.severity] - RANK[b.severity])

return { started: true, findings: confirmed, refutedCount: findings.length - confirmed.length, unreviewedSlices: lost, tests }
