#!/bin/sh
set -e
REPO="cloudaura-io/cloudaura-marketplace"
DEST="${INSTALL_DIR:-/usr/local/bin}"
ARCH=$(uname -m)
case "$ARCH" in x86_64) ARCH=x64;; aarch64|arm64) ARCH=arm64;; *) echo "Unsupported: $ARCH" && exit 1;; esac
TAG=$(curl -sI "https://github.com/$REPO/releases/latest" | grep -i location | sed 's/.*tag\///' | tr -d '\r\n')
URL="https://github.com/$REPO/releases/download/$TAG/conductor-tui-linux-$ARCH"
echo "Installing conductor-tui ($TAG, linux-$ARCH) to $DEST..."
curl -fSL "$URL" -o "$DEST/conductor-tui" && chmod +x "$DEST/conductor-tui"
echo "Done. Run: conductor-tui"
