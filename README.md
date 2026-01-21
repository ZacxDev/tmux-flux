# tmux-flux

A fast, keyboard-driven TUI for managing tmux sessions with grouping, vim-style navigation, fuzzy search, and live preview.

```
tmux-flux
─────────────────────────────────────────────────────────────────────
                              │
▼ work                        │ Preview: api-server
  ● api-server                │ ────────────────────────────────────
  ○ frontend                  │ $ npm run dev
  ○ database                  │ > api-server@1.0.0 dev
▼ personal                    │ > node src/index.js
  → dotfiles                  │
  ○ notes                     │ Server listening on port 3000
▶ archived                    │ Connected to database
                              │ Ready for connections...
                              │
j/k navigate  enter attach  / search  n new  ? help  q quit
```

## Why tmux-flux?

- **Organize chaos**: Group related sessions together (by project, client, or context)
- **Navigate fast**: Vim-style keys mean your hands never leave home row
- **Find instantly**: Fuzzy search across all sessions with `/`
- **Live preview**: See session content before switching
- **Works everywhere**: Runs inside or outside tmux seamlessly

## Installation

### From Source

```bash
git clone https://github.com/zchase/tmux-flux.git
cd tmux-flux
make build
make install  # Installs to ~/.local/bin/
```

### Go Install

```bash
go install github.com/zchase/tmux-flux/cmd/tmux-flux@latest
```

### Requirements

- Go 1.21+
- tmux 3.0+

### Tmux Integration (Ctrl+Q)

Add to your `~/.tmux.conf`:

```bash
# Option 1: Using the plugin script
run-shell /path/to/tmux-flux.tmux

# Option 2: Manual binding (if tmux-flux is in PATH)
bind-key -n C-q display-popup -E -w 80% -h 80% "tmux-flux"
```

Then reload tmux config:
```bash
tmux source-file ~/.tmux.conf
```

Now press `Ctrl+Q` anywhere in tmux to open the session selector.

## Quick Start

```bash
# Launch the TUI
tmux-flux

# Inside tmux-flux:
# - Press 'n' to create a new session
# - Use j/k to navigate
# - Press Enter to attach
# - Press 'm' to move session to a group
# - Press '/' to search
```

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `Ctrl+d` | Page down |
| `Ctrl+u` | Page up |

### Groups

| Key | Action |
|-----|--------|
| `N` | Create new group |
| `h` / `←` | Collapse group (or collapse parent if on session) |
| `l` / `→` | Expand group |
| `Tab` / `Space` | Toggle expand/collapse |

### Session Operations

| Key | Action |
|-----|--------|
| `Enter` | Attach to session (or switch if inside tmux) |
| `n` / `c` | Create new session |
| `r` | Rename session |
| `d` / `x` | Delete session (with confirmation) |
| `m` | Move session to a group |

### Search & Other

| Key | Action |
|-----|--------|
| `/` | Open fuzzy search |
| `R` | Refresh session list |
| `?` | Toggle help screen |
| `q` / `Ctrl+c` | Quit |

## Workflow Examples

### Organizing by Project

```
▼ project-alpha
  ○ alpha-api
  ○ alpha-frontend
  ○ alpha-db
▼ project-beta
  ○ beta-api
  ○ beta-tests
```

1. Create sessions: `n` → enter name
2. Move to group: `m` → enter group name (e.g., "project-alpha")
3. Collapse finished work: `h` on group header

### Quick Session Switching

1. Press `/` to search
2. Type part of session name (fuzzy matched)
3. Press `Enter` to attach

### Inside vs Outside tmux

| Context | Enter Behavior |
|---------|---------------|
| Outside tmux | Attaches to session (replaces terminal) |
| Inside tmux | Switches client to session (stays in tmux) |

### Session Preview

The preview pane shows the current content of the selected session's active pane:

- Preview updates automatically when navigating to a different session
- Only sessions show previews (groups show a placeholder)
- Preview is captured using `tmux capture-pane`

## Configuration

### Data Location

Sessions and groups are stored in:
```
~/.config/tmux-flux/sessions.json
```

### Storage Format

```json
[
  {
    "name": "api-server",
    "group": "work",
    "created_at": "2024-01-15T10:30:00Z"
  },
  {
    "name": "dotfiles",
    "group": "personal",
    "created_at": "2024-01-14T09:00:00Z"
  }
]
```

### Syncing with tmux

tmux-flux automatically syncs with tmux on startup and when you press `R`:
- New tmux sessions appear as "Ungrouped"
- Deleted tmux sessions are removed from the list
- Session state (attached, windows) is updated in real-time

## Architecture

```
tmux-flux/
├── cmd/tmux-flux/          # CLI entry point
│   └── main.go
├── internal/
│   ├── tmux/               # Tmux integration layer
│   │   └── tmux.go         # Session commands (list, create, kill, etc.)
│   ├── session/            # Business logic
│   │   ├── session.go      # Session & group management
│   │   └── storage.go      # JSON persistence
│   └── ui/                 # TUI components
│       ├── model.go        # Bubble Tea model
│       ├── styles.go       # Lipgloss styles (Tokyo Night theme)
│       └── keys.go         # Key bindings
├── docs/                   # Documentation
├── Makefile
└── go.mod
```

## Development

```bash
# Build
make build

# Run
make run

# Clean
make clean

# Tidy dependencies
make tidy

# Run tests
make test

# Run tests with coverage
go test ./... -cover

# Run tests with verbose output
go test ./... -v

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage

| Package | Coverage |
|---------|----------|
| `internal/ui` | 69.1% |
| `internal/tmux` | 63.8% |
| `internal/session` | 52.0% |
| **Total** | **63.4%** |

### Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - Reusable TUI components
- [Fuzzy](https://github.com/sahilm/fuzzy) - Fuzzy string matching

## Troubleshooting

### "no server running" error

tmux server isn't running. Start it with:
```bash
tmux new-session -d -s scratch
```

### Sessions not appearing

Press `R` to refresh, or check that sessions exist:
```bash
tmux list-sessions
```

### Can't attach to session

If inside tmux, tmux-flux uses `switch-client` instead of `attach`. This is expected behavior.

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make changes and test
4. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Credits

Inspired by [agent-deck](https://github.com/smtg-ai/agent-deck) and the excellent [charmbracelet](https://charm.sh/) ecosystem.
