package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

func (s *Selector) loadFiles() {
	entries, err := os.ReadDir(s.currentDir)
	if err != nil {
		// Handle error gracefully
		s.list.SetItems([]list.Item{})
		return
	}

	// Sort: Dirs first, then files alphabetically
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() && !entries[j].IsDir() {
			return true
		}
		if !entries[i].IsDir() && entries[j].IsDir() {
			return false
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	items := []list.Item{}

	// Add ".." if not root
	if filepath.Dir(s.currentDir) != s.currentDir {
		items = append(items, item{path: filepath.Dir(s.currentDir), isDir: true, info: nil})
	}

	// Count files and directories for summary
	dirCount := 0
	fileCount := 0

	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}

		// Skip hidden files and directories
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}

		if e.IsDir() {
			dirCount++
		} else {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if !s.allowedExts[ext] {
				continue
			}
			fileCount++
		}

		path := filepath.Join(s.currentDir, e.Name())
		items = append(items, item{
			path:     path,
			isDir:    e.IsDir(),
			info:     info,
			selected: s.selected[path],
		})
	}

	s.list.SetItems(items)

	// Update status bar with summary
	statusText := fmt.Sprintf("%d folders, %d files", dirCount, fileCount)
	s.list.SetStatusBarItemName(statusText, statusText)
}
