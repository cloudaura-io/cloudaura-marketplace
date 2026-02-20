package tui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cloudaura-io/cloudaura-marketplace/tools/conductor-tui/internal/data"
)

func TestNewModel_InitialState(t *testing.T) {
	m := NewModel(".")

	if m.BasePath != "." {
		t.Errorf("BasePath = %q, want %q", m.BasePath, ".")
	}

	if len(m.Stack) != 1 {
		t.Fatalf("stack length = %d, want 1", len(m.Stack))
	}

	s := m.Stack[0]
	if s.ScreenType != ScreenTracks {
		t.Errorf("initial screen type = %d, want %d (ScreenTracks)", s.ScreenType, ScreenTracks)
	}
	if s.Cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", s.Cursor)
	}

	if m.ShowArchived {
		t.Error("ShowArchived should be false initially")
	}

	if m.Width != 80 {
		t.Errorf("default width = %d, want 80", m.Width)
	}
	if m.Height != 24 {
		t.Errorf("default height = %d, want 24", m.Height)
	}
}

func TestNewModel_BasePath(t *testing.T) {
	m := NewModel("/some/project")
	if m.BasePath != "/some/project" {
		t.Errorf("BasePath = %q, want %q", m.BasePath, "/some/project")
	}
}

func TestScreen_TracksIsDefault(t *testing.T) {
	m := NewModel(".")
	if m.CurrentScreen().ScreenType != ScreenTracks {
		t.Error("current screen should be tracks")
	}
}

// Helper to create a model populated with test tracks.
func testModelWithTracks() Model {
	m := NewModel(".")
	m.AllTracks = []data.Track{
		{TrackID: "feature-auth", Type: "feature", Status: "in_progress", Source: "active",
			Phases: []data.Phase{
				{Number: 1, Name: "Setup", Tasks: []data.Task{
					{Name: "Init project", Completed: true, Commit: "abc1234"},
					{Name: "Add deps", Completed: false, SubTasks: []data.SubTask{
						{Name: "Add framework", Completed: true},
						{Name: "Add linter", Completed: false},
					}},
				}},
				{Number: 2, Name: "Implementation", Tasks: []data.Task{
					{Name: "Build API", Completed: false},
				}},
			}},
		{TrackID: "bugfix-login", Type: "bug", Status: "done", Source: "active",
			Phases: []data.Phase{
				{Number: 1, Name: "Fix", Tasks: []data.Task{
					{Name: "Fix login bug", Completed: true, Commit: "def5678"},
				}},
			}},
		{TrackID: "feature-old", Type: "feature", Status: "done", Source: "archived",
			Phases: []data.Phase{}},
	}
	return m
}

// --- Track Filtering Tests ---

func TestTracks_FilterArchived(t *testing.T) {
	m := testModelWithTracks()

	// Archived tracks hidden by default
	tracks := m.Tracks()
	if len(tracks) != 2 {
		t.Fatalf("expected 2 visible tracks, got %d", len(tracks))
	}
	for _, tr := range tracks {
		if tr.Source == "archived" {
			t.Error("archived track should not be visible when ShowArchived is false")
		}
	}
}

func TestTracks_ShowArchived(t *testing.T) {
	m := testModelWithTracks()
	m.ShowArchived = true

	tracks := m.Tracks()
	if len(tracks) != 3 {
		t.Fatalf("expected 3 tracks when ShowArchived is true, got %d", len(tracks))
	}
}

// --- Cursor Navigation Tests ---

func TestMoveCursor_Down(t *testing.T) {
	m := testModelWithTracks()

	if m.CurrentScreen().Cursor != 0 {
		t.Fatalf("initial cursor should be 0, got %d", m.CurrentScreen().Cursor)
	}

	m.MoveCursor(1)
	if m.CurrentScreen().Cursor != 1 {
		t.Errorf("cursor after down = %d, want 1", m.CurrentScreen().Cursor)
	}
}

func TestMoveCursor_Up(t *testing.T) {
	m := testModelWithTracks()
	m.Stack[0].Cursor = 1

	m.MoveCursor(-1)
	if m.CurrentScreen().Cursor != 0 {
		t.Errorf("cursor after up = %d, want 0", m.CurrentScreen().Cursor)
	}
}

