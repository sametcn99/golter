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
	// Parent directory item
	if i.info == nil && i.isDir {
		return iconBack + " .."
	}

	var prefix string
	if i.isDir {
		prefix = "  " // No checkbox for directories
	} else if i.selected {
		prefix = iconSelected
	} else {
		prefix = iconNotSelected
	}

	var icon string
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

// FileType represents the type of file (image, video, audio, document)
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeImage
	FileTypeVideo
	FileTypeAudio
	FileTypeDocument
)

// Highlight styles for list
var (
	listHighlightColor   = lipgloss.Color("#00D9FF") // Bright Cyan
	listHighlightBgColor = lipgloss.Color("#1E3A5F") // Dark Blue Background
)

type Selector struct {
	list             list.Model
	currentDir       string
	selected         map[string]bool // path -> true
	allowedExts      map[string]bool
	selectedFileType FileType // Track the type of first selected file
	width            int
	height           int
}

func NewSelector(startPath string, allowedExts []string) Selector {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	// Enhanced highlight styling for better UX - same padding for all to prevent shifting
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(listHighlightColor).
		Background(listHighlightBgColor).
		Bold(true)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(textColor)

	// Dimmed style for unselected items
	delegate.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(mutedColor)

	l := list.New(nil, delegate, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("item", "items")
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.Styles.Title = titleStyle
	l.SetFilteringEnabled(true)
	l.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(primaryLight).
		Bold(true)
	l.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(primaryLight)

	// Status bar styling
	l.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(dimTextColor).
		Padding(0, 0, 1, 2)

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
		width:            80,
		height:           24,
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
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a":
		return FileTypeAudio
	case ".pdf", ".md", ".html", ".epub", ".doc", ".docx", ".csv", ".xlsx", ".xls":
		return FileTypeDocument
	default:
		return FileTypeUnknown
	}
}

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

func (s *Selector) Init() tea.Cmd {
	return nil
}

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
