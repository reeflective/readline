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
	// Use the user-provided hint by default.
	// Useful when command/flag usage is given.
	if rl.HintText != nil {
		rl.hint = rl.HintText(rl.getCompletionLine())
	}

	// Remove the hint if we are autocompleting in insert mode.
	if rl.isAutoCompleting() && rl.main != vicmd {
		rl.resetHintText()
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

// generate a hint when no completion matches the prefix.
func (rl *Instance) hintNoMatches() {
	noMatches := seqDim + seqFgRed + "no matching "

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
}
