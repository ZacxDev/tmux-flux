package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zchase/tmux-flux/internal/session"
	"github.com/zchase/tmux-flux/internal/tmux"
)

// Mode represents the current UI mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeCreate
	ModeNewGroup
	ModeRename
	ModeDelete
	ModeGroup
	ModeHelp
)

// ListItem represents an item in the session list
type ListItem struct {
	Type     string // "group" or "session"
	Name     string
	Group    string
	Session  *session.Session
	Attached bool
}

// Model is the main application model
type Model struct {
	manager        *session.Manager
	keys           KeyMap
	width          int
	height         int
	mode           Mode
	cursor         int
	items          []ListItem
	searchInput    textinput.Model
	textInput      textinput.Model
	searchQuery    string
	message        string
	err            error
	gPressed       bool // For gg navigation
	insideTmux     bool
	previewContent string
	previewSession string // Track which session the preview is for
}

// NewModel creates a new application model
func NewModel(manager *session.Manager) Model {
	si := textinput.New()
	si.Placeholder = "Search sessions..."
	si.Prompt = "/ "
	si.PromptStyle = SearchStyle
	si.TextStyle = SearchInputStyle

	ti := textinput.New()
	ti.Prompt = "> "
	ti.PromptStyle = DialogTitleStyle

	return Model{
		manager:     manager,
		keys:        DefaultKeyMap(),
		mode:        ModeNormal,
		searchInput: si,
		textInput:   ti,
		insideTmux:  tmux.IsInsideTmux(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshCmd(),
		tea.SetWindowTitle("tmux-flux"),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle based on mode
		switch m.mode {
		case ModeSearch:
			return m.updateSearch(msg)
		case ModeCreate:
			return m.updateCreate(msg)
		case ModeNewGroup:
			return m.updateNewGroup(msg)
		case ModeRename:
			return m.updateRename(msg)
		case ModeDelete:
			return m.updateDelete(msg)
		case ModeGroup:
			return m.updateGroup(msg)
		case ModeHelp:
			if key.Matches(msg, m.keys.Escape) || key.Matches(msg, m.keys.Help) || msg.String() == "q" {
				m.mode = ModeNormal
			}
			return m, nil
		default:
			return m.updateNormal(msg)
		}

	case refreshMsg:
		m.rebuildItems()
		return m, m.previewCmd()

	case errMsg:
		m.err = msg.err
		m.message = ""
		return m, nil

	case msgMsg:
		m.message = msg.msg
		m.err = nil
		return m, nil

	case previewMsg:
		// Only update if this preview is for the currently selected session
		// This prevents race conditions when navigating quickly
		currentSession := ""
		if item := m.selectedItem(); item != nil && item.Type == "session" {
			currentSession = item.Name
		}
		if msg.session == currentSession || msg.session == "" {
			m.previewContent = msg.content
			m.previewSession = msg.session
		}
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	oldCursor := m.cursor

	switch {
	// Quit
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	// Help
	case key.Matches(msg, m.keys.Help):
		m.mode = ModeHelp
		return m, nil

	// Navigation
	case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.VimUp):
		m.cursorUp()
	case key.Matches(msg, m.keys.Down), key.Matches(msg, m.keys.VimDown):
		m.cursorDown()
	case key.Matches(msg, m.keys.PageUp):
		m.cursorPageUp()
	case key.Matches(msg, m.keys.PageDown):
		m.cursorPageDown()

	// Vim: gg to go to top
	case msg.String() == "g":
		if m.gPressed {
			m.cursor = 0
			m.gPressed = false
		} else {
			m.gPressed = true
			return m, nil
		}

	// Vim: G to go to bottom
	case key.Matches(msg, m.keys.GotoEnd):
		m.cursor = len(m.items) - 1
		m.gPressed = false

	// Expand/Collapse with h/l
	case key.Matches(msg, m.keys.VimLeft):
		m.collapseAtCursor()
	case key.Matches(msg, m.keys.VimRight):
		m.expandAtCursor()

	// Toggle expand/collapse with tab or space
	case key.Matches(msg, m.keys.Toggle), key.Matches(msg, m.keys.Select):
		m.toggleAtCursor()

	// Enter to attach/switch to session
	case key.Matches(msg, m.keys.Enter):
		return m.attachSelected()

	// Search
	case key.Matches(msg, m.keys.Search):
		m.mode = ModeSearch
		m.searchInput.SetValue("")
		m.searchQuery = ""
		m.searchInput.Focus()
		m.rebuildItems()
		return m, textinput.Blink

	// Create
	case key.Matches(msg, m.keys.Create):
		m.mode = ModeCreate
		m.textInput.SetValue("")
		m.textInput.Placeholder = "Session name"
		m.textInput.Focus()
		return m, textinput.Blink

	// New Group
	case key.Matches(msg, m.keys.NewGroup):
		m.mode = ModeNewGroup
		m.textInput.SetValue("")
		m.textInput.Placeholder = "Group name"
		m.textInput.Focus()
		return m, textinput.Blink

	// Rename
	case key.Matches(msg, m.keys.Rename):
		if item := m.selectedItem(); item != nil && item.Type == "session" {
			m.mode = ModeRename
			m.textInput.SetValue(item.Name)
			m.textInput.Placeholder = "New name"
			m.textInput.Focus()
			return m, textinput.Blink
		}

	// Delete
	case key.Matches(msg, m.keys.Delete):
		if item := m.selectedItem(); item != nil && item.Type == "session" {
			m.mode = ModeDelete
		}

	// Move to group
	case key.Matches(msg, m.keys.Group):
		if item := m.selectedItem(); item != nil && item.Type == "session" {
			m.mode = ModeGroup
			m.textInput.SetValue(item.Group)
			m.textInput.Placeholder = "Group name (empty for ungrouped)"
			m.textInput.Focus()
			return m, textinput.Blink
		}

	// Refresh
	case key.Matches(msg, m.keys.Refresh):
		return m, m.refreshCmd()

	default:
		m.gPressed = false
	}

	// Update preview if the selected session changed
	if m.cursor != oldCursor {
		newSession := ""
		if item := m.selectedItem(); item != nil && item.Type == "session" {
			newSession = item.Name
		}
		// Only fetch new preview if we're on a different session
		if newSession != m.previewSession {
			return m, m.previewCmd()
		}
	}

	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = ModeNormal
		m.searchQuery = ""
		m.rebuildItems()
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		m.mode = ModeNormal
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.searchQuery = m.searchInput.Value()
	m.rebuildItems()

	return m, cmd
}

func (m Model) updateCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = ModeNormal
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		name := strings.TrimSpace(m.textInput.Value())
		if name != "" {
			group := ""
			if item := m.selectedItem(); item != nil {
				if item.Type == "group" {
					group = item.Name
				} else {
					group = item.Group
				}
			}
			if err := m.manager.CreateSession(name, group, ""); err != nil {
				m.err = err
			} else {
				m.message = fmt.Sprintf("Created session: %s", name)
			}
		}
		m.mode = ModeNormal
		return m, m.refreshCmd()
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) updateNewGroup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = ModeNormal
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		name := strings.TrimSpace(m.textInput.Value())
		if name != "" {
			m.manager.CreateGroup(name)
			m.message = fmt.Sprintf("Created group: %s", name)
		}
		m.mode = ModeNormal
		return m, m.refreshCmd()
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) updateRename(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = ModeNormal
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		item := m.selectedItem()
		if item == nil || item.Type != "session" {
			m.mode = ModeNormal
			return m, nil
		}
		newName := strings.TrimSpace(m.textInput.Value())
		if newName != "" && newName != item.Name {
			if err := m.manager.RenameSession(item.Name, newName); err != nil {
				m.err = err
			} else {
				m.message = fmt.Sprintf("Renamed: %s → %s", item.Name, newName)
			}
		}
		m.mode = ModeNormal
		return m, m.refreshCmd()
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) updateDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape), msg.String() == "n":
		m.mode = ModeNormal
		return m, nil
	case msg.String() == "y":
		item := m.selectedItem()
		if item != nil && item.Type == "session" {
			if err := m.manager.DeleteSession(item.Name); err != nil {
				m.err = err
			} else {
				m.message = fmt.Sprintf("Deleted session: %s", item.Name)
			}
		}
		m.mode = ModeNormal
		return m, m.refreshCmd()
	}
	return m, nil
}

