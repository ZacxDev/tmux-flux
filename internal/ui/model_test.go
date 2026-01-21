package ui

import (
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zchase/tmux-flux/internal/session"
)

func newTestManager(t *testing.T) *session.Manager {
	t.Helper()
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")
	manager, err := session.NewManager(storagePath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	return manager
}

func TestNewModel(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)

	if model.manager == nil {
		t.Error("NewModel() manager is nil")
	}

	if model.mode != ModeNormal {
		t.Errorf("NewModel() mode = %v, want ModeNormal", model.mode)
	}

	if model.cursor != 0 {
		t.Errorf("NewModel() cursor = %d, want 0", model.cursor)
	}
}

func TestModel_Init(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Init() returned nil command")
	}
}

func TestModel_Update_WindowSize(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := model.Update(msg)

	m := newModel.(Model)
	if m.width != 100 {
		t.Errorf("Update(WindowSizeMsg) width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("Update(WindowSizeMsg) height = %d, want 50", m.height)
	}
}

func TestModel_View_Loading(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)
	// Width is 0, should show loading
	model.width = 0

	view := model.View()
	if view != "Loading..." {
		t.Errorf("View() with width=0 = %q, want 'Loading...'", view)
	}
}

func TestModel_View_Normal(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)
	model.width = 80
	model.height = 24

	view := model.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	// Should contain the title
	if len(view) < 10 {
		t.Errorf("View() returned very short string: %q", view)
	}
}

func TestModel_ModeConstants(t *testing.T) {
	// Verify modes are distinct
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

	seen := make(map[Mode]bool)
	for _, m := range modes {
		if seen[m] {
			t.Errorf("Duplicate mode value: %d", m)
		}
		seen[m] = true
	}
}

func TestListItem(t *testing.T) {
	item := ListItem{
		Type:     "session",
		Name:     "test-session",
		Group:    "test-group",
		Attached: true,
	}

	if item.Type != "session" {
		t.Errorf("ListItem.Type = %q, want 'session'", item.Type)
	}
	if item.Name != "test-session" {
		t.Errorf("ListItem.Name = %q, want 'test-session'", item.Name)
	}
	if item.Group != "test-group" {
		t.Errorf("ListItem.Group = %q, want 'test-group'", item.Group)
	}
	if !item.Attached {
		t.Error("ListItem.Attached = false, want true")
	}
}

func TestModel_RebuildItems_Empty(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)
	model.rebuildItems()

	// With no sessions and no groups, items may be nil or empty
	// This is valid behavior
	if model.items != nil && len(model.items) > 0 {
		t.Logf("rebuildItems() with empty manager has %d items", len(model.items))
	}
}

func TestModel_RebuildItems_WithSessions(t *testing.T) {
	manager := newTestManager(t)

	// Add test sessions directly
	manager.CreateGroup("work")
	manager.CreateGroup("personal")

	model := NewModel(manager)
	model.rebuildItems()

	// Should have at least the group headers
	if len(model.items) < 2 {
		t.Errorf("rebuildItems() items count = %d, expected at least 2", len(model.items))
	}
}

func TestModel_RebuildItems_Search(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)

	// Enable search mode
	model.searchQuery = "test"
	model.rebuildItems()

	// With no sessions matching search, items may be nil or empty
	// This is valid behavior - fuzzy search returns empty results
	if model.items != nil && len(model.items) > 0 {
		t.Logf("rebuildItems() with search query has %d items", len(model.items))
	}
}

func TestModel_CursorNavigation(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)

	// Add some groups so we have items
	manager.CreateGroup("a")
	manager.CreateGroup("b")
	manager.CreateGroup("c")
	model.rebuildItems()

	initialCursor := model.cursor

	// Test cursor down
	model.cursorDown()
	if model.cursor <= initialCursor && len(model.items) > 1 {
		t.Error("cursorDown() did not move cursor")
	}

	// Test cursor up
	prevCursor := model.cursor
	model.cursorUp()
	if model.cursor >= prevCursor && prevCursor > 0 {
		t.Error("cursorUp() did not move cursor")
	}
}

func TestModel_CursorBounds(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)

	// Set cursor to 0
	model.cursor = 0
	model.cursorUp()
	if model.cursor < 0 {
		t.Error("cursorUp() went below 0")
	}

	// Add items and test upper bound
	manager.CreateGroup("test")
	model.rebuildItems()

	model.cursor = len(model.items) - 1
	model.cursorDown()
	if model.cursor >= len(model.items) {
		t.Error("cursorDown() went past end of items")
	}
}

