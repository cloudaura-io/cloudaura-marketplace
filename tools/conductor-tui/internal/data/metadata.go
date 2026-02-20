package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// metadataJSON represents the raw JSON structure of metadata.json.
type metadataJSON struct {
	TrackID     string `json:"track_id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// LoadMetadata parses metadata.json bytes into a Track with fallback defaults.
func LoadMetadata(data []byte) (Track, error) {
	var raw metadataJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return Track{}, fmt.Errorf("invalid metadata JSON: %w", err)
	}

	t := Track{
		TrackID:     raw.TrackID,
		Type:        raw.Type,
		Status:      raw.Status,
		Description: raw.Description,
	}

	if t.Type == "" {
		t.Type = "unknown"
	}
	if t.Status == "" {
		t.Status = "unknown"
	}

	if raw.CreatedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, raw.CreatedAt); err == nil {
			t.CreatedAt = parsed
		}
	}
	if raw.UpdatedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, raw.UpdatedAt); err == nil {
			t.UpdatedAt = parsed
		}
	}

	return t, nil
}

// SaveMetadata writes a Track's metadata to the given path as JSON.
// It uses atomic write (write to temp file, then rename) and updates
// the updated_at timestamp to the current time.
func SaveMetadata(path string, track Track) error {
	createdAt := ""
	if !track.CreatedAt.IsZero() {
		createdAt = track.CreatedAt.UTC().Format(time.RFC3339)
	}

	raw := metadataJSON{
		TrackID:     track.TrackID,
		Type:        track.Type,
		Status:      track.Status,
		Description: track.Description,
		CreatedAt:   createdAt,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	data = append(data, '\n')

	// Atomic write: write to temp file, then rename
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".metadata-*.json.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
