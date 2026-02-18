package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel_InitialState(t *testing.T) {
	m := newModel(".")

	if m.basePath != "." {
		t.Errorf("basePath = %q, want %q", m.basePath, ".")
	}

	if len(m.stack) != 1 {
		t.Fatalf("stack length = %d, want 1", len(m.stack))
	}

	s := m.stack[0]
	if s.screenType != screenTracks {
		t.Errorf("initial screen type = %d, want %d (screenTracks)", s.screenType, screenTracks)
	}
	if s.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", s.cursor)
	}

	if m.showArchived {
		t.Error("showArchived should be false initially")
	}

	if m.width != 80 {
		t.Errorf("default width = %d, want 80", m.width)
	}
	if m.height != 24 {
		t.Errorf("default height = %d, want 24", m.height)
	}
}

func TestNewModel_BasePath(t *testing.T) {
	m := newModel("/some/project")
	if m.basePath != "/some/project" {
		t.Errorf("basePath = %q, want %q", m.basePath, "/some/project")
	}
}

func TestScreen_TracksIsDefault(t *testing.T) {
	m := newModel(".")
	if m.currentScreen().screenType != screenTracks {
		t.Error("current screen should be tracks")
	}
}

// Helper to create a model populated with test tracks.
func testModelWithTracks() tuiModel {
	m := newModel(".")
	m.allTracks = []Track{
		{TrackID: "feature-auth", Type: "feature", Status: "in_progress", Source: "active",
			Phases: []Phase{
				{Number: 1, Name: "Setup", Tasks: []Task{
					{Name: "Init project", Completed: true, Commit: "abc1234"},
					{Name: "Add deps", Completed: false, SubTasks: []SubTask{
						{Name: "Add framework", Completed: true},
						{Name: "Add linter", Completed: false},
					}},
				}},
				{Number: 2, Name: "Implementation", Tasks: []Task{
					{Name: "Build API", Completed: false},
				}},
			}},
		{TrackID: "bugfix-login", Type: "bug", Status: "done", Source: "active",
			Phases: []Phase{
				{Number: 1, Name: "Fix", Tasks: []Task{
					{Name: "Fix login bug", Completed: true, Commit: "def5678"},
				}},
			}},
		{TrackID: "feature-old", Type: "feature", Status: "done", Source: "archived",
			Phases: []Phase{}},
	}
	return m
}

// --- Track Filtering Tests ---

func TestTracks_FilterArchived(t *testing.T) {
	m := testModelWithTracks()

	// Archived tracks hidden by default
	tracks := m.tracks()
	if len(tracks) != 2 {
		t.Fatalf("expected 2 visible tracks, got %d", len(tracks))
	}
	for _, tr := range tracks {
		if tr.Source == "archived" {
			t.Error("archived track should not be visible when showArchived is false")
		}
	}
}

func TestTracks_ShowArchived(t *testing.T) {
	m := testModelWithTracks()
	m.showArchived = true

	tracks := m.tracks()
	if len(tracks) != 3 {
		t.Fatalf("expected 3 tracks when showArchived is true, got %d", len(tracks))
	}
}

// --- Cursor Navigation Tests ---

func TestMoveCursor_Down(t *testing.T) {
	m := testModelWithTracks()

	if m.currentScreen().cursor != 0 {
		t.Fatalf("initial cursor should be 0, got %d", m.currentScreen().cursor)
	}

	m.moveCursor(1)
	if m.currentScreen().cursor != 1 {
		t.Errorf("cursor after down = %d, want 1", m.currentScreen().cursor)
	}
}

func TestMoveCursor_Up(t *testing.T) {
	m := testModelWithTracks()
	m.stack[0].cursor = 1

	m.moveCursor(-1)
	if m.currentScreen().cursor != 0 {
		t.Errorf("cursor after up = %d, want 0", m.currentScreen().cursor)
	}
}

func TestMoveCursor_ClampAtTop(t *testing.T) {
	m := testModelWithTracks()
	m.moveCursor(-1)

	if m.currentScreen().cursor != 0 {
		t.Errorf("cursor should clamp at 0, got %d", m.currentScreen().cursor)
	}
}

func TestMoveCursor_ClampAtBottom(t *testing.T) {
	m := testModelWithTracks()
	// Only 2 visible tracks (archived hidden)
	m.moveCursor(1) // cursor = 1
	m.moveCursor(1) // should clamp at 1

	if m.currentScreen().cursor != 1 {
		t.Errorf("cursor should clamp at 1, got %d", m.currentScreen().cursor)
	}
}

func TestMoveCursor_EmptyList(t *testing.T) {
	m := newModel(".")
	// No tracks at all
	m.moveCursor(1)
	if m.currentScreen().cursor != 0 {
		t.Errorf("cursor should stay at 0 with no items, got %d", m.currentScreen().cursor)
	}
}

