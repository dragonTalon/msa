package config

import "github.com/charmbracelet/lipgloss"

// Styles for the config TUI
var (
	// Title style
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			MarginTop(1).
			MarginBottom(1)

	// Section title style
	SectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39")).
				MarginTop(1).
				MarginBottom(0)

	// Label style
	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")).
			MarginRight(1)

	// Value style
	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Selected item style
	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	// Dim style
	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")).
			MarginTop(1)

	// Border style
	BorderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238"))

	// Input style
	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	// Cursor style
	CursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)
)

// Helper function to render masked API key
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return apiKey
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
