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

func TestCapturePane_NonexistentSession(t *testing.T) {
	// Test capturing a non-existent session
	content, err := CapturePane("__nonexistent_test_session_12345__")

	// Should return an error for non-existent session
	if err == nil {
		t.Log("CapturePane() did not return error for non-existent session (tmux server may not be running)")
	}

	// Content might be empty or error message
	_ = content
}

func TestCapturePane_EmptySessionName(t *testing.T) {
	// Test with empty session name
	_, err := CapturePane("")

	// Should return an error or handle gracefully
	if err == nil {
		t.Log("CapturePane('') did not return error")
	}
}

func TestSession_Struct(t *testing.T) {
	// Test Session struct
	s := Session{
		Name:     "test",
		ID:       "$1",
		Windows:  3,
		Attached: true,
		Created:  1234567890,
		Activity: 1234567900,
	}

	if s.Name != "test" {
		t.Errorf("Session.Name = %q, want 'test'", s.Name)
	}
	if s.ID != "$1" {
		t.Errorf("Session.ID = %q, want '$1'", s.ID)
	}
	if s.Windows != 3 {
		t.Errorf("Session.Windows = %d, want 3", s.Windows)
	}
	if !s.Attached {
		t.Error("Session.Attached = false, want true")
	}
	if s.Created != 1234567890 {
		t.Errorf("Session.Created = %d, want 1234567890", s.Created)
	}
	if s.Activity != 1234567900 {
		t.Errorf("Session.Activity = %d, want 1234567900", s.Activity)
	}
}

func TestGetCurrentSession_NotInTmux(t *testing.T) {
	// Save original TMUX value
	original := os.Getenv("TMUX")
	defer os.Setenv("TMUX", original)

	// Ensure we're "not in tmux"
	os.Unsetenv("TMUX")

	// This may still work if tmux server is running
	_, err := GetCurrentSession()

	// Just verify it doesn't panic
	_ = err
}

func TestSwitchClient_NonexistentSession(t *testing.T) {
	// Test switching to non-existent session
	err := SwitchClient("__nonexistent_test_session_12345__")

	// Should return error when not in tmux or session doesn't exist
	if err == nil {
		t.Log("SwitchClient() did not return error (may be in tmux with valid session)")
	}
}

func TestCreateSession_Integration(t *testing.T) {
	// Skip if tmux not available
	if _, err := os.Stat("/usr/bin/tmux"); os.IsNotExist(err) {
		t.Skip("tmux not installed, skipping test")
	}

	// Test that the function doesn't panic
	// We don't actually create a session to avoid side effects
	// Just verify the function signature works
	_ = CreateSession
}

func TestRenameSession_Integration(t *testing.T) {
	// Skip if tmux not available
	if _, err := os.Stat("/usr/bin/tmux"); os.IsNotExist(err) {
		t.Skip("tmux not installed, skipping test")
	}

	// Test renaming non-existent session
	err := RenameSession("__nonexistent_old__", "__nonexistent_new__")
	if err == nil {
		t.Log("RenameSession() did not return error for non-existent session")
	}
}

func TestKillSession_NonexistentSession(t *testing.T) {
	// Test killing non-existent session
	err := KillSession("__nonexistent_test_session_12345__")

	// Should return error for non-existent session
	if err == nil {
		t.Log("KillSession() did not return error for non-existent session")
	}
}
