import { describe, expect, it } from "bun:test";
import { readFileSync } from "fs";
import { join } from "path";

const installSh = readFileSync(join(import.meta.dir, "install.sh"), "utf-8");

describe("install.sh macOS support", () => {
  it("detects Darwin OS via uname -s", () => {
    expect(installSh).toContain("uname -s");
  });

  it("sets platform to darwin for macOS", () => {
    expect(installSh).toContain("darwin");
  });

  it("builds URL using platform variable (not hardcoded linux)", () => {
    // URL should use $platform variable, not hardcode 'linux'
    expect(installSh).not.toContain('"https://github.com/$REPO/releases/download/$TAG/conductor-tui-linux-$ARCH"');
  });

  it("exits with unsupported OS message for unknown platforms", () => {
    expect(installSh).toContain("Unsupported");
  });
});
