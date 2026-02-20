package data

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMetadata_Valid(t *testing.T) {
	data, err := os.ReadFile("../../testdata/valid_metadata.json")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	track, err := LoadMetadata(data)
	if err != nil {
		t.Fatalf("LoadMetadata returned error: %v", err)
	}

	if track.TrackID != "feature-auth_20260101" {
		t.Errorf("TrackID = %q, want %q", track.TrackID, "feature-auth_20260101")
	}
	if track.Type != "feature" {
		t.Errorf("Type = %q, want %q", track.Type, "feature")
	}
	if track.Status != "in_progress" {
		t.Errorf("Status = %q, want %q", track.Status, "in_progress")
	}
	if track.Description != "Add authentication system" {
		t.Errorf("Description = %q, want %q", track.Description, "Add authentication system")
	}
}

func TestLoadMetadata_FallbackValues(t *testing.T) {
	data, err := os.ReadFile("../../testdata/partial_metadata.json")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	track, err := LoadMetadata(data)
	if err != nil {
		t.Fatalf("LoadMetadata returned error: %v", err)
	}

	if track.TrackID != "bugfix-login_20260105" {
		t.Errorf("TrackID = %q, want %q", track.TrackID, "bugfix-login_20260105")
	}
	if track.Type != "unknown" {
		t.Errorf("Type = %q, want %q (fallback)", track.Type, "unknown")
	}
	if track.Status != "unknown" {
		t.Errorf("Status = %q, want %q (fallback)", track.Status, "unknown")
	}
	if track.Description != "" {
		t.Errorf("Description = %q, want %q (fallback)", track.Description, "")
	}
}

func TestLoadMetadata_InvalidJSON(t *testing.T) {
	data, err := os.ReadFile("../../testdata/invalid_metadata.json")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	_, err = LoadMetadata(data)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoadMetadata_FallbackTrackID(t *testing.T) {
	// When track_id is empty, LoadMetadata should still succeed;
	// the caller (DiscoverTracks) handles fallback to directory name.
	data := []byte(`{"type": "bug"}`)
	track, err := LoadMetadata(data)
	if err != nil {
		t.Fatalf("LoadMetadata returned error: %v", err)
	}
	if track.TrackID != "" {
		t.Errorf("TrackID = %q, want empty string", track.TrackID)
	}
	if track.Type != "bug" {
		t.Errorf("Type = %q, want %q", track.Type, "bug")
	}
}

func TestLoadMetadata_CreatedAt(t *testing.T) {
	data, err := os.ReadFile("../../testdata/valid_metadata.json")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	track, err := LoadMetadata(data)
	if err != nil {
		t.Fatalf("LoadMetadata returned error: %v", err)
	}

	expected := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	if !track.CreatedAt.Equal(expected) {
		t.Errorf("CreatedAt = %v, want %v", track.CreatedAt, expected)
	}

	expectedUpdated := time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC)
	if !track.UpdatedAt.Equal(expectedUpdated) {
		t.Errorf("UpdatedAt = %v, want %v", track.UpdatedAt, expectedUpdated)
	}
}

func TestLoadMetadata_MissingCreatedAt(t *testing.T) {
	data := []byte(`{"track_id": "test", "type": "bug"}`)
	track, err := LoadMetadata(data)
	if err != nil {
		t.Fatalf("LoadMetadata returned error: %v", err)
	}

	if !track.CreatedAt.IsZero() {
		t.Errorf("CreatedAt should be zero value when missing, got %v", track.CreatedAt)
	}
	if !track.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt should be zero value when missing, got %v", track.UpdatedAt)
	}
}

