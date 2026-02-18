package main

// trunc truncates s to max characters, adding "..." if truncated.
func trunc(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return "..."
	}
	return s[:max-3] + "..."
}

// pad pads or truncates s to exactly n characters.
func pad(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + spaces(n-len(s))
}

func spaces(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

// statusColor returns a color name for the given status string.
func statusColor(s string) string {
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

// phaseStatus derives a status string from a Phase's task completion.
func phaseStatus(p Phase) string {
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
