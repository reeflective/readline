package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/strutil"
	"github.com/reeflective/readline/internal/term"
)

// Hint is in charge of printing the usage messages below the input line.
// Various other UI components have access to it so that they can feed
// specialized usage messages to it, like completions.
type Hint struct {
	text       []rune
	persistent []rune
	cleanup    bool
	temp       bool
	set        bool
}

// Set sets the hint message to the given text.
// Generally, this hint message will persist until either a command
// or the completion system overwrites it, or if hint.Reset() is called.
func (h *Hint) Set(hint string) {
	h.text = []rune(hint)
	h.set = true
}

// SetTemporary sets a hint message that will be cleared at the next keypress
// or command being run, which generally coincides with the next redisplay.
func (h *Hint) SetTemporary(hint string) {
	h.text = []rune(hint)
	h.set = true
	h.temp = true
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
	h.temp = false
	h.set = false
}

// ResetPersist drops the persistent hint section.
func (h *Hint) ResetPersist() {
	h.cleanup = len(h.persistent) > 0
	h.persistent = make([]rune, 0)
}

// DisplayHint prints the hint (persistent and/or temporary) sections.
func DisplayHint(hint *Hint) {
	if hint.temp && hint.set {
		hint.set = false
	} else if hint.temp {
		hint.Reset()
	}

	if len(hint.text) == 0 && len(hint.persistent) == 0 {
		if hint.cleanup {
			fmt.Print(term.ClearLineAfter)
		}

		hint.cleanup = false

		return
	}

	var text string

	// Add the various hints.
	if len(hint.persistent) > 0 {
		text += string(hint.persistent) + "\n"
	}

	if len(hint.text) > 0 {
		text += string(hint.text)
	}

	if strutil.RealLength(text) == 0 {
		return
	}

	text = "\r" + strings.TrimSuffix(text, "\n") + term.ClearLineAfter + string(inputrc.Newline) + color.Reset

	if len(text) > 0 {
		fmt.Print(text)
	}
}

// CoordinatesHint returns the number of terminal rows used by the hint.
func CoordinatesHint(hint *Hint) int {
	var text string

	// Add the various hints.
	if len(hint.persistent) > 0 {
		text += string(hint.persistent) + "\n"
	}

	if len(hint.text) > 0 {
		text += string(hint.text)
	}

	// Nothing to do if no real text
	text = strings.TrimSuffix(text, "\n")

	if strutil.RealLength(text) == 0 {
		return 0
	}

	// Otherwise compute the real length/span.
	line := color.Strip(text)
	line += string(inputrc.Newline)
	nl := regexp.MustCompile(string(inputrc.Newline))

	newlines := nl.FindAllStringIndex(line, -1)

	bpos := 0
	usedY := 0

	for i, newline := range newlines {
		bline := line[bpos:newline[0]]
		bpos = newline[0]

		x, lineY := strutil.LineSpan([]rune(bline), i, 0)

		if x != 0 || lineY == 0 {
			lineY++
		}

		usedY += lineY
	}

	return usedY
}
