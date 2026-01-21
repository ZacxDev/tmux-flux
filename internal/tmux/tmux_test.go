package tmux

import (
	"os"
	"testing"
)

func TestIsInsideTmux(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{
			name:     "inside tmux",
			envValue: "/tmp/tmux-1000/default,12345,0",
			want:     true,
		},
		{
			name:     "outside tmux",
			envValue: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := os.Getenv("TMUX")
			defer os.Setenv("TMUX", original)

			// Set test value
			if tt.envValue == "" {
				os.Unsetenv("TMUX")
			} else {
				os.Setenv("TMUX", tt.envValue)
			}

			got := IsInsideTmux()
			if got != tt.want {
				t.Errorf("IsInsideTmux() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSessionExists(t *testing.T) {
	// This test requires tmux to be running
	// Skip if tmux is not available
	if _, err := os.Stat("/usr/bin/tmux"); os.IsNotExist(err) {
		t.Skip("tmux not installed, skipping test")
	}

	// Test with a session that definitely doesn't exist
	exists := SessionExists("__nonexistent_test_session_12345__")
	if exists {
		t.Error("SessionExists() returned true for non-existent session")
	}
}

func TestAttachSession(t *testing.T) {
	cmd := AttachSession("test-session")

	if cmd == nil {
		t.Fatal("AttachSession() returned nil")
	}

	if cmd.Path == "" {
		t.Error("AttachSession() returned command with empty path")
	}

	// Check that the command has the right arguments
	args := cmd.Args
	if len(args) < 4 {
		t.Fatalf("Expected at least 4 args, got %d", len(args))
	}

	if args[1] != "attach-session" {
		t.Errorf("Expected 'attach-session', got %s", args[1])
	}

	if args[2] != "-t" {
		t.Errorf("Expected '-t', got %s", args[2])
	}

	if args[3] != "test-session" {
		t.Errorf("Expected 'test-session', got %s", args[3])
	}
}

func TestListSessions_NoServer(t *testing.T) {
	// This test checks behavior when no tmux server is running
	// We can't guarantee the server state, so we just verify the function doesn't panic
	sessions, err := ListSessions()

	// Either we get sessions or an empty list with no error (no server)
	// Both are valid outcomes
	if err != nil {
		t.Logf("ListSessions() returned error (may be expected): %v", err)
	}

	if sessions == nil {
		t.Error("ListSessions() returned nil slice, expected empty slice")
	}
}