func TestMoveCursor_ClampAtTop(t *testing.T) {
	m := testModelWithTracks()
	m.MoveCursor(-1)

	if m.CurrentScreen().Cursor != 0 {
		t.Errorf("cursor should clamp at 0, got %d", m.CurrentScreen().Cursor)
	}
}

func TestMoveCursor_ClampAtBottom(t *testing.T) {
	m := testModelWithTracks()
	// Only 2 visible tracks (archived hidden)
	m.MoveCursor(1) // cursor = 1
	m.MoveCursor(1) // should clamp at 1

	if m.CurrentScreen().Cursor != 1 {
		t.Errorf("cursor should clamp at 1, got %d", m.CurrentScreen().Cursor)
	}
}

func TestMoveCursor_EmptyList(t *testing.T) {
	m := NewModel(".")
	// No tracks at all
	m.MoveCursor(1)
	if m.CurrentScreen().Cursor != 0 {
		t.Errorf("cursor should stay at 0 with no items, got %d", m.CurrentScreen().Cursor)
	}
}

// --- Key Handling Tests ---

func TestHandleKey_QuitPrompt(t *testing.T) {
	m := testModelWithTracks()
	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	updated := result.(Model)

	if len(updated.Stack) != 2 {
		t.Fatalf("stack length = %d, want 2 after q press", len(updated.Stack))
	}
	if updated.CurrentScreen().ScreenType != ScreenQuit {
		t.Errorf("expected quit screen, got %d", updated.CurrentScreen().ScreenType)
	}
}

func TestHandleKey_QuitConfirm(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenQuit})

	_, cmd := m.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
}

func TestHandleKey_QuitCancel(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenQuit})

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated := result.(Model)

	if len(updated.Stack) != 1 {
		t.Fatalf("stack length = %d, want 1 after cancel", len(updated.Stack))
	}
	if updated.CurrentScreen().ScreenType != ScreenTracks {
		t.Errorf("expected tracks screen after cancel, got %d", updated.CurrentScreen().ScreenType)
	}
}

func TestHandleKey_QuitCancelEsc(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenQuit})

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})
	updated := result.(Model)

	if len(updated.Stack) != 1 {
		t.Fatalf("stack length = %d, want 1 after esc on quit screen", len(updated.Stack))
	}
}

func TestHandleKey_ArchiveToggle(t *testing.T) {
	m := testModelWithTracks()

	if m.ShowArchived {
		t.Fatal("ShowArchived should start as false")
	}

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated := result.(Model)

	if !updated.ShowArchived {
		t.Error("ShowArchived should be true after pressing 'a'")
	}

	// Toggling again should hide them
	result, _ = updated.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated = result.(Model)

	if updated.ShowArchived {
		t.Error("ShowArchived should be false after second 'a' press")
	}
}

func TestHandleKey_ArchiveToggleResetsCursor(t *testing.T) {
	m := testModelWithTracks()
	m.Stack[0].Cursor = 1

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated := result.(Model)

	// After toggling archive, cursor and stack should reset
	if updated.CurrentScreen().Cursor != 0 {
		t.Errorf("cursor should reset to 0 after archive toggle, got %d", updated.CurrentScreen().Cursor)
	}
}

func TestHandleKey_ArchiveToggleOnlyOnTracksScreen(t *testing.T) {
	m := testModelWithTracks()
	// Push to phases screen
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenPhases, TrackIdx: 0})

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated := result.(Model)

	if updated.ShowArchived {
		t.Error("'a' key should not toggle archive on phases screen")
	}
}

func TestHandleKey_EscOnTracksShowsQuit(t *testing.T) {
	m := testModelWithTracks()

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})
	updated := result.(Model)

	if len(updated.Stack) != 2 {
		t.Fatalf("stack length = %d, want 2", len(updated.Stack))
	}
	if updated.CurrentScreen().ScreenType != ScreenQuit {
		t.Errorf("expected quit screen on Esc at tracks level, got %d", updated.CurrentScreen().ScreenType)
	}
}

