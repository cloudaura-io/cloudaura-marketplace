package data

import (
	"encoding/json"
	"fmt"
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
