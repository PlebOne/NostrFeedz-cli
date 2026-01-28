package styles

import "github.com/charmbracelet/lipgloss"

// Colors - Dark terminal optimized
var (
	PrimaryColor   = lipgloss.Color("#A78BFA") // Brighter purple for dark bg
	SecondaryColor = lipgloss.Color("#9CA3AF")
	AccentColor    = lipgloss.Color("#60A5FA") // Brighter blue for dark bg
	ErrorColor     = lipgloss.Color("#F87171") // Brighter red
	SuccessColor   = lipgloss.Color("#34D399") // Brighter green
	WarningColor   = lipgloss.Color("#FBBF24") // Brighter yellow
	
	BackgroundColor = lipgloss.Color("#1F2937") // Dark background
	BorderColor     = lipgloss.Color("#4B5563")  // Gray border for dark bg
	TextColor       = lipgloss.Color("#F9FAFB")  // Light text for dark bg
	MutedTextColor  = lipgloss.Color("#9CA3AF")  // Muted gray text
)

// Styles
var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		PaddingLeft(2)

	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		Background(lipgloss.Color("#1E3A8A")). // Dark blue bg
		Padding(0, 1)

	FeedItemStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(TextColor)

	SelectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")). // White text for selection
		Background(AccentColor).                // Blue highlight
		PaddingLeft(2)

	UnreadBadge = lipgloss.NewStyle().
		Background(AccentColor).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Bold(true)

	FavoriteBadge = lipgloss.NewStyle().
		Foreground(WarningColor).
		Bold(true)

	PanelBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(1)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)

	MutedStyle = lipgloss.NewStyle().
		Foreground(MutedTextColor)

	StatusBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#111827")). // Darker background
		Foreground(MutedTextColor).
		Padding(0, 1)

	KeyStyle = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	ValueStyle = lipgloss.NewStyle().
		Foreground(TextColor)
)

// RenderKeyValue renders a key-value pair for status bar
func RenderKeyValue(key, value string) string {
	return KeyStyle.Render(key) + ": " + ValueStyle.Render(value)
}

// RenderError renders an error message
func RenderError(msg string) string {
	return ErrorStyle.Render("✗ " + msg)
}

// RenderSuccess renders a success message
func RenderSuccess(msg string) string {
	return SuccessStyle.Render("✓ " + msg)
}