// --- Key Handling Tests ---

func TestHandleKey_QuitPrompt(t *testing.T) {
	m := testModelWithTracks()
	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	updated := result.(tuiModel)

	if len(updated.stack) != 2 {
		t.Fatalf("stack length = %d, want 2 after q press", len(updated.stack))
	}
	if updated.currentScreen().screenType != screenQuit {
		t.Errorf("expected quit screen, got %d", updated.currentScreen().screenType)
	}
}

func TestHandleKey_QuitConfirm(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenQuit})

	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
}

func TestHandleKey_QuitCancel(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenQuit})

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated := result.(tuiModel)

	if len(updated.stack) != 1 {
		t.Fatalf("stack length = %d, want 1 after cancel", len(updated.stack))
	}
	if updated.currentScreen().screenType != screenTracks {
		t.Errorf("expected tracks screen after cancel, got %d", updated.currentScreen().screenType)
	}
}

func TestHandleKey_QuitCancelEsc(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenQuit})

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	updated := result.(tuiModel)

	if len(updated.stack) != 1 {
		t.Fatalf("stack length = %d, want 1 after esc on quit screen", len(updated.stack))
	}
}

func TestHandleKey_ArchiveToggle(t *testing.T) {
	m := testModelWithTracks()

	if m.showArchived {
		t.Fatal("showArchived should start as false")
	}

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated := result.(tuiModel)

	if !updated.showArchived {
		t.Error("showArchived should be true after pressing 'a'")
	}

	// Toggling again should hide them
	result, _ = updated.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated = result.(tuiModel)

	if updated.showArchived {
		t.Error("showArchived should be false after second 'a' press")
	}
}

func TestHandleKey_ArchiveToggleResetsCursor(t *testing.T) {
	m := testModelWithTracks()
	m.stack[0].cursor = 1

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated := result.(tuiModel)

	// After toggling archive, cursor and stack should reset
	if updated.currentScreen().cursor != 0 {
		t.Errorf("cursor should reset to 0 after archive toggle, got %d", updated.currentScreen().cursor)
	}
}

func TestHandleKey_ArchiveToggleOnlyOnTracksScreen(t *testing.T) {
	m := testModelWithTracks()
	// Push to phases screen
	m.stack = append(m.stack, screen{screenType: screenPhases, trackIdx: 0})

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated := result.(tuiModel)

	if updated.showArchived {
		t.Error("'a' key should not toggle archive on phases screen")
	}
}

func TestHandleKey_EscOnTracksShowsQuit(t *testing.T) {
	m := testModelWithTracks()

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	updated := result.(tuiModel)

	if len(updated.stack) != 2 {
		t.Fatalf("stack length = %d, want 2", len(updated.stack))
	}
	if updated.currentScreen().screenType != screenQuit {
		t.Errorf("expected quit screen on Esc at tracks level, got %d", updated.currentScreen().screenType)
	}
}

func TestHandleKey_EscOnPhasesGoesBack(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenPhases, trackIdx: 0})

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	updated := result.(tuiModel)

	if len(updated.stack) != 1 {
		t.Fatalf("stack length = %d, want 1 after Esc on phases", len(updated.stack))
	}
	if updated.currentScreen().screenType != screenTracks {
		t.Errorf("expected tracks screen, got %d", updated.currentScreen().screenType)
	}
}

func TestHandleKey_EnterOnTracksPushesPhases(t *testing.T) {
	m := testModelWithTracks()

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(tuiModel)

	if len(updated.stack) != 2 {
		t.Fatalf("stack length = %d, want 2 after Enter", len(updated.stack))
	}
	if updated.currentScreen().screenType != screenPhases {
		t.Errorf("expected phases screen, got %d", updated.currentScreen().screenType)
	}
	if updated.currentScreen().trackIdx != 0 {
		t.Errorf("trackIdx = %d, want 0", updated.currentScreen().trackIdx)
	}
}

func TestHandleKey_EnterOnPhasesPushesTasks(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenPhases, trackIdx: 0})

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(tuiModel)

	if len(updated.stack) != 3 {
		t.Fatalf("stack length = %d, want 3 after Enter on phases", len(updated.stack))
	}
	if updated.currentScreen().screenType != screenTasks {
		t.Errorf("expected tasks screen, got %d", updated.currentScreen().screenType)
	}
}

func TestHandleKey_EnterOnTasksPushesDetail(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenPhases, trackIdx: 0})
	m.stack = append(m.stack, screen{screenType: screenTasks, trackIdx: 0, phaseIdx: 0})

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(tuiModel)

	if len(updated.stack) != 4 {
		t.Fatalf("stack length = %d, want 4 after Enter on tasks", len(updated.stack))
	}
	if updated.currentScreen().screenType != screenDetail {
		t.Errorf("expected detail screen, got %d", updated.currentScreen().screenType)
	}
}