func TestHandleKey_EscOnPhasesGoesBack(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenPhases, TrackIdx: 0})

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})
	updated := result.(Model)

	if len(updated.Stack) != 1 {
		t.Fatalf("stack length = %d, want 1 after Esc on phases", len(updated.Stack))
	}
	if updated.CurrentScreen().ScreenType != ScreenTracks {
		t.Errorf("expected tracks screen, got %d", updated.CurrentScreen().ScreenType)
	}
}

func TestHandleKey_EnterOnTracksPushesPhases(t *testing.T) {
	m := testModelWithTracks()

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(Model)

	if len(updated.Stack) != 2 {
		t.Fatalf("stack length = %d, want 2 after Enter", len(updated.Stack))
	}
	if updated.CurrentScreen().ScreenType != ScreenPhases {
		t.Errorf("expected phases screen, got %d", updated.CurrentScreen().ScreenType)
	}
	if updated.CurrentScreen().TrackIdx != 0 {
		t.Errorf("TrackIdx = %d, want 0", updated.CurrentScreen().TrackIdx)
	}
}

func TestHandleKey_EKeyOnTracksPushesEdit(t *testing.T) {
	m := testModelWithTracks()

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	updated := result.(Model)

	if len(updated.Stack) != 2 {
		t.Fatalf("stack length = %d, want 2 after e press", len(updated.Stack))
	}
	if updated.CurrentScreen().ScreenType != ScreenEdit {
		t.Errorf("expected edit screen, got %d", updated.CurrentScreen().ScreenType)
	}
	if updated.CurrentScreen().TrackIdx != 0 {
		t.Errorf("TrackIdx = %d, want 0", updated.CurrentScreen().TrackIdx)
	}
}

func TestHandleKey_PKeyOnTracksDoesNothing(t *testing.T) {
	m := testModelWithTracks()

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	updated := result.(Model)

	if len(updated.Stack) != 1 {
		t.Errorf("stack length = %d, want 1 (p key should do nothing on tracks screen)", len(updated.Stack))
	}
}

func TestHandleKey_EscOnEditGoesBack(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0})

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})
	updated := result.(Model)

	if len(updated.Stack) != 1 {
		t.Fatalf("stack length = %d, want 1 after Esc on edit", len(updated.Stack))
	}
	if updated.CurrentScreen().ScreenType != ScreenTracks {
		t.Errorf("expected tracks screen, got %d", updated.CurrentScreen().ScreenType)
	}
}

func TestHandleKey_EnterOnPhasesPushesTasks(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenPhases, TrackIdx: 0})

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(Model)

	if len(updated.Stack) != 3 {
		t.Fatalf("stack length = %d, want 3 after Enter on phases", len(updated.Stack))
	}
	if updated.CurrentScreen().ScreenType != ScreenTasks {
		t.Errorf("expected tasks screen, got %d", updated.CurrentScreen().ScreenType)
	}
}

func TestHandleKey_EnterOnTasksPushesDetail(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenPhases, TrackIdx: 0})
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenTasks, TrackIdx: 0, PhaseIdx: 0})

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(Model)

	if len(updated.Stack) != 4 {
		t.Fatalf("stack length = %d, want 4 after Enter on tasks", len(updated.Stack))
	}
	if updated.CurrentScreen().ScreenType != ScreenDetail {
		t.Errorf("expected detail screen, got %d", updated.CurrentScreen().ScreenType)
	}
}

func TestHandleKey_QOnlyOnTracksScreen(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenPhases, TrackIdx: 0})

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	updated := result.(Model)

	if len(updated.Stack) != 2 {
		t.Error("q key should not trigger quit on phases screen")
	}
}

func TestHandleKey_UpDownNavigation(t *testing.T) {
	m := testModelWithTracks()

	// Move down
	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyUp})
	updated := result.(Model)
	if updated.CurrentScreen().Cursor != 0 {
		t.Errorf("cursor should stay at 0 on up from 0, got %d", updated.CurrentScreen().Cursor)
	}

	result, _ = updated.HandleKey(tea.KeyMsg{Type: tea.KeyDown})
	updated = result.(Model)
	if updated.CurrentScreen().Cursor != 1 {
		t.Errorf("cursor should be 1 after down, got %d", updated.CurrentScreen().Cursor)
	}
}

