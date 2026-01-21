package ui

import (
	"strings"
	"testing"
)

func TestRenderGroupHeader(t *testing.T) {
	tests := []struct {
		name      string
		collapsed bool
		selected  bool
	}{
		{"test-group", false, false},
		{"test-group", true, false},
		{"test-group", false, true},
		{"test-group", true, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := RenderGroupHeader(tt.name, tt.collapsed, tt.selected)
			if result == "" {
				t.Error("RenderGroupHeader() returned empty string")
			}
			// Should contain the group name
			if !strings.Contains(result, tt.name) {
				t.Errorf("RenderGroupHeader() = %q, should contain %q", result, tt.name)
			}
		})
	}
}

func TestRenderSession(t *testing.T) {
	tests := []struct {
		name     string
		attached bool
		selected bool
	}{
		{"test-session", false, false},
		{"test-session", true, false},
		{"test-session", false, true},
		{"test-session", true, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := RenderSession(tt.name, tt.attached, tt.selected)
			if result == "" {
				t.Error("RenderSession() returned empty string")
			}
			// Should contain the session name
			if !strings.Contains(result, tt.name) {
				t.Errorf("RenderSession() = %q, should contain %q", result, tt.name)
			}
		})
	}
}

func TestRenderHelp(t *testing.T) {
	items := []HelpItem{
		{"j/k", "navigate"},
		{"enter", "select"},
		{"q", "quit"},
	}

	result := RenderHelp(items)
	if result == "" {
		t.Error("RenderHelp() returned empty string")
	}

	// Should contain all keys
	for _, item := range items {
		if !strings.Contains(result, item.Key) {
			t.Errorf("RenderHelp() = %q, should contain key %q", result, item.Key)
		}
	}
}

func TestRenderHelp_Empty(t *testing.T) {
	result := RenderHelp([]HelpItem{})
	// Empty help is valid, just empty string
	if result != "" {
		t.Logf("RenderHelp([]) = %q", result)
	}
}

func TestIcons(t *testing.T) {
	// Verify icons are non-empty strings
	icons := []struct {
		name  string
		value string
	}{
		{"IconExpanded", IconExpanded},
		{"IconCollapsed", IconCollapsed},
		{"IconSession", IconSession},
		{"IconAttached", IconAttached},
		{"IconSelected", IconSelected},
	}

	for _, icon := range icons {
		t.Run(icon.name, func(t *testing.T) {
			if icon.value == "" {
				t.Errorf("%s is empty", icon.name)
			}
		})
	}
}

func TestColors(t *testing.T) {
	// Verify colors are defined (non-empty)
	colors := []struct {
		name  string
		value string
	}{
		{"ColorBg", string(ColorBg)},
		{"ColorFg", string(ColorFg)},
		{"ColorFgDim", string(ColorFgDim)},
		{"ColorAccent", string(ColorAccent)},
		{"ColorAccentDim", string(ColorAccentDim)},
		{"ColorGreen", string(ColorGreen)},
		{"ColorRed", string(ColorRed)},
		{"ColorYellow", string(ColorYellow)},
		{"ColorCyan", string(ColorCyan)},
		{"ColorPurple", string(ColorPurple)},
	}

	for _, color := range colors {
		t.Run(color.name, func(t *testing.T) {
			if color.value == "" {
				t.Errorf("%s is empty", color.name)
			}
			// Should be a hex color
			if !strings.HasPrefix(color.value, "#") {
				t.Errorf("%s = %q, should start with #", color.name, color.value)
			}
		})
	}
}

func TestStyles_NotNil(t *testing.T) {
	// Verify styles are initialized
	styles := []struct {
		name string
		fn   func() string
	}{
		{"BaseStyle", func() string { return BaseStyle.Render("test") }},
		{"TitleStyle", func() string { return TitleStyle.Render("test") }},
		{"GroupStyle", func() string { return GroupStyle.Render("test") }},
		{"GroupCollapsedStyle", func() string { return GroupCollapsedStyle.Render("test") }},
		{"SessionStyle", func() string { return SessionStyle.Render("test") }},
		{"SessionSelectedStyle", func() string { return SessionSelectedStyle.Render("test") }},
		{"SessionAttachedStyle", func() string { return SessionAttachedStyle.Render("test") }},
		{"StatusBarStyle", func() string { return StatusBarStyle.Render("test") }},
		{"HelpStyle", func() string { return HelpStyle.Render("test") }},
		{"HelpKeyStyle", func() string { return HelpKeyStyle.Render("test") }},
		{"SearchStyle", func() string { return SearchStyle.Render("test") }},
		{"SearchInputStyle", func() string { return SearchInputStyle.Render("test") }},
		{"DialogStyle", func() string { return DialogStyle.Render("test") }},
		{"DialogTitleStyle", func() string { return DialogTitleStyle.Render("test") }},
		{"ErrorStyle", func() string { return ErrorStyle.Render("test") }},
		{"WarningStyle", func() string { return WarningStyle.Render("test") }},
	}

	for _, style := range styles {
		t.Run(style.name, func(t *testing.T) {
			result := style.fn()
			if result == "" {
				t.Errorf("%s.Render() returned empty string", style.name)
			}
		})
	}
}

func TestHelpItem(t *testing.T) {
	item := HelpItem{Key: "j", Desc: "move down"}

	if item.Key != "j" {
		t.Errorf("HelpItem.Key = %q, want 'j'", item.Key)
	}

	if item.Desc != "move down" {
		t.Errorf("HelpItem.Desc = %q, want 'move down'", item.Desc)
	}
}

func TestPreviewStyles_NotNil(t *testing.T) {
	styles := []struct {
		name string
		fn   func() string
	}{
		{"PreviewStyle", func() string { return PreviewStyle.Render("test") }},
		{"PreviewTitleStyle", func() string { return PreviewTitleStyle.Render("test") }},
		{"PreviewBorderStyle", func() string { return PreviewBorderStyle.Render("test") }},
	}

	for _, style := range styles {
		t.Run(style.name, func(t *testing.T) {
			result := style.fn()
			if result == "" {
				t.Errorf("%s.Render() returned empty string", style.name)
			}
		})
	}
}

func TestJoinWithSep(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		sep      string
		expected int // expected length of result
	}{
		{"empty", []string{}, " ", 0},
		{"single", []string{"a"}, " ", 1},
		{"multiple", []string{"a", "b", "c"}, " ", 5}, // a, sep, b, sep, c
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinWithSep(tt.parts, tt.sep)
			if len(result) != tt.expected {
				t.Errorf("joinWithSep(%v, %q) len = %d, want %d", tt.parts, tt.sep, len(result), tt.expected)
			}
		})
	}
}
