#!/usr/bin/env bash
# Workaround for Claude Code issue #18949
# Creates symlinks to enable autocomplete for conductor skills
# https://github.com/anthropics/claude-code/issues/18949

set -e

SKILLS_DIR="${HOME}/.claude/skills"
MARKETPLACE_SKILLS="${HOME}/.claude/plugins/marketplaces/cloudaura-marketplace/plugins/conductor/skills"

# Check if marketplace is installed
if [[ ! -d "$MARKETPLACE_SKILLS" ]]; then
    echo "Error: Conductor plugin not found at $MARKETPLACE_SKILLS"
    echo "Install it first: /plugin install conductor@cloudaura-marketplace"
    exit 1
fi

# Create skills directory if needed
mkdir -p "$SKILLS_DIR"

# Skills to link (symlink name must match 'name' field in SKILL.md)
# Format: "skill_dir:symlink_name"
SKILLS=(
    "setup:conductor:setup"
    "implement:conductor:implement"
    "status:conductor:status"
    "new-track:conductor:new-track"
    "review:conductor:review"
    "revert:conductor:revert"
)

echo "Creating symlinks for conductor skills..."

for entry in "${SKILLS[@]}"; do
    skill_dir="${entry%%:*}"
    symlink_name="${entry#*:}"

    target="$SKILLS_DIR/$symlink_name"
    source="$MARKETPLACE_SKILLS/$skill_dir"

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
done

echo ""
echo "Done! Restart Claude Code to enable autocomplete."
echo ""
echo "Available commands (with autocomplete):"
echo "  /conductor:setup"
echo "  /conductor:implement"
echo "  /conductor:status"
echo "  /conductor:new-track"
echo "  /conductor:review"
echo "  /conductor:revert"
