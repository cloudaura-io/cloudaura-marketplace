package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMetadata_Valid(t *testing.T) {
	data, err := os.ReadFile("testdata/valid_metadata.json")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	track, err := loadMetadata(data)
	if err != nil {
		t.Fatalf("loadMetadata returned error: %v", err)
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
	data, err := os.ReadFile("testdata/partial_metadata.json")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	track, err := loadMetadata(data)
	if err != nil {
		t.Fatalf("loadMetadata returned error: %v", err)
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
	data, err := os.ReadFile("testdata/invalid_metadata.json")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	_, err = loadMetadata(data)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoadMetadata_FallbackTrackID(t *testing.T) {
	// When track_id is empty, loadMetadata should still succeed;
	// the caller (discoverTracks) handles fallback to directory name.
	data := []byte(`{"type": "bug"}`)
	track, err := loadMetadata(data)
	if err != nil {
		t.Fatalf("loadMetadata returned error: %v", err)
	}
	if track.TrackID != "" {
		t.Errorf("TrackID = %q, want empty string", track.TrackID)
	}
	if track.Type != "bug" {
		t.Errorf("Type = %q, want %q", track.Type, "bug")
	}
}

// --- Plan Parsing Tests ---

func TestParsePlan_FullPlan(t *testing.T) {
	data, err := os.ReadFile("testdata/full_plan.md")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	phases := parsePlan(string(data))

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
	data, err := os.ReadFile("testdata/empty_plan.md")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	phases := parsePlan(string(data))
	if len(phases) != 0 {
		t.Errorf("got %d phases, want 0", len(phases))
	}
}

func TestParsePlan_NoPhases(t *testing.T) {
	data, err := os.ReadFile("testdata/no_phases_plan.md")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	phases := parsePlan(string(data))
	if len(phases) != 0 {
		t.Errorf("got %d phases, want 0", len(phases))
	}
}

func TestParsePlan_EmptyString(t *testing.T) {
	phases := parsePlan("")
	if len(phases) != 0 {
		t.Errorf("got %d phases, want 0", len(phases))
	}
}

func TestParsePlan_PhaseWithoutCheckpoint(t *testing.T) {
	content := "## Phase 5: Deployment\n\n- [x] Task: Deploy to production `abc1234`\n"
	phases := parsePlan(content)
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
	phases := parsePlan(content)
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

// Placeholder to ensure testdata directory is accessible
func TestTestdataDirectoryExists(t *testing.T) {
	info, err := os.Stat("testdata")
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
		if _, err := os.Stat(filepath.Join("testdata", f)); err != nil {
			t.Errorf("expected file %q not found: %v", f, err)
		}
	}
}
