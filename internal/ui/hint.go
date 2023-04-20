package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/strutil"
)

// Hint is in charge of printing the usage messages below the input line.
// Various other UI components have access to it so that they can feed
// specialized usage messages to it, like completions.
type Hint struct {
	text       []rune
	persistent []rune
}

// Set sets the hint message to the given text.
func (h *Hint) Set(hint string) {
	h.text = []rune(hint)
}

// Persist adds a hint message to be persistently
// displayed until hint.ResetPersist() is called.
func (h *Hint) Persist(hint string) {
	h.persistent = []rune(hint)
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
}

// ResetPersist drops the persistent hint.
func (h *Hint) ResetPersist() {
	h.persistent = make([]rune, 0)
}

// Display prints the hint section.
func (h *Hint) Display() {
	if len(h.text) == 0 && len(h.persistent) == 0 {
		return
	}

	var text string

	// Add the various hints.
	if len(h.persistent) > 0 {
		text += string(h.persistent) + "\n"
	}

	if len(h.text) > 0 {
		text += string(h.text)
	}

	text = strings.TrimSuffix(text, "\n")

	if len(text) > 0 {
		fmt.Print("\r" + text + string(inputrc.Newline) + color.Reset)
	}
}

// Coordinates returns the number of terminal rows used by the hint.
func (h *Hint) Coordinates() int {
	var text string

	// Add the various hints.
	if len(h.persistent) > 0 {
		text += string(h.persistent) + "\n"
	}

	if len(h.text) > 0 {
		text += string(h.text)
	}

	text = strings.TrimSuffix(text, "\n")

	if text == "" {
		return 0
	}

	line := color.Strip(text)
	line += string(inputrc.Newline)
	nl := regexp.MustCompile(string(inputrc.Newline))

	newlines := nl.FindAllStringIndex(line, -1)

	bpos := 0
	usedY := 0

	for i, newline := range newlines {
		bline := line[bpos:newline[0]]
		bpos = newline[0]

		x, y := strutil.LineSpan([]rune(bline), i, 0)

		if x != 0 || y == 0 {
			y++
		}

		usedY += y
	}

	return usedY
}
