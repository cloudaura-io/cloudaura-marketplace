#!/usr/bin/env bun

import { useState, useEffect } from "react";
import { render, Box, Text, useInput, useApp } from "ink";
import { readdir, readFile } from "fs/promises";
import { join } from "path";
import pkg from "./package.json";

// === Types ===

interface Track {
  track_id: string;
  type: string;
  status: string;
  description: string;
  source: "active" | "archived";
  phases: Phase[];
}

interface Phase {
  number: number;
  name: string;
  checkpoint: string | null;
  tasks: Task[];
}

interface Task {
  name: string;
  completed: boolean;
  commit: string | null;
  subtasks: { name: string; completed: boolean }[];
}

type Screen =
  | { type: "tracks"; cursor: number }
  | { type: "phases"; trackIdx: number; cursor: number }
  | { type: "tasks"; trackIdx: number; phaseIdx: number; cursor: number }
  | { type: "detail"; trackIdx: number; phaseIdx: number; taskIdx: number; scroll: number }
  | { type: "quit" };

type StatusColor = "green" | "yellow" | "cyan" | "magenta" | "blue" | "red" | "gray" | undefined;

function statusColor(s: string): StatusColor {
  if (s === "completed" || s === "done") return "green";
  if (s === "in_progress" || s === "doing") return "yellow";
  if (s === "pending" || s === "todo") return "cyan";
  if (s === "new") return "magenta";
  if (s === "review") return "blue";
  if (s === "blocked") return "red";
  if (s === "archived") return "gray";
  return undefined;
}

function phaseStatus(p: Phase): string {
  if (p.tasks.length === 0) return "empty";
  const done = p.tasks.filter((t) => t.completed).length;
  if (done === p.tasks.length) return "completed";
  return done > 0 ? "in_progress" : "pending";
}

function trunc(s: string, n: number): string {
  return s.length <= n ? s : s.slice(0, n - 3) + "...";
}

function pad(s: string, n: number): string {
  return s.length >= n ? s.slice(0, n) : s + " ".repeat(n - s.length);
}

// === Data Loading ===

async function discoverTracks(basePath: string): Promise<Track[]> {
  const tracks: Track[] = [];
  const dirs = [
    { path: join(basePath, "conductor", "tracks"), source: "active" as const },
    { path: join(basePath, "conductor", "archive"), source: "archived" as const },
  ];

  for (const { path: dirPath, source } of dirs) {
    let entries: string[];
    try { entries = await readdir(dirPath); } catch { continue; }
    for (const entry of entries) {
      try {
        const meta = JSON.parse(await readFile(join(dirPath, entry, "metadata.json"), "utf-8"));
        let phases: Phase[] = [];
        try { phases = parsePlan(await readFile(join(dirPath, entry, "plan.md"), "utf-8")); } catch {}
        tracks.push({
          track_id: meta.track_id || entry,
          type: meta.type || "unknown",
          status: meta.status || "unknown",
          description: meta.description || "",
          source,
          phases,
        });
      } catch {}
    }
  }

  return tracks.sort((a, b) =>
    a.source !== b.source ? (a.source === "active" ? -1 : 1) : a.track_id.localeCompare(b.track_id)
  );
}

