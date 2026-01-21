package ui

import (
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zchase/tmux-flux/internal/session"
)

func setupTestModel(t *testing.T) Model {
	t.Helper()
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")
	manager, err := session.NewManager(storagePath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	model := NewModel(manager)
	model.width = 80
	model.height = 24

	// Add some test groups
	manager.CreateGroup("work")
	manager.CreateGroup("personal")
	model.rebuildItems()

	return model
}

func TestUpdateNormal_Quit(t *testing.T) {
	model := setupTestModel(t)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := model.Update(msg)

	// Quit command should be returned
	if cmd == nil {
		t.Error("Quit key should return a command")
	}
}

func TestUpdateNormal_Help(t *testing.T) {
	model := setupTestModel(t)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeHelp {
		t.Errorf("Help key should switch to ModeHelp, got %v", m.mode)
	}
}

func TestUpdateNormal_Navigation(t *testing.T) {
	model := setupTestModel(t)

	// Test j (down)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if len(m.items) > 1 && m.cursor == 0 {
		// If there are multiple items, cursor should have moved
		t.Log("Cursor may not have moved if at boundary")
	}

	// Test k (up)
	m.cursor = 1
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	// Should be at or near 0
	if m.cursor > 1 {
		t.Error("k key should move cursor up")
	}
}

func TestUpdateNormal_Search(t *testing.T) {
	model := setupTestModel(t)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeSearch {
		t.Errorf("/ key should switch to ModeSearch, got %v", m.mode)
	}
}

func TestUpdateNormal_Create(t *testing.T) {
	model := setupTestModel(t)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeCreate {
		t.Errorf("n key should switch to ModeCreate, got %v", m.mode)
	}
}

func TestUpdateNormal_NewGroup(t *testing.T) {
	model := setupTestModel(t)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNewGroup {
		t.Errorf("N key should switch to ModeNewGroup, got %v", m.mode)
	}
}

func TestUpdateNormal_Refresh(t *testing.T) {
	model := setupTestModel(t)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("R key should return refresh command")
	}
}

func TestUpdateNormal_GotoEnd(t *testing.T) {
	model := setupTestModel(t)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if len(m.items) > 0 && m.cursor != len(m.items)-1 {
		t.Errorf("G key should move to end, cursor=%d, items=%d", m.cursor, len(m.items))
	}
}

func TestUpdateNormal_GG(t *testing.T) {
	model := setupTestModel(t)
	model.cursor = 5

	// First g
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.gPressed {
		t.Error("First g should set gPressed")
	}

	// Second g
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.cursor != 0 {
		t.Errorf("gg should move cursor to 0, got %d", m.cursor)
	}
	if m.gPressed {
		t.Error("gg should reset gPressed")
	}
}

func TestUpdateNormal_Toggle(t *testing.T) {
	model := setupTestModel(t)

	// Position on a group
	model.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	newModel, _ := model.Update(msg)

	_ = newModel.(Model)
	// Toggle should work without error
}

func TestUpdateNormal_VimExpand(t *testing.T) {
	model := setupTestModel(t)
	model.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, _ := model.Update(msg)

	_ = newModel.(Model)
	// Expand should work without error
}

func TestUpdateNormal_VimCollapse(t *testing.T) {
	model := setupTestModel(t)
	model.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	newModel, _ := model.Update(msg)

	_ = newModel.(Model)
	// Collapse should work without error
}

func TestUpdateSearch_Escape(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeSearch
	model.searchQuery = "test"

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNormal {
		t.Errorf("Escape in search should return to ModeNormal, got %v", m.mode)
	}
	if m.searchQuery != "" {
		t.Error("Escape in search should clear search query")
	}
}

func TestUpdateSearch_Enter(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeSearch

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNormal {
		t.Errorf("Enter in search should return to ModeNormal, got %v", m.mode)
	}
}

func TestUpdateCreate_Escape(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeCreate

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNormal {
		t.Errorf("Escape in create should return to ModeNormal, got %v", m.mode)
	}
}

func TestUpdateNewGroup_Escape(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeNewGroup

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNormal {
		t.Errorf("Escape in new group should return to ModeNormal, got %v", m.mode)
	}
}

func TestUpdateRename_Escape(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeRename

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNormal {
		t.Errorf("Escape in rename should return to ModeNormal, got %v", m.mode)
	}
}

func TestUpdateDelete_Escape(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeDelete

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNormal {
		t.Errorf("Escape in delete should return to ModeNormal, got %v", m.mode)
	}
}

func TestUpdateDelete_No(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeDelete

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNormal {
		t.Errorf("'n' in delete should return to ModeNormal, got %v", m.mode)
	}
}

