import { describe, expect, it } from "bun:test";
import { readFileSync } from "fs";
import { join } from "path";
import { parse } from "yaml";

const ROOT = join(import.meta.dir, "../..");

describe("build.sh targets", () => {
  const buildSh = readFileSync(join(import.meta.dir, "build.sh"), "utf-8");

  it("includes bun-darwin-x64 target", () => {
    expect(buildSh).toContain('"bun-darwin-x64"');
  });

  it("includes bun-darwin-arm64 target", () => {
    expect(buildSh).toContain('"bun-darwin-arm64"');
  });
});

describe("release-conductor-tui.yml matrix", () => {
  const workflowPath = join(ROOT, ".github/workflows/release-conductor-tui.yml");
  const workflow = parse(readFileSync(workflowPath, "utf-8"));
  const matrix = workflow.jobs.build.strategy.matrix.include;

  it("includes darwin-x64 entry", () => {
    const entry = matrix.find(
      (e: { target: string }) => e.target === "bun-darwin-x64"
    );
    expect(entry).toBeDefined();
    expect(entry.artifact).toBe("conductor-tui-darwin-x64");
  });

  it("includes darwin-arm64 entry", () => {
    const entry = matrix.find(
      (e: { target: string }) => e.target === "bun-darwin-arm64"
    );
    expect(entry).toBeDefined();
    expect(entry.artifact).toBe("conductor-tui-darwin-arm64");
  });
});
