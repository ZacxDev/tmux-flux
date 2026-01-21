package session

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	manager, err := NewManager(storagePath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.sessions == nil {
		t.Error("NewManager() sessions map is nil")
	}

	if manager.groups == nil {
		t.Error("NewManager() groups map is nil")
	}
}

func TestManager_CreateGroup(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	manager, err := NewManager(storagePath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create a group
	manager.CreateGroup("test-group")

	// Verify group exists
	groups := manager.GetGroups()
	found := false
	for _, g := range groups {
		if g.Name == "test-group" {
			found = true
			break
		}
	}

	if !found {
		t.Error("CreateGroup() did not create the group")
	}

	// Creating same group again should not duplicate
	manager.CreateGroup("test-group")
	count := 0
	for _, g := range manager.GetGroups() {
		if g.Name == "test-group" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("CreateGroup() created duplicate groups, count = %d", count)
	}
}

func TestManager_GetGroups_Sorting(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	manager, err := NewManager(storagePath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create groups in non-alphabetical order
	manager.CreateGroup("zebra")
	manager.CreateGroup("alpha")
	manager.CreateGroup("beta")

	groups := manager.GetGroups()

	// Should be sorted alphabetically
	if len(groups) < 3 {
		t.Fatalf("Expected at least 3 groups, got %d", len(groups))
	}

	// Find the named groups (ungrouped may also exist)
	var namedGroups []string
	for _, g := range groups {
		if g.Name != "" {
			namedGroups = append(namedGroups, g.Name)
		}
	}

	if len(namedGroups) != 3 {
		t.Fatalf("Expected 3 named groups, got %d", len(namedGroups))
	}

	if namedGroups[0] != "alpha" {
		t.Errorf("Expected first group 'alpha', got '%s'", namedGroups[0])
	}
	if namedGroups[1] != "beta" {
		t.Errorf("Expected second group 'beta', got '%s'", namedGroups[1])
	}
	if namedGroups[2] != "zebra" {
		t.Errorf("Expected third group 'zebra', got '%s'", namedGroups[2])
	}
}

func TestManager_ToggleGroupCollapse(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	manager, err := NewManager(storagePath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	manager.CreateGroup("test-group")

	// Initially not collapsed
	groups := manager.GetGroups()
	var group *Group
	for _, g := range groups {
		if g.Name == "test-group" {
			group = g
			break
		}
	}

	if group == nil {
		t.Fatal("Could not find test-group")
	}

	if group.Collapsed {
		t.Error("Group should not be collapsed initially")
	}

	// Toggle collapse
	manager.ToggleGroupCollapse("test-group")

	groups = manager.GetGroups()
	for _, g := range groups {
		if g.Name == "test-group" {
			if !g.Collapsed {
				t.Error("Group should be collapsed after toggle")
			}
			break
		}
	}

	// Toggle again
	manager.ToggleGroupCollapse("test-group")

	groups = manager.GetGroups()
	for _, g := range groups {
		if g.Name == "test-group" {
			if g.Collapsed {
				t.Error("Group should not be collapsed after second toggle")
			}
			break
		}
	}
}

func TestManager_GetAllSessions(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	manager, err := NewManager(storagePath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Initially empty
	sessions := manager.GetAllSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions initially, got %d", len(sessions))
	}
}

func TestManager_FuzzySearch(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	manager, err := NewManager(storagePath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Add some test sessions directly to the manager
	sessions := []*Session{
		{Name: "api-server", Group: "work", CreatedAt: time.Now()},
		{Name: "frontend-app", Group: "work", CreatedAt: time.Now()},
		{Name: "database", Group: "work", CreatedAt: time.Now()},
		{Name: "personal-notes", Group: "personal", CreatedAt: time.Now()},
	}

	for _, s := range sessions {
		manager.sessions[s.Name] = s
		manager.addToGroup(s)
	}

	tests := []struct {
		query    string
		expected int
	}{
		{"", 4},           // Empty query returns all
		{"api", 1},        // Matches api-server
		{"app", 1},        // Matches frontend-app
		{"a", 4},          // Matches all (all contain 'a')
		{"xyz", 0},        // No matches
		{"front", 1},      // Matches frontend-app
		{"personal", 1},   // Matches personal-notes
	}

	for _, tt := range tests {
		t.Run("query_"+tt.query, func(t *testing.T) {
			results := manager.FuzzySearch(tt.query)
			if len(results) != tt.expected {
				t.Errorf("FuzzySearch(%q) returned %d results, want %d", tt.query, len(results), tt.expected)
			}
		})
	}
}

func TestManager_SessionSorting_ZeroFirst(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")

	manager, err := NewManager(storagePath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Add sessions with "0" not first
	sessions := []*Session{
		{Name: "zebra", Group: "", CreatedAt: time.Now()},
		{Name: "alpha", Group: "", CreatedAt: time.Now()},
		{Name: "0", Group: "", CreatedAt: time.Now()},
		{Name: "beta", Group: "", CreatedAt: time.Now()},
	}

	for _, s := range sessions {
		manager.sessions[s.Name] = s
		manager.addToGroup(s)
	}

	groups := manager.GetGroups()

	// Find ungrouped
	var ungrouped *Group
	for _, g := range groups {
		if g.Name == "" {
			ungrouped = g
			break
		}
	}

	if ungrouped == nil {
		t.Fatal("Could not find ungrouped group")
	}

	if len(ungrouped.Sessions) != 4 {
		t.Fatalf("Expected 4 sessions, got %d", len(ungrouped.Sessions))
	}

	// "0" should be first
	if ungrouped.Sessions[0].Name != "0" {
		t.Errorf("Expected '0' to be first, got '%s'", ungrouped.Sessions[0].Name)
	}

	// Rest should be alphabetical
	if ungrouped.Sessions[1].Name != "alpha" {
		t.Errorf("Expected 'alpha' to be second, got '%s'", ungrouped.Sessions[1].Name)
	}
}

func TestGroupDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"", "Ungrouped"},
		{"work", "work"},
		{"personal", "personal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GroupDisplayName(tt.name)
			if got != tt.expected {
				t.Errorf("GroupDisplayName(%q) = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestParseGroupPath(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"", nil},
		{"work", []string{"work"}},
		{"work/projects", []string{"work", "projects"}},
		{"a/b/c", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := ParseGroupPath(tt.path)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("ParseGroupPath(%q) = %v, want nil", tt.path, got)
				}
				return
			}
			if len(got) != len(tt.expected) {
				t.Errorf("ParseGroupPath(%q) = %v, want %v", tt.path, got, tt.expected)
				return
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("ParseGroupPath(%q)[%d] = %q, want %q", tt.path, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestFuzzyMatch(t *testing.T) {
	sessions := []*Session{
		{Name: "api-server"},
		{Name: "frontend"},
		{Name: "database"},
	}

	// Empty query returns nil
	matches := FuzzyMatch(sessions, "")
	if matches != nil {
		t.Errorf("FuzzyMatch with empty query should return nil, got %v", matches)
	}

	// Query with matches
	matches = FuzzyMatch(sessions, "api")
	if len(matches) != 1 {
		t.Errorf("FuzzyMatch('api') expected 1 match, got %d", len(matches))
	}

	// Query with no matches
	matches = FuzzyMatch(sessions, "xyz")
	if len(matches) != 0 {
		t.Errorf("FuzzyMatch('xyz') expected 0 matches, got %d", len(matches))
	}
}

func TestSessionSource(t *testing.T) {
	sessions := SessionSource([]*Session{
		{Name: "first"},
		{Name: "second"},
		{Name: "third"},
	})

	if sessions.Len() != 3 {
		t.Errorf("SessionSource.Len() = %d, want 3", sessions.Len())
	}

	if sessions.String(0) != "first" {
		t.Errorf("SessionSource.String(0) = %q, want 'first'", sessions.String(0))
	}

	if sessions.String(1) != "second" {
		t.Errorf("SessionSource.String(1) = %q, want 'second'", sessions.String(1))
	}
}
