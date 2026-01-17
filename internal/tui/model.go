package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sametcn99/golter/internal/converter"
	"github.com/sametcn99/golter/internal/version"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// State represents the current state of the application
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

// String returns a human-readable state name
func (s State) String() string {
	switch s {
	case StateSelecting:
		return "File Selection"
	case StateSelectingAction:
		return "Action Selection"
	case StateSelectingFormat:
		return "Format Selection"
	case StateSelectingQuality:
		return "Quality Selection"
	case StateConverting:
		return "Converting"
	case StateDone:
		return "Complete"
	case StateQuitting:
		return "Quitting"
	default:
		return "Unknown"
	}
}

type conversionResult struct {
	path       string
	outputPath string
	err        error
	duration   time.Duration
}

type batchResult struct {
	results  []conversionResult
	duration time.Duration
}

type progressMsg struct {
	current int
	total   int
	file    string
	status  string
}

type tickMsg time.Time

type checkUpdateMsg struct {
	latest    string
	url       string
	available bool
	err       error
}

// Model represents the main application model
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
	currentStatus   string
	output          string
	err             error
	quitting        bool
	manager         *converter.Manager
	targetFormat    string
	width           int
	height          int
	startTime       time.Time
	latestVersion   string
	updateUrl       string
	updateAvailable bool
}

// NewModel creates a new Model with initial configuration
func NewModel(initialPath string) Model {
	mgr := converter.NewManager()
	mgr.Register(&converter.ImageConverter{})
	mgr.Register(&converter.VideoConverter{})
	mgr.Register(&converter.DocumentConverter{})
	mgr.Register(&converter.AudioConverter{})

	if initialPath == "" {
		// Default to user's home directory (cross-platform)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home cannot be determined
			initialPath, _ = filepath.Abs(".")
		} else {
			initialPath = homeDir
		}
	}

	// Ensure initialPath is a directory
	info, err := statFile(initialPath)
	if err == nil && !info.IsDir() {
		initialPath = filepath.Dir(initialPath)
	}

	s := NewSelector(initialPath, mgr.SupportedExtensions())

	// Configure spinner with custom style
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	// Configure progress bar with gradient
	p := progress.New(
		progress.WithScaledGradient("#8B5CF6", "#10B981"),
		progress.WithoutPercentage(),
	)
	p.Width = 40

	return Model{
		state:    StateSelecting,
		selector: s,
		spinner:  sp,
		progress: p,
		manager:  mgr,
		actionOptions: []string{
			iconConvert + "  Convert Format",
			iconCompress + "  Compress Files",
		},
		qualityOptions: []string{
			"✨ High Quality   (Best visual quality, larger files)",
			"⚖️  Balanced      (Good quality, moderate size)",
			"📦 Compact        (Smaller files, reduced quality)",
		},
		width:  80,
		height: 24,
	}
}

// statFile is a helper to get file info
func statFile(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		checkUpdatesCmd,
	)
}