func (m Model) updateGroup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = ModeNormal
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		item := m.selectedItem()
		if item == nil || item.Type != "session" {
			m.mode = ModeNormal
			return m, nil
		}
		newGroup := strings.TrimSpace(m.textInput.Value())
		if newGroup != item.Group {
			if err := m.manager.SetSessionGroup(item.Name, newGroup); err != nil {
				m.err = err
			} else {
				groupDisplay := newGroup
				if groupDisplay == "" {
					groupDisplay = "Ungrouped"
				}
				m.message = fmt.Sprintf("Moved %s to: %s", item.Name, groupDisplay)
			}
		}
		m.mode = ModeNormal
		return m, m.refreshCmd()
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) attachSelected() (tea.Model, tea.Cmd) {
	item := m.selectedItem()
	if item == nil || item.Type != "session" {
		return m, nil
	}

	if m.insideTmux {
		// Switch client to session
		if err := tmux.SwitchClient(item.Name); err != nil {
			m.err = err
			return m, nil
		}
		return m, tea.Quit
	}

	// Not inside tmux, use exec to attach
	cmd := tmux.AttachSession(item.Name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	})
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Title
	title := TitleStyle.Render("tmux-flux")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Help mode
	if m.mode == ModeHelp {
		b.WriteString(m.renderHelp())
		return BaseStyle.Width(m.width).Height(m.height).Render(b.String())
	}

	// Search bar
	if m.mode == ModeSearch {
		b.WriteString(m.searchInput.View())
		b.WriteString("\n\n")
	} else if m.searchQuery != "" {
		b.WriteString(SearchStyle.Render("/ " + m.searchQuery))
		b.WriteString("\n\n")
	}

	// Calculate layout dimensions
	// Left panel: session list, Right panel: preview
	listWidth := m.width / 3
	if listWidth < 25 {
		listWidth = 25
	}
	previewWidth := m.width - listWidth - 3 // 3 for border/padding

	listHeight := m.height - 8
	if m.mode == ModeSearch {
		listHeight -= 2
	}

	// Render session list
	listContent := m.renderList(listHeight)

	// Render preview pane
	previewContent := m.renderPreview(previewWidth, listHeight)

	// Join left and right panels
	leftPanel := lipgloss.NewStyle().Width(listWidth).Height(listHeight).Render(listContent)
	rightPanel := PreviewBorderStyle.Width(previewWidth).Height(listHeight).Render(previewContent)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel))
	b.WriteString("\n")

	// Dialogs
	switch m.mode {
	case ModeCreate:
		b.WriteString("\n")
		b.WriteString(DialogTitleStyle.Render("New session: "))
		b.WriteString(m.textInput.View())
	case ModeNewGroup:
		b.WriteString("\n")
		b.WriteString(DialogTitleStyle.Render("New group: "))
		b.WriteString(m.textInput.View())
	case ModeRename:
		b.WriteString("\n")
		b.WriteString(DialogTitleStyle.Render("Rename: "))
		b.WriteString(m.textInput.View())
	case ModeDelete:
		item := m.selectedItem()
		if item != nil {
			b.WriteString("\n")
			b.WriteString(WarningStyle.Render(fmt.Sprintf("Delete session '%s'? (y/n)", item.Name)))
		}
	case ModeGroup:
		b.WriteString("\n")
		b.WriteString(DialogTitleStyle.Render("Move to group: "))
		b.WriteString(m.textInput.View())
	}

	// Error/message
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(ErrorStyle.Render("Error: " + m.err.Error()))
	} else if m.message != "" {
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render(m.message))
	}

	// Help bar
	b.WriteString("\n")
	b.WriteString(m.renderHelpBar())

	return BaseStyle.Width(m.width).Height(m.height).Render(b.String())
}