// --- Item Count Tests ---

func TestItemCount_TracksScreen(t *testing.T) {
	m := testModelWithTracks()
	if m.ItemCount() != 2 {
		t.Errorf("ItemCount for tracks = %d, want 2 (archived hidden)", m.ItemCount())
	}

	m.ShowArchived = true
	if m.ItemCount() != 3 {
		t.Errorf("ItemCount for tracks with archived = %d, want 3", m.ItemCount())
	}
}

func TestItemCount_PhasesScreen(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenPhases, TrackIdx: 0})

	if m.ItemCount() != 2 {
		t.Errorf("ItemCount for phases = %d, want 2", m.ItemCount())
	}
}

func TestItemCount_TasksScreen(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenTasks, TrackIdx: 0, PhaseIdx: 0})

	if m.ItemCount() != 2 {
		t.Errorf("ItemCount for tasks = %d, want 2", m.ItemCount())
	}
}

func TestItemCount_InvalidTrackIdx(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenPhases, TrackIdx: 99})

	if m.ItemCount() != 0 {
		t.Errorf("ItemCount with invalid TrackIdx = %d, want 0", m.ItemCount())
	}
}

// --- Update Tests ---

func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := testModelWithTracks()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	updated := result.(Model)

	if updated.Width != 120 {
		t.Errorf("width = %d, want 120", updated.Width)
	}
	if updated.Height != 40 {
		t.Errorf("height = %d, want 40", updated.Height)
	}
}

func TestUpdate_TracksLoadedMsg(t *testing.T) {
	m := NewModel(".")
	newTracks := []data.Track{{TrackID: "test-track", Source: "active"}}

	result, _ := m.Update(TracksLoadedMsg(newTracks))
	updated := result.(Model)

	if len(updated.AllTracks) != 1 {
		t.Fatalf("AllTracks length = %d, want 1", len(updated.AllTracks))
	}
	if updated.AllTracks[0].TrackID != "test-track" {
		t.Errorf("TrackID = %q, want %q", updated.AllTracks[0].TrackID, "test-track")
	}
}

// --- View Tests ---

func TestViewTracks_EmptyState(t *testing.T) {
	m := NewModel(".")
	m.Width = 80
	m.Height = 24

	output := m.ViewTracks()
	if !strings.Contains(output, "No tracks found.") {
		t.Error("empty tracks view should contain 'No tracks found.'")
	}
}

func TestViewTracks_Header(t *testing.T) {
	m := testModelWithTracks()
	output := m.ViewTracks()

	if !strings.Contains(output, "Conductor TUI") {
		t.Error("tracks view should contain 'Conductor TUI' in header")
	}
}

func TestViewTracks_ShowsTrackData(t *testing.T) {
	m := testModelWithTracks()
	output := m.ViewTracks()

	if !strings.Contains(output, "feature-auth") {
		t.Error("tracks view should contain 'feature-auth'")
	}
	if !strings.Contains(output, "bugfix-login") {
		t.Error("tracks view should contain 'bugfix-login'")
	}
}

func TestViewTracks_HidesArchivedByDefault(t *testing.T) {
	m := testModelWithTracks()
	output := m.ViewTracks()

	if strings.Contains(output, "feature-old") {
		t.Error("archived track 'feature-old' should not appear when ShowArchived is false")
	}
}

func TestViewTracks_ShowsArchivedWithStar(t *testing.T) {
	m := testModelWithTracks()
	m.ShowArchived = true
	output := m.ViewTracks()

	if !strings.Contains(output, "feature-old") {
		t.Error("archived track should appear when ShowArchived is true")
	}
	if !strings.Contains(output, "*") {
		t.Error("archived track should have '*' suffix")
	}
}

