package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Version is set at build time via -ldflags.
var Version = "0.2.2"

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			fmt.Printf("conductor-tui v%s\n", Version)
			os.Exit(0)
		}
	}

	basePath, _ := os.Getwd()
	p := tea.NewProgram(newModel(basePath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
