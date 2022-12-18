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
	rl.hintText = []rune(s)
	rl.renderHelpers()
}

func (rl *Instance) getHintText() {
	// Use the user-provided hint by default.
	// Useful when command/flag usage is given.
	rl.hintText = rl.HintText(rl.getCompletionLine())

	// Remove the hint if we are autocompleting in insert mode.
	if rl.isAutoCompleting() && rl.main != vicmd {
		rl.resetHintText()
	}

	// When completing history, we have a special hint
	if len(rl.histHint) > 0 {
		rl.hintText = rl.histHint
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
	noMatches := DIM + RED + "no matching "

	var groups []string
	for _, group := range rl.tcGroups {
		if group.Name == "" {
			continue
		}
		groups = append(groups, group.Name)
	}

	// History has no named group, so add it
	if len(groups) == 0 && len(rl.histHint) > 0 {
		groups = append(groups, rl.historyNames[rl.historySourcePos])
	}

	if len(groups) > 0 {
		groupsStr := strings.Join(groups, ", ")
		noMatches += "'" + groupsStr + "'"
	}

	rl.hintText = []rune(noMatches + " completions")
}

func (rl *Instance) isearchHint() {
	rl.hintText = []rune(BOLD + seqFgCyan + "isearch: " + seqReset)
	rl.hintText = append(rl.hintText, rl.tfLine...)

	if rl.regexSearch == nil && len(rl.tfLine) > 0 {
		rl.hintText = append(rl.hintText, []rune(Red(" ! failed to compile search regexp"))...)
	} else if rl.noCompletions() {
		rl.hintText = append(rl.hintText, []rune(RED+" ! no matches")...)
	}

	rl.hintText = append(rl.hintText, []rune(RESET)...)
}

// writeHintText - only writes the hint text and computes its offsets.
func (rl *Instance) writeHintText() {
	if len(rl.hintText) == 0 {
		rl.hintY = 0
		return
	}

	width := GetTermWidth()

	// Wraps the line, and counts the number of newlines in the string,
	// adjusting the offset as well.
	re := regexp.MustCompile(`\r?\n`)
	newlines := re.Split(string(rl.hintText), -1)
	offset := len(newlines)

	wrapped, hintLen := WrapText(string(rl.hintText), width)
	offset += hintLen
	rl.hintY = offset

	hintText := string(wrapped)

	if len(hintText) > 0 {
		print("\r" + rl.config.HintFormatting + string(hintText) + seqReset)
	}
}

func (rl *Instance) resetHintText() {
	rl.hintY = 0
	rl.hintText = []rune{}
}
