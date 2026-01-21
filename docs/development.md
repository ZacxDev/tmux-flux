# Development Guide

This guide covers setting up a development environment and contributing to tmux-flux.

## Prerequisites

- Go 1.21 or later
- tmux 3.0 or later
- Git

## Setup

### Clone the Repository

```bash
git clone https://github.com/zchase/tmux-flux.git
cd tmux-flux
```

### Install Dependencies

```bash
go mod download
```

### Build

```bash
make build
# or
go build -o tmux-flux ./cmd/tmux-flux/
```

### Run

```bash
make run
# or
./tmux-flux
```

## Project Structure

```
tmux-flux/
├── cmd/
│   └── tmux-flux/
│       └── main.go          # Entry point
├── internal/
│   ├── tmux/
│   │   └── tmux.go          # Tmux command execution
│   ├── session/
│   │   ├── session.go       # Session/group management
│   │   └── storage.go       # JSON persistence
│   └── ui/
│       ├── model.go         # Bubble Tea model
│       ├── styles.go        # Lipgloss styles
│       └── keys.go          # Key bindings
├── docs/
│   ├── architecture.md
│   ├── configuration.md
│   └── development.md
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Development Workflow

### Make Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/my-feature
   ```

2. Make your changes

3. Build and test:
   ```bash
   make build
   ./tmux-flux
   ```

4. Commit:
   ```bash
   git add .
   git commit -m "Add my feature"
   ```

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and small
- Add comments for exported functions

### Testing

```bash
# Run all tests
make test
# or
go test ./...

# Run specific package tests
go test ./internal/session/...

# Run with verbose output
go test -v ./...
```

## Key Components

### Adding a New Key Binding

1. **Define the binding** in `internal/ui/keys.go`:
   ```go
   type KeyMap struct {
       // ...existing bindings...
       MyAction key.Binding
   }

   func DefaultKeyMap() KeyMap {
       return KeyMap{
           // ...existing bindings...
           MyAction: key.NewBinding(
               key.WithKeys("M"),
               key.WithHelp("M", "my action"),
           ),
       }
   }
   ```

2. **Handle the key** in `internal/ui/model.go`:
   ```go
   func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
       switch {
       // ...existing cases...
       case key.Matches(msg, m.keys.MyAction):
           // Handle the action
           return m, nil
       }
   }
   ```

3. **Update help** in `renderHelpBar()` if needed.

### Adding a New Mode

1. **Add mode constant** in `internal/ui/model.go`:
   ```go
   const (
       ModeNormal Mode = iota
       // ...existing modes...
       ModeMyMode
   )
   ```

2. **Add update handler**:
   ```go
   func (m Model) updateMyMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
       switch {
       case key.Matches(msg, m.keys.Escape):
           m.mode = ModeNormal
           return m, nil
       // Handle mode-specific keys
       }
       return m, nil
   }
   ```

3. **Route to handler** in `Update()`:
   ```go
   switch m.mode {
   // ...existing cases...
   case ModeMyMode:
       return m.updateMyMode(msg)
   }
   ```

4. **Render mode** in `View()` if needed.

### Adding a Tmux Command

1. **Add function** in `internal/tmux/tmux.go`:
   ```go
   func MyTmuxCommand(args ...string) error {
       cmd := exec.Command("tmux", "my-command", args...)
       var stderr bytes.Buffer
       cmd.Stderr = &stderr

       if err := cmd.Run(); err != nil {
           return fmt.Errorf("my-command failed: %s", stderr.String())
       }
       return nil
   }
   ```

2. **Use in session manager** or UI model as needed.

### Adding a Session Operation

1. **Add manager method** in `internal/session/session.go`:
   ```go
   func (m *Manager) MyOperation(sessionName string) error {
       // Validate
       sess, ok := m.sessions[sessionName]
       if !ok {
           return fmt.Errorf("session not found: %s", sessionName)
       }

       // Execute tmux command
       if err := tmux.MyTmuxCommand(sessionName); err != nil {
           return err
       }

       // Update state
       // ...

       // Persist
       return m.save()
   }
   ```

2. **Call from UI** via key binding or command.

## Debugging

### Print Debugging

For quick debugging, use `tea.Println`:
```go
return m, tea.Println("Debug: cursor =", m.cursor)
```

Note: This prints to the alternate screen buffer.

### Log to File

For persistent logging:
```go
import "log"
import "os"

func init() {
    f, _ := os.OpenFile("/tmp/tmux-flux.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    log.SetOutput(f)
}

// Then use:
log.Printf("Debug: %v", value)
```

### Inspecting tmux State

```bash
# List all sessions with details
tmux list-sessions -F "#{session_name}\t#{session_windows}\t#{session_attached}"

# Check if tmux server is running
tmux has-session 2>/dev/null && echo "running" || echo "not running"
```

## Dependencies

### Core Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | Styling |
| `github.com/charmbracelet/bubbles` | UI components |
| `github.com/sahilm/fuzzy` | Fuzzy matching |

### Updating Dependencies

```bash
go get -u ./...
go mod tidy
```

## Release Process

1. Update version (if applicable)
2. Run tests: `make test`
3. Build: `make build`
4. Test manually
5. Create tag: `git tag v0.1.0`
6. Push: `git push origin main --tags`

## Common Issues

### "go: module not found"

```bash
go mod tidy
```

### "could not open TTY"

tmux-flux requires a terminal. It cannot run in:
- CI/CD pipelines (without TTY)
- IDE run configurations (without terminal)
- Piped input

### Build Errors After Dependency Update

```bash
go clean -modcache
go mod download
```

## Resources

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [tmux Manual](https://man.openbsd.org/tmux.1)
