package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cloudaura-io/conductor-claude-code/tools/conductor-tui/internal/tui"
)

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			fmt.Printf("conductor-tui v%s\n", tui.Version)
			os.Exit(0)
		}
	}

	basePath, _ := os.Getwd()
	p := tea.NewProgram(tui.NewModel(basePath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
