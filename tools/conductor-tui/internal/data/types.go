// Package data provides types and functions for loading and parsing
// Conductor track metadata and plan files.
package data

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
