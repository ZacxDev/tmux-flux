package tmux

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Session represents a tmux session
type Session struct {
	Name      string
	ID        string
	Windows   int
	Attached  bool
	Created   int64
	Activity  int64
}

// ListSessions returns all tmux sessions
func ListSessions() ([]Session, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F",
		"#{session_name}\t#{session_id}\t#{session_windows}\t#{session_attached}\t#{session_created}\t#{session_activity}")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// No sessions is not an error
		if strings.Contains(stderr.String(), "no server running") ||
			strings.Contains(stderr.String(), "no sessions") {
			return []Session{}, nil
		}
		return nil, fmt.Errorf("tmux list-sessions: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	sessions := make([]Session, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 6 {
			continue
		}

		var windows int
		var attached int
		var created, activity int64
		fmt.Sscanf(parts[2], "%d", &windows)
		fmt.Sscanf(parts[3], "%d", &attached)
		fmt.Sscanf(parts[4], "%d", &created)
		fmt.Sscanf(parts[5], "%d", &activity)

		sessions = append(sessions, Session{
			Name:     parts[0],
			ID:       parts[1],
			Windows:  windows,
			Attached: attached == 1,
			Created:  created,
			Activity: activity,
		})
	}

	return sessions, nil
}

// CreateSession creates a new tmux session
func CreateSession(name string, startDir string) error {
	args := []string{"new-session", "-d", "-s", name}
	if startDir != "" {
		args = append(args, "-c", startDir)
	}

	cmd := exec.Command("tmux", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create session %q: %s", name, stderr.String())
	}
	return nil
}

// RenameSession renames an existing tmux session
func RenameSession(oldName, newName string) error {
	cmd := exec.Command("tmux", "rename-session", "-t", oldName, newName)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rename session %q to %q: %s", oldName, newName, stderr.String())
	}
	return nil
}

// KillSession kills a tmux session
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kill session %q: %s", name, stderr.String())
	}
	return nil
}

// AttachSession attaches to a tmux session (for use with exec)
func AttachSession(name string) *exec.Cmd {
	return exec.Command("tmux", "attach-session", "-t", name)
}

// SwitchClient switches the current client to a session
func SwitchClient(name string) error {
	cmd := exec.Command("tmux", "switch-client", "-t", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("switch to session %q: %s", name, stderr.String())
	}
	return nil
}

// SessionExists checks if a session exists
func SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// IsInsideTmux returns true if we're running inside tmux
func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// GetCurrentSession returns the current tmux session name if inside tmux
func GetCurrentSession() (string, error) {
	cmd := exec.Command("tmux", "display-message", "-p", "#{session_name}")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}
