package main

import "testing"

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
