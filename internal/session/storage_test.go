package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStorage(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "subdir", "sessions.json")

	storage, err := NewStorage(storagePath)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}

	if storage == nil {
		t.Fatal("NewStorage() returned nil")
	}

	// Check that directory was created
	dir := filepath.Dir(storagePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("NewStorage() did not create directory")
	}
}

func TestStorage_LoadEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	storage, err := NewStorage(storagePath)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}

	// Load from non-existent file should return empty slice
	sessions, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if sessions == nil {
		t.Error("Load() returned nil, expected empty slice")
	}

	if len(sessions) != 0 {
		t.Errorf("Load() returned %d sessions, expected 0", len(sessions))
	}
}

func TestStorage_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	storage, err := NewStorage(storagePath)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}

	// Create test sessions
	now := time.Now().Truncate(time.Second) // Truncate for comparison
	sessions := []*Session{
		{Name: "session1", Group: "group1", CreatedAt: now},
		{Name: "session2", Group: "group2", CreatedAt: now},
		{Name: "session3", Group: "", CreatedAt: now},
	}

	// Save
	err = storage.Save(sessions)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		t.Fatal("Save() did not create file")
	}

	// Load
	loaded, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded) != len(sessions) {
		t.Fatalf("Load() returned %d sessions, expected %d", len(loaded), len(sessions))
	}

	// Verify content
	for i, s := range loaded {
		if s.Name != sessions[i].Name {
			t.Errorf("Session %d Name = %q, want %q", i, s.Name, sessions[i].Name)
		}
		if s.Group != sessions[i].Group {
			t.Errorf("Session %d Group = %q, want %q", i, s.Group, sessions[i].Group)
		}
	}
}

func TestStorage_LoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	// Write invalid JSON
	err := os.WriteFile(storagePath, []byte("not valid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	storage, err := NewStorage(storagePath)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}

	_, err = storage.Load()
	if err == nil {
		t.Error("Load() should return error for invalid JSON")
	}
}

func TestStorage_SaveCreatesValidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	storage, err := NewStorage(storagePath)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}

	sessions := []*Session{
		{Name: "test", Group: "group", CreatedAt: time.Now()},
	}

	err = storage.Save(sessions)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Read raw file and verify it's valid JSON
	data, err := os.ReadFile(storagePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var parsed []map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Errorf("Save() created invalid JSON: %v", err)
	}

	if len(parsed) != 1 {
		t.Errorf("Expected 1 session in JSON, got %d", len(parsed))
	}
}

func TestGetDefaultStoragePath(t *testing.T) {
	path := GetDefaultStoragePath()

	if path == "" {
		t.Error("GetDefaultStoragePath() returned empty string")
	}

	// Should end with sessions.json
	if filepath.Base(path) != "sessions.json" {
		t.Errorf("GetDefaultStoragePath() = %q, should end with sessions.json", path)
	}

	// Should contain tmux-flux directory
	if filepath.Base(filepath.Dir(path)) != "tmux-flux" {
		t.Errorf("GetDefaultStoragePath() = %q, should be in tmux-flux directory", path)
	}
}

func TestStorage_SaveEmptySlice(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	storage, err := NewStorage(storagePath)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}

	// Save empty slice
	err = storage.Save([]*Session{})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load should return empty slice
	loaded, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded) != 0 {
		t.Errorf("Load() returned %d sessions after saving empty slice", len(loaded))
	}
}

func TestStorage_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	storage, err := NewStorage(storagePath)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}

	// Save first set
	sessions1 := []*Session{
		{Name: "first", Group: "a", CreatedAt: time.Now()},
		{Name: "second", Group: "b", CreatedAt: time.Now()},
	}
	err = storage.Save(sessions1)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Save second set (different)
	sessions2 := []*Session{
		{Name: "third", Group: "c", CreatedAt: time.Now()},
	}
	err = storage.Save(sessions2)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load should return second set only
	loaded, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("Load() returned %d sessions, expected 1", len(loaded))
	}

	if loaded[0].Name != "third" {
		t.Errorf("Load() returned session %q, expected 'third'", loaded[0].Name)
	}
}
