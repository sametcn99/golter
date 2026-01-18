package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// statFile is a helper to get file info
func statFile(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func getFormatIcon(format string) string {
	switch strings.ToLower(format) {
	case ".jpg", ".jpeg":
		return "🖼️ "
	case ".png":
		return "🖼️ "
	case ".webp":
		return "🖼️ "
	case ".gif":
		return iconGIF
	case ".mp4", ".avi", ".mkv", ".webm", ".mov":
		return iconVideo
	case ".pdf":
		return iconPDF
	case ".md":
		return iconMarkdown
	case ".html":
		return iconHTML
	case ".epub":
		return iconEPUB
	case ".mobi", ".azw", ".azw3", ".fb2":
		return iconEPUB
	case ".csv":
		return iconCSV
	case ".xlsx", ".xls":
		return iconExcel
	case ".mp3", ".wav", ".ogg", ".flac", ".m4a", ".aac":
		return iconAudio
	default:
		return iconFile
	}
}

// formatDuration returns a human-readable duration string
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
}

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
	case ".pdf", ".md", ".html", ".epub", ".mobi", ".azw", ".azw3", ".fb2", ".doc", ".docx", ".txt", ".csv", ".xlsx", ".xls":
		return "document"
	default:
		return "other"
	}
}
