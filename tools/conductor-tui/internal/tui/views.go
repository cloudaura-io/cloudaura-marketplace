package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/cloudaura-io/cloudaura-marketplace/tools/conductor-tui/internal/util"
)

// View renders the current screen.
func (m Model) View() string {
	s := m.CurrentScreen()

	if s.ScreenType == ScreenQuit {
		return m.ViewQuit()
	}

	switch s.ScreenType {
	case ScreenTracks:
		return m.ViewTracks()
	case ScreenPhases:
		return m.ViewPhases()
	case ScreenTasks:
		return m.ViewTasks()
	case ScreenDetail:
		return m.ViewDetail()
	}
	return ""
}

// ViewQuit renders the quit confirmation prompt.
func (m Model) ViewQuit() string {
	prompt := BoldStyle.Render("Quit Conductor TUI? ") + DimStyle.Render("[y/n]")
	lines := make([]string, m.Height)
	mid := m.Height / 2
	for i := range lines {
		if i == mid {
			w := lipgloss.Width(prompt)
			leftPad := (m.Width - w) / 2
			if leftPad < 0 {
				leftPad = 0
			}
			lines[i] = strings.Repeat(" ", leftPad) + prompt
		} else {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}

// ViewTracks renders the tracks list screen.
func (m Model) ViewTracks() string {
	tracks := m.Tracks()
	s := m.CurrentScreen()

	var b strings.Builder
	b.WriteString(m.RenderHeader(nil, "[q] Quit"))

	if len(tracks) == 0 {
		b.WriteString(" " + DimStyle.Render("No tracks found.") + "\n")
		footer := fmt.Sprintf("[q] Quit")
		b.WriteString(m.RenderFooter(footer))
		return b.String()
	}

	maxVis := m.Height - 6
	if maxVis < 1 {
		maxVis = 1
	}

	scroll := s.Cursor - maxVis + 1
	if scroll < 0 {
		scroll = 0
	}
	end := scroll + maxVis
	if end > len(tracks) {
		end = len(tracks)
	}
	visible := tracks[scroll:end]

	descW := m.Width - 64
	if descW < 8 {
		descW = 8
	}

	// Column headers
	b.WriteString(DimStyle.Render("  "+util.Pad("Track ID", 28)+util.Pad("Type", 10)+util.Pad("Status", 14)+util.Pad("Phases", 8)+"Description") + "\n")

	for i, t := range visible {
		idx := scroll + i
		sel := idx == s.Cursor

		tag := ""
		if t.Source == "archived" {
			tag = " *"
		}

		prefix := "  "
		if sel {
			prefix = CursorStyle.Render("> ")
		}

		statusStr := t.Status + tag
		statusRendered := ColorStyle(util.StatusColor(t.Status)).Render(util.Pad(statusStr, 14))

		line := prefix +
			util.Pad(util.Trunc(t.TrackID, 26), 28) +
			util.Pad(t.Type, 10) +
			statusRendered +
			util.Pad(fmt.Sprintf("%d", len(t.Phases)), 8) +
			util.Trunc(t.Description, descW)

		if sel {
			line = BoldStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}

	archiveHint := "Show"
	if m.ShowArchived {
		archiveHint = "Hide"
	}
	footer := fmt.Sprintf("[↑↓] Navigate  [Enter] View phases  [a] %s archived  [q] Quit", archiveHint)
	b.WriteString(m.RenderFooter(footer))
	return b.String()
}

// ViewPhases renders the phases list for a selected track.
func (m Model) ViewPhases() string {
	tracks := m.Tracks()
	s := m.CurrentScreen()

	if s.TrackIdx >= len(tracks) {
		return ""
	}
	track := tracks[s.TrackIdx]

	var b strings.Builder
	b.WriteString(m.RenderHeader([]string{track.TrackID}, "[Esc] Back"))
	b.WriteString(" " + DimStyle.Render(track.Description) + "\n")

	maxVis := m.Height - 7
	if maxVis < 1 {
		maxVis = 1
	}

	scroll := s.Cursor - maxVis + 1
	if scroll < 0 {
		scroll = 0
	}
	end := scroll + maxVis
	if end > len(track.Phases) {
		end = len(track.Phases)
	}
	visible := track.Phases[scroll:end]

	b.WriteString(DimStyle.Render("  "+util.Pad("#", 4)+util.Pad("Phase", 34)+util.Pad("Tasks", 10)+"Status") + "\n")

	for i, p := range visible {
		idx := scroll + i
		sel := idx == s.Cursor

		done := 0
		for _, t := range p.Tasks {
			if t.Completed {
				done++
			}
		}
		st := util.PhaseStatus(p)

		prefix := "  "
		if sel {
			prefix = CursorStyle.Render("> ")
		}

		statusRendered := ColorStyle(util.StatusColor(st)).Render(st)

		line := prefix +
			util.Pad(fmt.Sprintf("%d", p.Number), 4) +
			util.Pad(util.Trunc(p.Name, 32), 34) +
			util.Pad(fmt.Sprintf("%d/%d", done, len(p.Tasks)), 10) +
			statusRendered

		if sel {
			line = BoldStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString(m.RenderFooter("[↑↓] Navigate  [Enter] View tasks  [Esc] Back"))
	return b.String()
}

// ViewTasks renders the tasks list for a selected phase.
func (m Model) ViewTasks() string {
	tracks := m.Tracks()
	s := m.CurrentScreen()

	if s.TrackIdx >= len(tracks) || s.PhaseIdx >= len(tracks[s.TrackIdx].Phases) {
		return ""
	}
	track := tracks[s.TrackIdx]
	phase := track.Phases[s.PhaseIdx]

	var b strings.Builder
	b.WriteString(m.RenderHeader(
		[]string{util.Trunc(track.TrackID, 20), fmt.Sprintf("Phase %d", phase.Number)},
		"[Esc] Back",
	))
	b.WriteString(" " + DimStyle.Render(phase.Name) + "\n")

	maxVis := m.Height - 7
	if maxVis < 1 {
		maxVis = 1
	}

	scroll := s.Cursor - maxVis + 1
	if scroll < 0 {
		scroll = 0
	}
	end := scroll + maxVis
	if end > len(phase.Tasks) {
		end = len(phase.Tasks)
	}
	visible := phase.Tasks[scroll:end]

	b.WriteString(DimStyle.Render("  "+util.Pad("#", 4)+util.Pad("Task", 42)+util.Pad("Status", 10)+"Commit") + "\n")

	for i, t := range visible {
		idx := scroll + i
		sel := idx == s.Cursor

		st := "pending"
		if t.Completed {
			st = "done"
		}

		commit := "—"
		if t.Commit != "" {
			commit = t.Commit
		}

		prefix := "  "
		if sel {
			prefix = CursorStyle.Render("> ")
		}

		statusRendered := ColorStyle(util.StatusColor(st)).Render(util.Pad(st, 10))

		line := prefix +
			util.Pad(fmt.Sprintf("%d", idx+1), 4) +
			util.Pad(util.Trunc(t.Name, 40), 42) +
			statusRendered +
			commit

		if sel {
			line = BoldStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString(m.RenderFooter("[↑↓] Navigate  [Enter] View detail  [Esc] Back"))
	return b.String()
}

// ViewDetail renders the detail view for a selected task.
func (m Model) ViewDetail() string {
	tracks := m.Tracks()
	s := m.CurrentScreen()

	if s.TrackIdx >= len(tracks) ||
		s.PhaseIdx >= len(tracks[s.TrackIdx].Phases) ||
		s.TaskIdx >= len(tracks[s.TrackIdx].Phases[s.PhaseIdx].Tasks) {
		return ""
	}
	track := tracks[s.TrackIdx]
	phase := track.Phases[s.PhaseIdx]
	task := phase.Tasks[s.TaskIdx]

	st := "pending"
	if task.Completed {
		st = "completed"
	}

	var b strings.Builder
	b.WriteString(m.RenderHeader(
		[]string{
			util.Trunc(track.TrackID, 16),
			fmt.Sprintf("Phase %d", phase.Number),
			"Task: " + util.Trunc(task.Name, 30),
		},
		"[Esc] Back",
	))

	b.WriteString(" " + BoldStyle.Render("Task: ") + task.Name + "\n")

	statusLine := " Status: " + ColorStyle(util.StatusColor(st)).Render(st)
	if task.Commit != "" {
		statusLine += "          Commit: " + BoldStyle.Render(task.Commit)
	}
	b.WriteString(statusLine + "\n\n")

	if len(task.SubTasks) == 0 {
		b.WriteString(" " + DimStyle.Render("No sub-tasks.") + "\n")
	} else {
		b.WriteString(" " + BoldStyle.Render("Sub-tasks:") + "\n")

		maxSub := m.Height - 10
		if maxSub < 1 {
			maxSub = 1
		}

		scrollIdx := s.Scroll
		maxScroll := len(task.SubTasks) - maxSub
		if maxScroll < 0 {
			maxScroll = 0
		}
		if scrollIdx > maxScroll {
			scrollIdx = maxScroll
		}

		end := scrollIdx + maxSub
		if end > len(task.SubTasks) {
			end = len(task.SubTasks)
		}
		visibleSubs := task.SubTasks[scrollIdx:end]

		for _, sub := range visibleSubs {
			check := "[ ]"
			if sub.Completed {
				check = ColorStyle("green").Render("[x]")
			}
			b.WriteString("    " + check + " " + sub.Name + "\n")
		}
	}

	footerText := "[Esc] Back"
	if len(task.SubTasks) > m.Height-10 {
		footerText = "[↑↓] Scroll  [Esc] Back"
	}
	b.WriteString(m.RenderFooter(footerText))
	return b.String()
}
