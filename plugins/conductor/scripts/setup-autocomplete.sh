#!/usr/bin/env bash
# Workaround for Claude Code issue #18949
# Creates symlinks to enable autocomplete for conductor skills
# https://github.com/anthropics/claude-code/issues/18949

set -e

SKILLS_DIR="${HOME}/.claude/skills"
# Resolve skills relative to this script, so it works for both a GitHub
# marketplace and a local directory one (`claude plugin marketplace add <path>`)
PLUGIN_SKILLS="$(cd "$(dirname "$0")/../skills" && pwd)"

# Create skills directory if needed
mkdir -p "$SKILLS_DIR"

echo "Creating symlinks for conductor skills..."

names=()
for skill_md in "$PLUGIN_SKILLS"/*/SKILL.md; do
    source="$(dirname "$skill_md")"
    # Symlink name must match 'name' field in SKILL.md
    symlink_name="$(sed -n 's/^name: *//p' "$skill_md" | head -1)"
    if [[ -z "$symlink_name" ]]; then
        echo "  Skipping: $source (no name field in SKILL.md)"
        continue
    fi

    target="$SKILLS_DIR/$symlink_name"

    if [[ -L "$target" ]]; then
        echo "  Updating: $symlink_name"
        rm "$target"
    elif [[ -e "$target" ]]; then
        echo "  Skipping: $symlink_name (file exists, not a symlink)"
        continue
    else
        echo "  Creating: $symlink_name"
    fi

    ln -s "$source" "$target"
    names+=("$symlink_name")
done

echo ""
echo "Done! Restart Claude Code to enable autocomplete."
echo ""
echo "Available commands (with autocomplete):"
for name in "${names[@]}"; do
    echo "  /$name"
done
