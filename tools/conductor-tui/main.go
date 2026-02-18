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

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// placeholder model for initial build verification
type model struct{}

func initialModel() model { return model{} }

func (m model) Init() tea.Cmd { return tea.Quit }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, tea.Quit }

func (m model) View() string { return "Conductor TUI placeholder\n" }
