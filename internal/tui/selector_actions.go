package tui

import (
	"path/filepath"
	"sort"
)

func (s *Selector) SelectedFiles() []string {
	files := make([]string, 0, len(s.selected))
	for f := range s.selected {
		files = append(files, f)
	}
	// Sort for consistent ordering
	sort.Strings(files)
	return files
}

func (s *Selector) ClearSelection() {
	s.selected = make(map[string]bool)
	s.selectedFileType = FileTypeUnknown
	s.loadFiles()
}

// GetSelectedFileType returns the current selected file type
func (s *Selector) GetSelectedFileType() FileType {
	return s.selectedFileType
}

// TotalSelectableFiles returns the count of selectable files in current directory
func (s *Selector) TotalSelectableFiles() int {
	count := 0
	for _, listItem := range s.list.Items() {
		if i, ok := listItem.(item); ok && !i.isDir && i.info != nil {
			ext := filepath.Ext(i.path)
			fileType := getFileType(ext)
			if s.selectedFileType == FileTypeUnknown || fileType == s.selectedFileType {
				count++
			}
		}
	}
	return count
}
