package readline

import (
	"regexp"
	"strings"
)

// SetHint - a nasty function to force writing a new hint text.
// It does not update helpers, it just renders them, so the hint
// will survive until the helpers (thus including the hint) will
// be updated/recomputed.
func (rl *Instance) SetHint(s string) {
	rl.hint = []rune(s)
	rl.renderHelpers()
}

func (rl *Instance) getHintText() {
	// Before entering this function, some completer might have
	// been called, which might have already populated the hint
	// area (with either an error, a usage, etc).
	// Some of the branchings below will overwrite it.

	// Use the user-provided hint by default, if not empty.
	if rl.HintText != nil {
		userHint := rl.HintText(rl.getCompletionLine())
		if len(userHint) > 0 {
			rl.hint = rl.HintText(rl.getCompletionLine())
		}
	}

	// When completing history, we have a special hint
	if len(rl.histHint) > 0 {
		rl.hint = append([]rune{}, rl.histHint...)
	}

	// But the local keymap, especially completions,
	// overwrites the user hint, since either:
	// - We have some completions, which are self-describing
	// - We didn't match any, thus we have a new error hint.
	switch rl.local {
	case isearch:
		rl.isearchHint()
	case menuselect:
		if rl.noCompletions() {
			rl.hintNoMatches()
		}
	}
}

// hintCompletions generates a hint string from all the
// usage/message strings yielded by completions.
func (rl *Instance) hintCompletions(comps Completions) {
	rl.hint = []rune{}

	// First add the command/flag usage string if any,
	// and only if we don't have completions.
	if len(comps.values) == 0 {
		rl.hint = append([]rune(seqDim), []rune(comps.usage)...)
	}

	// And all further messages
	for _, message := range comps.messages.Get() {
		if message == "" {
			continue
		}

		rl.hint = append(rl.hint, []rune(message+"\n")...)
	}

	// Remove the last newline
	if len(rl.hint) > 0 && rl.hint[len(rl.hint)-1] == '\n' {
		rl.hint = rl.hint[:len(rl.hint)-2]
	}

	// And remove the coloring if no hint
	if string(rl.hint) == seqDim {
		rl.hint = []rune{}
	}
}

// generate a hint when no completion matches the prefix.
func (rl *Instance) hintNoMatches() {
	noMatches := seqDim + "no matching "

	var groups []string
	for _, group := range rl.tcGroups {
		if group.tag == "" {
			continue
		}
		groups = append(groups, group.tag)
	}

	// History has no named group, so add it
	if len(groups) == 0 && len(rl.histHint) > 0 {
		groups = append(groups, rl.historyNames[rl.historySourcePos])
	}

	if len(groups) > 0 {
		groupsStr := strings.Join(groups, ", ")
		noMatches += "'" + groupsStr + "'"
	}

	rl.hint = []rune(noMatches + " completions")
}

// writeHintText - only writes the hint text and computes its offsets.
func (rl *Instance) writeHintText() {
	if len(rl.hint) == 0 {
		rl.hintY = 0
		return
	}

	// Wraps the line, and counts the number of newlines
	// in the string, adjusting the offset as well.
	re := regexp.MustCompile(`\r?\n`)
	newlines := re.Split(string(rl.hint), -1)
	offset := len(newlines)

	wrapped, hintLen := wrapText(string(rl.hint), GetTermWidth())
	offset += hintLen
	rl.hintY = offset

	hintText := string(wrapped)

	if len(hintText) > 0 {
		print("\r" + string(hintText) + seqReset)
	}
}

func (rl *Instance) resetHintText() {
	rl.hintY = 0
	rl.hint = []rune{}
	rl.histHint = []rune{}
}
