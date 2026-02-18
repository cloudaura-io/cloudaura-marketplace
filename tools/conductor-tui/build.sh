#!/bin/bash
set -e

OUTDIR="dist"
NAME="conductor-tui"
MODULE_DIR="$(cd "$(dirname "$0")" && pwd)"

rm -rf "$OUTDIR"
mkdir -p "$OUTDIR"

targets=(
  "linux:amd64:conductor-tui-linux-x64"
  "linux:arm64:conductor-tui-linux-arm64"
  "darwin:amd64:conductor-tui-darwin-x64"
  "darwin:arm64:conductor-tui-darwin-arm64"
  "windows:amd64:conductor-tui-windows-x64.exe"
)

for entry in "${targets[@]}"; do
  IFS=: read -r goos goarch outfile <<< "$entry"
  echo "Building ${goos}/${goarch}..."
  GOOS="$goos" GOARCH="$goarch" go build -ldflags="-s -w" -o "$OUTDIR/$outfile" "$MODULE_DIR"
done

echo ""
ls -lh "$OUTDIR"/
echo ""
echo "Done."
