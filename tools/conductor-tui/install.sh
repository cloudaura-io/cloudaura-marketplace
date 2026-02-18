#!/bin/sh
set -e
REPO="cloudaura-io/conductor-claude-code"
DEST="${INSTALL_DIR:-/usr/local/bin}"
OS=$(uname -s)
ARCH=$(uname -m)
case "$OS" in Linux) PLATFORM=linux;; Darwin) PLATFORM=darwin;; *) echo "Unsupported OS: $OS" && exit 1;; esac
case "$ARCH" in x86_64) ARCH=x64;; aarch64|arm64) ARCH=arm64;; *) echo "Unsupported: $ARCH" && exit 1;; esac
TAG=$(curl -sI "https://github.com/$REPO/releases/latest" | grep -i location | sed 's/.*tag\///' | tr -d '\r\n')
URL="https://github.com/$REPO/releases/download/$TAG/conductor-tui-$PLATFORM-$ARCH"
echo "Installing conductor-tui ($TAG, $PLATFORM-$ARCH) to $DEST..."
curl -fSL "$URL" -o "$DEST/conductor-tui" && chmod +x "$DEST/conductor-tui"
echo "Done. Run: conductor-tui"
