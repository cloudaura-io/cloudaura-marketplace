package util

import (
	"testing"

	"github.com/cloudaura-io/conductor-claude-code/tools/conductor-tui/internal/data"
)

func TestTrunc(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello..."},
		{"hello world", 3, "..."},
		{"", 5, ""},
		{"ab", 1, "..."},
	}
	for _, tt := range tests {
		got := Trunc(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("Trunc(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestPad(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"hi", 5, "hi   "},
		{"hello", 5, "hello"},
		{"hello world", 5, "hello"},
		{"", 3, "   "},
	}
	for _, tt := range tests {
		got := Pad(tt.input, tt.n)
		if got != tt.want {
			t.Errorf("Pad(%q, %d) = %q, want %q", tt.input, tt.n, got, tt.want)
		}
	}
}

func TestStatusColor(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"completed", "green"},
		{"done", "green"},
		{"in_progress", "yellow"},
		{"doing", "yellow"},
		{"pending", "cyan"},
		{"todo", "cyan"},
		{"new", "magenta"},
		{"review", "blue"},
		{"blocked", "red"},
		{"archived", "gray"},
		{"unknown", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := StatusColor(tt.status)
		if got != tt.want {
			t.Errorf("StatusColor(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestPhaseStatus(t *testing.T) {
	tests := []struct {
		name  string
		phase data.Phase
		want  string
	}{
		{
			name:  "empty phase",
			phase: data.Phase{Tasks: []data.Task{}},
			want:  "empty",
		},
		{
			name: "all completed",
			phase: data.Phase{Tasks: []data.Task{
				{Completed: true},
				{Completed: true},
			}},
			want: "completed",
		},
		{
			name: "some completed",
			phase: data.Phase{Tasks: []data.Task{
				{Completed: true},
				{Completed: false},
			}},
			want: "in_progress",
		},
		{
			name: "none completed",
			phase: data.Phase{Tasks: []data.Task{
				{Completed: false},
				{Completed: false},
			}},
			want: "pending",
		},
	}
	for _, tt := range tests {
		got := PhaseStatus(tt.phase)
		if got != tt.want {
			t.Errorf("PhaseStatus(%s) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
