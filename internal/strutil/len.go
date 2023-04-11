package strutil

import (
	"strings"
	"unicode/utf8"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/term"
)

// FormatTabs replaces all '\t' occurrences in a string with 6 spaces each.
func FormatTabs(s string) string {
	return strings.ReplaceAll(s, "\t", "     ")
}

// RealLength returns the real length of a string, trimming all ANSI escaped codes.
func RealLength(s string) int {
	colors := color.Strip(s)                          // Remove colors
	tabs := strings.ReplaceAll(colors, "\t", "     ") // Remove spaces

	return utf8.RuneCountInString(tabs)
	// return utf8.RuneCountInString(colorStripped)
}

// LineSpan computes the number of columns and lines that are needed for a given line,
// accounting for any ANSI escapes/color codes and wrapping with the current term width.
func LineSpan(line []rune, idx, indent int) (x, y int) {
	termWidth := term.GetWidth()
	lineLen := RealLength(string(line))
	lineLen += indent

	cursorY := lineLen / termWidth
	cursorX := lineLen % termWidth

	// Empty lines are still considered a line.
	if cursorY == 0 && idx != 0 {
		cursorY++
	}

	return cursorX, cursorY
}
