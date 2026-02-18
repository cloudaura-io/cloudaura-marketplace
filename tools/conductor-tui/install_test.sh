#!/bin/sh
# Tests for install.sh
# Validates script content and behavior without network access.
set -e

PASS=0
FAIL=0
SCRIPT="$(dirname "$0")/install.sh"

pass() { PASS=$((PASS + 1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  FAIL: $1"; }

echo "=== install.sh tests ==="

# T1: Default INSTALL_DIR should be ~/.local/bin, not /usr/local/bin
if grep -q 'INSTALL_DIR:-.*\.local/bin' "$SCRIPT"; then
  pass "Default INSTALL_DIR is ~/.local/bin"
else
  fail "Default INSTALL_DIR should be ~/.local/bin"
fi

# T2: Must NOT reference /usr/local/bin anywhere
if grep -q '/usr/local/bin' "$SCRIPT"; then
  fail "Script should not reference /usr/local/bin"
else
  pass "No /usr/local/bin reference"
fi

# T3: Must create install dir with mkdir -p
if grep -q 'mkdir -p' "$SCRIPT"; then
  pass "Creates install dir with mkdir -p"
else
  fail "Missing mkdir -p for install dir"
fi

# T4: Must detect shell from SHELL variable
if grep -q '$SHELL\|"$SHELL"' "$SCRIPT"; then
  pass "Detects shell from SHELL variable"
else
  fail "Missing shell detection from SHELL variable"
fi

# T5: Must handle bash profile
if grep -q '\.bashrc' "$SCRIPT"; then
  pass "Handles .bashrc for bash"
else
  fail "Missing .bashrc handling for bash"
fi

# T6: Must handle zsh profile
if grep -q '\.zshrc' "$SCRIPT"; then
  pass "Handles .zshrc for zsh"
else
  fail "Missing .zshrc handling for zsh"
fi

# T7: Must handle fish config
if grep -q 'config\.fish' "$SCRIPT"; then
  pass "Handles config.fish for fish"
else
  fail "Missing config.fish handling for fish"
fi

# T8: Must add PATH export line to shell profile
if grep -q 'export PATH.*\.local/bin' "$SCRIPT" || grep -q 'fish_add_path\|set -gx PATH' "$SCRIPT"; then
  pass "Adds PATH export to shell profile"
else
  fail "Missing PATH export logic"
fi

# T9: Must check if PATH line already exists before adding
if grep -q 'grep.*\.local/bin\|already' "$SCRIPT"; then
  pass "Checks if PATH line already present"
else
  fail "Missing check for existing PATH line"
fi

# T10: Must print message about PATH update / terminal restart
if grep -q 'restart\|reload\|source\|new terminal\|new session' "$SCRIPT"; then
  pass "Prints terminal restart message"
else
  fail "Missing terminal restart message"
fi

# T11: INSTALL_DIR env var override must still work
if grep -q 'INSTALL_DIR' "$SCRIPT"; then
  pass "INSTALL_DIR override supported"
else
  fail "Missing INSTALL_DIR override support"
fi

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ] || exit 1