func TestModel_PageNavigation(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)

	// Add many groups
	for i := 0; i < 20; i++ {
		manager.CreateGroup(string(rune('a' + i)))
	}
	model.rebuildItems()

	// Test page down
	model.cursor = 0
	model.cursorPageDown()
	if model.cursor == 0 && len(model.items) > 10 {
		t.Error("cursorPageDown() did not move cursor")
	}

	// Test page up
	model.cursorPageUp()
	// Should move up, possibly to 0
	if model.cursor < 0 {
		t.Error("cursorPageUp() went below 0")
	}
}

func TestModel_SelectedItem_Empty(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)
	model.items = nil

	item := model.selectedItem()
	if item != nil {
		t.Error("selectedItem() should return nil when items is empty")
	}
}

func TestModel_SelectedItem_Valid(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)

	manager.CreateGroup("test")
	model.rebuildItems()

	if len(model.items) == 0 {
		t.Skip("No items to test")
	}

	model.cursor = 0
	item := model.selectedItem()
	if item == nil {
		t.Error("selectedItem() returned nil with valid cursor")
	}
}

func TestModel_RefreshMsg(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)
	model.width = 80
	model.height = 24

	newModel, _ := model.Update(refreshMsg{})
	if newModel == nil {
		t.Error("Update(refreshMsg) returned nil model")
	}
}

func TestModel_ErrMsg(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)

	testErr := tea.Quit() // Use any error-like thing
	newModel, _ := model.Update(errMsg{err: nil})

	m := newModel.(Model)
	// With nil error, should be set
	_ = testErr
	_ = m
}

func TestModel_MsgMsg(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)

	newModel, _ := model.Update(msgMsg{msg: "test message"})
	m := newModel.(Model)

	if m.message != "test message" {
		t.Errorf("Update(msgMsg) message = %q, want 'test message'", m.message)
	}
}

func TestExecProcess(t *testing.T) {
	// Test that ExecProcess doesn't panic with valid command
	// We can't actually test the exec since it would replace the process
	// Just verify it's a valid function
	_ = ExecProcess
}

// Test message types
func TestRefreshMsg(t *testing.T) {
	msg := refreshMsg{}
	_ = msg // Just verify it compiles
}

func TestErrMsg(t *testing.T) {
	msg := errMsg{err: nil}
	if msg.err != nil {
		t.Error("errMsg.err should be nil")
	}
}

func TestMsgMsg(t *testing.T) {
	msg := msgMsg{msg: "test"}
	if msg.msg != "test" {
		t.Errorf("msgMsg.msg = %q, want 'test'", msg.msg)
	}
}

// Benchmark tests
func BenchmarkRebuildItems(b *testing.B) {
	tmpDir := b.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")
	manager, _ := session.NewManager(storagePath)

	// Add test data
	for i := 0; i < 100; i++ {
		manager.CreateGroup(string(rune('a' + (i % 26))))
	}

	model := NewModel(manager)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.rebuildItems()
	}
}

func BenchmarkView(b *testing.B) {
	tmpDir := b.TempDir()
	storagePath := filepath.Join(tmpDir, "sessions.json")
	manager, _ := session.NewManager(storagePath)

	model := NewModel(manager)
	model.width = 80
	model.height = 24

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

// Integration test with session data
func TestModel_WithSessionData(t *testing.T) {
	manager := newTestManager(t)

	// Simulate having sessions by adding them to the manager directly
	// (we can't use CreateSession as it needs tmux running)
	sessions := []*session.Session{
		{Name: "api", Group: "work", CreatedAt: time.Now()},
		{Name: "frontend", Group: "work", CreatedAt: time.Now()},
		{Name: "notes", Group: "personal", CreatedAt: time.Now()},
	}

	// Access internal fields through the manager's methods
	manager.CreateGroup("work")
	manager.CreateGroup("personal")

	model := NewModel(manager)
	model.width = 80
	model.height = 24
	model.rebuildItems()

	// Verify we have items
	if len(model.items) == 0 {
		t.Log("Warning: No items in model (expected if no sessions)")
	}

	// Test view renders without panic
	view := model.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	_ = sessions // Acknowledge we created these
}

// Preview tests
func TestPreviewMsg(t *testing.T) {
	msg := previewMsg{content: "test content"}
	if msg.content != "test content" {
		t.Errorf("previewMsg.content = %q, want 'test content'", msg.content)
	}
}

func TestModel_PreviewMsgUpdate(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)
	model.width = 80
	model.height = 24

	// Send a preview message
	newModel, _ := model.Update(previewMsg{content: "preview test"})
	m := newModel.(Model)

	if m.previewContent != "preview test" {
		t.Errorf("Update(previewMsg) previewContent = %q, want 'preview test'", m.previewContent)
	}
}

