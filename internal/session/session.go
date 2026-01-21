package session

import (
	"sort"
	"strings"
	"time"

	"github.com/sahilm/fuzzy"
	"github.com/zchase/tmux-flux/internal/tmux"
)

// Session represents a tmux session with group metadata
type Session struct {
	Name      string    `json:"name"`
	Group     string    `json:"group"`
	CreatedAt time.Time `json:"created_at"`
	// Runtime fields (not persisted)
	Windows   int   `json:"-"`
	Attached  bool  `json:"-"`
	Activity  int64 `json:"-"`
}

// Group represents a collection of sessions
type Group struct {
	Name      string
	Sessions  []*Session
	Collapsed bool
}

// Manager handles session and group operations
type Manager struct {
	sessions map[string]*Session // name -> session
	groups   map[string]*Group   // group name -> group
	storage  *Storage
}

// NewManager creates a new session manager
func NewManager(storagePath string) (*Manager, error) {
	storage, err := NewStorage(storagePath)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		sessions: make(map[string]*Session),
		groups:   make(map[string]*Group),
		storage:  storage,
	}

	// Load persisted sessions
	if err := m.loadFromStorage(); err != nil {
		return nil, err
	}

	return m, nil
}

// Refresh syncs with tmux and updates session states
func (m *Manager) Refresh() error {
	tmuxSessions, err := tmux.ListSessions()
	if err != nil {
		return err
	}

	// Build lookup map
	tmuxMap := make(map[string]tmux.Session)
	for _, ts := range tmuxSessions {
		tmuxMap[ts.Name] = ts
	}

	// Update existing sessions with tmux state
	for name, sess := range m.sessions {
		if ts, ok := tmuxMap[name]; ok {
			sess.Windows = ts.Windows
			sess.Attached = ts.Attached
			sess.Activity = ts.Activity
		} else {
			// Session no longer exists in tmux, remove from manager
			delete(m.sessions, name)
			if g, ok := m.groups[sess.Group]; ok {
				g.removeSession(name)
			}
		}
	}

	// Add new tmux sessions (ungrouped by default)
	for _, ts := range tmuxSessions {
		if _, exists := m.sessions[ts.Name]; !exists {
			sess := &Session{
				Name:      ts.Name,
				Group:     "", // Ungrouped
				CreatedAt: time.Unix(ts.Created, 0),
				Windows:   ts.Windows,
				Attached:  ts.Attached,
				Activity:  ts.Activity,
			}
			m.sessions[ts.Name] = sess
			m.addToGroup(sess)
		}
	}

	return m.save()
}

// GetGroups returns all groups sorted by name
func (m *Manager) GetGroups() []*Group {
	groups := make([]*Group, 0, len(m.groups))
	for _, g := range m.groups {
		groups = append(groups, g)
	}

	sort.Slice(groups, func(i, j int) bool {
		// Ungrouped always last
		if groups[i].Name == "" {
			return false
		}
		if groups[j].Name == "" {
			return true
		}
		return groups[i].Name < groups[j].Name
	})

	// Sort sessions within each group
	for _, g := range groups {
		sort.Slice(g.Sessions, func(i, j int) bool {
			// "0" always first
			if g.Sessions[i].Name == "0" {
				return true
			}
			if g.Sessions[j].Name == "0" {
				return false
			}
			return g.Sessions[i].Name < g.Sessions[j].Name
		})
	}

	return groups
}

// GetAllSessions returns all sessions flat
func (m *Manager) GetAllSessions() []*Session {
	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// SetSessionGroup moves a session to a group
func (m *Manager) SetSessionGroup(sessionName, groupName string) error {
	sess, ok := m.sessions[sessionName]
	if !ok {
		return nil
	}

	// Remove from old group
	if oldGroup, ok := m.groups[sess.Group]; ok {
		oldGroup.removeSession(sessionName)
	}

	// Update and add to new group
	sess.Group = groupName
	m.addToGroup(sess)

	return m.save()
}

// CreateSession creates a new tmux session
func (m *Manager) CreateSession(name, group, startDir string) error {
	if err := tmux.CreateSession(name, startDir); err != nil {
		return err
	}

	sess := &Session{
		Name:      name,
		Group:     group,
		CreatedAt: time.Now(),
	}
	m.sessions[name] = sess
	m.addToGroup(sess)

	return m.save()
}

// RenameSession renames a session
func (m *Manager) RenameSession(oldName, newName string) error {
	if err := tmux.RenameSession(oldName, newName); err != nil {
		return err
	}

	if sess, ok := m.sessions[oldName]; ok {
		// Remove from group with old name
		if g, ok := m.groups[sess.Group]; ok {
			g.removeSession(oldName)
		}

		// Update session
		delete(m.sessions, oldName)
		sess.Name = newName
		m.sessions[newName] = sess

		// Re-add to group
		m.addToGroup(sess)
	}

	return m.save()
}

// DeleteSession kills and removes a session
func (m *Manager) DeleteSession(name string) error {
	if err := tmux.KillSession(name); err != nil {
		return err
	}

	if sess, ok := m.sessions[name]; ok {
		if g, ok := m.groups[sess.Group]; ok {
			g.removeSession(name)
		}
		delete(m.sessions, name)
	}

	return m.save()
}

// ToggleGroupCollapse toggles a group's collapsed state
func (m *Manager) ToggleGroupCollapse(groupName string) {
	if g, ok := m.groups[groupName]; ok {
		g.Collapsed = !g.Collapsed
	}
}

// FuzzySearch performs fuzzy search on session names
func (m *Manager) FuzzySearch(query string) []*Session {
	if query == "" {
		return m.GetAllSessions()
	}

	sessions := m.GetAllSessions()
	names := make([]string, len(sessions))
	for i, s := range sessions {
		names[i] = s.Name
	}

	matches := fuzzy.Find(query, names)
	results := make([]*Session, len(matches))
	for i, match := range matches {
		results[i] = sessions[match.Index]
	}

	return results
}

// Helper methods

func (m *Manager) addToGroup(sess *Session) {
	groupName := sess.Group
	g, ok := m.groups[groupName]
	if !ok {
		g = &Group{Name: groupName}
		m.groups[groupName] = g
	}
	g.Sessions = append(g.Sessions, sess)
}

func (g *Group) removeSession(name string) {
	for i, s := range g.Sessions {
		if s.Name == name {
			g.Sessions = append(g.Sessions[:i], g.Sessions[i+1:]...)
			return
		}
	}
}

func (m *Manager) loadFromStorage() error {
	sessions, err := m.storage.Load()
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		m.sessions[sess.Name] = sess
		m.addToGroup(sess)
	}

	return nil
}

func (m *Manager) save() error {
	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return m.storage.Save(sessions)
}

// Implement fuzzy.Source interface for sessions
type SessionSource []*Session

func (s SessionSource) String(i int) string {
	return s[i].Name
}

func (s SessionSource) Len() int {
	return len(s)
}

// FuzzyMatch performs fuzzy matching with match info
func FuzzyMatch(sessions []*Session, query string) []fuzzy.Match {
	if query == "" {
		return nil
	}

	source := SessionSource(sessions)
	return fuzzy.FindFrom(query, source)
}

// GroupDisplayName returns a display name for a group
func GroupDisplayName(name string) string {
	if name == "" {
		return "Ungrouped"
	}
	return name
}

// ParseGroupPath splits a group path (e.g., "work/projects")
func ParseGroupPath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
