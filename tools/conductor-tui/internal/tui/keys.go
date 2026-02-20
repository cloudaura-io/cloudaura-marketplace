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
		} else if s.ScreenType == ScreenEdit {
			m.MoveEditField(-1)
		} else {
			m.MoveCursor(-1)
		}
	case "down":
		if s.ScreenType == ScreenDetail {
			m.MoveScroll(1)
		} else if s.ScreenType == ScreenEdit {
			m.MoveEditField(1)
		} else {
			m.MoveCursor(1)
		}
	case "enter":
		if s.ScreenType == ScreenEdit {
			m.cycleEditField(1)
			m.saveCurrentTrack()
		} else {
			m.handleEnter(tracks)
		}
	case "right":
		if s.ScreenType == ScreenEdit {
			m.cycleEditField(1)
			m.saveCurrentTrack()
		}
	case "left":
		if s.ScreenType == ScreenEdit {
			m.cycleEditField(-1)
			m.saveCurrentTrack()
		}
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
	case "p":
		if s.ScreenType == ScreenTracks {
			if len(tracks) > 0 && s.Cursor < len(tracks) {
				m.Stack = append(m.Stack, Screen{ScreenType: ScreenPhases, TrackIdx: s.Cursor})
			}
		}
	case "q":
		if s.ScreenType == ScreenTracks {
			m.Stack = append(m.Stack, Screen{ScreenType: ScreenQuit})
		}
	}
	return m, nil
}

// cycleEditField cycles the value of the currently selected edit field.
func (m *Model) cycleEditField(delta int) {
	s := m.CurrentScreen()
	tracks := m.Tracks()
	if s.TrackIdx >= len(tracks) {
		return
	}
	track := &m.AllTracks[m.resolveTrackIndex(s.TrackIdx)]

	switch s.EditFieldIdx {
	case 0: // Status
		track.Status = CycleValue(StatusValues, track.Status, delta)
	case 1: // Type
		track.Type = CycleValue(TypeValues, track.Type, delta)
	}
}

// saveCurrentTrack persists the current track's metadata to disk.
func (m *Model) saveCurrentTrack() {
	s := m.CurrentScreen()
	path := m.MetadataPath(s.TrackIdx)
	if path == "" {
		return
	}
	tracks := m.Tracks()
	if s.TrackIdx >= len(tracks) {
		return
	}
	track := tracks[s.TrackIdx]
	// Best-effort save; errors are silently ignored in the TUI
	_ = data.SaveMetadata(path, track)
}

// resolveTrackIndex maps a filtered track index to the AllTracks index.
func (m *Model) resolveTrackIndex(filteredIdx int) int {
	tracks := m.Tracks()
	if filteredIdx >= len(tracks) {
		return 0
	}
	target := tracks[filteredIdx]
	for i, t := range m.AllTracks {
		if t.TrackID == target.TrackID {
			return i
		}
	}
	return 0
}

func (m *Model) handleEnter(tracks []data.Track) {
	s := m.CurrentScreen()
	switch s.ScreenType {
	case ScreenTracks:
		if len(tracks) > 0 && s.Cursor < len(tracks) {
			m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: s.Cursor})
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