func checkUpdatesCmd() tea.Msg {
	latest, url, available, err := version.CheckForUpdates()
	return checkUpdateMsg{latest, url, available, err}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update progress bar width responsively
		m.progress.Width = msg.Width - 30
		if m.progress.Width > 50 {
			m.progress.Width = 50
		}
		if m.progress.Width < 20 {
			m.progress.Width = 20
		}

	case tea.KeyMsg:
		if m.state == StateQuitting {
			switch msg.String() {
			case "y", "Y", "enter":
				m.quitting = true
				return m, tea.Quit
			case "n", "N", "esc":
				m.state = m.previousState
				return m, nil
			}
			return m, nil
		}

		// Universal back handling
		if msg.String() == "esc" || msg.String() == "backspace" {
			switch m.state {
			case StateSelectingAction:
				m.state = StateSelecting
				m.cursor = 0
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
				m.progressCurrent = 0
				m.progressTotal = 0
				return m, nil
			}
		}

		// Quit handling
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
		m.currentStatus = msg.status
		return m, nil

	case batchResult:
		m.state = StateDone
		// Aggregate results
		successCount := 0
		var errs []string
		var successFiles []string
		var totalSaved int64

		for _, res := range msg.results {
			if res.err != nil {
				errs = append(errs, fmt.Sprintf("  %s %s: %v", iconError, filepath.Base(res.path), res.err))
			} else {
				successCount++
				durationStr := ""
				if res.duration > 0 {
					durationStr = fmt.Sprintf(" (%s)", formatDuration(res.duration))
				}
				successFiles = append(successFiles, fmt.Sprintf("  %s %s %s %s%s",
					iconSuccess,
					filepath.Base(res.path),
					iconArrowRight,
					filepath.Base(res.outputPath),
					durationStr,
				))
			}
		}

		totalDuration := formatDuration(msg.duration)

		if len(errs) > 0 && successCount > 0 {
			// Partial success
			m.output = fmt.Sprintf("Converted %d/%d files in %s\n\n%s\n\nErrors:\n%s",
				successCount,
				len(msg.results),
				totalDuration,
				strings.Join(successFiles, "\n"),
				strings.Join(errs, "\n"),
			)
		} else if len(errs) > 0 {
			// All failed
			m.err = fmt.Errorf("All conversions failed:\n%s", strings.Join(errs, "\n"))
		} else {
			// All success
			savedStr := ""
			if totalSaved > 0 {
				savedStr = fmt.Sprintf(" (saved %s)", FormatSize(totalSaved))
			}
			m.output = fmt.Sprintf("Successfully converted %d files in %s%s\n\n%s",
				successCount,
				totalDuration,
				savedStr,
				strings.Join(successFiles, "\n"),
			)
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case checkUpdateMsg:
		if msg.err == nil && msg.available {
			m.latestVersion = msg.latest
			m.updateUrl = msg.url
			m.updateAvailable = true
		}
		return m, nil
	}

	// State-specific handling
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
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.actionOptions)-1 {
					m.cursor++
				}
			case "enter":
				selectedAction := m.actionOptions[m.cursor]

				if strings.Contains(selectedAction, "Convert Format") {
					ext := filepath.Ext(m.selectedFiles[0])
					m.targetFormats = m.manager.GetSupportedTargetFormats(ext)

					if len(m.targetFormats) == 0 {
						m.err = fmt.Errorf("no supported target formats for %s", ext)
						m.state = StateDone
						return m, nil
					}

					m.state = StateSelectingFormat
					m.cursor = 0
				} else {
					// Compress Files - keep original format
					m.targetFormat = ""
					m.state = StateSelectingQuality
					m.cursor = 0
				}
				return m, nil
			}
		}

	case StateSelectingFormat:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
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
				m.startTime = time.Now()
				m.currentStatus = "Starting conversion..."
				return m, tea.Batch(
					m.spinner.Tick,
					convertFilesWithProgress(m.selectedFiles, m.targetFormat, "High", m.manager),
				)
			}
		}

	case StateSelectingQuality:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
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
				m.startTime = time.Now()
				m.currentStatus = "Starting compression..."
				return m, tea.Batch(
					m.spinner.Tick,
					convertFilesWithProgress(m.selectedFiles, m.targetFormat, quality, m.manager),
				)
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
		startTime := time.Now()
		var results []conversionResult
		var mu sync.Mutex
		var completed int32

		opts := converter.Options{
			"quality": quality,
		}

		// Determine optimal concurrency based on file count
		concurrency := 4
		if len(files) < concurrency {
			concurrency = len(files)
		}

		// Process files concurrently with semaphore
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, concurrency)

		for _, path := range files {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				fileStart := time.Now()

				ext := filepath.Ext(path)
				effectiveTargetExt := targetExt
				if effectiveTargetExt == "" {
					effectiveTargetExt = ext
				}

				conv, err := mgr.FindConverter(ext, effectiveTargetExt)
				if err != nil {
					mu.Lock()
					results = append(results, conversionResult{
						path:     path,
						err:      err,
						duration: time.Since(fileStart),
					})
					atomic.AddInt32(&completed, 1)
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
				duration := time.Since(fileStart)

				mu.Lock()
				results = append(results, conversionResult{
					path:       path,
					outputPath: outputPath,
					err:        err,
					duration:   duration,
				})
				atomic.AddInt32(&completed, 1)
				mu.Unlock()
			}(path)
		}

		wg.Wait()
		return batchResult{
			results:  results,
			duration: time.Since(startTime),
		}
	}
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Header with branding
	versionInfo := fmt.Sprintf(" v%s", version.Current)
	if m.updateAvailable {
		linkText := fmt.Sprintf("(Update available: v%s)", m.latestVersion)
		link := fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", m.updateUrl, linkText)
		versionInfo += " " + link
	}
	header := headerStyle.Render(" " + iconConvert + " Golter - File Converter" + versionInfo + " ")
	s.WriteString("\n" + header + "\n\n")

	switch m.state {
	case StateSelecting:
		m.renderSelectingState(&s)

	case StateSelectingAction:
		m.renderActionState(&s)

	case StateSelectingFormat:
		m.renderFormatState(&s)

	case StateSelectingQuality:
		m.renderQualityState(&s)

	case StateConverting:
		m.renderConvertingState(&s)

	case StateDone:
		m.renderDoneState(&s)

	case StateQuitting:
		m.renderQuittingState(&s)
	}

	// Footer with keyboard shortcuts
	if m.state != StateQuitting && m.state != StateConverting {
		s.WriteString("\n")
		footer := m.renderFooter()
		s.WriteString(footer)
	}

	return s.String()
}