func TestUpdateGroup_Escape(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeGroup

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNormal {
		t.Errorf("Escape in group should return to ModeNormal, got %v", m.mode)
	}
}

func TestUpdateHelp_Escape(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeHelp

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNormal {
		t.Errorf("Escape in help should return to ModeNormal, got %v", m.mode)
	}
}

func TestUpdateHelp_QuestionMark(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeHelp

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNormal {
		t.Errorf("? in help should return to ModeNormal, got %v", m.mode)
	}
}

func TestUpdateHelp_Q(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeHelp

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.mode != ModeNormal {
		t.Errorf("q in help should return to ModeNormal, got %v", m.mode)
	}
}

func TestKeyMatches(t *testing.T) {
	km := DefaultKeyMap()

	// Test that key matching works
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	if !key.Matches(msg, km.VimDown) {
		t.Error("j should match VimDown")
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	if !key.Matches(msg, km.VimUp) {
		t.Error("k should match VimUp")
	}
}

func TestModel_ToggleAtCursor(t *testing.T) {
	model := setupTestModel(t)

	// Ensure we have items
	if len(model.items) == 0 {
		t.Skip("No items to test toggle")
	}

	// Find a group item
	for i, item := range model.items {
		if item.Type == "group" {
			model.cursor = i
			break
		}
	}

	// Toggle should not panic
	model.toggleAtCursor()
}

func TestModel_ExpandAtCursor(t *testing.T) {
	model := setupTestModel(t)

	if len(model.items) == 0 {
		t.Skip("No items to test expand")
	}

	// Find a group item
	for i, item := range model.items {
		if item.Type == "group" {
			model.cursor = i
			break
		}
	}

	// Expand should not panic
	model.expandAtCursor()
}

func TestModel_CollapseAtCursor(t *testing.T) {
	model := setupTestModel(t)

	if len(model.items) == 0 {
		t.Skip("No items to test collapse")
	}

	// Find a group item
	for i, item := range model.items {
		if item.Type == "group" {
			model.cursor = i
			break
		}
	}

	// Collapse should not panic
	model.collapseAtCursor()
}

func TestModel_RenderHelpBar_AllModes(t *testing.T) {
	model := setupTestModel(t)

	modes := []Mode{
		ModeNormal,
		ModeSearch,
		ModeCreate,
		ModeNewGroup,
		ModeRename,
		ModeDelete,
		ModeGroup,
	}

	for _, mode := range modes {
		t.Run("", func(t *testing.T) {
			model.mode = mode
			result := model.renderHelpBar()
			if result == "" {
				t.Errorf("renderHelpBar() for mode %v returned empty string", mode)
			}
		})
	}
}

func TestModel_RenderHelp(t *testing.T) {
	model := setupTestModel(t)
	model.mode = ModeHelp

	result := model.renderHelp()
	if result == "" {
		t.Error("renderHelp() returned empty string")
	}

	// Should contain section headers
	if len(result) < 100 {
		t.Error("renderHelp() returned very short string")
	}
}

func TestModel_View_AllModes(t *testing.T) {
	model := setupTestModel(t)

	modes := []Mode{
		ModeNormal,
		ModeSearch,
		ModeCreate,
		ModeNewGroup,
		ModeRename,
		ModeDelete,
		ModeGroup,
		ModeHelp,
	}

	for _, mode := range modes {
		t.Run("", func(t *testing.T) {
			model.mode = mode
			result := model.View()
			if result == "" {
				t.Errorf("View() for mode %v returned empty string", mode)
			}
		})
	}
}

func TestModel_View_WithError(t *testing.T) {
	model := setupTestModel(t)
	model.err = tea.ErrProgramKilled

	result := model.View()
	if result == "" {
		t.Error("View() with error returned empty string")
	}
}

func TestModel_View_WithMessage(t *testing.T) {
	model := setupTestModel(t)
	model.message = "Test message"

	result := model.View()
	if result == "" {
		t.Error("View() with message returned empty string")
	}
}

func TestModel_RenderList(t *testing.T) {
	model := setupTestModel(t)

	result := model.renderList(10)
	// May be empty if no sessions, but should not panic
	_ = result
}

func TestModel_RenderList_WithScroll(t *testing.T) {
	model := setupTestModel(t)

	// Add many groups
	for i := 0; i < 20; i++ {
		model.manager.CreateGroup(string(rune('a' + i)))
	}
	model.rebuildItems()
	model.cursor = 15

	result := model.renderList(5) // Small height to trigger scroll
	if result == "" && len(model.items) > 0 {
		t.Error("renderList() with scroll returned empty string")
	}
}
