package tui

import (
	"time"
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

type checkUpdateMsg struct {
	latest    string
	url       string
	available bool
	err       error
}