func TestSaveMetadata_WritesFile(t *testing.T) {
	dir := t.TempDir()
	metaPath := filepath.Join(dir, "metadata.json")

	track := Track{
		TrackID:     "test-track",
		Type:        "feature",
		Status:      "in_progress",
		Description: "Test track",
		CreatedAt:   time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
	}

	err := SaveMetadata(metaPath, track)
	if err != nil {
		t.Fatalf("SaveMetadata returned error: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	loaded, err := LoadMetadata(data)
	if err != nil {
		t.Fatalf("LoadMetadata of written file returned error: %v", err)
	}

	if loaded.TrackID != "test-track" {
		t.Errorf("TrackID = %q, want %q", loaded.TrackID, "test-track")
	}
	if loaded.Type != "feature" {
		t.Errorf("Type = %q, want %q", loaded.Type, "feature")
	}
	if loaded.Status != "in_progress" {
		t.Errorf("Status = %q, want %q", loaded.Status, "in_progress")
	}
	if loaded.Description != "Test track" {
		t.Errorf("Description = %q, want %q", loaded.Description, "Test track")
	}
}

func TestSaveMetadata_UpdatesTimestamp(t *testing.T) {
	dir := t.TempDir()
	metaPath := filepath.Join(dir, "metadata.json")

	track := Track{
		TrackID:   "test-track",
		Type:      "bug",
		Status:    "new",
		CreatedAt: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
	}

	before := time.Now().UTC().Truncate(time.Second)
	err := SaveMetadata(metaPath, track)
	if err != nil {
		t.Fatalf("SaveMetadata returned error: %v", err)
	}

	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	loaded, err := LoadMetadata(data)
	if err != nil {
		t.Fatalf("LoadMetadata returned error: %v", err)
	}

	// updated_at should be set to current time (truncated to seconds due to RFC3339)
	if loaded.UpdatedAt.Before(before) {
		t.Errorf("UpdatedAt = %v, expected >= %v", loaded.UpdatedAt, before)
	}
}

func TestSaveMetadata_PreservesCreatedAt(t *testing.T) {
	dir := t.TempDir()
	metaPath := filepath.Join(dir, "metadata.json")

	createdAt := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	track := Track{
		TrackID:   "test-track",
		Type:      "chore",
		Status:    "completed",
		CreatedAt: createdAt,
	}

	err := SaveMetadata(metaPath, track)
	if err != nil {
		t.Fatalf("SaveMetadata returned error: %v", err)
	}

	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	loaded, err := LoadMetadata(data)
	if err != nil {
		t.Fatalf("LoadMetadata returned error: %v", err)
	}

	if !loaded.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v (should be preserved)", loaded.CreatedAt, createdAt)
	}
}

func TestSaveMetadata_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	metaPath := filepath.Join(dir, "metadata.json")

	// Write initial content
	track := Track{TrackID: "test", Type: "feature", Status: "new"}
	err := SaveMetadata(metaPath, track)
	if err != nil {
		t.Fatalf("SaveMetadata returned error: %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	_, err = LoadMetadata(data)
	if err != nil {
		t.Fatalf("written file is not valid metadata: %v", err)
	}
}

// --- Plan Parsing Tests ---

func TestParsePlan_FullPlan(t *testing.T) {
	data, err := os.ReadFile("../../testdata/full_plan.md")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	phases := ParsePlan(string(data))

	if len(phases) != 3 {
		t.Fatalf("got %d phases, want 3", len(phases))
	}

	// Phase 1: with checkpoint
	p1 := phases[0]
	if p1.Number != 1 {
		t.Errorf("Phase 1 Number = %d, want 1", p1.Number)
	}
	if p1.Name != "Setup" {
		t.Errorf("Phase 1 Name = %q, want %q", p1.Name, "Setup")
	}
	if p1.Checkpoint != "abc1234" {
		t.Errorf("Phase 1 Checkpoint = %q, want %q", p1.Checkpoint, "abc1234")
	}
	if len(p1.Tasks) != 2 {
		t.Fatalf("Phase 1 has %d tasks, want 2", len(p1.Tasks))
	}

	// Phase 1, Task 1: completed with commit and sub-tasks
	task1 := p1.Tasks[0]
	if task1.Name != "Create project structure" {
		t.Errorf("Task 1 Name = %q, want %q", task1.Name, "Create project structure")
	}
	if !task1.Completed {
		t.Error("Task 1 should be completed")
	}
	if task1.Commit != "def5678" {
		t.Errorf("Task 1 Commit = %q, want %q", task1.Commit, "def5678")
	}
	if len(task1.SubTasks) != 2 {
		t.Fatalf("Task 1 has %d sub-tasks, want 2", len(task1.SubTasks))
	}
	if !task1.SubTasks[0].Completed {
		t.Error("Task 1 SubTask 0 should be completed")
	}
	if task1.SubTasks[0].Name != "Create directory layout" {
		t.Errorf("Task 1 SubTask 0 Name = %q, want %q", task1.SubTasks[0].Name, "Create directory layout")
	}

	// Phase 1, Task 2: incomplete without commit, mixed sub-tasks
	task2 := p1.Tasks[1]
	if task2.Name != "Add dependencies" {
		t.Errorf("Task 2 Name = %q, want %q", task2.Name, "Add dependencies")
	}
	if task2.Completed {
		t.Error("Task 2 should not be completed")
	}
	if task2.Commit != "" {
		t.Errorf("Task 2 Commit = %q, want empty", task2.Commit)
	}
	if len(task2.SubTasks) != 2 {
		t.Fatalf("Task 2 has %d sub-tasks, want 2", len(task2.SubTasks))
	}
	if task2.SubTasks[0].Completed {
		t.Error("Task 2 SubTask 0 should not be completed")
	}
	if !task2.SubTasks[1].Completed {
		t.Error("Task 2 SubTask 1 should be completed")
	}

	// Phase 2: without checkpoint
	p2 := phases[1]
	if p2.Number != 2 {
		t.Errorf("Phase 2 Number = %d, want 2", p2.Number)
	}
	if p2.Name != "Implementation" {
		t.Errorf("Phase 2 Name = %q, want %q", p2.Name, "Implementation")
	}
	if p2.Checkpoint != "" {
		t.Errorf("Phase 2 Checkpoint = %q, want empty", p2.Checkpoint)
	}
	if len(p2.Tasks) != 2 {
		t.Fatalf("Phase 2 has %d tasks, want 2", len(p2.Tasks))
	}

	// Phase 3: fully completed
	p3 := phases[2]
	if len(p3.Tasks) != 2 {
		t.Fatalf("Phase 3 has %d tasks, want 2", len(p3.Tasks))
	}
	if !p3.Tasks[0].Completed || !p3.Tasks[1].Completed {
		t.Error("Phase 3 tasks should all be completed")
	}
}

func TestParsePlan_Empty(t *testing.T) {
	data, err := os.ReadFile("../../testdata/empty_plan.md")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	phases := ParsePlan(string(data))
	if len(phases) != 0 {
		t.Errorf("got %d phases, want 0", len(phases))
	}
}

func TestParsePlan_NoPhases(t *testing.T) {
	data, err := os.ReadFile("../../testdata/no_phases_plan.md")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	phases := ParsePlan(string(data))
	if len(phases) != 0 {
		t.Errorf("got %d phases, want 0", len(phases))
	}
}

func TestParsePlan_EmptyString(t *testing.T) {
	phases := ParsePlan("")
	if len(phases) != 0 {
		t.Errorf("got %d phases, want 0", len(phases))
	}
}

func TestParsePlan_PhaseWithoutCheckpoint(t *testing.T) {
	content := "## Phase 5: Deployment\n\n- [x] Task: Deploy to production `abc1234`\n"
	phases := ParsePlan(content)
	if len(phases) != 1 {
		t.Fatalf("got %d phases, want 1", len(phases))
	}
	if phases[0].Checkpoint != "" {
		t.Errorf("Checkpoint = %q, want empty", phases[0].Checkpoint)
	}
	if phases[0].Number != 5 {
		t.Errorf("Number = %d, want 5", phases[0].Number)
	}
}

func TestParsePlan_TaskWithoutCommit(t *testing.T) {
	content := "## Phase 1: Setup\n\n- [ ] Task: Initialize project\n"
	phases := ParsePlan(content)
	if len(phases) != 1 {
		t.Fatalf("got %d phases, want 1", len(phases))
	}
	if len(phases[0].Tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(phases[0].Tasks))
	}
	task := phases[0].Tasks[0]
	if task.Completed {
		t.Error("task should not be completed")
	}
	if task.Commit != "" {
		t.Errorf("Commit = %q, want empty", task.Commit)
	}
}

// --- Track Discovery Tests ---

func TestDiscoverTracks_AllTracks(t *testing.T) {
	tracks := DiscoverTracks("../../testdata/discovery")

	if len(tracks) != 3 {
		t.Fatalf("got %d tracks, want 3", len(tracks))
	}

	// Active tracks should come first, sorted by creation date (newest first)
	// bugfix-beta (2026-01-02) is newer than feature-alpha (2026-01-01)
	if tracks[0].TrackID != "bugfix-beta_20260102" {
		t.Errorf("tracks[0].TrackID = %q, want %q (newest active first)", tracks[0].TrackID, "bugfix-beta_20260102")
	}
	if tracks[0].Source != "active" {
		t.Errorf("tracks[0].Source = %q, want %q", tracks[0].Source, "active")
	}

	if tracks[1].TrackID != "feature-alpha_20260101" {
		t.Errorf("tracks[1].TrackID = %q, want %q", tracks[1].TrackID, "feature-alpha_20260101")
	}
	if tracks[1].Source != "active" {
		t.Errorf("tracks[1].Source = %q, want %q", tracks[1].Source, "active")
	}

	// Archived track should be last
	if tracks[2].TrackID != "feature-gamma_20250601" {
		t.Errorf("tracks[2].TrackID = %q, want %q", tracks[2].TrackID, "feature-gamma_20250601")
	}
	if tracks[2].Source != "archived" {
		t.Errorf("tracks[2].Source = %q, want %q", tracks[2].Source, "archived")
	}
}

func TestDiscoverTracks_ActiveWithPlan(t *testing.T) {
	tracks := DiscoverTracks("../../testdata/discovery")

	// feature-alpha has a plan.md with 1 phase and 2 tasks
	var alpha Track
	for _, tr := range tracks {
		if tr.TrackID == "feature-alpha_20260101" {
			alpha = tr
			break
		}
	}
	if len(alpha.Phases) != 1 {
		t.Fatalf("alpha has %d phases, want 1", len(alpha.Phases))
	}
	if len(alpha.Phases[0].Tasks) != 2 {
		t.Errorf("alpha phase 1 has %d tasks, want 2", len(alpha.Phases[0].Tasks))
	}
}

func TestDiscoverTracks_MissingDirectory(t *testing.T) {
	// Should handle missing base directory gracefully
	tracks := DiscoverTracks("../../testdata/nonexistent")
	if len(tracks) != 0 {
		t.Errorf("got %d tracks, want 0 for missing directory", len(tracks))
	}
}

func TestDiscoverTracks_SortOrder(t *testing.T) {
	tracks := DiscoverTracks("../../testdata/discovery")

	// Verify active tracks come before archived
	foundArchived := false
	for _, tr := range tracks {
		if tr.Source == "archived" {
			foundArchived = true
		}
		if foundArchived && tr.Source == "active" {
			t.Error("found active track after archived track; sorting is broken")
		}
	}
}

func TestDiscoverTracks_FilterArchived(t *testing.T) {
	allTracks := DiscoverTracks("../../testdata/discovery")

	// Simulate filtering archived tracks (as done in TUI)
	var active []Track
	for _, tr := range allTracks {
		if tr.Source != "archived" {
			active = append(active, tr)
		}
	}

	if len(active) != 2 {
		t.Errorf("got %d active tracks, want 2", len(active))
	}
	for _, tr := range active {
		if tr.Source == "archived" {
			t.Error("found archived track after filtering")
		}
	}
}

func TestDiscoverTracks_SortByCreationDate(t *testing.T) {
	tracks := DiscoverTracks("../../testdata/discovery")

	// Active tracks should be sorted newest first
	var activeTracks []Track
	for _, tr := range tracks {
		if tr.Source == "active" {
			activeTracks = append(activeTracks, tr)
		}
	}

	if len(activeTracks) < 2 {
		t.Fatalf("expected at least 2 active tracks, got %d", len(activeTracks))
	}

	// bugfix-beta (2026-01-02) should come before feature-alpha (2026-01-01)
	if activeTracks[0].TrackID != "bugfix-beta_20260102" {
		t.Errorf("first active track = %q, want %q (newest first)", activeTracks[0].TrackID, "bugfix-beta_20260102")
	}
	if activeTracks[1].TrackID != "feature-alpha_20260101" {
		t.Errorf("second active track = %q, want %q", activeTracks[1].TrackID, "feature-alpha_20260101")
	}
}

func TestDiscoverTracks_MissingCreatedAtSortedToBottom(t *testing.T) {
	// Tracks without created_at should sort to the bottom of their group.
	// In the current testdata, all tracks have created_at, so this tests the
	// sort logic via a unit test on the sort function directly.
	tracks := []Track{
		{TrackID: "has-date", Source: "active", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{TrackID: "no-date", Source: "active"},
		{TrackID: "newer-date", Source: "active", CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	sorted := SortTracks(tracks)

	// newer-date first, then has-date, then no-date (zero time at bottom)
	if sorted[0].TrackID != "newer-date" {
		t.Errorf("sorted[0] = %q, want %q", sorted[0].TrackID, "newer-date")
	}
	if sorted[1].TrackID != "has-date" {
		t.Errorf("sorted[1] = %q, want %q", sorted[1].TrackID, "has-date")
	}
	if sorted[2].TrackID != "no-date" {
		t.Errorf("sorted[2] = %q, want %q", sorted[2].TrackID, "no-date")
	}
}

// Placeholder to ensure testdata directory is accessible
func TestTestdataDirectoryExists(t *testing.T) {
	info, err := os.Stat("../../testdata")
	if err != nil {
		t.Fatalf("testdata directory not found: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("testdata is not a directory")
	}

	// Verify all expected test files exist
	expectedFiles := []string{
		"valid_metadata.json",
		"partial_metadata.json",
		"invalid_metadata.json",
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join("../../testdata", f)); err != nil {
			t.Errorf("expected file %q not found: %v", f, err)
		}
	}
}
