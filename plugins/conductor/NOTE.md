# Known Limitations

> [!WARNING]
> **Autocomplete not working for custom marketplace skills**
>
> Skills from custom marketplace plugins may not appear in Claude Code's slash command autocomplete menu.
> This is a known issue affecting all custom marketplace plugins.
>
> **Tracking issue:** [anthropics/claude-code#18949](https://github.com/anthropics/claude-code/issues/18949)
>
> See also: [anthropics/claude-code#20802](https://github.com/anthropics/claude-code/issues/20802)

## Workaround

Until the issue is resolved, you can use one of these approaches:

### Option 1: Type commands manually

The skills work correctly when typed manually:

- `/conductor:setup` - Initialize Conductor in your project
- `/conductor:new-track` - Create a new development track
- `/conductor:implement` - Implement tasks from the current track
- `/conductor:status` - View project progress
- `/conductor:review` - Review completed work
- `/conductor:revert` - Revert changes from a track

### Option 2: Run setup script (Unix/macOS/Linux)

Run the included script to enable autocomplete:

```bash
~/.claude/plugins/marketplaces/cloudaura-marketplace/plugins/conductor/scripts/setup-autocomplete.sh
```

Then restart Claude Code. Skills will appear in autocomplete as `/conductor:setup`, `/conductor:implement`, etc.

**Note:** You may need to re-run this script after Claude Code updates the plugin.

---

*This note will be removed once the upstream issue is resolved.*