function parsePlan(content: string): Phase[] {
  const phases: Phase[] = [];
  let phase: Phase | null = null;
  let task: Task | null = null;

  for (const line of content.split("\n")) {
    const pm = line.match(/^## Phase (\d+): (.+?)(?:\s*\[checkpoint:\s*([a-f0-9]+)\])?\s*$/);
    if (pm) {
      if (phase) phases.push(phase);
      phase = { number: parseInt(pm[1]!), name: pm[2]!.trim(), checkpoint: pm[3] ?? null, tasks: [] };
      task = null;
      continue;
    }
    const tm = line.match(/^- \[([ x])\] Task: (.+?)(?:\s+`([a-f0-9]{7,})`)?$/);
    if (tm && phase) {
      task = { name: tm[2]!.trim(), completed: tm[1] === "x", commit: tm[3] ?? null, subtasks: [] };
      phase.tasks.push(task);
      continue;
    }
    const sm = line.match(/^    - \[([ x])\] (.+)$/);
    if (sm && task) {
      task.subtasks.push({ name: sm[2]!.trim(), completed: sm[1] === "x" });
    }
  }
  if (phase) phases.push(phase);
  return phases;
}

// === Components ===

function Header({ breadcrumbs, hint }: { breadcrumbs: string[]; hint: string }) {
  const w = process.stdout.columns || 80;
  return (
    <Box flexDirection="column" paddingLeft={1}>
      <Box justifyContent="space-between" width={w - 2}>
        <Text>
          <Text bold>Conductor TUI</Text><Text dimColor> v{pkg.version}</Text>
          {breadcrumbs.map((b, i) => (
            <Text key={i}> <Text dimColor>&gt;</Text> {b}</Text>
          ))}
        </Text>
        <Text dimColor>{hint}</Text>
      </Box>
      <Text dimColor>{"─".repeat(w - 4)}</Text>
    </Box>
  );
}

function Footer({ text }: { text: string }) {
  return <Box paddingLeft={1}><Text dimColor>{text}</Text></Box>;
}

// === App ===

function App({ basePath }: { basePath: string }) {
  const [tracks, setTracks] = useState<Track[]>([]);

  useEffect(() => {
    const load = () => { discoverTracks(basePath).then(setTracks); };
    load();
    const id = setInterval(load, 2000);
    return () => clearInterval(id);
  }, [basePath]);
  const { exit } = useApp();
  const [stack, setStack] = useState<Screen[]>([{ type: "tracks", cursor: 0 }]);
  const screen = stack[stack.length - 1]!;

  const push = (s: Screen) => setStack((prev) => [...prev, s]);
  const pop = () =>
    setStack((prev) => (prev.length > 1 ? prev.slice(0, -1) : [...prev, { type: "quit" }]));

  const moveCursor = (delta: number, max: number) => {
    setStack((prev) => {
      const cur = prev[prev.length - 1]!;
      if (!("cursor" in cur)) return prev;
      const next = Math.max(0, Math.min(max - 1, cur.cursor + delta));
      if (next === cur.cursor) return prev;
      return [...prev.slice(0, -1), { ...cur, cursor: next }];
    });
  };

  const moveScroll = (delta: number) => {
    setStack((prev) => {
      const cur = prev[prev.length - 1]!;
      if (cur.type !== "detail") return prev;
      return [...prev.slice(0, -1), { ...cur, scroll: Math.max(0, cur.scroll + delta) }];
    });
  };

  function itemCount(): number {
    switch (screen.type) {
      case "tracks": return tracks.length;
      case "phases": return tracks[screen.trackIdx]?.phases.length ?? 0;
      case "tasks": return tracks[screen.trackIdx]?.phases[screen.phaseIdx]?.tasks.length ?? 0;
      default: return 0;
    }
  }

  function handleEnter() {
    if (screen.type === "tracks" && tracks.length > 0) {
      push({ type: "phases", trackIdx: screen.cursor, cursor: 0 });
    } else if (screen.type === "phases") {
      const t = tracks[screen.trackIdx];
      if (t && t.phases.length > 0)
        push({ type: "tasks", trackIdx: screen.trackIdx, phaseIdx: screen.cursor, cursor: 0 });
    } else if (screen.type === "tasks") {
      const p = tracks[screen.trackIdx]?.phases[screen.phaseIdx];
      if (p && p.tasks.length > 0)
        push({ type: "detail", trackIdx: screen.trackIdx, phaseIdx: screen.phaseIdx, taskIdx: screen.cursor, scroll: 0 });
    }
  }

  useInput((input, key) => {
    if (screen.type === "quit") {
      if (input === "y") exit();
      if (input === "n" || key.escape) setStack((prev) => prev.slice(0, -1));
      return;
    }
    if (key.upArrow) {
      screen.type === "detail" ? moveScroll(-1) : moveCursor(-1, itemCount());
    }
    if (key.downArrow) {
      screen.type === "detail" ? moveScroll(1) : moveCursor(1, itemCount());
    }
    if (key.return) handleEnter();
    if (key.escape) pop();
    if (input === "q" && screen.type === "tracks") push({ type: "quit" });
  });

  // --- Screens ---

  const maxVis = (process.stdout.rows || 24) - 6;

  if (screen.type === "quit") {
    return (
      <Box justifyContent="center" alignItems="center" height={process.stdout.rows}>
        <Text><Text bold>Quit Conductor TUI? </Text><Text dimColor>[y/n]</Text></Text>
      </Box>
    );
  }

  if (screen.type === "tracks") {
    if (tracks.length === 0) {
      return (
        <Box flexDirection="column">
          <Header breadcrumbs={[]} hint="[q] Quit" />
          <Box paddingLeft={1}><Text dimColor>No tracks found.</Text></Box>
          <Footer text="[q] Quit" />
        </Box>
      );
    }
    const scroll = Math.max(0, screen.cursor - maxVis + 1);
    const visible = tracks.slice(scroll, scroll + maxVis);
    const w = process.stdout.columns || 80;
    const descW = Math.max(8, w - 64);
    return (
      <Box flexDirection="column">
        <Header breadcrumbs={[]} hint="[q] Quit" />
        <Text dimColor>  {pad("Track ID", 28)}{pad("Type", 10)}{pad("Status", 14)}{pad("Phases", 8)}Description</Text>
        {visible.map((t, i) => {
          const idx = scroll + i;
          const sel = idx === screen.cursor;
          const tag = t.source === "archived" ? " *" : "";
          return (
            <Text key={t.track_id} bold={sel}>
              <Text color={sel ? "blue" : undefined}>{sel ? "> " : "  "}</Text>
              {pad(trunc(t.track_id, 26), 28)}
              {pad(t.type, 10)}
              <Text color={statusColor(t.status)}>{pad(t.status + tag, 14)}</Text>
              {pad(String(t.phases.length), 8)}
              {trunc(t.description, descW)}
            </Text>
          );
        })}
        <Footer text="[↑↓] Navigate  [Enter] View phases  [q] Quit" />
      </Box>
    );
  }

  if (screen.type === "phases") {
    const track = tracks[screen.trackIdx]!;
    const scroll = Math.max(0, screen.cursor - (maxVis - 1) + 1);
    const visible = track.phases.slice(scroll, scroll + maxVis - 1);
    return (
      <Box flexDirection="column">
        <Header breadcrumbs={[track.track_id]} hint="[Esc] Back" />
        <Box paddingLeft={1}><Text dimColor>{track.description}</Text></Box>
        <Text dimColor>  {pad("#", 4)}{pad("Phase", 34)}{pad("Tasks", 10)}Status</Text>
        {visible.map((p, i) => {
          const idx = scroll + i;
          const sel = idx === screen.cursor;
          const done = p.tasks.filter((t) => t.completed).length;
          const st = phaseStatus(p);
          return (
            <Text key={p.number} bold={sel}>
              <Text color={sel ? "blue" : undefined}>{sel ? "> " : "  "}</Text>
              {pad(String(p.number), 4)}
              {pad(trunc(p.name, 32), 34)}
              {pad(`${done}/${p.tasks.length}`, 10)}
              <Text color={statusColor(st)}>{st}</Text>
            </Text>
          );
        })}
        <Footer text="[↑↓] Navigate  [Enter] View tasks  [Esc] Back" />
      </Box>
    );
  }

  if (screen.type === "tasks") {
    const track = tracks[screen.trackIdx]!;
    const phase = track.phases[screen.phaseIdx]!;
    const scroll = Math.max(0, screen.cursor - (maxVis - 1) + 1);
    const visible = phase.tasks.slice(scroll, scroll + maxVis - 1);
    return (
      <Box flexDirection="column">
        <Header
          breadcrumbs={[trunc(track.track_id, 20), `Phase ${phase.number}`]}
          hint="[Esc] Back"
        />
        <Box paddingLeft={1}><Text dimColor>{phase.name}</Text></Box>
        <Text dimColor>  {pad("#", 4)}{pad("Task", 42)}{pad("Status", 10)}Commit</Text>
        {visible.map((t, i) => {
          const idx = scroll + i;
          const sel = idx === screen.cursor;
          const st = t.completed ? "done" : "pending";
          return (
            <Text key={idx} bold={sel}>
              <Text color={sel ? "blue" : undefined}>{sel ? "> " : "  "}</Text>
              {pad(String(idx + 1), 4)}
              {pad(trunc(t.name, 40), 42)}
              <Text color={statusColor(st)}>{pad(st, 10)}</Text>
              {t.commit ?? "—"}
            </Text>
          );
        })}
        <Footer text="[↑↓] Navigate  [Enter] View detail  [Esc] Back" />
      </Box>
    );
  }

  if (screen.type === "detail") {
    const track = tracks[screen.trackIdx]!;
    const phase = track.phases[screen.phaseIdx]!;
    const task = phase.tasks[screen.taskIdx]!;
    const st = task.completed ? "completed" : "pending";
    const maxSub = (process.stdout.rows || 24) - 10;
    const scrollIdx = Math.min(screen.scroll, Math.max(0, task.subtasks.length - maxSub));
    const visibleSubs = task.subtasks.slice(scrollIdx, scrollIdx + maxSub);

    return (
      <Box flexDirection="column">
        <Header
          breadcrumbs={[trunc(track.track_id, 16), `Phase ${phase.number}`, `Task: ${trunc(task.name, 30)}`]}
          hint="[Esc] Back"
        />
        <Box paddingLeft={1} flexDirection="column" gap={1}>
          <Text><Text bold>Task: </Text>{task.name}</Text>
          <Text>
            <Text>Status: </Text><Text color={statusColor(st)}>{st}</Text>
            {task.commit && <Text>          Commit: <Text bold>{task.commit}</Text></Text>}
          </Text>
        </Box>
        <Box paddingLeft={1} marginTop={1} flexDirection="column">
          {task.subtasks.length > 0 ? (
            <>
              <Text bold>Sub-tasks:</Text>
              {visibleSubs.map((s, i) => (
                <Text key={scrollIdx + i}>
                  {"    "}{s.completed ? <Text color="green">[x]</Text> : <Text>[ ]</Text>}
                  {" "}{s.name}
                </Text>
              ))}
            </>
          ) : (
            <Text dimColor>No sub-tasks.</Text>
          )}
        </Box>
        <Footer
          text={task.subtasks.length > maxSub ? "[↑↓] Scroll  [Esc] Back" : "[Esc] Back"}
        />
      </Box>
    );
  }

  return null;
}

// === Main ===

if (process.argv.includes("--version") || process.argv.includes("-v")) {
  console.log(`conductor-tui v${pkg.version}`);
  process.exit(0);
}

render(<App basePath={process.cwd()} />);