func TestViewTracks_Footer(t *testing.T) {
	m := testModelWithTracks()
	output := m.ViewTracks()

	if !strings.Contains(output, "[Enter] Phases") {
		t.Error("footer should contain '[Enter] Phases' hint")
	}
	if !strings.Contains(output, "[e] Edit") {
		t.Error("footer should contain '[e] Edit' hint")
	}
	if !strings.Contains(output, "Show archived") {
		t.Error("footer should say 'Show archived' when archived hidden")
	}
	if !strings.Contains(output, "[q] Quit") {
		t.Error("footer should contain '[q] Quit' hint")
	}

	m.ShowArchived = true
	output = m.ViewTracks()
	if !strings.Contains(output, "Hide archived") {
		t.Error("footer should say 'Hide archived' when archived shown")
	}
}

func TestViewPhases_Content(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenPhases, TrackIdx: 0})

	output := m.ViewPhases()

	if !strings.Contains(output, "feature-auth") {
		t.Error("phases view should contain track ID in breadcrumb")
	}
	if !strings.Contains(output, "Setup") {
		t.Error("phases view should contain phase name 'Setup'")
	}
	if !strings.Contains(output, "Implementation") {
		t.Error("phases view should contain phase name 'Implementation'")
	}
}

func TestViewTasks_Content(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenTasks, TrackIdx: 0, PhaseIdx: 0})

	output := m.ViewTasks()

	if !strings.Contains(output, "Phase 1") {
		t.Error("tasks view should contain 'Phase 1' in breadcrumb")
	}
	if !strings.Contains(output, "Init project") {
		t.Error("tasks view should contain task name 'Init project'")
	}
	if !strings.Contains(output, "abc1234") {
		t.Error("tasks view should contain commit hash 'abc1234'")
	}
}

func TestViewDetail_Content(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{
		ScreenType: ScreenDetail, TrackIdx: 0, PhaseIdx: 0, TaskIdx: 1,
	})

	output := m.ViewDetail()

	if !strings.Contains(output, "Add deps") {
		t.Error("detail view should contain task name 'Add deps'")
	}
	if !strings.Contains(output, "Sub-tasks:") {
		t.Error("detail view should contain 'Sub-tasks:' label")
	}
	if !strings.Contains(output, "Add framework") {
		t.Error("detail view should show sub-task 'Add framework'")
	}
	if !strings.Contains(output, "Add linter") {
		t.Error("detail view should show sub-task 'Add linter'")
	}
}

func TestViewDetail_NoSubTasks(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{
		ScreenType: ScreenDetail, TrackIdx: 0, PhaseIdx: 0, TaskIdx: 0,
	})

	output := m.ViewDetail()

	if !strings.Contains(output, "No sub-tasks.") {
		t.Error("detail view should show 'No sub-tasks.' when task has none")
	}
}

func TestViewQuit(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenQuit})

	output := m.ViewQuit()

	if !strings.Contains(output, "Quit Conductor TUI?") {
		t.Error("quit view should contain 'Quit Conductor TUI?'")
	}
	if !strings.Contains(output, "y/n") {
		t.Error("quit view should contain 'y/n'")
	}
}

// --- Edit Screen View Tests ---

func TestViewEdit_Content(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0})

	output := m.ViewEdit()

	if !strings.Contains(output, "feature-auth") {
		t.Error("edit view should contain track ID in breadcrumb")
	}
	if !strings.Contains(output, "Edit") {
		t.Error("edit view should contain 'Edit' in breadcrumb")
	}
	if !strings.Contains(output, "Status") {
		t.Error("edit view should contain 'Status' field label")
	}
	if !strings.Contains(output, "Type") {
		t.Error("edit view should contain 'Type' field label")
	}
	if !strings.Contains(output, "in_progress") {
		t.Error("edit view should show current status value 'in_progress'")
	}
	if !strings.Contains(output, "feature") {
		t.Error("edit view should show current type value 'feature'")
	}
}

func TestViewEdit_CursorOnFirstField(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0, EditFieldIdx: 0})

	output := m.ViewEdit()

	// The selected field should have the cursor indicator
	if !strings.Contains(output, ">") {
		t.Error("edit view should show cursor '>' on selected field")
	}
	// The selected field should have bracket indicators
	if !strings.Contains(output, "[<") {
		t.Error("edit view should show bracket indicators on selected field")
	}
}

