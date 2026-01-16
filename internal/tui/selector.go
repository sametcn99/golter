package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	path     string
	isDir    bool
	info     os.FileInfo
	selected bool
}

func (i item) Title() string {
	if i.info == nil && i.isDir {
		return "📁 .."
	}

	var prefix string
	if i.selected {
		prefix = iconSelected
	} else {
		prefix = iconNotSelected
	}

	icon := iconFile
	if i.isDir {
		icon = iconFolder
	} else {
		ext := strings.ToLower(filepath.Ext(i.info.Name()))
		icon = getFileIcon(ext)
	}

	name := i.info.Name()
	if i.isDir {
		return fmt.Sprintf("%s %s %s", prefix, icon, name)
	}

	// Add file size for files
	size := FormatSize(i.info.Size())
	return fmt.Sprintf("%s %s %s (%s)", prefix, icon, name, size)
}

func (i item) Description() string {
	return ""
}

func (i item) FilterValue() string {
	if i.info == nil {
		return ".."
	}
	return i.info.Name()
}

// FileType represents the type of file (image, video, etc.)
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeImage
	FileTypeVideo
)

type Selector struct {
	list             list.Model
	currentDir       string
	selected         map[string]bool // path -> true
	allowedExts      map[string]bool
	selectedFileType FileType // Track the type of first selected file
}

func NewSelector(startPath string, allowedExts []string) Selector {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	// Highlight color for selected item - bright cyan with background
	highlightColor := lipgloss.Color("#00FFFF") // Cyan
	highlightBg := lipgloss.Color("#1E3A5F")    // Dark blue background

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(highlightColor).
		Background(highlightBg).
		Bold(true).
		Padding(0, 1, 0, 2)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(textColor).
		Padding(0, 0, 0, 2)

	l := list.New(nil, delegate, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.Styles.Title = titleStyle
	l.SetFilteringEnabled(true)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(primaryColor)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(primaryColor)

	exts := make(map[string]bool)
	for _, e := range allowedExts {
		exts[strings.ToLower(e)] = true
	}

	s := Selector{
		list:             l,
		currentDir:       startPath,
		selected:         make(map[string]bool),
		allowedExts:      exts,
		selectedFileType: FileTypeUnknown,
	}
	s.loadFiles()
	return s
}

// getFileType returns the file type based on extension
func getFileType(ext string) FileType {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".tiff":
		return FileTypeImage
	case ".mp4", ".avi", ".mkv", ".webm", ".mov", ".wmv", ".flv":
		return FileTypeVideo
	default:
		return FileTypeUnknown
	}
}

func (s *Selector) loadFiles() {
	entries, _ := os.ReadDir(s.currentDir)

	// Sort: Dirs first, then files
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() && !entries[j].IsDir() {
			return true
		}
		if !entries[i].IsDir() && entries[j].IsDir() {
			return false
		}
		return entries[i].Name() < entries[j].Name()
	})

	items := []list.Item{}

	// Add ".." if not root
	if filepath.Dir(s.currentDir) != s.currentDir {
		items = append(items, item{path: filepath.Dir(s.currentDir), isDir: true, info: nil})
	}

	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}

		if !e.IsDir() {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if !s.allowedExts[ext] {
				continue
			}
		}
		// Skip hidden files
		if strings.HasPrefix(e.Name(), ".") {
			continue
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
}

func (s *Selector) Init() tea.Cmd {
	return nil
}

func (s *Selector) Update(msg tea.Msg) (Selector, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.list.SetWidth(msg.Width)
		s.list.SetHeight(msg.Height - 8) // Adjust for header/footer
		return *s, nil

	case tea.KeyMsg:
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
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return *s, cmd
}

func (s *Selector) View() string {
	// Show current directory path
	dirPath := mutedStyle.Render("  📂 " + s.currentDir)
	return dirPath + "\n\n" + s.list.View()
}

func (s *Selector) SelectedFiles() []string {
	files := make([]string, 0, len(s.selected))
	for f := range s.selected {
		files = append(files, f)
	}
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

// GetFileTypeName returns a human-readable name for the file type
func (ft FileType) String() string {
	switch ft {
	case FileTypeImage:
		return "images"
	case FileTypeVideo:
		return "videos"
	default:
		return "files"
	}
}

func getFileIcon(ext string) string {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".tiff":
		return iconImage
	case ".mp4", ".avi", ".mkv", ".webm", ".mov", ".wmv", ".flv":
		return iconVideo
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		return "🎵"
	case ".pdf":
		return "📕"
	case ".doc", ".docx":
		return "📝"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "📦"
	default:
		return iconFile
	}
}
