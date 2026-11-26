package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sametcn99/golter/internal/converter"
)

func selectionOptions(files []string, manager *converter.Manager) (targets, actions []string) {
	if len(files) == 0 {
		return
	}
	targets = manager.GetSupportedTargetFormats(filepath.Ext(files[0]))
	compress := true
	for _, file := range files {
		ext := filepath.Ext(file)
		available := manager.GetSupportedTargetFormats(ext)
		targets = slices.DeleteFunc(targets, func(target string) bool { return !slices.Contains(available, target) })
		if _, err := manager.FindConverter(ext, ext); err != nil {
			compress = false
		}
	}
	slices.Sort(targets)
	if len(targets) > 0 {
		actions = append(actions, iconConvert+"  Convert Format")
	}
	if compress {
		actions = append(actions, iconCompress+"  Compress Files")
	}
	return
}

func outputPath(file, target string) string {
	ext := filepath.Ext(file)
	stem := strings.TrimSuffix(file, ext)
	if target == "" {
		return stem + "_compressed" + ext
	}
	if strings.EqualFold(ext, target) {
		return stem + "_converted" + target
	}
	return stem + target
}

func destinationKey(path string) string {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return strings.ToLower(filepath.Clean(path))
	}
	return filepath.Clean(path)
}

func planBatch(files []string, target string) []conversionResult {
	jobs := make([]conversionResult, len(files))
	counts := make(map[string]int, len(files))
	for i, file := range files {
		dst := outputPath(file, target)
		jobs[i] = conversionResult{path: file, outputPath: dst}
		counts[destinationKey(dst)]++
	}
	for i := range jobs {
		job := &jobs[i]
		if counts[destinationKey(job.outputPath)] > 1 {
			job.err = fmt.Errorf("multiple inputs target %s", job.outputPath)
			continue
		}
		_, err := os.Lstat(job.outputPath)
		switch {
		case err == nil:
			job.err = fmt.Errorf("output already exists: %s", job.outputPath)
		case !errors.Is(err, os.ErrNotExist):
			job.err = err
		}
	}
	return jobs
}

func (m Model) startBatch(target, quality string) (tea.Model, tea.Cmd) {
	m.state = StateConverting
	m.startTime = time.Now()
	m.currentStatus = "Starting conversion..."
	m.currentFile = ""
	m.quality = quality
	m.progressCurrent = 0
	m.progressTotal = len(m.selectedFiles)
	m.batchCtx, m.cancelCtx = context.WithCancel(context.Background())
	m.results = planBatch(m.selectedFiles, target)
	m.nextIndex = 0
	var cmds []tea.Cmd
	for range min(4, len(m.results)) {
		cmds = append(cmds, m.nextFile())
	}
	return m, tea.Batch(cmds...)
}

// Bubble Tea schedules another file only after a worker reports completion.
func (m *Model) nextFile() tea.Cmd {
	if m.nextIndex == len(m.results) {
		return nil
	}
	index := m.nextIndex
	m.nextIndex++
	job := m.results[index]
	ctx, manager, quality := m.batchCtx, m.manager, m.quality
	return func() tea.Msg {
		started := time.Now()
		if job.err == nil {
			conv, err := manager.FindConverter(filepath.Ext(job.path), filepath.Ext(job.outputPath))
			if err == nil {
				err = converter.ConvertNew(ctx, conv, job.path, job.outputPath, converter.Options{Quality: quality})
			}
			job.err = err
		}
		job.duration = time.Since(started)
		return fileFinished{index, job}
	}
}
