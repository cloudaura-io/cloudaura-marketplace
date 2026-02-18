package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cloudaura-io/cloudaura-marketplace/tools/conductor-tui/internal/data"
)

// HandleKey processes key messages and returns the updated model and command.
func (m Model) HandleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := m.CurrentScreen()

	// Quit confirmation screen
	if s.ScreenType == ScreenQuit {
		switch msg.String() {
		case "y":
			return m, tea.Quit
		case "n", "esc":
			m.Stack = m.Stack[:len(m.Stack)-1]
			return m, nil
		}
		return m, nil
	}

	tracks := m.Tracks()

	switch msg.String() {
	case "up":
		if s.ScreenType == ScreenDetail {
			m.MoveScroll(-1)
		} else {
			m.MoveCursor(-1)
		}
	case "down":
		if s.ScreenType == ScreenDetail {
			m.MoveScroll(1)
		} else {
			m.MoveCursor(1)
		}
	case "enter":
		m.handleEnter(tracks)
	case "esc":
		if len(m.Stack) > 1 {
			m.Stack = m.Stack[:len(m.Stack)-1]
		} else {
			m.Stack = append(m.Stack, Screen{ScreenType: ScreenQuit})
		}
	case "a":
		if s.ScreenType == ScreenTracks {
			m.ShowArchived = !m.ShowArchived
			m.Stack = []Screen{{ScreenType: ScreenTracks}}
		}
	case "q":
		if s.ScreenType == ScreenTracks {
			m.Stack = append(m.Stack, Screen{ScreenType: ScreenQuit})
		}
	}
	return m, nil
}

func (m *Model) handleEnter(tracks []data.Track) {
	s := m.CurrentScreen()
	switch s.ScreenType {
	case ScreenTracks:
		if len(tracks) > 0 && s.Cursor < len(tracks) {
			m.Stack = append(m.Stack, Screen{ScreenType: ScreenPhases, TrackIdx: s.Cursor})
		}
	case ScreenPhases:
		if s.TrackIdx < len(tracks) {
			phases := tracks[s.TrackIdx].Phases
			if len(phases) > 0 && s.Cursor < len(phases) {
				m.Stack = append(m.Stack, Screen{
					ScreenType: ScreenTasks,
					TrackIdx:   s.TrackIdx,
					PhaseIdx:   s.Cursor,
				})
			}
		}
	case ScreenTasks:
		if s.TrackIdx < len(tracks) && s.PhaseIdx < len(tracks[s.TrackIdx].Phases) {
			tasks := tracks[s.TrackIdx].Phases[s.PhaseIdx].Tasks
			if len(tasks) > 0 && s.Cursor < len(tasks) {
				m.Stack = append(m.Stack, Screen{
					ScreenType: ScreenDetail,
					TrackIdx:   s.TrackIdx,
					PhaseIdx:   s.PhaseIdx,
					TaskIdx:    s.Cursor,
				})
			}
		}
	}
}
