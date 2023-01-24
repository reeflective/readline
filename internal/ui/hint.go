package ui

import (
	"regexp"
	"strings"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/term"
)

// Hint is in charge of printing the usage messages below the input line.
// Various other UI components have access to it so that they can feed
// specialized usage messages to it, like completions.
type Hint struct {
	text  []rune
	usedY int
}

// Set sets the hint message to the given text.
func (h *Hint) Set(hint string) {
	h.text = []rune(hint)
}

// Text returns the current hint text.
func (h *Hint) Text() string {
	return string(h.text)
}

// Len returns the length of the current hint.
// This is generally used by consumers to know if there already
// is an active hint, in which case they might want to append to
// it instead of overwriting it altogether (like in isearch mode).
func (h *Hint) Len() int {
	return len(h.text)
}

// Reset removes the hint message.
func (h *Hint) Reset() {
	h.text = make([]rune, 0)
	h.usedY = 0
}

// Display prints the hint section.
func (h *Hint) Display() {
	if len(h.text) == 0 {
		h.usedY = 0
		return
	}

	// Wraps the line, and counts the number of newlines
	// in the string, adjusting the offset as well.
	re := regexp.MustCompile(`\r?\n`)
	newlines := re.Split(string(h.text), -1)
	offset := len(newlines)

	termWidth := term.GetWidth()

	_, actual := wrapText(color.Strip(string(h.text)), termWidth)
	wrapped, _ := wrapText(string(h.text), termWidth)

	offset += actual
	h.usedY = offset - 1

	if len(wrapped) > 0 {
		print("\n")
		print("\r" + wrapped + color.Reset)
	}
}

// Coordinates returns the number of terminal rows used by the hint.
func (h *Hint) Coordinates() int {
	return h.usedY
}

// wrapText - Wraps a text given a specified width, and returns the formatted
// string as well the number of lines it will occupy.
func wrapText(text string, lineWidth int) (wrapped string, lines int) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}

	wrapped = words[0]
	spaceLeft := lineWidth - len(wrapped)

	// There must be at least a line
	if text != "" {
		lines++
	}

	for _, word := range words[1:] {
		if len(color.Strip(word))+1 > spaceLeft {
			lines++

			wrapped += "\n" + word
			spaceLeft = lineWidth - len(word)
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}

	return
}
