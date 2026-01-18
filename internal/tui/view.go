package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sametcn99/golter/internal/version"
)

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