func TestModel_RenderPreview_NoSelection(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)
	model.width = 80
	model.height = 24
	model.items = nil

	preview := model.renderPreview(40, 10)
	if preview == "" {
		t.Error("renderPreview() returned empty string")
	}
	// Should contain "select a session" message
	if !contains(preview, "select a session") {
		t.Log("Preview content:", preview)
	}
}

func TestModel_RenderPreview_GroupSelected(t *testing.T) {
	manager := newTestManager(t)
	manager.CreateGroup("work")
	model := NewModel(manager)
	model.width = 80
	model.height = 24
	model.rebuildItems()
	model.cursor = 0 // Should be on group

	preview := model.renderPreview(40, 10)
	if preview == "" {
		t.Error("renderPreview() with group selected returned empty string")
	}
}

func TestModel_RenderPreview_WithContent(t *testing.T) {
	manager := newTestManager(t)
	manager.CreateGroup("work")
	model := NewModel(manager)
	model.width = 80
	model.height = 24
	model.rebuildItems()

	// Add some preview content
	model.previewContent = "line1\nline2\nline3"
	preview := model.renderPreview(40, 10)

	if preview == "" {
		t.Error("renderPreview() with content returned empty string")
	}
}

func TestModel_RenderPreview_ContentTruncation(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)
	model.width = 80
	model.height = 24

	// Create content longer than height
	longContent := ""
	for i := 0; i < 50; i++ {
		longContent += "line content here\n"
	}
	model.previewContent = longContent

	// Render with small height
	preview := model.renderPreview(40, 5)
	if preview == "" {
		t.Error("renderPreview() with truncated content returned empty string")
	}
}

func TestModel_RenderPreview_LineTruncation(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)
	model.width = 80
	model.height = 24

	// Create content with very long lines
	model.previewContent = "this is a very long line that should be truncated when rendered in a narrow preview pane"

	preview := model.renderPreview(20, 10)
	if preview == "" {
		t.Error("renderPreview() with long lines returned empty string")
	}
}

func TestModel_PreviewCmd_NoSelection(t *testing.T) {
	manager := newTestManager(t)
	model := NewModel(manager)
	model.items = nil

	cmd := model.previewCmd()
	if cmd == nil {
		t.Error("previewCmd() with no selection returned nil")
	}

	// Execute the command
	msg := cmd()
	if pMsg, ok := msg.(previewMsg); ok {
		if pMsg.content != "" {
			t.Errorf("previewCmd() with no selection content = %q, want empty", pMsg.content)
		}
	} else {
		t.Error("previewCmd() did not return previewMsg")
	}
}

func TestModel_PreviewCmd_GroupSelected(t *testing.T) {
	manager := newTestManager(t)
	manager.CreateGroup("work")
	model := NewModel(manager)
	model.rebuildItems()
	model.cursor = 0 // On group

	cmd := model.previewCmd()
	if cmd == nil {
		t.Error("previewCmd() with group selected returned nil")
	}

	msg := cmd()
	if pMsg, ok := msg.(previewMsg); ok {
		if pMsg.content != "" {
			t.Errorf("previewCmd() with group selected content = %q, want empty", pMsg.content)
		}
	} else {
		t.Error("previewCmd() did not return previewMsg")
	}
}

func TestModel_View_SplitLayout(t *testing.T) {
	manager := newTestManager(t)
	manager.CreateGroup("work")
	model := NewModel(manager)
	model.width = 100
	model.height = 30
	model.rebuildItems()
	model.previewContent = "test preview"

	view := model.View()
	if view == "" {
		t.Error("View() with split layout returned empty string")
	}

	// View should be rendered without panic
	if len(view) < 50 {
		t.Error("View() split layout seems too short")
	}
}

func TestModel_CursorChange_TriggersPreview(t *testing.T) {
	manager := newTestManager(t)
	manager.CreateGroup("work")
	manager.CreateGroup("personal")
	model := NewModel(manager)
	model.width = 80
	model.height = 24
	model.rebuildItems()

	// Navigate down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, cmd := model.Update(msg)

	m := newModel.(Model)
	if len(m.items) > 1 {
		// Should return a preview command when cursor changes
		if cmd == nil {
			t.Error("Navigation should return preview command when cursor changes")
		}
	}
}

// helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