func (m Model) renderList(maxHeight int) string {
	if len(m.items) == 0 {
		return HelpStyle.Render("No sessions. Press 'n' to create one.")
	}

	var b strings.Builder
	start := 0
	end := len(m.items)

	// Scroll handling
	if maxHeight > 0 && len(m.items) > maxHeight {
		if m.cursor >= maxHeight {
			start = m.cursor - maxHeight + 1
		}
		end = start + maxHeight
		if end > len(m.items) {
			end = len(m.items)
		}
	}

	for i := start; i < end; i++ {
		item := m.items[i]
		selected := i == m.cursor

		var line string
		if item.Type == "group" {
			group := m.manager.GetGroups()[0]
			for _, g := range m.manager.GetGroups() {
				if g.Name == item.Name {
					group = g
					break
				}
			}
			line = RenderGroupHeader(session.GroupDisplayName(item.Name), group.Collapsed, selected)
		} else {
			line = RenderSession(item.Name, item.Attached, selected)
		}

		b.WriteString(line)
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) renderPreview(width, height int) string {
	item := m.selectedItem()

	// Title
	var title string
	if item != nil && item.Type == "session" {
		title = PreviewTitleStyle.Render("Preview: " + item.Name)
	} else {
		title = PreviewTitleStyle.Render("Preview")
	}

	// Content
	content := m.previewContent
	if content == "" {
		if item == nil || item.Type != "session" {
			content = HelpStyle.Render("(select a session to preview)")
		} else {
			content = HelpStyle.Render("(loading...)")
		}
	} else {
		// Truncate lines to fit width and limit height
		lines := strings.Split(content, "\n")
		maxLines := height - 2 // Account for title
		if maxLines < 1 {
			maxLines = 1
		}
		if len(lines) > maxLines {
			lines = lines[:maxLines]
		}

		// Truncate each line to fit width
		for i, line := range lines {
			if len(line) > width-2 {
				lines[i] = line[:width-2]
			}
		}
		content = PreviewStyle.Render(strings.Join(lines, "\n"))
	}

	return title + "\n" + content
}

func (m Model) renderHelpBar() string {
	var items []HelpItem
	switch m.mode {
	case ModeSearch:
		items = []HelpItem{
			{"enter", "apply"},
			{"esc", "cancel"},
		}
	case ModeCreate, ModeNewGroup, ModeRename, ModeGroup:
		items = []HelpItem{
			{"enter", "confirm"},
			{"esc", "cancel"},
		}
	case ModeDelete:
		items = []HelpItem{
			{"y", "yes"},
			{"n", "no"},
		}
	default:
		items = []HelpItem{
			{"j/k", "navigate"},
			{"enter", "attach"},
			{"/", "search"},
			{"n", "new"},
			{"r", "rename"},
			{"d", "delete"},
			{"?", "help"},
			{"q", "quit"},
		}
	}
	return StatusBarStyle.Render(RenderHelp(items))
}

func (m Model) renderHelp() string {
	var b strings.Builder
	b.WriteString(DialogTitleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	sections := []struct {
		title string
		keys  []HelpItem
	}{
		{
			"Navigation",
			[]HelpItem{
				{"j / ↓", "Move down"},
				{"k / ↑", "Move up"},
				{"gg", "Go to top"},
				{"G", "Go to bottom"},
				{"Ctrl+d", "Page down"},
				{"Ctrl+u", "Page up"},
			},
		},
		{
			"Groups",
			[]HelpItem{
				{"N", "Create new group"},
				{"h / ←", "Collapse group"},
				{"l / →", "Expand group"},
				{"Tab/Space", "Toggle group"},
			},
		},
		{
			"Sessions",
			[]HelpItem{
				{"Enter", "Attach to session"},
				{"n / c", "Create new session"},
				{"r", "Rename session"},
				{"d / x", "Delete session"},
				{"m", "Move to group"},
			},
		},
		{
			"Other",
			[]HelpItem{
				{"/", "Search sessions"},
				{"R", "Refresh list"},
				{"?", "Toggle help"},
				{"q", "Quit"},
			},
		},
	}

	for _, section := range sections {
		b.WriteString(GroupStyle.Render(section.title))
		b.WriteString("\n")
		for _, k := range section.keys {
			b.WriteString("  ")
			b.WriteString(HelpKeyStyle.Render(fmt.Sprintf("%-10s", k.Key)))
			b.WriteString(" ")
			b.WriteString(HelpStyle.Render(k.Desc))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(HelpStyle.Render("Press ? or Esc to close"))

	return b.String()
}

// Helper methods

func (m *Model) rebuildItems() {
	m.items = nil

	if m.searchQuery != "" {
		// Search mode - flat list of matching sessions
		sessions := m.manager.FuzzySearch(m.searchQuery)
		for _, s := range sessions {
			m.items = append(m.items, ListItem{
				Type:     "session",
				Name:     s.Name,
				Group:    s.Group,
				Session:  s,
				Attached: s.Attached,
			})
		}
	} else {
		// Normal mode - grouped
		groups := m.manager.GetGroups()
		for _, g := range groups {
			// Add group header
			m.items = append(m.items, ListItem{
				Type:  "group",
				Name:  g.Name,
				Group: g.Name,
			})

			// Add sessions if not collapsed
			if !g.Collapsed {
				for _, s := range g.Sessions {
					m.items = append(m.items, ListItem{
						Type:     "session",
						Name:     s.Name,
						Group:    g.Name,
						Session:  s,
						Attached: s.Attached,
					})
				}
			}
		}
	}

	// Ensure cursor is valid
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) selectedItem() *ListItem {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		return &m.items[m.cursor]
	}
	return nil
}

func (m *Model) cursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *Model) cursorDown() {
	if m.cursor < len(m.items)-1 {
		m.cursor++
	}
}

func (m *Model) cursorPageUp() {
	m.cursor -= 10
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) cursorPageDown() {
	m.cursor += 10
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
}

func (m *Model) toggleAtCursor() {
	item := m.selectedItem()
	if item == nil {
		return
	}

	if item.Type == "group" {
		m.manager.ToggleGroupCollapse(item.Name)
		m.rebuildItems()
	}
}

func (m *Model) expandAtCursor() {
	item := m.selectedItem()
	if item == nil {
		return
	}

	if item.Type == "group" {
		for _, g := range m.manager.GetGroups() {
			if g.Name == item.Name && g.Collapsed {
				m.manager.ToggleGroupCollapse(item.Name)
				m.rebuildItems()
				break
			}
		}
	}
}

func (m *Model) collapseAtCursor() {
	item := m.selectedItem()
	if item == nil {
		return
	}

	if item.Type == "group" {
		for _, g := range m.manager.GetGroups() {
			if g.Name == item.Name && !g.Collapsed {
				m.manager.ToggleGroupCollapse(item.Name)
				m.rebuildItems()
				break
			}
		}
	} else if item.Type == "session" {
		// If on a session, collapse its parent group and move cursor to group
		for i, it := range m.items {
			if it.Type == "group" && it.Name == item.Group {
				m.manager.ToggleGroupCollapse(item.Group)
				m.cursor = i
				m.rebuildItems()
				break
			}
		}
	}
}

func (m Model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		if err := m.manager.Refresh(); err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	}
}

func (m Model) previewCmd() tea.Cmd {
	item := m.selectedItem()
	if item == nil || item.Type != "session" {
		return func() tea.Msg {
			return previewMsg{session: "", content: ""}
		}
	}
	sessionName := item.Name
	return func() tea.Msg {
		content, err := tmux.CapturePane(sessionName)
		if err != nil {
			return previewMsg{session: sessionName, content: "(unable to capture preview)"}
		}
		return previewMsg{session: sessionName, content: content}
	}
}

// Messages
type refreshMsg struct{}
type errMsg struct{ err error }
type msgMsg struct{ msg string }
type previewMsg struct {
	session string // Which session this preview is for
	content string
}

// ExecProcess is a helper for executing processes with proper cleanup
func ExecProcess(name string, args ...string) error {
	binary, err := exec.LookPath(name)
	if err != nil {
		return err
	}
	return syscall.Exec(binary, append([]string{name}, args...), os.Environ())
}