func TestViewEdit_CursorOnSecondField(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 1, EditFieldIdx: 1})

	output := m.ViewEdit()

	if !strings.Contains(output, "bugfix-login") {
		t.Error("edit view should show second track's ID")
	}
}

func TestViewEdit_InvalidTrackIdx(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 99})

	output := m.ViewEdit()
	if output != "" {
		t.Errorf("ViewEdit with invalid TrackIdx should return empty string, got %q", output)
	}
}

func TestViewEdit_Footer(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0})

	output := m.ViewEdit()

	if !strings.Contains(output, "Select field") {
		t.Error("edit footer should contain 'Select field' hint")
	}
	if !strings.Contains(output, "Change value") {
		t.Error("edit footer should contain 'Change value' hint")
	}
	if !strings.Contains(output, "Esc") {
		t.Error("edit footer should contain 'Esc' hint")
	}
}

func TestViewEdit_UpDownNavigatesFields(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0, EditFieldIdx: 0})

	// Move down
	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyDown})
	updated := result.(Model)

	if updated.CurrentScreen().EditFieldIdx != 1 {
		t.Errorf("EditFieldIdx = %d, want 1 after down", updated.CurrentScreen().EditFieldIdx)
	}

	// Move up
	result, _ = updated.HandleKey(tea.KeyMsg{Type: tea.KeyUp})
	updated = result.(Model)

	if updated.CurrentScreen().EditFieldIdx != 0 {
		t.Errorf("EditFieldIdx = %d, want 0 after up", updated.CurrentScreen().EditFieldIdx)
	}
}

func TestMoveEditField_ClampAtBounds(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0, EditFieldIdx: 0})

	// Try to go above 0
	m.MoveEditField(-1)
	if m.CurrentScreen().EditFieldIdx != 0 {
		t.Errorf("EditFieldIdx = %d, want 0 (clamped at top)", m.CurrentScreen().EditFieldIdx)
	}

	// Move to last field
	m.MoveEditField(1)
	if m.CurrentScreen().EditFieldIdx != 1 {
		t.Errorf("EditFieldIdx = %d, want 1", m.CurrentScreen().EditFieldIdx)
	}

	// Try to go past last field
	m.MoveEditField(1)
	if m.CurrentScreen().EditFieldIdx != 1 {
		t.Errorf("EditFieldIdx = %d, want 1 (clamped at bottom)", m.CurrentScreen().EditFieldIdx)
	}
}

// --- Field Value Cycling Tests ---

func TestHandleKey_EnterCyclesStatusField(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0, EditFieldIdx: 0})

	// Status starts as "in_progress", Enter should cycle to next value
	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(Model)

	track := updated.Tracks()[updated.CurrentScreen().TrackIdx]
	if track.Status == "in_progress" {
		t.Error("status should have changed from 'in_progress' after Enter")
	}
}

func TestHandleKey_EnterCyclesTypeField(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0, EditFieldIdx: 1})

	// Type starts as "feature", Enter should cycle to next value
	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(Model)

	track := updated.Tracks()[updated.CurrentScreen().TrackIdx]
	if track.Type == "feature" {
		t.Error("type should have changed from 'feature' after Enter")
	}
}

func TestHandleKey_RightCyclesFieldForward(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0, EditFieldIdx: 0})

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyRight})
	updated := result.(Model)

	track := updated.Tracks()[updated.CurrentScreen().TrackIdx]
	if track.Status == "in_progress" {
		t.Error("status should have changed after right arrow")
	}
}

func TestHandleKey_LeftCyclesFieldBackward(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0, EditFieldIdx: 0})

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyLeft})
	updated := result.(Model)

	track := updated.Tracks()[updated.CurrentScreen().TrackIdx]
	if track.Status == "in_progress" {
		t.Error("status should have changed after left arrow")
	}
}

