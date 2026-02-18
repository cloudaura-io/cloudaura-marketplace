package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	DimStyle    = lipgloss.NewStyle().Faint(true)
	BoldStyle   = lipgloss.NewStyle().Bold(true)
	CursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4")) // blue
)

// ColorStyle returns a lipgloss style for the given color name.
func ColorStyle(c string) lipgloss.Style {
	switch c {
	case "green":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	case "yellow":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	case "cyan":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	case "magenta":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	case "blue":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	case "red":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	case "gray":
		return lipgloss.NewStyle().Faint(true)
	default:
		return lipgloss.NewStyle()
	}
}

// RenderHeader renders the header bar with breadcrumbs and hint text.
func (m Model) RenderHeader(breadcrumbs []string, hint string) string {
	var b strings.Builder

	title := BoldStyle.Render("Conductor TUI") + DimStyle.Render(" v"+Version)
	for _, bc := range breadcrumbs {
		title += " " + DimStyle.Render(">") + " " + bc
	}

	hintStr := DimStyle.Render(hint)
	gap := m.Width - lipgloss.Width(title) - lipgloss.Width(hintStr) - 2
	if gap < 1 {
		gap = 1
	}

	b.WriteString(" " + title + strings.Repeat(" ", gap) + hintStr + "\n")
	sep := strings.Repeat("─", m.Width-4)
	b.WriteString(" " + DimStyle.Render(sep) + "\n")
	return b.String()
}

// RenderFooter renders the footer bar with help text.
func (m Model) RenderFooter(text string) string {
	return " " + DimStyle.Render(text) + "\n"
}
