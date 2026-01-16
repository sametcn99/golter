package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golter/internal/converter"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	StateSelecting State = iota
	StateSelectingAction
	StateSelectingFormat
	StateSelectingQuality
	StateConverting
	StateDone
	StateQuitting
)

type conversionResult struct {
	path       string
	outputPath string
	err        error
}

type batchResult struct {
	results []conversionResult
}

type progressMsg struct {
	current int
	total   int
	file    string
}

type Model struct {
	state           State
	previousState   State
	selector        Selector
	selectedFiles   []string
	targetFormats   []string
	actionOptions   []string
	qualityOptions  []string
	cursor          int
	spinner         spinner.Model
	progress        progress.Model
	progressCurrent int
	progressTotal   int
	currentFile     string
	output          string
	err             error
	quitting        bool
	manager         *converter.Manager
	targetFormat    string
	width           int
	height          int
}

func NewModel(initialPath string) Model {
	mgr := converter.NewManager()
	mgr.Register(&converter.ImageConverter{})
	mgr.Register(&converter.VideoConverter{})

	if initialPath == "" {
		initialPath, _ = os.UserHomeDir()
	}
	// Ensure initialPath is a directory
	info, err := os.Stat(initialPath)
	if err == nil && !info.IsDir() {
		initialPath = filepath.Dir(initialPath)
	}

	s := NewSelector(initialPath, mgr.SupportedExtensions())

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = spinnerStyle

	p := progress.New(progress.WithDefaultGradient())
	p.Width = 40

	return Model{
		state:          StateSelecting,
		selector:       s,
		spinner:        sp,
		progress:       p,
		manager:        mgr,
		actionOptions:  []string{iconConvert + " Convert Format", iconCompress + " Compress Files"},
		qualityOptions: []string{"✨ High (Best Quality)", "⚖️  Medium (Balanced)", "📦 Low (Smallest Size)"},
		width:          80,
		height:         24,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 20
		if m.progress.Width > 60 {
			m.progress.Width = 60
		}

	case tea.KeyMsg:
		if m.state == StateQuitting {
			switch msg.String() {
			case "y", "Y":
				m.quitting = true
				return m, tea.Quit
			case "n", "N":
				m.state = m.previousState
				return m, nil
			}
			return m, nil
		}

		// Back handling
		if msg.String() == "esc" || msg.String() == "backspace" {
			switch m.state {
			case StateSelectingAction:
				m.state = StateSelecting
				return m, nil
			case StateSelectingFormat:
				m.state = StateSelectingAction
				m.cursor = 0
				return m, nil
			case StateSelectingQuality:
				m.state = StateSelectingAction
				m.cursor = 0
				return m, nil
			case StateDone:
				m.state = StateSelecting
				m.selectedFiles = nil
				m.selector.ClearSelection()
				m.err = nil
				m.output = ""
				return m, nil
			}
		}

		if m.state != StateConverting {
			switch msg.String() {
			case "ctrl+c", "q":
				m.previousState = m.state
				m.state = StateQuitting
				return m, nil
			}
		} else {
			if msg.String() == "ctrl+c" {
				m.quitting = true
				return m, tea.Quit
			}
		}

	case progressMsg:
		m.progressCurrent = msg.current
		m.progressTotal = msg.total
		m.currentFile = msg.file
		return m, nil

	case batchResult:
		m.state = StateDone
		// Aggregate results
		successCount := 0
		var errs []string
		var successFiles []string

		for _, res := range msg.results {
			if res.err != nil {
				errs = append(errs, fmt.Sprintf("%s %s: %v", iconError, filepath.Base(res.path), res.err))
			} else {
				successCount++
				successFiles = append(successFiles, fmt.Sprintf("%s %s %s %s", iconSuccess, filepath.Base(res.path), iconArrowRight, filepath.Base(res.outputPath)))
			}
		}

		if len(errs) > 0 {
			m.err = fmt.Errorf("Converted %d files. Errors:\n%s", successCount, strings.Join(errs, "\n"))
		} else {
			m.output = fmt.Sprintf("Successfully converted %d files:\n\n%s", successCount, strings.Join(successFiles, "\n"))
		}
		return m, nil
	}

	switch m.state {
	case StateSelecting:
		var cmd tea.Cmd
		m.selector, cmd = m.selector.Update(msg)

		// Check for "c" key to confirm selection
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "c" {
			files := m.selector.SelectedFiles()
			if len(files) > 0 {
				m.selectedFiles = files
				m.state = StateSelectingAction
				m.cursor = 0
				return m, nil
			}
		}
		return m, cmd

	case StateSelectingAction:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.actionOptions)-1 {
					m.cursor++
				}
			case "enter":
				if m.cursor == 0 { // Convert Format
					ext := filepath.Ext(m.selectedFiles[0])
					m.targetFormats = m.manager.GetSupportedTargetFormats(ext)

					if len(m.targetFormats) == 0 {
						m.err = fmt.Errorf("no supported target formats for %s", ext)
						m.state = StateDone
						return m, nil
					}

					m.state = StateSelectingFormat
					m.cursor = 0
				} else { // Compress Files
					m.targetFormat = "" // Empty means keep original format
					m.state = StateSelectingQuality
					m.cursor = 0
				}
				return m, nil
			}
		}

	case StateSelectingFormat:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.targetFormats)-1 {
					m.cursor++
				}
			case "enter":
				m.targetFormat = m.targetFormats[m.cursor]
				m.state = StateConverting
				m.progressCurrent = 0
				m.progressTotal = len(m.selectedFiles)
				// Use default quality for conversion
				return m, tea.Batch(m.spinner.Tick, convertFilesWithProgress(m.selectedFiles, m.targetFormat, "High", m.manager))
			}
		}

	case StateSelectingQuality:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.qualityOptions)-1 {
					m.cursor++
				}
			case "enter":
				quality := m.qualityOptions[m.cursor]
				m.state = StateConverting
				m.progressCurrent = 0
				m.progressTotal = len(m.selectedFiles)
				return m, tea.Batch(m.spinner.Tick, convertFilesWithProgress(m.selectedFiles, m.targetFormat, quality, m.manager))
			}
		}

	case StateConverting:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func convertFilesWithProgress(files []string, targetExt string, quality string, mgr *converter.Manager) tea.Cmd {
	return func() tea.Msg {
		var results []conversionResult
		var mu sync.Mutex
		opts := converter.Options{
			"quality": quality,
		}

		// Process files concurrently for better performance
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, 4) // Limit concurrent conversions

		for _, path := range files {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				ext := filepath.Ext(path)
				effectiveTargetExt := targetExt
				if effectiveTargetExt == "" {
					effectiveTargetExt = ext
				}

				conv, err := mgr.FindConverter(ext, effectiveTargetExt)
				if err != nil {
					mu.Lock()
					results = append(results, conversionResult{path: path, err: err})
					mu.Unlock()
					return
				}

				var outputPath string
				if targetExt == "" {
					// Compression mode
					outputPath = strings.TrimSuffix(path, ext) + "_compressed" + ext
				} else {
					// Conversion mode
					outputPath = strings.TrimSuffix(path, ext) + targetExt
					if outputPath == path {
						outputPath = strings.TrimSuffix(path, ext) + "_converted" + targetExt
					}
				}

				err = conv.Convert(path, outputPath, opts)
				mu.Lock()
				results = append(results, conversionResult{path: path, outputPath: outputPath, err: err})
				mu.Unlock()
			}(path)
		}

		wg.Wait()
		return batchResult{results: results}
	}
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Header
	header := headerStyle.Render(" 🔄 Golter - File Converter ")
	s.WriteString("\n" + header + "\n\n")

	switch m.state {
	case StateSelecting:
		s.WriteString(subtitleStyle.Render("  Select files to convert:") + "\n")
		s.WriteString(m.selector.View())
		s.WriteString("\n")

		// Selected files counter with better styling
		selectedCount := len(m.selector.SelectedFiles())
		fileType := m.selector.GetSelectedFileType()
		if selectedCount > 0 {
			typeStr := fileType.String()
			s.WriteString(selectedFileStyle.Render(fmt.Sprintf("  %s %d %s selected", iconSelected, selectedCount, typeStr)) + "\n")
			s.WriteString(infoStyle.Render("  Press 'c' to continue") + "\n")
		} else {
			s.WriteString(mutedStyle.Render("  No files selected") + "\n")
			s.WriteString(mutedStyle.Render("  (Select only images OR videos - cannot mix types)") + "\n")
		}

	case StateSelectingAction:
		s.WriteString(subtitleStyle.Render(fmt.Sprintf("  %d file(s) selected. Choose action:", len(m.selectedFiles))) + "\n\n")
		for i, action := range m.actionOptions {
			if m.cursor == i {
				s.WriteString(selectedMenuItemStyle.Render("  ▸ "+action) + "\n")
			} else {
				s.WriteString(menuItemStyle.Render("    "+action) + "\n")
			}
		}

	case StateSelectingFormat:
		s.WriteString(subtitleStyle.Render(fmt.Sprintf("  Converting %d file(s). Choose target format:", len(m.selectedFiles))) + "\n\n")
		for i, format := range m.targetFormats {
			icon := getFormatIcon(format)
			if m.cursor == i {
				s.WriteString(selectedMenuItemStyle.Render(fmt.Sprintf("  ▸ %s %s", icon, format)) + "\n")
			} else {
				s.WriteString(menuItemStyle.Render(fmt.Sprintf("    %s %s", icon, format)) + "\n")
			}
		}

	case StateSelectingQuality:
		s.WriteString(subtitleStyle.Render("  Choose compression quality:") + "\n\n")
		for i, q := range m.qualityOptions {
			if m.cursor == i {
				s.WriteString(selectedMenuItemStyle.Render("  ▸ "+q) + "\n")
			} else {
				s.WriteString(menuItemStyle.Render("    "+q) + "\n")
			}
		}

	case StateConverting:
		s.WriteString(subtitleStyle.Render("  Converting files...") + "\n\n")
		s.WriteString(fmt.Sprintf("  %s Processing %d file(s)...\n\n", m.spinner.View(), len(m.selectedFiles)))

		// Progress bar
		if m.progressTotal > 0 {
			progressPercent := float64(m.progressCurrent) / float64(m.progressTotal)
			s.WriteString("  " + m.progress.ViewAs(progressPercent) + "\n")
			s.WriteString(mutedStyle.Render(fmt.Sprintf("  %d / %d files", m.progressCurrent, m.progressTotal)) + "\n")
		}

		if m.currentFile != "" {
			s.WriteString(mutedStyle.Render(fmt.Sprintf("  Current: %s", filepath.Base(m.currentFile))) + "\n")
		}

	case StateDone:
		if m.err != nil {
			s.WriteString(boxStyle.BorderForeground(errorColor).Render(
				errorStyle.Render(fmt.Sprintf("%s Error\n\n", iconError))+m.err.Error(),
			) + "\n")
		} else {
			s.WriteString(boxStyle.BorderForeground(successColor).Render(
				successStyle.Render(fmt.Sprintf("%s Conversion Complete!\n\n", iconSuccess))+m.output,
			) + "\n")
		}

	case StateQuitting:
		s.WriteString(boxStyle.Render(
			infoStyle.Render(fmt.Sprintf("%s Are you sure you want to quit?\n\n", iconWarning))+
				mutedStyle.Render("Press 'y' to confirm, 'n' to cancel"),
		) + "\n")
	}

	// Footer with keyboard shortcuts
	if m.state != StateQuitting {
		s.WriteString("\n")
		footer := m.renderFooter()
		s.WriteString(helpStyle.Render(footer))
	}

	return s.String()
}

func (m Model) renderFooter() string {
	var shortcuts []string

	switch m.state {
	case StateSelecting:
		shortcuts = []string{"↑↓ Navigate", "Space Select", "a All", "d Clear", "Enter Open", "c Confirm", "q Quit"}
	case StateSelectingAction, StateSelectingFormat, StateSelectingQuality:
		shortcuts = []string{"↑↓ Navigate", "Enter Select", "Esc Back", "q Quit"}
	case StateConverting:
		shortcuts = []string{"Ctrl+C Cancel"}
	case StateDone:
		shortcuts = []string{"Esc New conversion", "q Quit"}
	}

	return "  " + strings.Join(shortcuts, " │ ")
}

func getFormatIcon(format string) string {
	switch strings.ToLower(format) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return iconImage
	case ".mp4", ".avi", ".mkv", ".webm", ".mov":
		return iconVideo
	case ".gif":
		return "🎞️ "
	default:
		return iconFile
	}
}
