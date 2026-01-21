package session

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Storage handles session persistence
type Storage struct {
	path string
}

// NewStorage creates a new storage instance
func NewStorage(path string) (*Storage, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &Storage{path: path}, nil
}

// Load reads sessions from storage
func (s *Storage) Load() ([]*Session, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Session{}, nil
		}
		return nil, err
	}

	var sessions []*Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// Save writes sessions to storage
func (s *Storage) Save(sessions []*Session) error {
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

// GetDefaultStoragePath returns the default storage path
func GetDefaultStoragePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".config", "tmux-flux", "sessions.json")
}
