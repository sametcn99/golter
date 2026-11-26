package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// FileType represents the type of file (image, video, audio, document)
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeImage
	FileTypeVideo
	FileTypeAudio
	FileTypeDocument
)

// GetFileTypeName returns a human-readable name for the file type
func (ft FileType) String() string {
	switch ft {
	case FileTypeImage:
		return "images"
	case FileTypeVideo:
		return "videos"
	case FileTypeAudio:
		return "audio files"
	case FileTypeDocument:
		return "documents"
	default:
		return "files"
	}
}

// getFileType returns the file type based on extension
func getFileType(ext string) FileType {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return FileTypeImage
	case ".mp4", ".avi", ".mkv", ".webm", ".mov", ".gif":
		return FileTypeVideo
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a":
		return FileTypeAudio
	case ".pdf", ".md", ".html", ".epub", ".mobi", ".azw", ".azw3", ".fb2", ".docx", ".csv", ".xlsx", ".json", ".yaml", ".yml", ".toml", ".xml":
		return FileTypeDocument
	default:
		return FileTypeUnknown
	}
}

func getFileIcon(ext string) string {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".bmp", ".tiff":
		return iconImage
	case ".gif":
		return iconGIF
	case ".mp4", ".avi", ".mkv", ".webm", ".mov", ".wmv", ".flv":
		return iconVideo
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a":
		return iconAudio
	case ".pdf":
		return iconPDF
	case ".md":
		return iconMarkdown
	case ".html":
		return iconHTML
	case ".epub":
		return iconEPUB
	case ".csv":
		return iconCSV
	case ".xlsx":
		return iconExcel
	case ".doc", ".docx":
		return iconDocument
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return iconArchive
	default:
		return iconFile
	}
}

// truncatePath shortens a path to fit within maxWidth
func truncatePath(path string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	width := ansi.StringWidth(path)
	if width <= maxWidth {
		return path
	}
	if maxWidth == 1 {
		return "…"
	}
	return ansi.Truncate(ansi.TruncateLeft(path, width-maxWidth+1, "…"), maxWidth, "")
}
