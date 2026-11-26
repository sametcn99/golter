package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sametcn99/golter/internal/version"

	"github.com/charmbracelet/bubbles/list"
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
		if m.state == StateSelecting && m.selector.list.FilterState() == list.Filtering && msg.String() != "ctrl+c" {
			var cmd tea.Cmd
			m.selector, cmd = m.selector.Update(msg)
			return m, cmd
		}
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
				m.currentStatus = "Cancelling..."
				if m.cancelCtx != nil {
					m.cancelCtx()
				}
				return m, nil
			}
		}

	case fileFinished:
		m.results[msg.index] = msg.result
		m.progressCurrent++
		m.currentFile = msg.result.path
		if m.batchCtx.Err() == nil {
			m.currentStatus = "Converting files..."
		}
		if m.progressCurrent == m.progressTotal {
			finished := batchResult{m.results, time.Since(m.startTime)}
			m.cancelCtx()
			m.cancelCtx = nil
			m.batchCtx = nil
			return m.Update(finished)
		}
		cmd := m.nextFile()
		return m, cmd

	case batchResult:
		m.state = StateDone
		// Aggregate results
		successCount := 0
		var errs []string
		var successFiles []string

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
			m.output = fmt.Sprintf("Successfully converted %d files in %s\n\n%s",
				successCount, totalDuration,
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

				m.targetFormats, m.actionOptions = selectionOptions(files, m.manager)
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
				if len(m.actionOptions) == 0 {
					return m, nil
				}
				selectedAction := m.actionOptions[m.cursor]

				if strings.Contains(selectedAction, "Convert Format") {
					m.state = StateSelectingFormat
					m.cursor = 0
				} else {
					// Compress Files - keep original format
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
				return m.startBatch(m.targetFormats[m.cursor], "High")
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
				return m.startBatch("", m.qualityOptions[m.cursor])
			}
		}

	case StateConverting:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}