func TestHandleKey_QOnlyOnTracksScreen(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenPhases, trackIdx: 0})

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	updated := result.(tuiModel)

	if len(updated.stack) != 2 {
		t.Error("q key should not trigger quit on phases screen")
	}
}

func TestHandleKey_UpDownNavigation(t *testing.T) {
	m := testModelWithTracks()

	// Move down
	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	updated := result.(tuiModel)
	if updated.currentScreen().cursor != 0 {
		t.Errorf("cursor should stay at 0 on up from 0, got %d", updated.currentScreen().cursor)
	}

	result, _ = updated.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	updated = result.(tuiModel)
	if updated.currentScreen().cursor != 1 {
		t.Errorf("cursor should be 1 after down, got %d", updated.currentScreen().cursor)
	}
}

// --- Item Count Tests ---

func TestItemCount_TracksScreen(t *testing.T) {
	m := testModelWithTracks()
	if m.itemCount() != 2 {
		t.Errorf("itemCount for tracks = %d, want 2 (archived hidden)", m.itemCount())
	}

	m.showArchived = true
	if m.itemCount() != 3 {
		t.Errorf("itemCount for tracks with archived = %d, want 3", m.itemCount())
	}
}

func TestItemCount_PhasesScreen(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenPhases, trackIdx: 0})

	if m.itemCount() != 2 {
		t.Errorf("itemCount for phases = %d, want 2", m.itemCount())
	}
}

func TestItemCount_TasksScreen(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenTasks, trackIdx: 0, phaseIdx: 0})

	if m.itemCount() != 2 {
		t.Errorf("itemCount for tasks = %d, want 2", m.itemCount())
	}
}

func TestItemCount_InvalidTrackIdx(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenPhases, trackIdx: 99})

	if m.itemCount() != 0 {
		t.Errorf("itemCount with invalid trackIdx = %d, want 0", m.itemCount())
	}
}

// --- Update Tests ---

func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := testModelWithTracks()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	updated := result.(tuiModel)

	if updated.width != 120 {
		t.Errorf("width = %d, want 120", updated.width)
	}
	if updated.height != 40 {
		t.Errorf("height = %d, want 40", updated.height)
	}
}

func TestUpdate_TracksLoadedMsg(t *testing.T) {
	m := newModel(".")
	newTracks := []Track{{TrackID: "test-track", Source: "active"}}

	result, _ := m.Update(tracksLoadedMsg(newTracks))
	updated := result.(tuiModel)

	if len(updated.allTracks) != 1 {
		t.Fatalf("allTracks length = %d, want 1", len(updated.allTracks))
	}
	if updated.allTracks[0].TrackID != "test-track" {
		t.Errorf("TrackID = %q, want %q", updated.allTracks[0].TrackID, "test-track")
	}
}

// --- View Tests ---

func TestViewTracks_EmptyState(t *testing.T) {
	m := newModel(".")
	m.width = 80
	m.height = 24

	output := m.viewTracks()
	if !strings.Contains(output, "No tracks found.") {
		t.Error("empty tracks view should contain 'No tracks found.'")
	}
}

func TestViewTracks_Header(t *testing.T) {
	m := testModelWithTracks()
	output := m.viewTracks()

	if !strings.Contains(output, "Conductor TUI") {
		t.Error("tracks view should contain 'Conductor TUI' in header")
	}
}

func TestViewTracks_ShowsTrackData(t *testing.T) {
	m := testModelWithTracks()
	output := m.viewTracks()

	if !strings.Contains(output, "feature-auth") {
		t.Error("tracks view should contain 'feature-auth'")
	}
	if !strings.Contains(output, "bugfix-login") {
		t.Error("tracks view should contain 'bugfix-login'")
	}
}

func TestViewTracks_HidesArchivedByDefault(t *testing.T) {
	m := testModelWithTracks()
	output := m.viewTracks()

	if strings.Contains(output, "feature-old") {
		t.Error("archived track 'feature-old' should not appear when showArchived is false")
	}
}

func TestViewTracks_ShowsArchivedWithStar(t *testing.T) {
	m := testModelWithTracks()
	m.showArchived = true
	output := m.viewTracks()

	if !strings.Contains(output, "feature-old") {
		t.Error("archived track should appear when showArchived is true")
	}
	if !strings.Contains(output, "*") {
		t.Error("archived track should have '*' suffix")
	}
}

