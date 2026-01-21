package ui

import "github.com/charmbracelet/lipgloss"

// Theme colors (Tokyo Night inspired)
var (
	ColorBg        = lipgloss.Color("#1a1b26")
	ColorFg        = lipgloss.Color("#c0caf5")
	ColorFgDim     = lipgloss.Color("#565f89")
	ColorAccent    = lipgloss.Color("#7aa2f7")
	ColorAccentDim = lipgloss.Color("#3d59a1")
	ColorGreen     = lipgloss.Color("#9ece6a")
	ColorRed       = lipgloss.Color("#f7768e")
	ColorYellow    = lipgloss.Color("#e0af68")
	ColorCyan      = lipgloss.Color("#7dcfff")
	ColorPurple    = lipgloss.Color("#bb9af7")
)

// Styles
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Background(ColorBg).
			Foreground(ColorFg)

	// Title bar
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent).
			Padding(0, 1)

	// Group styles
	GroupStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPurple)

	GroupCollapsedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorFgDim)

	// Session styles
	SessionStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			PaddingLeft(2)

	SessionSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorBg).
				Background(ColorAccent).
				Bold(true).
				PaddingLeft(2)

	SessionAttachedStyle = lipgloss.NewStyle().
				Foreground(ColorGreen).
				PaddingLeft(2)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorFgDim).
			Padding(0, 1)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorFgDim)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	// Search
	SearchStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true)

	SearchInputStyle = lipgloss.NewStyle().
				Foreground(ColorFg)

	// Dialog styles
	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			Padding(1, 2)

	DialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAccent)

	// Error/Warning
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)

	// Preview pane
	PreviewStyle = lipgloss.NewStyle().
			Foreground(ColorFgDim).
			PaddingLeft(1)

	PreviewTitleStyle = lipgloss.NewStyle().
				Foreground(ColorCyan).
				Bold(true).
				PaddingBottom(1)

	PreviewBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(ColorAccentDim).
				PaddingLeft(1)

	// Icons
	IconExpanded  = "▼"
	IconCollapsed = "▶"
	IconSession   = "○"
	IconAttached  = "●"
	IconSelected  = "→"
)

// Render helpers

func RenderGroupHeader(name string, collapsed bool, selected bool) string {
	icon := IconExpanded
	style := GroupStyle
	if collapsed {
		icon = IconCollapsed
		style = GroupCollapsedStyle
	}
	if selected {
		style = style.Background(ColorAccentDim)
	}
	return style.Render(icon + " " + name)
}

func RenderSession(name string, attached bool, selected bool) string {
	icon := IconSession
	style := SessionStyle
	if attached {
		icon = IconAttached
		style = SessionAttachedStyle
	}
	if selected {
		style = SessionSelectedStyle
		icon = IconSelected
	}
	return style.Render(icon + " " + name)
}

func RenderHelp(keys []HelpItem) string {
	var parts []string
	for _, item := range keys {
		key := HelpKeyStyle.Render(item.Key)
		desc := HelpStyle.Render(item.Desc)
		parts = append(parts, key+" "+desc)
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, joinWithSep(parts, "  ")...)
}

type HelpItem struct {
	Key  string
	Desc string
}

func joinWithSep(parts []string, sep string) []string {
	if len(parts) == 0 {
		return parts
	}
	result := make([]string, 0, len(parts)*2-1)
	for i, p := range parts {
		result = append(result, p)
		if i < len(parts)-1 {
			result = append(result, HelpStyle.Render(sep))
		}
	}
	return result
}
