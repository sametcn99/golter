package tui

import (
	"fmt"
	"strings"
)

func (s *Selector) View() string {
	var b strings.Builder

	// Current directory path with breadcrumb style
	path := truncatePath(s.currentDir, s.width-10)
	dirHeader := dirPathStyle.Render(iconFolderOpen + " " + path)
	b.WriteString(dirHeader + "\n")

	// Show selection summary if files are selected
	if len(s.selected) > 0 {
		typeIcon := s.getSelectedTypeIcon()
		summary := selectedFileStyle.Render(fmt.Sprintf("  %s %d %s selected", typeIcon, len(s.selected), s.selectedFileType.String()))
		b.WriteString(summary + "\n")
	}

	b.WriteString("\n")
	b.WriteString(s.list.View())

	return b.String()
}

// getSelectedTypeIcon returns an icon for the currently selected file type
func (s *Selector) getSelectedTypeIcon() string {
	switch s.selectedFileType {
	case FileTypeImage:
		return iconImage
	case FileTypeVideo:
		return iconVideo
	case FileTypeAudio:
		return iconAudio
	case FileTypeDocument:
		return iconDocument
	default:
		return iconFile
	}
}
