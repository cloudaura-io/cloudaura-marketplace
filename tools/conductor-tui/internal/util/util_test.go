package util

import (
	"fmt"
	"testing"

	"github.com/cloudaura-io/cloudaura-marketplace/tools/conductor-tui/internal/data"
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

func TestWrap(t *testing.T) {
	tests := []struct {
		input  string
		width  int
		indent string
		want   string
	}{
		{"short", 20, "", "short"},
		{"hello world foo bar", 11, "", "hello world\nfoo bar"},
		{"hello world foo bar", 11, "  ", "hello world\n  foo bar"},
		{"one two three four five", 10, "", "one two\nthree four\nfive"},
		{"", 10, "", ""},
		{"already fits", 50, "", "already fits"},
		{"superlongword next", 5, "", "superlongword\nnext"},
	}
	for _, tt := range tests {
		got := Wrap(tt.input, tt.width, tt.indent)
		if got != tt.want {
			t.Errorf("Wrap(%q, %d, %q) = %q, want %q", tt.input, tt.width, tt.indent, got, tt.want)
		}
	}
}

// --- CalcViewport Tests ---

func TestCalcViewport_ZeroItems(t *testing.T) {
	vp := CalcViewport(0, 0, 10)
	if vp.Start != 0 || vp.End != 0 || vp.MoreAbove != 0 || vp.MoreBelow != 0 {
		t.Errorf("CalcViewport(0,0,10) = %+v, want all zeros", vp)
	}
}

func TestCalcViewport_ItemsFitExactly(t *testing.T) {
	// 5 items, cursor at 0, maxVisible=5: all fit, no indicators
	vp := CalcViewport(5, 0, 5)
	if vp.Start != 0 || vp.End != 5 || vp.MoreAbove != 0 || vp.MoreBelow != 0 {
		t.Errorf("CalcViewport(5,0,5) = %+v, want Start=0 End=5 MoreAbove=0 MoreBelow=0", vp)
	}
}

func TestCalcViewport_OverflowByOne_MoreBelow(t *testing.T) {
	// 6 items, cursor at 0, maxVisible=5: items overflow by 1
	// "more below" indicator is shown, so capacity is reduced by 1 -> 4 visible items
	vp := CalcViewport(6, 0, 5)
	if vp.Start != 0 {
		t.Errorf("Start = %d, want 0", vp.Start)
	}
	if vp.MoreAbove != 0 {
		t.Errorf("MoreAbove = %d, want 0", vp.MoreAbove)
	}
	if vp.MoreBelow < 1 {
		t.Errorf("MoreBelow = %d, want >= 1", vp.MoreBelow)
	}
	// Cursor must be within [Start, End)
	if 0 < vp.Start || 0 >= vp.End {
		t.Errorf("cursor 0 not in [%d, %d)", vp.Start, vp.End)
	}
}

func TestCalcViewport_CursorAtBottom_MoreAbove(t *testing.T) {
	// 10 items, cursor at 9 (last), maxVisible=5
	// "more above" indicator should be shown, cursor must be visible
	vp := CalcViewport(10, 9, 5)
	if vp.MoreAbove < 1 {
		t.Errorf("MoreAbove = %d, want >= 1", vp.MoreAbove)
	}
	// Cursor must be within [Start, End)
	if 9 < vp.Start || 9 >= vp.End {
		t.Errorf("cursor 9 not in [%d, %d)", vp.Start, vp.End)
	}
	if vp.End > 10 {
		t.Errorf("End = %d, should not exceed total=10", vp.End)
	}
}

func TestCalcViewport_CursorInMiddle_BothIndicators(t *testing.T) {
	// 20 items, cursor at 10, maxVisible=5
	// Both indicators should be shown, capacity reduced by 2
	vp := CalcViewport(20, 10, 5)
	if vp.MoreAbove < 1 {
		t.Errorf("MoreAbove = %d, want >= 1", vp.MoreAbove)
	}
	if vp.MoreBelow < 1 {
		t.Errorf("MoreBelow = %d, want >= 1", vp.MoreBelow)
	}
	// Cursor must be within [Start, End)
	if 10 < vp.Start || 10 >= vp.End {
		t.Errorf("cursor 10 not in [%d, %d)", vp.Start, vp.End)
	}
	// Visible count should be maxVisible - 2 (both indicators)
	visible := vp.End - vp.Start
	if visible != 3 {
		t.Errorf("visible items = %d, want 3 (maxVisible=5 minus 2 indicators)", visible)
	}
}

func TestCalcViewport_OneItem_MaxVisibleOne(t *testing.T) {
	// 1 item, cursor at 0, maxVisible=1: no indicators needed
	vp := CalcViewport(1, 0, 1)
	if vp.Start != 0 || vp.End != 1 || vp.MoreAbove != 0 || vp.MoreBelow != 0 {
		t.Errorf("CalcViewport(1,0,1) = %+v, want Start=0 End=1 no indicators", vp)
	}
}

func TestCalcViewport_MaxVisibleZero_ClampedToOne(t *testing.T) {
	// maxVisible <= 0 should be clamped to at least 1
	vp := CalcViewport(5, 0, 0)
	if vp.End-vp.Start < 1 {
		t.Errorf("CalcViewport with maxVisible=0: visible items = %d, want >= 1", vp.End-vp.Start)
	}
	// cursor must be visible
	if 0 < vp.Start || 0 >= vp.End {
		t.Errorf("cursor 0 not in [%d, %d)", vp.Start, vp.End)
	}
}

func TestCalcViewport_MaxVisibleNegative_ClampedToOne(t *testing.T) {
	vp := CalcViewport(3, 1, -5)
	if vp.End-vp.Start < 1 {
		t.Errorf("CalcViewport with maxVisible=-5: visible items = %d, want >= 1", vp.End-vp.Start)
	}
	if 1 < vp.Start || 1 >= vp.End {
		t.Errorf("cursor 1 not in [%d, %d)", vp.Start, vp.End)
	}
}

func TestCalcViewport_CursorAlwaysVisible(t *testing.T) {
	// Exhaustive sweep: for various combinations, cursor must always be in [Start, End)
	for total := 0; total <= 15; total++ {
		for maxVis := 1; maxVis <= 10; maxVis++ {
			for cursor := 0; cursor < total; cursor++ {
				vp := CalcViewport(total, cursor, maxVis)
				if cursor < vp.Start || cursor >= vp.End {
					t.Errorf("CalcViewport(%d, %d, %d): cursor %d not in [%d, %d)",
						total, cursor, maxVis, cursor, vp.Start, vp.End)
				}
				if vp.Start < 0 {
					t.Errorf("CalcViewport(%d, %d, %d): Start=%d < 0", total, cursor, maxVis, vp.Start)
				}
				if vp.End > total {
					t.Errorf("CalcViewport(%d, %d, %d): End=%d > total=%d", total, cursor, maxVis, vp.End, total)
				}
			}
		}
	}
}

func TestCalcViewport_IndicatorCounts(t *testing.T) {
	// Verify MoreAbove equals Start and MoreBelow equals total-End
	tests := []struct {
		total, cursor, maxVis int
	}{
		{10, 0, 5},
		{10, 5, 5},
		{10, 9, 5},
		{20, 10, 5},
		{3, 1, 2},
		{100, 50, 10},
	}
	for _, tt := range tests {
		vp := CalcViewport(tt.total, tt.cursor, tt.maxVis)
		if vp.MoreAbove != vp.Start {
			t.Errorf("CalcViewport(%d,%d,%d): MoreAbove=%d, want Start=%d",
				tt.total, tt.cursor, tt.maxVis, vp.MoreAbove, vp.Start)
		}
		if vp.MoreBelow != tt.total-vp.End {
			t.Errorf("CalcViewport(%d,%d,%d): MoreBelow=%d, want total-End=%d",
				tt.total, tt.cursor, tt.maxVis, vp.MoreBelow, tt.total-vp.End)
		}
	}
}

func TestCalcViewport_LastItemReachable(t *testing.T) {
	// AC-2: The last item in any list is reachable and fully visible when the cursor is on it
	for total := 1; total <= 20; total++ {
		for maxVis := 1; maxVis <= 10; maxVis++ {
			cursor := total - 1
			vp := CalcViewport(total, cursor, maxVis)
			if cursor < vp.Start || cursor >= vp.End {
				t.Errorf("CalcViewport(%d, %d, %d): last item cursor=%d not in [%d, %d)",
					total, cursor, maxVis, cursor, vp.Start, vp.End)
			}
		}
	}
}

func TestCalcViewport_VisibleCountRespectsBudget(t *testing.T) {
	// When maxVisible >= 3, the visible items + indicator lines must fit
	// within the maxVisible budget. For very small maxVisible (1-2),
	// we always guarantee at least 1 visible item, which may exceed the
	// budget when indicators are also needed -- that's correct behavior.
	for total := 0; total <= 15; total++ {
		for maxVis := 3; maxVis <= 10; maxVis++ {
			for cursor := 0; cursor < total; cursor++ {
				vp := CalcViewport(total, cursor, maxVis)
				visible := vp.End - vp.Start
				indicators := 0
				if vp.MoreAbove > 0 {
					indicators++
				}
				if vp.MoreBelow > 0 {
					indicators++
				}
				totalUsed := visible + indicators
				if totalUsed > maxVis {
					t.Errorf("CalcViewport(%d,%d,%d): visible=%d + indicators=%d = %d > maxVisible=%d",
						total, cursor, maxVis, visible, indicators, totalUsed, maxVis)
				}
			}
		}
	}
}

func TestCalcViewport_SmallMaxVisible_StillShowsOneItem(t *testing.T) {
	// When maxVisible is very small (1 or 2), we must still show at least 1 item.
	// Indicators may cause the total lines to exceed maxVisible; that's acceptable.
	for total := 1; total <= 10; total++ {
		for cursor := 0; cursor < total; cursor++ {
			for maxVis := 1; maxVis <= 2; maxVis++ {
				vp := CalcViewport(total, cursor, maxVis)
				visible := vp.End - vp.Start
				if visible < 1 {
					t.Errorf("CalcViewport(%d,%d,%d): must show at least 1 item, got %d",
						total, cursor, maxVis, visible)
				}
			}
		}
	}
}

// Suppress unused import warning for fmt
var _ = fmt.Sprintf

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
