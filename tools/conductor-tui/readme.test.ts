import { describe, expect, it } from "bun:test";
import { readFileSync } from "fs";
import { join } from "path";

const readme = readFileSync(join(import.meta.dir, "README.md"), "utf-8");

describe("README macOS documentation", () => {
  it("mentions macOS as a supported platform", () => {
    expect(readme.toLowerCase()).toContain("macos");
  });

  it("includes macOS install command", () => {
    expect(readme).toContain("install.sh");
  });
});
