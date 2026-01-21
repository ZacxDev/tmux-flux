# Architecture

This document describes the internal architecture of tmux-flux.

## Overview

tmux-flux follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────┐
│                    CLI                       │
│              (cmd/tmux-flux)                 │
├─────────────────────────────────────────────┤
│                    UI                        │
│               (internal/ui)                  │
│   ┌─────────┐ ┌─────────┐ ┌─────────┐      │
│   │  Model  │ │ Styles  │ │  Keys   │      │
│   └─────────┘ └─────────┘ └─────────┘      │
├─────────────────────────────────────────────┤
│              Session Layer                   │
│            (internal/session)                │
│   ┌─────────────┐ ┌─────────────┐          │
│   │   Manager   │ │   Storage   │          │
│   └─────────────┘ └─────────────┘          │
├─────────────────────────────────────────────┤
│              Tmux Layer                      │
│             (internal/tmux)                  │
│   ┌─────────────────────────────┐          │
│   │      Tmux Commands          │          │
│   └─────────────────────────────┘          │
└─────────────────────────────────────────────┘
```

## Layers

### CLI Layer (`cmd/tmux-flux/`)

The entry point that:
- Initializes the session manager
- Creates the Bubble Tea program
- Handles startup and shutdown

### UI Layer (`internal/ui/`)

The terminal user interface built with Bubble Tea.

#### Model (`model.go`)

The core Bubble Tea model implementing:
- `Init()` - Initial commands (refresh, set title)
- `Update()` - Message handling and state transitions
- `View()` - Rendering the UI

**Modes:**
```go
type Mode int

const (
    ModeNormal Mode = iota  // Default navigation
    ModeSearch              // Fuzzy search active
    ModeCreate              // Creating new session
    ModeRename              // Renaming session
    ModeDelete              // Delete confirmation
    ModeGroup               // Moving to group
    ModeHelp                // Help screen
)
```

**State:**
```go
type Model struct {
    manager     *session.Manager  // Business logic
    keys        KeyMap            // Key bindings
    mode        Mode              // Current UI mode
    cursor      int               // Selected item index
    items       []ListItem        // Flattened view items
    searchQuery string            // Current search filter
    // ... input fields, dimensions, etc.
}
```

#### Styles (`styles.go`)

Lipgloss style definitions using Tokyo Night color palette:
- `ColorBg`, `ColorFg` - Base colors
- `ColorAccent` - Highlights (blue)
- `ColorGreen`, `ColorRed`, `ColorYellow` - Status colors

#### Keys (`keys.go`)

Key binding definitions using Bubbles key package:
- Navigation: `j/k`, arrows, `gg/G`, `Ctrl+u/d`
- Actions: `Enter`, `n`, `r`, `d`, `m`
- Modes: `/`, `?`, `Esc`

### Session Layer (`internal/session/`)

Business logic for session and group management.

#### Manager (`session.go`)

Central coordinator that:
- Maintains session registry (`map[string]*Session`)
- Maintains group registry (`map[string]*Group`)
- Syncs with tmux state
- Handles CRUD operations
- Provides fuzzy search

**Key Operations:**
```go
func (m *Manager) Refresh() error           // Sync with tmux
func (m *Manager) CreateSession(...) error  // Create new session
func (m *Manager) RenameSession(...) error  // Rename session
func (m *Manager) DeleteSession(...) error  // Kill and remove
func (m *Manager) SetSessionGroup(...) error // Move to group
func (m *Manager) FuzzySearch(query) []*Session
```

#### Storage (`storage.go`)

JSON persistence for session metadata:
- Location: `~/.config/tmux-flux/sessions.json`
- Stores: name, group, created_at
- Does NOT store: runtime state (windows, attached)

### Tmux Layer (`internal/tmux/`)

Low-level tmux command execution.

**Commands Used:**
| Function | Tmux Command |
|----------|--------------|
| `ListSessions()` | `tmux list-sessions -F "..."` |
| `CreateSession()` | `tmux new-session -d -s NAME` |
| `RenameSession()` | `tmux rename-session -t OLD NEW` |
| `KillSession()` | `tmux kill-session -t NAME` |
| `AttachSession()` | `tmux attach-session -t NAME` |
| `SwitchClient()` | `tmux switch-client -t NAME` |
| `SessionExists()` | `tmux has-session -t NAME` |
| `IsInsideTmux()` | `tmux display-message -p ...` |

## Data Flow

### Startup

```
main.go
  │
  ├─► NewManager(storagePath)
  │     ├─► NewStorage()
  │     └─► loadFromStorage()
  │
  ├─► manager.Refresh()
  │     ├─► tmux.ListSessions()
  │     ├─► Update session states
  │     └─► storage.Save()
  │
  └─► tea.NewProgram(model).Run()
```

### User Action (e.g., Create Session)

```
KeyMsg{key: "n"}
  │
  ├─► Model.Update() → ModeCreate
  │
  ├─► KeyMsg{key: "enter"}
  │     │
  │     └─► manager.CreateSession(name, group, "")
  │           ├─► tmux.CreateSession(name, "")
  │           ├─► Add to sessions map
  │           ├─► Add to group
  │           └─► storage.Save()
  │
  └─► refreshCmd() → rebuildItems()
```

### Refresh Cycle

```
manager.Refresh()
  │
  ├─► tmux.ListSessions()
  │     └─► Returns []tmux.Session
  │
  ├─► For each existing session:
  │     ├─► If in tmux: Update runtime state
  │     └─► If not in tmux: Remove from manager
  │
  ├─► For each new tmux session:
  │     └─► Add to manager (ungrouped)
  │
  └─► storage.Save()
```

## Key Design Decisions

### 1. Tmux as Source of Truth

The tmux server is authoritative for:
- Session existence
- Runtime state (attached, windows, activity)

tmux-flux only persists:
- Group assignments
- Creation timestamps

This prevents state drift and ensures consistency.

### 2. Flat Item List

Groups and sessions are flattened into a single `[]ListItem` for rendering:

```go
type ListItem struct {
    Type     string  // "group" or "session"
    Name     string
    Group    string
    Session  *Session
    Attached bool
}
```

Benefits:
- Simple cursor navigation (just an index)
- Easy scrolling calculation
- Straightforward rendering loop

### 3. Mode-Based Input Handling

Each mode has its own update function:
- `updateNormal()` - Navigation and commands
- `updateSearch()` - Search input
- `updateCreate()` - Session name input
- `updateRename()` - New name input
- `updateDelete()` - Confirmation
- `updateGroup()` - Group name input

This keeps input handling clean and predictable.

### 4. Smart Attach vs Switch

```go
if m.insideTmux {
    tmux.SwitchClient(name)  // Stay in tmux
} else {
    tea.ExecProcess(tmux.AttachSession(name), ...)  // Replace process
}
```

Detected at startup via `tmux display-message`.

## Extension Points

### Adding New Session Operations

1. Add key binding in `keys.go`
2. Add mode constant if needed
3. Add update handler in `model.go`
4. Add manager method in `session.go`
5. Add tmux command in `tmux.go` if needed

### Adding Configuration

1. Create `internal/config/config.go`
2. Add TOML parsing
3. Load in `main.go`
4. Pass to Model and Manager

### Adding Themes

1. Define color palette in `styles.go`
2. Add theme selection logic
3. Store preference in config
