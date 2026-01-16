package main

import (
	"fmt"
	"os"

	"golter/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Get initial path from args or use current directory
	initialPath := ""
	if len(os.Args) > 1 {
		initialPath = os.Args[1]
	}

	model := tui.NewModel(initialPath)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
