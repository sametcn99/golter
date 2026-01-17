package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	m := NewModel(".")
	if m.state != StateSelecting {
		t.Errorf("Initial state should be StateSelecting, got %v", m.state)
	}
	if m.err != nil {
		t.Errorf("Initial error should be nil, got %v", m.err)
	}
	if m.width != 80 || m.height != 24 {
		t.Errorf("Initial dimensions should be 80x24, got %dx%d", m.width, m.height)
	}
}

func TestModel_Init(t *testing.T) {
	m := NewModel(".")
	cmd := m.Init()
	if cmd == nil {
		// Init usually returns nil or a command
	}
}

func TestModel_Update(t *testing.T) {
	m := NewModel(".")

	// Test window resize
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, cmd := m.Update(msg)

	newM, ok := updatedModel.(Model)
	if !ok {
		t.Fatal("Update did not return Model")
	}

	if newM.width != 100 || newM.height != 50 {
		t.Errorf("Window size not updated. Got %dx%d", newM.width, newM.height)
	}
	if cmd != nil {
		// Resize usually doesn't return a command
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "500ms"},
		{1500 * time.Millisecond, "1.5s"},
		{65 * time.Second, "1m 5s"},
	}

	for _, tt := range tests {
		if got := formatDuration(tt.d); got != tt.want {
			t.Errorf("formatDuration(%v) = %v, want %v", tt.d, got, tt.want)
		}
	}
}
