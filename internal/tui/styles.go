package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	accentColor    = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	successColor   = lipgloss.Color("#10B981") // Green
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	textColor      = lipgloss.Color("#F9FAFB") // Light text
	dimTextColor   = lipgloss.Color("#9CA3AF") // Dim text
	bgColor        = lipgloss.Color("#1F2937") // Dark background
)

// Styles
var (
	// Title styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1).
			Padding(0, 1)

	// Header style for the app
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 2).
			MarginBottom(1)

	// Subtitle style
	subtitleStyle = lipgloss.NewStyle().
			Foreground(dimTextColor).
			MarginBottom(1)

	// Menu item styles
	menuItemStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 2)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Padding(0, 2)

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			MarginTop(1)

	// Success style
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Info style
	infoStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	// Muted style
	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Help style
	helpStyle = lipgloss.NewStyle().
			Foreground(dimTextColor).
			MarginTop(1)

	// File selected indicator
	selectedFileStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	// Progress bar style
	progressBarStyle = lipgloss.NewStyle().
				Foreground(primaryColor)

	// Spinner style
	spinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	// Divider style
	dividerStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
)

// Icons
const (
	iconFolder      = "📁"
	iconFile        = "📄"
	iconImage       = "🖼️ "
	iconVideo       = "🎬"
	iconSelected    = "✓"
	iconNotSelected = "○"
	iconArrowRight  = "→"
	iconSuccess     = "✅"
	iconError       = "❌"
	iconWarning     = "⚠️"
	iconSpinner     = "◐"
	iconConvert     = "🔄"
	iconCompress    = "📦"
)

// FormatSize returns a human-readable file size
func FormatSize(size int64) string {
	const (
		KB int64 = 1024
		MB       = KB * 1024
		GB       = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}
