#!/bin/sh
# Tests for install.ps1 - verifies script correctness without running PowerShell.
set -e

PASS=0
FAIL=0
SCRIPT="$(dirname "$0")/install.ps1"

pass() { PASS=$((PASS + 1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  FAIL: $1"; }

echo "=== install.ps1 tests ==="

# T1: Default destination uses LOCALAPPDATA (no admin needed)
if grep -q 'LOCALAPPDATA' "$SCRIPT"; then
  pass "Default destination uses LOCALAPPDATA"
else
  fail "Default destination should use LOCALAPPDATA"
fi

# T2: PATH addition uses User scope (not Machine)
if grep -q '"User"' "$SCRIPT"; then
  pass "PATH modification uses User scope"
else
  fail "PATH modification should use User scope"
fi

# T3: Checks for existing PATH entry before adding
if grep -q 'notlike.*\$dest' "$SCRIPT"; then
  pass "Checks for existing PATH entry"
else
  fail "Should check for existing PATH entry"
fi

# T4: INSTALL_DIR env var override is supported
if grep -q 'INSTALL_DIR' "$SCRIPT"; then
  pass "INSTALL_DIR override supported"
else
  fail "Missing INSTALL_DIR override support"
fi

# T5: Creates destination directory
if grep -q 'New-Item.*Directory.*Force' "$SCRIPT"; then
  pass "Creates destination directory"
else
  fail "Should create destination directory"
fi

# T6: Does not reference admin/elevated commands
if grep -qi 'RunAs\|Administrator\|Machine' "$SCRIPT"; then
  fail "Script should not require admin privileges"
else
  pass "No admin privilege references"
fi

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ] || exit 1
