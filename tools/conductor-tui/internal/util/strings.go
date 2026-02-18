// Package util provides string helpers and status utilities for the Conductor TUI.
package util

// Trunc truncates s to max characters, adding "..." if truncated.
func Trunc(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return "..."
	}
	return s[:max-3] + "..."
}

// Pad pads or truncates s to exactly n characters.
func Pad(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + Spaces(n-len(s))
}

// Spaces returns a string of n space characters.
func Spaces(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}
