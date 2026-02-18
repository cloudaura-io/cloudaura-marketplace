package main

import "testing"

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
		got := trunc(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("trunc(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
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
		got := pad(tt.input, tt.n)
		if got != tt.want {
			t.Errorf("pad(%q, %d) = %q, want %q", tt.input, tt.n, got, tt.want)
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
		got := statusColor(tt.status)
		if got != tt.want {
			t.Errorf("statusColor(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestPhaseStatus(t *testing.T) {
	tests := []struct {
		name  string
		phase Phase
		want  string
	}{
		{
			name:  "empty phase",
			phase: Phase{Tasks: []Task{}},
			want:  "empty",
		},
		{
			name: "all completed",
			phase: Phase{Tasks: []Task{
				{Completed: true},
				{Completed: true},
			}},
			want: "completed",
		},
		{
			name: "some completed",
			phase: Phase{Tasks: []Task{
				{Completed: true},
				{Completed: false},
			}},
			want: "in_progress",
		},
		{
			name: "none completed",
			phase: Phase{Tasks: []Task{
				{Completed: false},
				{Completed: false},
			}},
			want: "pending",
		},
	}
	for _, tt := range tests {
		got := phaseStatus(tt.phase)
		if got != tt.want {
			t.Errorf("phaseStatus(%s) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
