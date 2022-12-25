package readline

import (
	"strings"
)

// Insert the current completion candidate into the input line.
// This candidate might either be the currently selected one (white frame),
// or the only candidate available, if the total number of candidates is 1.
func (rl *Instance) insertCandidate() {
	cur := rl.currentGroup()
	if cur == nil {
		return
	}

	completion := cur.selected().Value
	prefix := len(rl.tcPrefix)

	// Special case for the only special escape, which
	// if not handled, will make us insert the first
	// character of our actual rl.tcPrefix in the candidate.
	if strings.HasPrefix(string(rl.tcPrefix), "%") {
		prefix++
	}

	// Ensure no indexing error happens with prefix
	if len(completion) >= prefix {
		rl.insert([]rune(completion[prefix:]))
		rl.compSuffix = cur.noSpace
		rl.compSuffix.pos = rl.pos - 1
	}
}

func (rl *Instance) removeSuffixInserted() {
	// We need a line, and no virtual candidate.
	if len(rl.line) == 0 || rl.pos == 0 || len(rl.comp) != 0 {
		return
	}

	// If our suffix matcher was registered at a different
	// place in our line, then it's an orphan.
	if rl.compSuffix.pos != rl.pos-1 {
		rl.compSuffix = suffixMatcher{}
		return
	}

	suffix := string(rl.line[rl.pos-1])

	// Special case when completing paths: if the comp is ended
	// by a slash, only remove this slash if the inserted key is
	// one of the suffix matchers, otherwise keep it.
	if suffix == "/" && rl.keys != " " {
		for _, s := range rl.compSuffix.string {
			if s == '/' && keyIsNotMatcher(rl.keys, rl.compSuffix.string) {
				return
			}
		}
	}

	if rl.compSuffix.Matches(suffix) {
		rl.deletex()

		// Only remove the matcher if the key we inserted
		// is not the same as the one we removed. This is
		// so that if we just inserted a slash after an
		// inserted directory, the net result is actually
		// nil, so that a space then entered should still
		// trigger the same behavior.
		if rl.keys != suffix {
			rl.compSuffix = suffixMatcher{}
		}
	}
}

// insertCandidateVirtual - When a completion candidate is selected, we insert it virtually in the input line:
// this will not trigger further firltering against the other candidates. Each time this function
// is called, any previous candidate is dropped, after being used for moving the cursor around.
func (rl *Instance) insertCandidateVirtual(candidate []rune) {
	for {
		// I don't really understand why `0` is creaping in at the end of the
		// array but it only happens with unicode characters.
		if len(candidate) > 1 && candidate[len(candidate)-1] == 0 {
			candidate = candidate[:len(candidate)-1]
			continue
		}
		break
	}

	// We place the cursor back at the beginning of the previous virtual candidate
	if rl.pos > 0 {
		rl.pos -= len(rl.comp)
	}

	// We delete the previous virtual completion, just
	// like we would delete a word in vim editing mode.
	if len(rl.comp) == 1 {
		rl.deleteVirtual()
	} else if len(rl.comp) > 0 {
		rl.viDeleteByAdjustVirtual(len(rl.comp))
	}

	// We then keep a reference to the new candidate
	rl.comp = candidate

	// We should not have a remaining virtual completion
	// line, so it is now identical to the real line.
	rl.compLine = rl.line

	// Insert the new candidate in the virtual line.
	switch {
	case len(rl.compLine) == 0:
		rl.compLine = candidate
	case rl.pos == 0:
		rl.compLine = append(candidate, rl.compLine...)
	case rl.pos < len(rl.compLine):
		r := append(candidate, rl.compLine[rl.pos:]...)
		rl.compLine = append(rl.compLine[:rl.pos], r...)
	default:
		rl.compLine = append(rl.compLine, candidate...)
	}

	// We place the cursor at the end of our new virtually completed item
	rl.pos += len(candidate)
}

// updateVirtualComp - Either insert the current completion
// candidate virtually, or on the real line.
func (rl *Instance) updateVirtualComp() {
	cur := rl.currentGroup()
	if cur == nil {
		return
	}

	completion := cur.selected().Value
	prefix := len(rl.tcPrefix)

	// If the total number of completions is one, automatically insert it.
	if rl.hasUniqueCandidate() {
		rl.insertCandidate()
		// Quit the tab completion mode to avoid asking to the user to press
		// Enter twice to actually run the command
		// Refresh first, and then quit the completion mode
		rl.undoSkipAppend = true
		rl.resetCompletion()
	} else {

		// Special case for the only special escape, which
		// if not handled, will make us insert the first
		// character of our actual rl.tcPrefix in the candidate.
		// TODO: This should be changed.
		if strings.HasPrefix(string(rl.tcPrefix), "%") {
			prefix++
		}

		// Or insert it virtually.
		if len(completion) >= prefix {
			rl.insertCandidateVirtual([]rune(completion[prefix:]))
		}
	}
}

