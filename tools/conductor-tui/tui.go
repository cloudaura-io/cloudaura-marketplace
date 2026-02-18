package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen types
const (
	screenTracks = iota
	screenPhases
	screenTasks
	screenDetail
	screenQuit
)

// screen represents a navigation state in the screen stack.
type screen struct {
	screenType int
	cursor     int
	scroll     int
	trackIdx   int
	phaseIdx   int
	taskIdx    int
}

// tuiModel is the Bubble Tea model for the Conductor TUI.
type tuiModel struct {
	basePath     string
	allTracks    []Track
	showArchived bool
	stack        []screen
	width        int
	height       int
}

// tickMsg triggers a data refresh.
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func newModel(basePath string) tuiModel {
	return tuiModel{
		basePath: basePath,
		stack:    []screen{{screenType: screenTracks}},
		width:    80,
		height:   24,
	}
}

func (m tuiModel) currentScreen() screen {
	return m.stack[len(m.stack)-1]
}

func (m tuiModel) tracks() []Track {
	if m.showArchived {
		return m.allTracks
	}
	var filtered []Track
	for _, t := range m.allTracks {
		if t.Source != "archived" {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// Init starts the first data load and the tick timer.
func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(m.loadTracks(), tickCmd())
}

func (m tuiModel) loadTracks() tea.Cmd {
	return func() tea.Msg {
		return tracksLoadedMsg(discoverTracks(m.basePath))
	}
}

type tracksLoadedMsg []Track

// Update handles all messages.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tracksLoadedMsg:
		m.allTracks = []Track(msg)
		return m, nil

	case tickMsg:
		return m, tea.Batch(m.loadTracks(), tickCmd())

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m tuiModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := m.currentScreen()

	// Quit confirmation screen
	if s.screenType == screenQuit {
		switch msg.String() {
		case "y":
			return m, tea.Quit
		case "n", "esc":
			m.stack = m.stack[:len(m.stack)-1]
			return m, nil
		}
		return m, nil
	}

	tracks := m.tracks()

	switch msg.String() {
	case "up":
		if s.screenType == screenDetail {
			m.moveScroll(-1)
		} else {
			m.moveCursor(-1)
		}
	case "down":
		if s.screenType == screenDetail {
			m.moveScroll(1)
		} else {
			m.moveCursor(1)
		}
	case "enter":
		m.handleEnter(tracks)
	case "esc":
		if len(m.stack) > 1 {
			m.stack = m.stack[:len(m.stack)-1]
		} else {
			m.stack = append(m.stack, screen{screenType: screenQuit})
		}
	case "a":
		if s.screenType == screenTracks {
			m.showArchived = !m.showArchived
			m.stack = []screen{{screenType: screenTracks}}
		}
	case "q":
		if s.screenType == screenTracks {
			m.stack = append(m.stack, screen{screenType: screenQuit})
		}
	}
	return m, nil
}

func (m *tuiModel) moveCursor(delta int) {
	s := &m.stack[len(m.stack)-1]
	max := m.itemCount()
	next := s.cursor + delta
	if next < 0 {
		next = 0
	}
	if next >= max {
		next = max - 1
	}
	if next < 0 {
		next = 0
	}
	s.cursor = next
}

func (m *tuiModel) moveScroll(delta int) {
	s := &m.stack[len(m.stack)-1]
	next := s.scroll + delta
	if next < 0 {
		next = 0
	}
	s.scroll = next
}

func (m tuiModel) itemCount() int {
	s := m.currentScreen()
	tracks := m.tracks()
	switch s.screenType {
	case screenTracks:
		return len(tracks)
	case screenPhases:
		if s.trackIdx < len(tracks) {
			return len(tracks[s.trackIdx].Phases)
		}
	case screenTasks:
		if s.trackIdx < len(tracks) && s.phaseIdx < len(tracks[s.trackIdx].Phases) {
			return len(tracks[s.trackIdx].Phases[s.phaseIdx].Tasks)
		}
	}
	return 0
}

func (m *tuiModel) handleEnter(tracks []Track) {
	s := m.currentScreen()
	switch s.screenType {
	case screenTracks:
		if len(tracks) > 0 && s.cursor < len(tracks) {
			m.stack = append(m.stack, screen{screenType: screenPhases, trackIdx: s.cursor})
		}
	case screenPhases:
		if s.trackIdx < len(tracks) {
			phases := tracks[s.trackIdx].Phases
			if len(phases) > 0 && s.cursor < len(phases) {
				m.stack = append(m.stack, screen{
					screenType: screenTasks,
					trackIdx:   s.trackIdx,
					phaseIdx:   s.cursor,
				})
			}
		}
	case screenTasks:
		if s.trackIdx < len(tracks) && s.phaseIdx < len(tracks[s.trackIdx].Phases) {
			tasks := tracks[s.trackIdx].Phases[s.phaseIdx].Tasks
			if len(tasks) > 0 && s.cursor < len(tasks) {
				m.stack = append(m.stack, screen{
					screenType: screenDetail,
					trackIdx:   s.trackIdx,
					phaseIdx:   s.phaseIdx,
					taskIdx:    s.cursor,
				})
			}
		}
	}
}

// Styles
var (
	dimStyle    = lipgloss.NewStyle().Faint(true)
	boldStyle   = lipgloss.NewStyle().Bold(true)
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4")) // blue
)

func colorStyle(c string) lipgloss.Style {
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

// View renders the current screen.
func (m tuiModel) View() string {
	s := m.currentScreen()

	if s.screenType == screenQuit {
		return m.viewQuit()
	}

	switch s.screenType {
	case screenTracks:
		return m.viewTracks()
	case screenPhases:
		return m.viewPhases()
	case screenTasks:
		return m.viewTasks()
	case screenDetail:
		return m.viewDetail()
	}
	return ""
}

func (m tuiModel) renderHeader(breadcrumbs []string, hint string) string {
	var b strings.Builder

	title := boldStyle.Render("Conductor TUI") + dimStyle.Render(" v"+Version)
	for _, bc := range breadcrumbs {
		title += " " + dimStyle.Render(">") + " " + bc
	}

	hintStr := dimStyle.Render(hint)
	gap := m.width - lipgloss.Width(title) - lipgloss.Width(hintStr) - 2
	if gap < 1 {
		gap = 1
	}

	b.WriteString(" " + title + strings.Repeat(" ", gap) + hintStr + "\n")
	sep := strings.Repeat("─", m.width-4)
	b.WriteString(" " + dimStyle.Render(sep) + "\n")
	return b.String()
}

func (m tuiModel) renderFooter(text string) string {
	return " " + dimStyle.Render(text) + "\n"
}

func (m tuiModel) viewQuit() string {
	// Center the quit prompt
	prompt := boldStyle.Render("Quit Conductor TUI? ") + dimStyle.Render("[y/n]")
	lines := make([]string, m.height)
	mid := m.height / 2
	for i := range lines {
		if i == mid {
			w := lipgloss.Width(prompt)
			leftPad := (m.width - w) / 2
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

func (m tuiModel) viewTracks() string {
	tracks := m.tracks()
	s := m.currentScreen()

	var b strings.Builder
	b.WriteString(m.renderHeader(nil, "[q] Quit"))

	if len(tracks) == 0 {
		b.WriteString(" " + dimStyle.Render("No tracks found.") + "\n")
		footer := fmt.Sprintf("[q] Quit")
		b.WriteString(m.renderFooter(footer))
		return b.String()
	}

	maxVis := m.height - 6
	if maxVis < 1 {
		maxVis = 1
	}

	scroll := s.cursor - maxVis + 1
	if scroll < 0 {
		scroll = 0
	}
	end := scroll + maxVis
	if end > len(tracks) {
		end = len(tracks)
	}
	visible := tracks[scroll:end]

	descW := m.width - 64
	if descW < 8 {
		descW = 8
	}

	// Column headers
	b.WriteString(dimStyle.Render("  "+pad("Track ID", 28)+pad("Type", 10)+pad("Status", 14)+pad("Phases", 8)+"Description") + "\n")

	for i, t := range visible {
		idx := scroll + i
		sel := idx == s.cursor

		tag := ""
		if t.Source == "archived" {
			tag = " *"
		}

		prefix := "  "
		if sel {
			prefix = cursorStyle.Render("> ")
		}

		statusStr := t.Status + tag
		statusRendered := colorStyle(statusColor(t.Status)).Render(pad(statusStr, 14))

		line := prefix +
			pad(trunc(t.TrackID, 26), 28) +
			pad(t.Type, 10) +
			statusRendered +
			pad(fmt.Sprintf("%d", len(t.Phases)), 8) +
			trunc(t.Description, descW)

		if sel {
			line = boldStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}

	archiveHint := "Show"
	if m.showArchived {
		archiveHint = "Hide"
	}
	footer := fmt.Sprintf("[↑↓] Navigate  [Enter] View phases  [a] %s archived  [q] Quit", archiveHint)
	b.WriteString(m.renderFooter(footer))
	return b.String()
}

func (m tuiModel) viewPhases() string {
	tracks := m.tracks()
	s := m.currentScreen()

	if s.trackIdx >= len(tracks) {
		return ""
	}
	track := tracks[s.trackIdx]

	var b strings.Builder
	b.WriteString(m.renderHeader([]string{track.TrackID}, "[Esc] Back"))
	b.WriteString(" " + dimStyle.Render(track.Description) + "\n")

	maxVis := m.height - 7
	if maxVis < 1 {
		maxVis = 1
	}

	scroll := s.cursor - maxVis + 1
	if scroll < 0 {
		scroll = 0
	}
	end := scroll + maxVis
	if end > len(track.Phases) {
		end = len(track.Phases)
	}
	visible := track.Phases[scroll:end]

	b.WriteString(dimStyle.Render("  "+pad("#", 4)+pad("Phase", 34)+pad("Tasks", 10)+"Status") + "\n")

	for i, p := range visible {
		idx := scroll + i
		sel := idx == s.cursor

		done := 0
		for _, t := range p.Tasks {
			if t.Completed {
				done++
			}
		}
		st := phaseStatus(p)

		prefix := "  "
		if sel {
			prefix = cursorStyle.Render("> ")
		}

		statusRendered := colorStyle(statusColor(st)).Render(st)

		line := prefix +
			pad(fmt.Sprintf("%d", p.Number), 4) +
			pad(trunc(p.Name, 32), 34) +
			pad(fmt.Sprintf("%d/%d", done, len(p.Tasks)), 10) +
			statusRendered

		if sel {
			line = boldStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString(m.renderFooter("[↑↓] Navigate  [Enter] View tasks  [Esc] Back"))
	return b.String()
}

func (m tuiModel) viewTasks() string {
	tracks := m.tracks()
	s := m.currentScreen()

	if s.trackIdx >= len(tracks) || s.phaseIdx >= len(tracks[s.trackIdx].Phases) {
		return ""
	}
	track := tracks[s.trackIdx]
	phase := track.Phases[s.phaseIdx]

	var b strings.Builder
	b.WriteString(m.renderHeader(
		[]string{trunc(track.TrackID, 20), fmt.Sprintf("Phase %d", phase.Number)},
		"[Esc] Back",
	))
	b.WriteString(" " + dimStyle.Render(phase.Name) + "\n")

	maxVis := m.height - 7
	if maxVis < 1 {
		maxVis = 1
	}

	scroll := s.cursor - maxVis + 1
	if scroll < 0 {
		scroll = 0
	}
	end := scroll + maxVis
	if end > len(phase.Tasks) {
		end = len(phase.Tasks)
	}
	visible := phase.Tasks[scroll:end]

	b.WriteString(dimStyle.Render("  "+pad("#", 4)+pad("Task", 42)+pad("Status", 10)+"Commit") + "\n")

	for i, t := range visible {
		idx := scroll + i
		sel := idx == s.cursor

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
			prefix = cursorStyle.Render("> ")
		}

		statusRendered := colorStyle(statusColor(st)).Render(pad(st, 10))

		line := prefix +
			pad(fmt.Sprintf("%d", idx+1), 4) +
			pad(trunc(t.Name, 40), 42) +
			statusRendered +
			commit

		if sel {
			line = boldStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString(m.renderFooter("[↑↓] Navigate  [Enter] View detail  [Esc] Back"))
	return b.String()
}

func (m tuiModel) viewDetail() string {
	tracks := m.tracks()
	s := m.currentScreen()

	if s.trackIdx >= len(tracks) ||
		s.phaseIdx >= len(tracks[s.trackIdx].Phases) ||
		s.taskIdx >= len(tracks[s.trackIdx].Phases[s.phaseIdx].Tasks) {
		return ""
	}
	track := tracks[s.trackIdx]
	phase := track.Phases[s.phaseIdx]
	task := phase.Tasks[s.taskIdx]

	st := "pending"
	if task.Completed {
		st = "completed"
	}

	var b strings.Builder
	b.WriteString(m.renderHeader(
		[]string{
			trunc(track.TrackID, 16),
			fmt.Sprintf("Phase %d", phase.Number),
			"Task: " + trunc(task.Name, 30),
		},
		"[Esc] Back",
	))

	b.WriteString(" " + boldStyle.Render("Task: ") + task.Name + "\n")

	statusLine := " Status: " + colorStyle(statusColor(st)).Render(st)
	if task.Commit != "" {
		statusLine += "          Commit: " + boldStyle.Render(task.Commit)
	}
	b.WriteString(statusLine + "\n\n")

	if len(task.SubTasks) == 0 {
		b.WriteString(" " + dimStyle.Render("No sub-tasks.") + "\n")
	} else {
		b.WriteString(" " + boldStyle.Render("Sub-tasks:") + "\n")

		maxSub := m.height - 10
		if maxSub < 1 {
			maxSub = 1
		}

		scrollIdx := s.scroll
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
				check = colorStyle("green").Render("[x]")
			}
			b.WriteString("    " + check + " " + sub.Name + "\n")
		}
	}

	footerText := "[Esc] Back"
	if len(task.SubTasks) > m.height-10 {
		footerText = "[↑↓] Scroll  [Esc] Back"
	}
	b.WriteString(m.renderFooter(footerText))
	return b.String()
}
