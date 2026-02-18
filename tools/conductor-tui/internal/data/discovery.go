package data

import (
	"os"
	"path/filepath"
	"sort"
)

// DiscoverTracks scans the conductor/tracks and conductor/archive directories
// for tracks, loading metadata and parsing plans for each.
func DiscoverTracks(basePath string) []Track {
	var tracks []Track

	dirs := []struct {
		path   string
		source string
	}{
		{filepath.Join(basePath, "conductor", "tracks"), "active"},
		{filepath.Join(basePath, "conductor", "archive"), "archived"},
	}

	for _, d := range dirs {
		entries, err := os.ReadDir(d.path)
		if err != nil {
			continue // directory may not exist
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			metaPath := filepath.Join(d.path, entry.Name(), "metadata.json")
			metaData, err := os.ReadFile(metaPath)
			if err != nil {
				continue
			}

			track, err := LoadMetadata(metaData)
			if err != nil {
				continue
			}

			if track.TrackID == "" {
				track.TrackID = entry.Name()
			}
			track.Source = d.source

			// Try to parse plan.md
			planPath := filepath.Join(d.path, entry.Name(), "plan.md")
			planData, err := os.ReadFile(planPath)
			if err == nil {
				track.Phases = ParsePlan(string(planData))
			}

			tracks = append(tracks, track)
		}
	}

	// Sort: active first, then alphabetical by track_id within each group
	sort.Slice(tracks, func(i, j int) bool {
		if tracks[i].Source != tracks[j].Source {
			return tracks[i].Source == "active"
		}
		return tracks[i].TrackID < tracks[j].TrackID
	})

	return tracks
}
