package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	// Title styles - Hero section
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryLight).
			MarginBottom(1).
			Padding(0, 1)

	// Header style for the app - Gradient-like appearance
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 3).
			MarginBottom(1)

	// Menu item styles - List items (same padding to prevent shifting)
	menuItemStyle = lipgloss.NewStyle().
			Foreground(dimTextColor).
			PaddingLeft(4)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(primaryLight).
				Bold(true).
				PaddingLeft(4).
				Background(highlightBg)

	// Box styles - Container elements
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 3).
			MarginTop(1).
			MarginBottom(1)

	// Result box styles
	successBoxStyle = boxStyle.
			BorderForeground(successColor).
			Background(lipgloss.Color("#052E16")) // Very dark green

	errorBoxStyle = boxStyle.
			BorderForeground(errorColor).
			Background(lipgloss.Color("#450A0A")) // Very dark red

	// Success style
	successStyle = lipgloss.NewStyle().
			Foreground(successLight).
			Bold(true)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(errorLight).
			Bold(true)

	// Info style
	infoStyle = lipgloss.NewStyle().
			Foreground(accentLight)

	// Muted style - Secondary text
	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Help style - Footer hints
	helpStyle = lipgloss.NewStyle().
			Foreground(dimTextColor).
			MarginTop(1).
			PaddingLeft(2)

	// Help key style - Keyboard shortcuts
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(primaryLight).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(dimTextColor)

	helpSeparatorStyle = lipgloss.NewStyle().
				Foreground(borderColor)

	// File selected indicator
	selectedFileStyle = lipgloss.NewStyle().
				Foreground(secondaryLight).
				Bold(true)

	// Spinner style
	spinnerStyle = lipgloss.NewStyle().
			Foreground(primaryLight)

	// Progress bar custom styling
	progressFullStyle = lipgloss.NewStyle().
				Foreground(primaryColor)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(borderColor)

	// Directory path style
	dirPathStyle = lipgloss.NewStyle().
			Foreground(infoColor).
			Bold(true).
			PaddingLeft(2)

	// State title style for different states
	stateTitleStyle = lipgloss.NewStyle().
			Foreground(primaryLight).
			Bold(true).
			MarginBottom(1).
			PaddingLeft(2)

	// Confirmation dialog style
	confirmStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(warningColor).
			Padding(1, 3).
			MarginTop(1).
			Align(lipgloss.Center)

	// Progress text style
	progressTextStyle = lipgloss.NewStyle().
				Foreground(dimTextColor).
				Italic(true)

	// Current file processing style
	currentFileStyle = lipgloss.NewStyle().
				Foreground(accentLight).
				Italic(true)
)
