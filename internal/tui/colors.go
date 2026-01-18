package tui

import "github.com/charmbracelet/lipgloss"

// Color palette - Modern dark theme with gradient accents
var (
	// Primary colors
	primaryColor = lipgloss.Color("#8B5CF6") // Vivid Purple
	// primaryDark    = lipgloss.Color("#6D28D9") // Deep Purple
	primaryLight = lipgloss.Color("#A78BFA") // Light Purple
	// secondaryColor = lipgloss.Color("#10B981") // Emerald Green
	secondaryLight = lipgloss.Color("#34D399") // Light Emerald
	// accentColor    = lipgloss.Color("#F59E0B") // Amber
	accentLight = lipgloss.Color("#FBBF24") // Light Amber

	// Status colors
	errorColor   = lipgloss.Color("#EF4444") // Red
	errorLight   = lipgloss.Color("#F87171") // Light Red
	successColor = lipgloss.Color("#10B981") // Green
	successLight = lipgloss.Color("#34D399") // Light Green
	warningColor = lipgloss.Color("#F59E0B") // Amber
	// warningLight = lipgloss.Color("#FBBF24") // Light Amber
	infoColor = lipgloss.Color("#3B82F6") // Blue
	// infoLight    = lipgloss.Color("#60A5FA") // Light Blue

	// Neutral colors
	// bgColor      = lipgloss.Color("#0F172A") // Dark Slate
	// bgLightColor = lipgloss.Color("#1E293B") // Slate
	borderColor  = lipgloss.Color("#334155") // Slate Border
	mutedColor   = lipgloss.Color("#64748B") // Slate Gray
	textColor    = lipgloss.Color("#F1F5F9") // Light Slate
	dimTextColor = lipgloss.Color("#94A3B8") // Dim Slate
	highlightBg  = lipgloss.Color("#1E3A5F") // Dark Blue Background

	// Special colors for file types
	imageColor    = lipgloss.Color("#EC4899") // Pink
	videoColor    = lipgloss.Color("#F43F5E") // Rose
	audioColor    = lipgloss.Color("#8B5CF6") // Purple
	documentColor = lipgloss.Color("#3B82F6") // Blue
)
