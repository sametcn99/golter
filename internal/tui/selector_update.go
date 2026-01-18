package tui

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (s *Selector) Update(msg tea.Msg) (Selector, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		// Reserve space for header, footer, and path display
		listHeight := msg.Height - 10
		if listHeight < 5 {
			listHeight = 5
		}
		s.list.SetWidth(msg.Width)
		s.list.SetHeight(listHeight)
		return *s, nil

	case tea.KeyMsg:
		// Don't handle keys when filtering
		if s.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "enter":
			i, ok := s.list.SelectedItem().(item)
			if ok {
				if i.isDir {
					s.currentDir = i.path
					s.loadFiles()
					s.list.ResetSelected()
					return *s, nil
				}
			}
		case " ":
			i, ok := s.list.SelectedItem().(item)
			if ok && !i.isDir {
				path := i.path
				ext := filepath.Ext(path)
				fileType := getFileType(ext)

				if s.selected[path] {
					// Deselecting
					delete(s.selected, path)
					i.selected = false
					// If no files left, reset the file type
					if len(s.selected) == 0 {
						s.selectedFileType = FileTypeUnknown
					}
				} else {
					// Check if we can select this file type
					if s.selectedFileType == FileTypeUnknown {
						// First file selected, set the type
						s.selectedFileType = fileType
						s.selected[path] = true
						i.selected = true
					} else if s.selectedFileType == fileType {
						// Same type, allow selection
						s.selected[path] = true
						i.selected = true
					}
					// If different type, don't allow selection (silently ignore)
				}
				// Update item in list
				s.list.SetItem(s.list.Index(), i)
				return *s, nil
			}
		case "a":
			// Select all files of the same type in current directory
			items := s.list.Items()
			selectedCount := 0
			for idx, listItem := range items {
				if i, ok := listItem.(item); ok && !i.isDir && i.info != nil {
					ext := filepath.Ext(i.path)
					fileType := getFileType(ext)

					// If no type selected yet, use the first file's type
					if s.selectedFileType == FileTypeUnknown && !s.selected[i.path] {
						s.selectedFileType = fileType
					}

					// Only select files of the same type
					if fileType == s.selectedFileType && !s.selected[i.path] {
						s.selected[i.path] = true
						i.selected = true
						s.list.SetItem(idx, i)
						selectedCount++
					}
				}
			}
			return *s, nil
		case "d":
			// Deselect all files
			items := s.list.Items()
			for idx, listItem := range items {
				if i, ok := listItem.(item); ok && !i.isDir && i.info != nil {
					if s.selected[i.path] {
						delete(s.selected, i.path)
						i.selected = false
						s.list.SetItem(idx, i)
					}
				}
			}
			s.selectedFileType = FileTypeUnknown
			return *s, nil
		case "h", "left":
			// Go to parent directory
			parent := filepath.Dir(s.currentDir)
			if parent != s.currentDir {
				s.currentDir = parent
				s.loadFiles()
				s.list.ResetSelected()
				return *s, nil
			}
		case "l", "right":
			// Enter directory if selected
			i, ok := s.list.SelectedItem().(item)
			if ok && i.isDir {
				s.currentDir = i.path
				s.loadFiles()
				s.list.ResetSelected()
				return *s, nil
			}
		case "g":
			// Go to top
			s.list.Select(0)
			return *s, nil
		case "G":
			// Go to bottom
			s.list.Select(len(s.list.Items()) - 1)
			return *s, nil
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return *s, cmd
}
