package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sametcn99/golter/internal/converter"
	"github.com/sametcn99/golter/internal/version"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

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
