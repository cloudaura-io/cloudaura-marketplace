// Package tui implements the Bubble Tea model, views, and key handling
// for the Conductor TUI application.
package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cloudaura-io/conductor-claude-code/tools/conductor-tui/internal/data"
)

// Version is set at build time via -ldflags.
var Version = "0.2.2"

// Screen types
const (
	ScreenTracks = iota
	ScreenPhases
	ScreenTasks
	ScreenDetail
	ScreenQuit
)

// Screen represents a navigation state in the screen stack.
type Screen struct {
	ScreenType int
	Cursor     int
	Scroll     int
	TrackIdx   int
	PhaseIdx   int
	TaskIdx    int
}

// Model is the Bubble Tea model for the Conductor TUI.
type Model struct {
	BasePath     string
	AllTracks    []data.Track
	ShowArchived bool
	Stack        []Screen
	Width        int
	Height       int
}

// TracksLoadedMsg carries newly loaded tracks.
type TracksLoadedMsg []data.Track

// tickMsg triggers a data refresh.
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// NewModel creates a new Model with default settings.
func NewModel(basePath string) Model {
	return Model{
		BasePath: basePath,
		Stack:    []Screen{{ScreenType: ScreenTracks}},
		Width:    80,
		Height:   24,
	}
}

// CurrentScreen returns the topmost screen on the stack.
func (m Model) CurrentScreen() Screen {
	return m.Stack[len(m.Stack)-1]
}

// Tracks returns the list of tracks filtered by archive visibility.
func (m Model) Tracks() []data.Track {
	if m.ShowArchived {
		return m.AllTracks
	}
	var filtered []data.Track
	for _, t := range m.AllTracks {
		if t.Source != "archived" {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// Init starts the first data load and the tick timer.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.LoadTracks(), tickCmd())
}

// LoadTracks returns a command that discovers tracks from the filesystem.
func (m Model) LoadTracks() tea.Cmd {
	return func() tea.Msg {
		return TracksLoadedMsg(data.DiscoverTracks(m.BasePath))
	}
}

// Update handles all messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case TracksLoadedMsg:
		m.AllTracks = []data.Track(msg)
		return m, nil

	case tickMsg:
		return m, tea.Batch(m.LoadTracks(), tickCmd())

	case tea.KeyMsg:
		return m.HandleKey(msg)
	}
	return m, nil
}

// ItemCount returns the number of items in the current screen's list.
func (m Model) ItemCount() int {
	s := m.CurrentScreen()
	tracks := m.Tracks()
	switch s.ScreenType {
	case ScreenTracks:
		return len(tracks)
	case ScreenPhases:
		if s.TrackIdx < len(tracks) {
			return len(tracks[s.TrackIdx].Phases)
		}
	case ScreenTasks:
		if s.TrackIdx < len(tracks) && s.PhaseIdx < len(tracks[s.TrackIdx].Phases) {
			return len(tracks[s.TrackIdx].Phases[s.PhaseIdx].Tasks)
		}
	}
	return 0
}

// MoveCursor moves the cursor by delta, clamping to valid range.
func (m *Model) MoveCursor(delta int) {
	s := &m.Stack[len(m.Stack)-1]
	max := m.ItemCount()
	next := s.Cursor + delta
	if next < 0 {
		next = 0
	}
	if next >= max {
		next = max - 1
	}
	if next < 0 {
		next = 0
	}
	s.Cursor = next
}

// MoveScroll moves the scroll offset by delta, clamping at zero.
func (m *Model) MoveScroll(delta int) {
	s := &m.Stack[len(m.Stack)-1]
	next := s.Scroll + delta
	if next < 0 {
		next = 0
	}
	s.Scroll = next
}
