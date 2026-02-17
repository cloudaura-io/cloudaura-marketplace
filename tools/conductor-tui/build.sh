#!/bin/bash
set -e

OUTDIR="dist"
SRC="conductor-tui/tui.tsx"
NAME="conductor-tui"

rm -rf "$OUTDIR"
mkdir -p "$OUTDIR"

targets=(
  "bun-linux-x64"
  "bun-linux-arm64"
  "bun-darwin-x64"
  "bun-darwin-arm64"
  "bun-windows-x64"
)

for target in "${targets[@]}"; do
  echo "Building $target..."
  ext=""
  [[ "$target" == *windows* ]] && ext=".exe"
  outfile="$OUTDIR/${NAME}-${target#bun-}${ext}"
  bun build --compile --target="$target" "$SRC" --outfile "$outfile"
done

echo ""
ls -lh "$OUTDIR"/
echo ""
echo "Done."
