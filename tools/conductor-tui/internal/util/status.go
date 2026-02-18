package util

import "github.com/cloudaura-io/conductor-claude-code/tools/conductor-tui/internal/data"

// StatusColor returns a color name for the given status string.
func StatusColor(s string) string {
	switch s {
	case "completed", "done":
		return "green"
	case "in_progress", "doing":
		return "yellow"
	case "pending", "todo":
		return "cyan"
	case "new":
		return "magenta"
	case "review":
		return "blue"
	case "blocked":
		return "red"
	case "archived":
		return "gray"
	default:
		return ""
	}
}

// PhaseStatus derives a status string from a Phase's task completion.
func PhaseStatus(p data.Phase) string {
	if len(p.Tasks) == 0 {
		return "empty"
	}
	done := 0
	for _, t := range p.Tasks {
		if t.Completed {
			done++
		}
	}
	if done == len(p.Tasks) {
		return "completed"
	}
	if done > 0 {
		return "in_progress"
	}
	return "pending"
}
