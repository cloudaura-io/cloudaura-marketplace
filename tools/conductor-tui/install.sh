#!/bin/sh
set -e
REPO="cloudaura-io/cloudaura-marketplace"
DEST="${INSTALL_DIR:-$HOME/.local/bin}"
OS=$(uname -s)
ARCH=$(uname -m)
case "$OS" in Linux) PLATFORM=linux;; Darwin) PLATFORM=darwin;; *) echo "Unsupported OS: $OS" && exit 1;; esac
case "$ARCH" in x86_64) ARCH=x64;; aarch64|arm64) ARCH=arm64;; *) echo "Unsupported: $ARCH" && exit 1;; esac
TAG=$(curl -sI "https://github.com/$REPO/releases/latest" | grep -i location | sed 's/.*tag\///' | tr -d '\r\n')
URL="https://github.com/$REPO/releases/download/$TAG/conductor-tui-$PLATFORM-$ARCH"
mkdir -p "$DEST"
echo "Installing conductor-tui ($TAG, $PLATFORM-$ARCH) to $DEST..."
curl -fSL "$URL" -o "$DEST/conductor-tui" && chmod +x "$DEST/conductor-tui"

# Add DEST to PATH if not already present
case ":$PATH:" in
  *:"$DEST":*) ;;
  *)
    SHELL_NAME=$(basename "$SHELL")
    case "$SHELL_NAME" in
      bash) PROFILE="$HOME/.bashrc" ;;
      zsh)  PROFILE="$HOME/.zshrc" ;;
      fish) PROFILE="$HOME/.config/fish/config.fish" ;;
      *)    PROFILE="" ;;
    esac
    if [ -n "$PROFILE" ]; then
      if [ "$SHELL_NAME" = "fish" ]; then
        EXPORT_LINE="set -gx PATH \"$DEST\" \$PATH"
      else
        EXPORT_LINE="export PATH=\"$DEST:\$PATH\""
      fi
      if ! grep -qF "$DEST" "$PROFILE" 2>/dev/null; then
        echo "" >> "$PROFILE"
        echo "$EXPORT_LINE" >> "$PROFILE"
        echo "Added $DEST to PATH in $PROFILE"
        echo "Please restart your terminal or run: source $PROFILE"
      fi
    else
      echo "Could not detect shell profile. Please add $DEST to your PATH manually."
    fi
    ;;
esac

echo "Done. Run: conductor-tui"