func (m *Model) renderSelectingState(s *strings.Builder) {
	s.WriteString(stateTitleStyle.Render("Select files to convert") + "\n")
	s.WriteString(m.selector.View())
	s.WriteString("\n")

	// Selected files info
	selectedCount := len(m.selector.SelectedFiles())
	if selectedCount < 0 {
		s.WriteString(mutedStyle.Render("  (Files of the same type only can be selected together)") + "\n")
		s.WriteString(mutedStyle.Render("  Select files using Space, then press 'c' to continue") + "\n")
	}
}

func (m *Model) renderActionState(s *strings.Builder) {
	// Get file type info
	ext := filepath.Ext(m.selectedFiles[0])
	supportedTargets := m.manager.GetSupportedTargetFormats(ext)

	// Rebuild action options dynamically
	availableActions := []string{}
	if len(supportedTargets) > 0 {
		availableActions = append(availableActions, iconConvert+"  Convert Format")
	}
	availableActions = append(availableActions, iconCompress+"  Compress Files")
	m.actionOptions = availableActions

	if len(m.actionOptions) == 0 {
		s.WriteString(errorStyle.Render("  No actions available for this file type") + "\n")
		return
	}

	// Header
	fileCount := len(m.selectedFiles)
	fileType := m.selector.GetSelectedFileType()
	s.WriteString(stateTitleStyle.Render(fmt.Sprintf("Choose action for %d %s", fileCount, fileType.String())) + "\n\n")

	// Action list with visual styling (consistent spacing to prevent shift)
	for i, action := range m.actionOptions {
		if m.cursor == i {
			s.WriteString(selectedMenuItemStyle.Render(action) + "\n")
		} else {
			s.WriteString(menuItemStyle.Render(action) + "\n")
		}
	}
}

func (m *Model) renderFormatState(s *strings.Builder) {
	fileCount := len(m.selectedFiles)
	srcExt := filepath.Ext(m.selectedFiles[0])

	s.WriteString(stateTitleStyle.Render(fmt.Sprintf("Select target format (from %s)", srcExt)) + "\n")
	s.WriteString(mutedStyle.Render(fmt.Sprintf("  Converting %d file(s)", fileCount)) + "\n\n")

	for i, format := range m.targetFormats {
		icon := getFormatIcon(format)
		formatDisplay := fmt.Sprintf("%s  %s", icon, format)

		if m.cursor == i {
			s.WriteString(selectedMenuItemStyle.Render(formatDisplay) + "\n")
		} else {
			s.WriteString(menuItemStyle.Render(formatDisplay) + "\n")
		}
	}
}

func (m *Model) renderQualityState(s *strings.Builder) {
	s.WriteString(stateTitleStyle.Render("Select compression quality") + "\n")
	s.WriteString(mutedStyle.Render(fmt.Sprintf("  Compressing %d file(s)", len(m.selectedFiles))) + "\n\n")

	for i, q := range m.qualityOptions {
		if m.cursor == i {
			s.WriteString(selectedMenuItemStyle.Render(q) + "\n")
		} else {
			s.WriteString(menuItemStyle.Render(q) + "\n")
		}
	}
}

