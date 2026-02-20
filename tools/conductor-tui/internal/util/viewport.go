package util

// ViewportResult holds the computed viewport window for a scrollable list.
type ViewportResult struct {
	Start     int // first visible item index (inclusive)
	End       int // last visible item index (exclusive)
	MoreAbove int // number of items hidden above the viewport
	MoreBelow int // number of items hidden below the viewport
}

// CalcViewport computes a viewport window that keeps the cursor visible.
//
// It reserves space for "more above" / "more below" indicator lines when
// items are hidden outside the viewport. The cursor is guaranteed to be
// within [Start, End).
//
// Parameters:
//   - total: total number of items in the list
//   - cursor: current cursor position (0-based index)
//   - maxVisible: maximum number of lines available for items + indicators
func CalcViewport(total, cursor, maxVisible int) ViewportResult {
	if maxVisible <= 0 {
		maxVisible = 1
	}
	if total <= 0 {
		return ViewportResult{}
	}

	// Clamp cursor to valid range.
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}

	// If everything fits, no indicators needed.
	if total <= maxVisible {
		return ViewportResult{
			Start:     0,
			End:       total,
			MoreAbove: 0,
			MoreBelow: 0,
		}
	}

	// We need scrolling. Iteratively determine the window.
	// Start by centering the cursor in the available capacity, then adjust
	// for indicator lines.
	//
	// We may need 0, 1, or 2 indicator lines. Each indicator reduces the
	// capacity available for actual items by 1.

	for attempt := 0; attempt < 3; attempt++ {
		// Estimate how many indicator lines we need based on cursor position.
		needAbove := false
		needBelow := false

		// Tentative capacity for items.
		capacity := maxVisible

		// First pass: assume indicators based on cursor position.
		// If cursor is not at the start, we might need "above".
		// If cursor is not at the end, we might need "below".
		// We'll refine after computing the window.

		// Position window so cursor is at the bottom of the visible range.
		start := cursor - capacity + 1
		if start < 0 {
			start = 0
		}
		end := start + capacity
		if end > total {
			end = total
			start = end - capacity
			if start < 0 {
				start = 0
			}
		}

		// Determine which indicators are needed.
		needAbove = start > 0
		needBelow = end < total

		// Reduce capacity for indicators.
		indicators := 0
		if needAbove {
			indicators++
		}
		if needBelow {
			indicators++
		}

		itemCapacity := maxVisible - indicators
		if itemCapacity < 1 {
			itemCapacity = 1
		}

		// Recompute window with adjusted capacity.
		start = cursor - itemCapacity + 1
		if start < 0 {
			start = 0
		}
		end = start + itemCapacity
		if end > total {
			end = total
			start = end - itemCapacity
			if start < 0 {
				start = 0
			}
		}

		// Verify indicators are still correct after recomputation.
		actualAbove := start > 0
		actualBelow := end < total

		if actualAbove == needAbove && actualBelow == needBelow {
			// Stable result.
			moreAbove := start
			moreBelow := total - end
			return ViewportResult{
				Start:     start,
				End:       end,
				MoreAbove: moreAbove,
				MoreBelow: moreBelow,
			}
		}
		// Indicators changed; retry with the corrected assumption.
	}

	// Fallback: should not reach here, but ensure cursor is visible.
	itemCapacity := maxVisible - 2 // assume both indicators
	if itemCapacity < 1 {
		itemCapacity = 1
	}
	start := cursor - itemCapacity + 1
	if start < 0 {
		start = 0
	}
	end := start + itemCapacity
	if end > total {
		end = total
		start = end - itemCapacity
		if start < 0 {
			start = 0
		}
	}
	return ViewportResult{
		Start:     start,
		End:       end,
		MoreAbove: start,
		MoreBelow: total - end,
	}
}