func TestViewTracks_Footer(t *testing.T) {
	m := testModelWithTracks()
	output := m.viewTracks()

	if !strings.Contains(output, "Navigate") {
		t.Error("footer should contain navigation hint")
	}
	if !strings.Contains(output, "Show archived") {
		t.Error("footer should say 'Show archived' when archived hidden")
	}

	m.showArchived = true
	output = m.viewTracks()
	if !strings.Contains(output, "Hide archived") {
		t.Error("footer should say 'Hide archived' when archived shown")
	}
}

func TestViewPhases_Content(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenPhases, trackIdx: 0})

	output := m.viewPhases()

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
	m.stack = append(m.stack, screen{screenType: screenTasks, trackIdx: 0, phaseIdx: 0})

	output := m.viewTasks()

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
	m.stack = append(m.stack, screen{
		screenType: screenDetail, trackIdx: 0, phaseIdx: 0, taskIdx: 1,
	})

	output := m.viewDetail()

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
	m.stack = append(m.stack, screen{
		screenType: screenDetail, trackIdx: 0, phaseIdx: 0, taskIdx: 0,
	})

	output := m.viewDetail()

	if !strings.Contains(output, "No sub-tasks.") {
		t.Error("detail view should show 'No sub-tasks.' when task has none")
	}
}

func TestViewQuit(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenQuit})

	output := m.viewQuit()

	if !strings.Contains(output, "Quit Conductor TUI?") {
		t.Error("quit view should contain 'Quit Conductor TUI?'")
	}
	if !strings.Contains(output, "y/n") {
		t.Error("quit view should contain 'y/n'")
	}
}

func TestViewPhases_InvalidTrackIdx(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenPhases, trackIdx: 99})

	output := m.viewPhases()
	if output != "" {
		t.Errorf("viewPhases with invalid trackIdx should return empty string, got %q", output)
	}
}

func TestViewTasks_InvalidIdx(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenTasks, trackIdx: 99})

	output := m.viewTasks()
	if output != "" {
		t.Errorf("viewTasks with invalid trackIdx should return empty string, got %q", output)
	}
}

func TestViewDetail_InvalidIdx(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenDetail, trackIdx: 99})

	output := m.viewDetail()
	if output != "" {
		t.Errorf("viewDetail with invalid trackIdx should return empty string, got %q", output)
	}
}

// --- Scroll Tests (Detail Screen) ---

func TestMoveScroll_Down(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenDetail, trackIdx: 0, phaseIdx: 0, taskIdx: 1})

	m.moveScroll(1)
	if m.currentScreen().scroll != 1 {
		t.Errorf("scroll = %d, want 1", m.currentScreen().scroll)
	}
}

func TestMoveScroll_ClampAtTop(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenDetail, trackIdx: 0, phaseIdx: 0, taskIdx: 1})

	m.moveScroll(-1)
	if m.currentScreen().scroll != 0 {
		t.Errorf("scroll should clamp at 0, got %d", m.currentScreen().scroll)
	}
}

func TestHandleKey_UpDownOnDetailScrolls(t *testing.T) {
	m := testModelWithTracks()
	m.stack = append(m.stack, screen{screenType: screenDetail, trackIdx: 0, phaseIdx: 0, taskIdx: 1})

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	updated := result.(tuiModel)

	if updated.currentScreen().scroll != 1 {
		t.Errorf("scroll = %d, want 1 after down on detail", updated.currentScreen().scroll)
	}
}

// --- View Dispatch Test ---

func TestView_DispatchesCorrectScreen(t *testing.T) {
	m := testModelWithTracks()

	output := m.View()
	if !strings.Contains(output, "Conductor TUI") {
		t.Error("View() on tracks screen should show Conductor TUI header")
	}

	m.stack = append(m.stack, screen{screenType: screenQuit})
	output = m.View()
	if !strings.Contains(output, "Quit Conductor TUI?") {
		t.Error("View() on quit screen should show quit prompt")
	}
}

// --- Enter on Empty Tracks ---

func TestHandleEnter_NoTracksNoOp(t *testing.T) {
	m := newModel(".")
	// No tracks loaded

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(tuiModel)

	if len(updated.stack) != 1 {
		t.Errorf("stack should remain at 1 when pressing Enter with no tracks, got %d", len(updated.stack))
	}
}

// --- Color Style Test ---

func TestColorStyle_ReturnsStyleForKnownColors(t *testing.T) {
	colors := []string{"green", "yellow", "cyan", "magenta", "blue", "red", "gray"}
	for _, c := range colors {
		style := colorStyle(c)
		// Verify style was created (non-nil return)
		_ = style.Render("test")
	}
}

func TestColorStyle_UnknownReturnsDefault(t *testing.T) {
	style := colorStyle("nonexistent")
	result := style.Render("test")
	if !strings.Contains(result, "test") {
		t.Error("unknown color should still render text")
	}
}
