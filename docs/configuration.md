# Configuration

This document describes tmux-flux configuration options and data storage.

## Data Storage

### Location

tmux-flux stores session data in:
```
~/.config/tmux-flux/sessions.json
```

The directory is created automatically on first run.

### File Format

```json
[
  {
    "name": "my-project",
    "group": "work",
    "created_at": "2024-01-15T10:30:00Z"
  },
  {
    "name": "dotfiles",
    "group": "personal",
    "created_at": "2024-01-14T09:00:00Z"
  },
  {
    "name": "scratch",
    "group": "",
    "created_at": "2024-01-16T14:00:00Z"
  }
]
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Session name (matches tmux session name) |
| `group` | string | Group name (empty string = ungrouped) |
| `created_at` | string | ISO 8601 timestamp |

### What's NOT Stored

Runtime state is fetched from tmux on each refresh:
- Number of windows
- Attached status
- Last activity time

This ensures tmux-flux always reflects the true tmux state.

## Groups

### Group Names

Groups are identified by their name string:
- Empty string (`""`) = Ungrouped sessions
- Any non-empty string = Named group

### Group Operations

**Move to group:**
1. Select a session
2. Press `m`
3. Enter group name (or empty for ungrouped)
4. Press Enter

**Collapse/Expand:**
- `h` or `←` to collapse
- `l` or `→` to expand
- `Tab` or `Space` to toggle

### Group Display Order

Groups are displayed alphabetically, with "Ungrouped" always last.

## Syncing with tmux

### Automatic Sync

tmux-flux syncs with tmux:
1. On startup
2. When pressing `R` (refresh)
3. After session operations (create, rename, delete)

### Sync Behavior

| Scenario | tmux-flux Behavior |
|----------|-------------------|
| New tmux session found | Added as "Ungrouped" |
| tmux session deleted | Removed from list |
| Session renamed in tmux | Name updated (group preserved) |
| Session created via tmux-flux | Added with specified group |

### Manual Sync

If sessions get out of sync:
```bash
# Inside tmux-flux
R  # Press R to refresh

# Or restart tmux-flux
q  # Quit
tmux-flux  # Restart
```

## Environment

### Inside vs Outside tmux

tmux-flux detects its context automatically:

| Context | Detection | Attach Behavior |
|---------|-----------|-----------------|
| Outside tmux | `tmux display-message` fails | `tmux attach-session` |
| Inside tmux | `tmux display-message` succeeds | `tmux switch-client` |

### Working Directory

When creating sessions, tmux-flux uses:
- The current working directory (if specified)
- Otherwise, the default tmux behavior (user's home)

## Customization (Future)

The following customization options are planned for future releases:

### Planned: Config File

```toml
# ~/.config/tmux-flux/config.toml

[ui]
theme = "tokyo-night"  # or "dracula", "gruvbox", etc.

[keys]
# Custom key bindings
create = "a"
delete = "D"

[behavior]
confirm_delete = true
auto_refresh = true
refresh_interval = 5  # seconds
```

### Planned: Themes

Custom color schemes via config:
```toml
[theme.custom]
background = "#1a1b26"
foreground = "#c0caf5"
accent = "#7aa2f7"
```

### Planned: Session Templates

Pre-defined session configurations:
```toml
[[templates]]
name = "dev-project"
windows = ["editor", "server", "shell"]
layout = "main-vertical"
```

## Backup and Restore

### Backup

```bash
cp ~/.config/tmux-flux/sessions.json ~/backup/
```

### Restore

```bash
cp ~/backup/sessions.json ~/.config/tmux-flux/
```

### Reset

```bash
rm ~/.config/tmux-flux/sessions.json
# Next run will start fresh
```

## Troubleshooting

### Corrupt Data File

If the JSON file becomes corrupted:
```bash
# Backup current file
mv ~/.config/tmux-flux/sessions.json ~/.config/tmux-flux/sessions.json.bak

# Start fresh - existing tmux sessions will appear as "Ungrouped"
tmux-flux
```

### Permission Issues

```bash
# Check permissions
ls -la ~/.config/tmux-flux/

# Fix if needed
chmod 755 ~/.config/tmux-flux/
chmod 644 ~/.config/tmux-flux/sessions.json
```

### Data Location Override

Currently, the data location is hardcoded. To use a different location, modify `session.GetDefaultStoragePath()` in the source code.

Future versions may support:
```bash
TMUX_FLUX_DATA_DIR=~/my/path tmux-flux
```
