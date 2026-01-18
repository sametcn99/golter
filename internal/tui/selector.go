package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func (s *Selector) Init() tea.Cmd {
	return nil
}
