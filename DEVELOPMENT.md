# Development

## Local Testing

To test local changes without reinstalling the plugin:

```bash
# Replace marketplace plugin with symlink to local repo
rm -rf ~/.claude/plugins/marketplaces/cloudaura-marketplace/plugins/conductor
ln -s $(pwd)/plugins/conductor ~/.claude/plugins/marketplaces/cloudaura-marketplace/plugins/conductor
```

Restart Claude Code session to load changes.

> **Note:** Running `claude plugin update` will replace the symlink with the marketplace version.
