package data

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// PhaseRe matches phase headings like "## Phase 1: Setup [checkpoint: abc1234]"
	PhaseRe = regexp.MustCompile(`^## Phase (\d+): (.+?)(?:\s*\[checkpoint:\s*([a-f0-9]+)\])?\s*$`)
	// TaskRe matches task lines like "- [x] Task: Create project structure `def5678`"
	TaskRe = regexp.MustCompile(`^- \[([ x~])\] Task: (.+?)(?:\s+` + "`" + `([a-f0-9]{7,})` + "`" + `)?\s*$`)
	// SubtaskRe matches sub-task lines like "    - [x] Create directory layout"
	SubtaskRe = regexp.MustCompile(`^    - \[([ x])\] (.+)$`)
)

// ParsePlan parses a plan.md file into a list of phases with tasks and sub-tasks.
func ParsePlan(content string) []Phase {
	var phases []Phase
	var currentPhase *Phase
	var currentTask *Task

	for _, line := range strings.Split(content, "\n") {
		if m := PhaseRe.FindStringSubmatch(line); m != nil {
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

		if m := TaskRe.FindStringSubmatch(line); m != nil && currentPhase != nil {
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

		if m := SubtaskRe.FindStringSubmatch(line); m != nil && currentTask != nil {
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
