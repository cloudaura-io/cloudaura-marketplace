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

// Wrap wraps text to fit within width, breaking at word boundaries.
// Each continuation line is indented by the given indent string.
func Wrap(s string, width int, indent string) string {
	if width <= 0 || len(s) <= width {
		return s
	}

	var result []byte
	line := ""
	firstLine := true

	for _, word := range splitWords(s) {
		lineWidth := width
		if !firstLine {
			lineWidth = width - len(indent)
		}
		if lineWidth < 1 {
			lineWidth = 1
		}

		if line == "" {
			line = word
		} else if len(line)+1+len(word) <= lineWidth {
			line += " " + word
		} else {
			if len(result) > 0 {
				result = append(result, '\n')
				result = append(result, indent...)
			}
			result = append(result, line...)
			line = word
			firstLine = false
		}
	}
	if line != "" {
		if len(result) > 0 {
			result = append(result, '\n')
			result = append(result, indent...)
		}
		result = append(result, line...)
	}
	return string(result)
}

// splitWords splits a string by whitespace into words.
func splitWords(s string) []string {
	var words []string
	word := ""
	for _, c := range s {
		if c == ' ' || c == '\t' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(c)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}
