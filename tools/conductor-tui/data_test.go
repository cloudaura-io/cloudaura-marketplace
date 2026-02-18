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
