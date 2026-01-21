package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Test that all key bindings are set
	bindings := []struct {
		name    string
		binding key.Binding
	}{
		{"Up", km.Up},
		{"Down", km.Down},
		{"PageUp", km.PageUp},
		{"PageDown", km.PageDown},
		{"Home", km.Home},
		{"End", km.End},
		{"VimUp", km.VimUp},
		{"VimDown", km.VimDown},
		{"VimLeft", km.VimLeft},
		{"VimRight", km.VimRight},
		{"GotoTop", km.GotoTop},
		{"GotoEnd", km.GotoEnd},
		{"Enter", km.Enter},
		{"Select", km.Select},
		{"Toggle", km.Toggle},
		{"Search", km.Search},
		{"Create", km.Create},
		{"NewGroup", km.NewGroup},
		{"Rename", km.Rename},
		{"Delete", km.Delete},
		{"Group", km.Group},
		{"Refresh", km.Refresh},
		{"Help", km.Help},
		{"Quit", km.Quit},
		{"Escape", km.Escape},
		{"Confirm", km.Confirm},
		{"Cancel", km.Cancel},
		{"Tab", km.Tab},
	}

	for _, b := range bindings {
		t.Run(b.name, func(t *testing.T) {
			if len(b.binding.Keys()) == 0 {
				t.Errorf("%s has no keys defined", b.name)
			}
		})
	}
}

func TestKeyMap_VimNavigation(t *testing.T) {
	km := DefaultKeyMap()

	// Check vim navigation keys
	tests := []struct {
		name     string
		binding  key.Binding
		expected string
	}{
		{"VimUp", km.VimUp, "k"},
		{"VimDown", km.VimDown, "j"},
		{"VimLeft", km.VimLeft, "h"},
		{"VimRight", km.VimRight, "l"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.binding.Keys()
			found := false
			for _, k := range keys {
				if k == tt.expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s should include key %q, got %v", tt.name, tt.expected, keys)
			}
		})
	}
}

func TestKeyMap_CreateKeys(t *testing.T) {
	km := DefaultKeyMap()

	// Create should have both 'n' and 'c'
	createKeys := km.Create.Keys()
	hasN := false
	hasC := false
	for _, k := range createKeys {
		if k == "n" {
			hasN = true
		}
		if k == "c" {
			hasC = true
		}
	}

	if !hasN {
		t.Error("Create binding should include 'n'")
	}
	if !hasC {
		t.Error("Create binding should include 'c'")
	}
}

func TestKeyMap_NewGroupKey(t *testing.T) {
	km := DefaultKeyMap()

	// NewGroup should have 'N' (capital)
	keys := km.NewGroup.Keys()
	hasN := false
	for _, k := range keys {
		if k == "N" {
			hasN = true
			break
		}
	}

	if !hasN {
		t.Error("NewGroup binding should include 'N'")
	}
}

func TestKeyMap_DeleteKeys(t *testing.T) {
	km := DefaultKeyMap()

	// Delete should have both 'd' and 'x'
	deleteKeys := km.Delete.Keys()
	hasD := false
	hasX := false
	for _, k := range deleteKeys {
		if k == "d" {
			hasD = true
		}
		if k == "x" {
			hasX = true
		}
	}

	if !hasD {
		t.Error("Delete binding should include 'd'")
	}
	if !hasX {
		t.Error("Delete binding should include 'x'")
	}
}

func TestKeyMap_QuitKeys(t *testing.T) {
	km := DefaultKeyMap()

	// Quit should have 'q' and 'ctrl+c'
	quitKeys := km.Quit.Keys()
	hasQ := false
	hasCtrlC := false
	for _, k := range quitKeys {
		if k == "q" {
			hasQ = true
		}
		if k == "ctrl+c" {
			hasCtrlC = true
		}
	}

	if !hasQ {
		t.Error("Quit binding should include 'q'")
	}
	if !hasCtrlC {
		t.Error("Quit binding should include 'ctrl+c'")
	}
}

func TestKeyMap_SearchKey(t *testing.T) {
	km := DefaultKeyMap()

	// Search should be '/'
	keys := km.Search.Keys()
	hasSlash := false
	for _, k := range keys {
		if k == "/" {
			hasSlash = true
			break
		}
	}

	if !hasSlash {
		t.Error("Search binding should include '/'")
	}
}

func TestKeyMap_ShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.ShortHelp()

	if len(help) == 0 {
		t.Error("ShortHelp() returned empty slice")
	}

	// Should include essential bindings
	if len(help) < 5 {
		t.Errorf("ShortHelp() returned %d bindings, expected at least 5", len(help))
	}
}

func TestKeyMap_FullHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.FullHelp()

	if len(help) == 0 {
		t.Error("FullHelp() returned empty slice")
	}

	// Should have multiple categories
	if len(help) < 3 {
		t.Errorf("FullHelp() returned %d categories, expected at least 3", len(help))
	}

	// Each category should have bindings
	for i, category := range help {
		if len(category) == 0 {
			t.Errorf("FullHelp() category %d is empty", i)
		}
	}
}