// resetVirtualComp - This function is called before most of our readline key handlers,
// and makes sure that the current completion (virtually inserted) is either inserted or dropped,
// and that all related parameters are reinitialized.
func (rl *Instance) resetVirtualComp(drop bool) {
	if len(rl.comp) == 0 {
		return
	}

	// Get the current candidate and its group.
	// It contains info on how we must process it
	cur := rl.currentGroup()
	if cur == nil {
		return
	}

	completion := cur.selected().Value

	// Avoid problems with empty completions
	if completion == "" {
		rl.clearVirtualComp()
		return
	}

	// We will only insert the net difference
	// between prefix and completion.
	prefix := len(rl.tcPrefix)

	// Special case for the only special escape, which
	// if not handled, will make us insert the first
	// character of our actual rl.tcPrefix in the candidate.
	if strings.HasPrefix(string(rl.tcPrefix), "%") {
		prefix++
	}

	// If we are asked to drop the completion,
	// move it away from the line and return.
	if drop {
		rl.pos -= len([]rune(completion[prefix:]))
		rl.compLine = rl.line
		rl.clearVirtualComp()
		return
	}

	// Trim any suffix when found, except for a few cases.
	completion = rl.removeSuffixCandidate(cur, prefix)

	// Insert the current candidate and keep the suffix matcher
	// for this candidate in case a space is inserted after it.
	rl.insertCandidateVirtual([]rune(completion[prefix:]))
	rl.clearVirtualComp()
}

func (rl *Instance) removeSuffixCandidate(cur *comps, prefix int) (comp string) {
	comp = cur.selected().Value

	// When the completion has a size of 1, don't remove anything:
	// stacked flags, for example, will never be inserted otherwise.
	if len(comp[prefix:]) == 1 {
		return
	}

	// If we are to even consider removing a suffix, we keep the suffix
	// matcher for later: whatever the decision we take here will be identical
	// to the one we take while removing suffix in "non-virtual comp" mode.
	rl.compSuffix = cur.noSpace
	rl.compSuffix.pos = rl.pos - 1

	// Add a space to suffix matcher when empty.
	if cur.noSpace.string == "" {
		cur.noSpace.Add([]rune{' '}...)
	}

	// When the suffix matcher is a wildcard, that just means
	// it's a noSpace directive: if the currently inserted key
	// is a space, don't remove anything, but keep it for later.
	if cur.noSpace.string == "*" && rl.keys == " " && comp[len(comp)-1] != ' ' {
		return
	}

	// Special case when completing paths: if the comp is ended
	// by a slash, only remove this slash if the inserted key is
	// one of the suffix matchers and not a space, otherwise keep it.
	if strings.HasSuffix(comp, "/") && rl.keys != " " {
		for _, s := range cur.noSpace.string {
			if s == '/' && keyIsNotMatcher(rl.keys, cur.noSpace.string) {
				return
			}
		}
	}

	// Else if the suffix matches a pattern, remove
	if cur.noSpace.Matches(comp) {
		comp = comp[:len(comp)-1]
	}

	return comp
}

// viDeleteByAdjustVirtual - Same as viDeleteByAdjust, but for our virtually completed input line.
func (rl *Instance) viDeleteByAdjustVirtual(adjust int) {
	var newLine []rune

	// Avoid doing anything if input line is empty.
	if len(rl.compLine) == 0 {
		return
	}

	switch {
	case adjust == 0:
		rl.undoSkipAppend = true
		return
	case rl.pos+adjust == len(rl.compLine)-1:
		newLine = rl.compLine[:rl.pos]
	case rl.pos+adjust == 0:
		newLine = rl.compLine[rl.pos:]
	case adjust < 0:
		newLine = append(rl.compLine[:rl.pos+adjust], rl.compLine[rl.pos:]...)
	default:
		newLine = append(rl.compLine[:rl.pos], rl.compLine[rl.pos+adjust:]...)
	}

	// We have our new line completed
	rl.compLine = newLine

	if adjust < 0 {
		rl.moveCursorByAdjust(adjust)
	}
}

// viJumpEVirtual - Same as viJumpE, but for our virtually completed input line.
func (rl *Instance) viJumpEVirtual(tokeniser func([]rune, int) ([]string, int, int)) (adjust int) {
	split, index, pos := tokeniser(rl.compLine, rl.pos)
	if len(split) == 0 {
		return
	}

	word := rTrimWhiteSpace(split[index])

	switch {
	case len(split) == 0:
		return
	case index == len(split)-1 && pos >= len(word)-1:
		return
	case pos >= len(word)-1:
		word = rTrimWhiteSpace(split[index+1])
		adjust = len(split[index]) - pos
		adjust += len(word) - 1
	default:
		adjust = len(word) - pos - 1
	}

	return
}

func (rl *Instance) deleteVirtual() {
	switch {
	case len(rl.compLine) == 0:
		return
	case rl.pos == 0:
		rl.compLine = rl.compLine[1:]
	case rl.pos > len(rl.compLine):
	case rl.pos == len(rl.compLine):
		rl.compLine = rl.compLine[:rl.pos]
	default:
		rl.compLine = append(rl.compLine[:rl.pos], rl.compLine[rl.pos+1:]...)
	}
}

// We are done with the current virtual completion candidate.
// Get ready for the next one
func (rl *Instance) clearVirtualComp() {
	rl.line = rl.compLine
	rl.comp = []rune{}
}

func keyIsNotMatcher(key, matchers string) bool {
	for _, r := range matchers {
		if string(r) == key {
			return false
		}
	}

	return true
}