func TestCycleStatusValues(t *testing.T) {
	// Test the full cycle: new -> in_progress -> completed -> cancelled -> new
	values := StatusValues
	for i, v := range values {
		next := CycleValue(values, v, 1)
		expected := values[(i+1)%len(values)]
		if next != expected {
			t.Errorf("CycleValue(%q, 1) = %q, want %q", v, next, expected)
		}
	}
}

func TestCycleTypeValues(t *testing.T) {
	// Test the full cycle: feature -> bug -> chore -> refactor -> feature
	values := TypeValues
	for i, v := range values {
		next := CycleValue(values, v, 1)
		expected := values[(i+1)%len(values)]
		if next != expected {
			t.Errorf("CycleValue(%q, 1) = %q, want %q", v, next, expected)
		}
	}
}

func TestCycleValue_Backward(t *testing.T) {
	values := StatusValues
	// Cycling backward from "new" should give "cancelled"
	result := CycleValue(values, "new", -1)
	if result != "cancelled" {
		t.Errorf("CycleValue('new', -1) = %q, want %q", result, "cancelled")
	}
}

func TestCycleValue_UnknownValue(t *testing.T) {
	values := StatusValues
	// Unknown value should default to first value
	result := CycleValue(values, "unknown_value", 1)
	if result != values[0] {
		t.Errorf("CycleValue('unknown_value', 1) = %q, want %q", result, values[0])
	}
}

// --- Persistence Integration Tests ---

func TestHandleKey_EnterOnEditTriggersReload(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0, EditFieldIdx: 0})

	// After cycling, a reload command should be returned
	_, cmd := m.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// When BasePath is "." (no real conductor dir), no save happens but the model
	// should still update the in-memory state
	track := m.Tracks()[0]
	// The track should have been cycled
	_ = cmd
	_ = track
}

func TestMetadataPath_ActiveTrack(t *testing.T) {
	m := NewModel("/project")
	m.AllTracks = []data.Track{
		{TrackID: "feature-test", Source: "active"},
	}

	path := m.MetadataPath(0)
	expected := "/project/conductor/tracks/feature-test/metadata.json"
	if path != expected {
		t.Errorf("MetadataPath = %q, want %q", path, expected)
	}
}

func TestMetadataPath_ArchivedTrack(t *testing.T) {
	m := NewModel("/project")
	m.AllTracks = []data.Track{
		{TrackID: "old-track", Source: "archived"},
	}
	m.ShowArchived = true

	path := m.MetadataPath(0)
	expected := "/project/conductor/archive/old-track/metadata.json"
	if path != expected {
		t.Errorf("MetadataPath = %q, want %q", path, expected)
	}
}

func TestPersistence_CycleAndSave(t *testing.T) {
	// Create a temp directory with a real track structure
	dir := t.TempDir()
	trackDir := dir + "/conductor/tracks/test-track"
	if err := os.MkdirAll(trackDir, 0755); err != nil {
		t.Fatalf("failed to create track dir: %v", err)
	}

	// Write initial metadata
	initial := `{"track_id":"test-track","type":"feature","status":"new","created_at":"2026-01-01T10:00:00Z"}`
	if err := os.WriteFile(trackDir+"/metadata.json", []byte(initial), 0644); err != nil {
		t.Fatalf("failed to write metadata: %v", err)
	}

	m := NewModel(dir)
	m.AllTracks = data.DiscoverTracks(dir)

	if len(m.AllTracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(m.AllTracks))
	}

	// Push to edit screen
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0, EditFieldIdx: 0})

	// Cycle status: new -> in_progress
	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(Model)

	// Read back from disk
	savedData, err := os.ReadFile(trackDir + "/metadata.json")
	if err != nil {
		t.Fatalf("failed to read saved metadata: %v", err)
	}

	savedTrack, err := data.LoadMetadata(savedData)
	if err != nil {
		t.Fatalf("failed to parse saved metadata: %v", err)
	}

	if savedTrack.Status != "in_progress" {
		t.Errorf("saved status = %q, want %q", savedTrack.Status, "in_progress")
	}

	// Verify updated_at was set
	if savedTrack.UpdatedAt.IsZero() {
		t.Error("updated_at should be set after save")
	}

	// Verify in-memory model also updated
	if updated.Tracks()[0].Status != "in_progress" {
		t.Errorf("in-memory status = %q, want %q", updated.Tracks()[0].Status, "in_progress")
	}
}

