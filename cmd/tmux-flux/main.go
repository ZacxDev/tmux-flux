package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zchase/tmux-flux/internal/session"
	"github.com/zchase/tmux-flux/internal/ui"
)

func main() {
	// Initialize session manager
	storagePath := session.GetDefaultStoragePath()
	manager, err := session.NewManager(storagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing session manager: %v\n", err)
		os.Exit(1)
	}

	// Initial refresh to sync with tmux
	if err := manager.Refresh(); err != nil {
		fmt.Fprintf(os.Stderr, "Error syncing with tmux: %v\n", err)
		os.Exit(1)
	}

	// Create and run the TUI
	model := ui.NewModel(manager)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
