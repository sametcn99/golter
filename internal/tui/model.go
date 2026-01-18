package tui

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sametcn99/golter/internal/converter"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
)

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
	mgr.Register(&converter.DocDataConverter{})
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
