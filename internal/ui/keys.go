package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	// Vim navigation
	VimUp    key.Binding
	VimDown  key.Binding
	VimLeft  key.Binding
	VimRight key.Binding
	GotoTop  key.Binding
	GotoEnd  key.Binding

	// Actions
	Enter    key.Binding
	Select   key.Binding
	Toggle   key.Binding
	Search   key.Binding
	Create   key.Binding
	Rename   key.Binding
	Delete   key.Binding
	Group    key.Binding
	Refresh  key.Binding
	Help     key.Binding
	Quit     key.Binding
	Escape   key.Binding
	Confirm  key.Binding
	Cancel   key.Binding
	Tab      key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "go to bottom"),
		),

		// Vim navigation
		VimUp: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "up"),
		),
		VimDown: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "down"),
		),
		VimLeft: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "collapse"),
		),
		VimRight: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "expand"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("gg", "go to top"),
		),
		GotoEnd: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "go to bottom"),
		),

		// Actions
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "attach"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Create: key.NewBinding(
			key.WithKeys("n", "c"),
			key.WithHelp("n", "new session"),
		),
		Rename: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d", "x"),
			key.WithHelp("d", "delete"),
		),
		Group: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "move to group"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "esc"),
			key.WithHelp("n", "cancel"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
	}
}

// ShortHelp returns key bindings for help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.VimDown, k.VimUp, k.Enter, k.Search, k.Create, k.Delete, k.Help, k.Quit}
}

// FullHelp returns extended key bindings for help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.VimUp, k.VimDown, k.VimLeft, k.VimRight},
		{k.GotoTop, k.GotoEnd, k.PageUp, k.PageDown},
		{k.Enter, k.Toggle, k.Search},
		{k.Create, k.Rename, k.Delete, k.Group},
		{k.Refresh, k.Help, k.Quit},
	}
}