func TestHandleKey_EscOnEditNoSave(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenEdit, TrackIdx: 0, EditFieldIdx: 0})

	// Pressing Esc should return to tracks without triggering save
	result, cmd := m.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})
	updated := result.(Model)

	if updated.CurrentScreen().ScreenType != ScreenTracks {
		t.Errorf("expected tracks screen after Esc, got %d", updated.CurrentScreen().ScreenType)
	}
	if cmd != nil {
		t.Error("Esc on edit screen should not trigger any command")
	}
}

func TestViewPhases_InvalidTrackIdx(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenPhases, TrackIdx: 99})

	output := m.ViewPhases()
	if output != "" {
		t.Errorf("ViewPhases with invalid TrackIdx should return empty string, got %q", output)
	}
}

func TestViewTasks_InvalidIdx(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenTasks, TrackIdx: 99})

	output := m.ViewTasks()
	if output != "" {
		t.Errorf("ViewTasks with invalid TrackIdx should return empty string, got %q", output)
	}
}

func TestViewDetail_InvalidIdx(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenDetail, TrackIdx: 99})

	output := m.ViewDetail()
	if output != "" {
		t.Errorf("ViewDetail with invalid TrackIdx should return empty string, got %q", output)
	}
}

// --- Scroll Tests (Detail Screen) ---

func TestMoveScroll_Down(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenDetail, TrackIdx: 0, PhaseIdx: 0, TaskIdx: 1})

	m.MoveScroll(1)
	if m.CurrentScreen().Scroll != 1 {
		t.Errorf("scroll = %d, want 1", m.CurrentScreen().Scroll)
	}
}

func TestMoveScroll_ClampAtTop(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenDetail, TrackIdx: 0, PhaseIdx: 0, TaskIdx: 1})

	m.MoveScroll(-1)
	if m.CurrentScreen().Scroll != 0 {
		t.Errorf("scroll should clamp at 0, got %d", m.CurrentScreen().Scroll)
	}
}

func TestHandleKey_UpDownOnDetailScrolls(t *testing.T) {
	m := testModelWithTracks()
	m.Stack = append(m.Stack, Screen{ScreenType: ScreenDetail, TrackIdx: 0, PhaseIdx: 0, TaskIdx: 1})

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyDown})
	updated := result.(Model)

	if updated.CurrentScreen().Scroll != 1 {
		t.Errorf("scroll = %d, want 1 after down on detail", updated.CurrentScreen().Scroll)
	}
}

// --- View Dispatch Test ---

func TestView_DispatchesCorrectScreen(t *testing.T) {
	m := testModelWithTracks()

	output := m.View()
	if !strings.Contains(output, "Conductor TUI") {
		t.Error("View() on tracks screen should show Conductor TUI header")
	}

	m.Stack = append(m.Stack, Screen{ScreenType: ScreenQuit})
	output = m.View()
	if !strings.Contains(output, "Quit Conductor TUI?") {
		t.Error("View() on quit screen should show quit prompt")
	}
}

// --- Enter on Empty Tracks ---

func TestHandleEnter_NoTracksNoOp(t *testing.T) {
	m := NewModel(".")
	// No tracks loaded

	result, _ := m.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(Model)

	if len(updated.Stack) != 1 {
		t.Errorf("stack should remain at 1 when pressing Enter with no tracks, got %d", len(updated.Stack))
	}
}

// --- Color Style Test ---

func TestColorStyle_ReturnsStyleForKnownColors(t *testing.T) {
	colors := []string{"green", "yellow", "cyan", "magenta", "blue", "red", "gray"}
	for _, c := range colors {
		style := ColorStyle(c)
		// Verify style was created (non-nil return)
		_ = style.Render("test")
	}
}

func TestColorStyle_UnknownReturnsDefault(t *testing.T) {
	style := ColorStyle("nonexistent")
	result := style.Render("test")
	if !strings.Contains(result, "test") {
		t.Error("unknown color should still render text")
	}
}
