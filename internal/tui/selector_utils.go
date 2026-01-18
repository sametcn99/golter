package tui

import (
	"path/filepath"
	"strings"
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
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".tiff":
		return FileTypeImage
	case ".mp4", ".avi", ".mkv", ".webm", ".mov", ".wmv", ".flv":
		return FileTypeVideo
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a":
		return FileTypeAudio
	case ".pdf", ".md", ".html", ".epub", ".mobi", ".azw", ".azw3", ".fb2", ".doc", ".docx", ".csv", ".xlsx", ".xls":
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
	case ".xlsx", ".xls":
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
	if len(path) <= maxWidth {
		return path
	}

	// Try to show the last few path components
	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) <= 2 {
		return "..." + path[len(path)-maxWidth+3:]
	}

	// Build from the end
	result := parts[len(parts)-1]
	for i := len(parts) - 2; i >= 0; i-- {
		if len(parts[i])+len(result)+4 > maxWidth {
			break
		}
		result = parts[i] + string(filepath.Separator) + result
	}

	return "..." + string(filepath.Separator) + result
}
