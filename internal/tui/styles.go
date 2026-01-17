package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palette - Modern dark theme with gradient accents
var (
	// Primary colors
	primaryColor   = lipgloss.Color("#8B5CF6") // Vivid Purple
	primaryDark    = lipgloss.Color("#6D28D9") // Deep Purple
	primaryLight   = lipgloss.Color("#A78BFA") // Light Purple
	secondaryColor = lipgloss.Color("#10B981") // Emerald Green
	secondaryLight = lipgloss.Color("#34D399") // Light Emerald
	accentColor    = lipgloss.Color("#F59E0B") // Amber
	accentLight    = lipgloss.Color("#FBBF24") // Light Amber

	// Status colors
	errorColor   = lipgloss.Color("#EF4444") // Red
	errorLight   = lipgloss.Color("#F87171") // Light Red
	successColor = lipgloss.Color("#10B981") // Green
	successLight = lipgloss.Color("#34D399") // Light Green
	warningColor = lipgloss.Color("#F59E0B") // Amber
	warningLight = lipgloss.Color("#FBBF24") // Light Amber
	infoColor    = lipgloss.Color("#3B82F6") // Blue
	infoLight    = lipgloss.Color("#60A5FA") // Light Blue

	// Neutral colors
	bgColor      = lipgloss.Color("#0F172A") // Dark Slate
	bgLightColor = lipgloss.Color("#1E293B") // Slate
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

// Base style - can be embedded in others
var baseStyle = lipgloss.NewStyle().
	Foreground(textColor)

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

	// Subtitle style - Secondary information
	subtitleStyle = lipgloss.NewStyle().
			Foreground(dimTextColor).
			MarginBottom(1).
			Italic(true)

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

	// File size style
	fileSizeStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// State title style for different states
	stateTitleStyle = lipgloss.NewStyle().
			Foreground(primaryLight).
			Bold(true).
			MarginBottom(1).
			PaddingLeft(2)

	// Badge styles for file types
	imageBadgeStyle = lipgloss.NewStyle().
			Foreground(imageColor).
			Bold(true)

	videoBadgeStyle = lipgloss.NewStyle().
			Foreground(videoColor).
			Bold(true)

	audioBadgeStyle = lipgloss.NewStyle().
			Foreground(audioColor).
			Bold(true)

	documentBadgeStyle = lipgloss.NewStyle().
				Foreground(documentColor).
				Bold(true)

	// Confirmation dialog style
	confirmStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(warningColor).
			Padding(1, 3).
			MarginTop(1).
			Align(lipgloss.Center)

	// Breadcrumb style for navigation
	breadcrumbStyle = lipgloss.NewStyle().
			Foreground(dimTextColor).
			PaddingLeft(2)

	breadcrumbActiveStyle = lipgloss.NewStyle().
				Foreground(primaryLight).
				Bold(true)

	// Counter/badge style
	counterStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 1).
			Bold(true)

	// Progress text style
	progressTextStyle = lipgloss.NewStyle().
				Foreground(dimTextColor).
				Italic(true)

	// Current file processing style
	currentFileStyle = lipgloss.NewStyle().
				Foreground(accentLight).
				Italic(true)
)

// Icons - Nerd Font compatible with fallbacks
const (
	// Navigation icons
	iconFolder       = "📁"
	iconFolderOpen   = "📂"
	iconFile         = "📄"
	iconBack         = "⬅️ "
	iconArrowRight   = "→"
	iconArrowDown    = "↓"
	iconChevronRight = "›"
	iconChevronDown  = "▾"

	// File type icons
	iconImage    = "🖼️ "
	iconVideo    = "🎬"
	iconAudio    = "🎵"
	iconDocument = "📄"
	iconPDF      = "📕"
	iconMarkdown = "📝"
	iconHTML     = "🌐"
	iconEPUB     = "📚"
	iconArchive  = "📦"
	iconGIF      = "🎞️ "
	iconCSV      = "📊"
	iconExcel    = "📗"

	// Status icons
	iconSelected    = "●"
	iconNotSelected = "○"
	iconSuccess     = "✅"
	iconError       = "❌"
	iconWarning     = "⚠️ "
	iconInfo        = "ℹ️ "
	iconSpinner     = "◐"
	iconLoading     = "⏳"
	iconDone        = "✓"

	// Action icons
	iconConvert  = "🔄"
	iconCompress = "📦"
	iconSettings = "⚙️ "
	iconQuit     = "🚪"

	// Decoration
	iconStar   = "★"
	iconDot    = "•"
	iconPipe   = "│"
	iconCorner = "└"
	iconTee    = "├"
	iconHLine  = "─"
)

// Animated spinner frames
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Progress bar characters
const (
	progressBarFull  = "█"
	progressBarEmpty = "░"
	progressBarHead  = "▓"
)

// FormatSize returns a human-readable file size with appropriate unit
func FormatSize(size int64) string {
	const (
		KB int64 = 1024
		MB       = KB * 1024
		GB       = MB * 1024
		TB       = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// RenderHelpKey renders a keyboard shortcut in a consistent style
func RenderHelpKey(key, desc string) string {
	return helpKeyStyle.Render(key) + helpDescStyle.Render(" "+desc)
}

// RenderProgressBar creates a custom progress bar
func RenderProgressBar(current, total, width int) string {
	if total == 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))
	empty := width - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += progressFullStyle.Render(progressBarFull)
	}
	if filled < width && filled > 0 {
		bar += progressFullStyle.Render(progressBarHead)
		empty--
	}
	for i := 0; i < empty; i++ {
		bar += progressEmptyStyle.Render(progressBarEmpty)
	}

	percentStr := fmt.Sprintf(" %3.0f%%", percent*100)
	return bar + mutedStyle.Render(percentStr)
}

// GetFileTypeColor returns the appropriate color for a file type
func GetFileTypeColor(ext string) lipgloss.Color {
	switch GetFileCategory(ext) {
	case "image":
		return imageColor
	case "video":
		return videoColor
	case "audio":
		return audioColor
	case "document":
		return documentColor
	default:
		return textColor
	}
}

// GetFileCategory returns the category of a file based on its extension
func GetFileCategory(ext string) string {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".tiff":
		return "image"
	case ".mp4", ".avi", ".mkv", ".webm", ".mov", ".wmv", ".flv":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a":
		return "audio"
	case ".pdf", ".md", ".html", ".epub", ".doc", ".docx", ".txt", ".csv", ".xlsx", ".xls":
		return "document"
	default:
		return "other"
	}
}
