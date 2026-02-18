package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// SubTask represents a sub-task within a task.
type SubTask struct {
	Name      string
	Completed bool
}

// Task represents a task within a phase.
type Task struct {
	Name      string
	Completed bool
	Commit    string // short SHA or empty
	SubTasks  []SubTask
}

// Phase represents a phase within a plan.
type Phase struct {
	Number     int
	Name       string
	Checkpoint string // checkpoint SHA or empty
	Tasks      []Task
}

// Track represents a discovered track with metadata and parsed plan.
type Track struct {
	TrackID     string
	Type        string
	Status      string
	Description string
	Source      string // "active" or "archived"
	Phases      []Phase
}

// metadataJSON represents the raw JSON structure of metadata.json.
type metadataJSON struct {
	TrackID     string `json:"track_id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// loadMetadata parses metadata.json bytes into a Track with fallback defaults.
func loadMetadata(data []byte) (Track, error) {
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

	return t, nil
}

// discoverTracks scans the conductor/tracks and conductor/archive directories
// for tracks, loading metadata and parsing plans for each.
func discoverTracks(basePath string) []Track {
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

			track, err := loadMetadata(metaData)
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
				track.Phases = parsePlan(string(planData))
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

var (
	phaseRe   = regexp.MustCompile(`^## Phase (\d+): (.+?)(?:\s*\[checkpoint:\s*([a-f0-9]+)\])?\s*$`)
	taskRe    = regexp.MustCompile(`^- \[([ x~])\] Task: (.+?)(?:\s+` + "`" + `([a-f0-9]{7,})` + "`" + `)?\s*$`)
	subtaskRe = regexp.MustCompile(`^    - \[([ x])\] (.+)$`)
)

// parsePlan parses a plan.md file into a list of phases with tasks and sub-tasks.
func parsePlan(content string) []Phase {
	var phases []Phase
	var currentPhase *Phase
	var currentTask *Task

	for _, line := range strings.Split(content, "\n") {
		if m := phaseRe.FindStringSubmatch(line); m != nil {
			if currentPhase != nil {
				phases = append(phases, *currentPhase)
			}
			num, _ := strconv.Atoi(m[1])
			checkpoint := ""
			if len(m) > 3 {
				checkpoint = m[3]
			}
			currentPhase = &Phase{
				Number:     num,
				Name:       strings.TrimSpace(m[2]),
				Checkpoint: checkpoint,
			}
			currentTask = nil
			continue
		}

		if m := taskRe.FindStringSubmatch(line); m != nil && currentPhase != nil {
			commit := ""
			if len(m) > 3 {
				commit = m[3]
			}
			task := Task{
				Name:      strings.TrimSpace(m[2]),
				Completed: m[1] == "x",
				Commit:    commit,
			}
			currentPhase.Tasks = append(currentPhase.Tasks, task)
			currentTask = &currentPhase.Tasks[len(currentPhase.Tasks)-1]
			continue
		}

		if m := subtaskRe.FindStringSubmatch(line); m != nil && currentTask != nil {
			currentTask.SubTasks = append(currentTask.SubTasks, SubTask{
				Name:      strings.TrimSpace(m[2]),
				Completed: m[1] == "x",
			})
		}
	}

	if currentPhase != nil {
		phases = append(phases, *currentPhase)
	}

	return phases
}