func (m *Model) renderConvertingState(s *strings.Builder) {
	s.WriteString(stateTitleStyle.Render("Converting files...") + "\n\n")

	// Spinner and status
	s.WriteString(fmt.Sprintf("  %s %s\n\n", m.spinner.View(), m.currentStatus))

	// Progress bar
	if m.progressTotal > 0 {
		progressPercent := float64(m.progressCurrent) / float64(m.progressTotal)
		s.WriteString("  " + m.progress.ViewAs(progressPercent) + "\n")
		s.WriteString(progressTextStyle.Render(fmt.Sprintf("  %d / %d files processed", m.progressCurrent, m.progressTotal)) + "\n\n")
	}

	// Current file being processed
	if m.currentFile != "" {
		s.WriteString(currentFileStyle.Render(fmt.Sprintf("  Processing: %s", filepath.Base(m.currentFile))) + "\n")
	}

	// Elapsed time
	elapsed := time.Since(m.startTime)
	s.WriteString(mutedStyle.Render(fmt.Sprintf("\n  Elapsed: %s", formatDuration(elapsed))) + "\n")

	// Cancel hint
	s.WriteString("\n" + mutedStyle.Render("  Press Ctrl+C to cancel") + "\n")
}

func (m *Model) renderDoneState(s *strings.Builder) {
	if m.err != nil {
		// Error box
		errorBox := errorBoxStyle.Render(
			errorStyle.Render(fmt.Sprintf("%s Conversion Failed\n\n", iconError)) +
				m.err.Error(),
		)
		s.WriteString(errorBox + "\n")
	} else {
		// Success box
		successBox := successBoxStyle.Render(
			successStyle.Render(fmt.Sprintf("%s Conversion Complete!\n\n", iconSuccess)) +
				m.output,
		)
		s.WriteString(successBox + "\n")
	}
}

func (m *Model) renderQuittingState(s *strings.Builder) {
	confirmBox := confirmStyle.Render(
		infoStyle.Render(iconWarning+" Quit Confirmation") + "\n\n" +
			"Are you sure you want to quit?\n\n" +
			mutedStyle.Render("Press 'y' or Enter to confirm, 'n' or Esc to cancel"),
	)
	s.WriteString(confirmBox + "\n")
}

func (m Model) renderFooter() string {
	var shortcuts []string

	switch m.state {
	case StateSelecting:
		shortcuts = []string{
			RenderHelpKey("↑↓/jk", "Navigate"),
			RenderHelpKey("Space", "Select"),
			RenderHelpKey("a", "Select all"),
			RenderHelpKey("d", "Deselect"),
			RenderHelpKey("Enter", "Open folder"),
			RenderHelpKey("c", "Confirm"),
			RenderHelpKey("/", "Filter"),
			RenderHelpKey("q", "Quit"),
		}
	case StateSelectingAction, StateSelectingFormat, StateSelectingQuality:
		shortcuts = []string{
			RenderHelpKey("↑↓/jk", "Navigate"),
			RenderHelpKey("Enter", "Select"),
			RenderHelpKey("Esc", "Back"),
			RenderHelpKey("q", "Quit"),
		}
	case StateDone:
		shortcuts = []string{
			RenderHelpKey("Esc", "New conversion"),
			RenderHelpKey("q", "Quit"),
		}
	}

	separator := helpSeparatorStyle.Render(" │ ")
	return helpStyle.Render(strings.Join(shortcuts, separator))
}

func getFormatIcon(format string) string {
	switch strings.ToLower(format) {
	case ".jpg", ".jpeg":
		return "🖼️ "
	case ".png":
		return "🖼️ "
	case ".webp":
		return "🖼️ "
	case ".gif":
		return iconGIF
	case ".mp4", ".avi", ".mkv", ".webm", ".mov":
		return iconVideo
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
	case ".mp3", ".wav", ".ogg", ".flac", ".m4a", ".aac":
		return iconAudio
	default:
		return iconFile
	}
}

// formatDuration returns a human-readable duration string
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
}
