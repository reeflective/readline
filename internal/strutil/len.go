package strutil

import (
	"unicode/utf8"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/term"
)

// RealLength returns the real length of a string, trimming all ANSI escaped codes.
func RealLength(s string) int {
	colorStripped := color.Strip(s)
	return utf8.RuneCountInString(colorStripped)
}

// LineSpan computes the number of columns and lines that are needed for a given line,
// accounting for any ANSI escapes/color codes and wrapping with the current term width.
func LineSpan(line []rune, idx, indent int) (x, y int) {
	termWidth := term.GetWidth()
	lineLen := RealLength(string(line))
	lineLen += indent

	cursorY := lineLen / termWidth
	cursorX := lineLen % termWidth

	// The very first (unreal) line counts for nothing,
	// so by opposition all others count for one more.
	if idx == 0 {
		cursorY--
	}

	// Any excess wrap means a newline.
	if cursorX > 0 {
		cursorY++
	}

	// Empty lines are still considered a line.
	if cursorY == 0 && idx != 0 {
		cursorY++
	}

	return cursorX, cursorY
}
