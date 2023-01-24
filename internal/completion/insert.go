package completion

import (
	"strings"

	"github.com/xo/inputrc"
)

// acceptCandidate transfers the virtually inserted candidate into the real line.
func (e *Engine) acceptCandidate() {
	cur := e.currentGroup()
	if cur == nil {
		return
	}

	value := cur.selected().Value
	length := len(e.prefix)

	if len(value) >= length {
		e.line.Insert(e.cursor.Pos(), []rune(value[length:])...)
		e.cursor.Inc()

		e.suffix = cur.noSpace
		e.suffix.pos = e.cursor.Pos()
	}
}

// removeSuffixAccepted removes a suffix from the real input line.
func (e *Engine) removeSuffixAccepted() {
	if len(e.comp) > 0 {
		return
	}

	// If our suffix matcher was registered at a different
	// place in our line, then it's an orphan.
	if e.suffix.pos != e.cursor.Pos()-1 {
		e.suffix = suffixMatcher{}
		return
	}

	suf := (*e.line)[e.cursor.Pos()-1]
	key, _ := e.keys.Peek()

	// Special case when completing paths: if the comp is ended
	// by a slash, only remove this slash if the inserted key is
	// one of the suffix matchers, otherwise keep it.
	if suf == '/' && key != ' ' && notMatcher(key, e.suffix.string) {
		return
	}

	// Only remove the matcher if the key we inserted
	// is not the same as the one we removed. This is
	// so that if we just inserted a slash after an
	// inserted directory, the net result is actually
	// nil, so that a space then entered should still
	// trigger the same behavior.
	if e.suffix.Matches(string(suf)) {
		e.line.CutRune(e.cursor.Pos())

		if key != suf {
			e.suffix = suffixMatcher{}
		}
	}
}

// insertCandidate inserts a completion candidate into the virtual line.
func (e *Engine) insertCandidate(candidate []rune) {
	// Remove the current candidate by reusing the current line.
	e.completed.Set(*e.line...)
	e.compCursor.Set(e.cursor.Pos())

	// Insert the new candidate and keep it.
	e.completed.Insert(e.cursor.Pos(), candidate...)
	e.compCursor.Move(len(candidate))
}

// removeSuffixInserted removes a suffix from the virtual input line.
func (e *Engine) removeSuffixInserted() (comp string) {
	cur := e.currentGroup()
	if cur == nil {
		return
	}

	comp = cur.selected().Value
	prefix := len(e.prefix)

	// When the completion has a size of 1, don't remove anything:
	// stacked flags, for example, will never be inserted otherwise.
	if len(comp) > 0 && len(comp[prefix:]) <= 1 {
		return
	}

	suffix := rune(comp[len(comp)-1])
	key, _ := e.keys.Peek()

	// If we are to even consider removing a suffix, we keep the suffix
	// matcher for later: whatever the decision we take here will be identical
	// to the one we take while removing suffix in "non-virtual comp" mode.
	e.suffix = cur.noSpace
	e.suffix.pos = e.cursor.Pos() - 1

	// Add a space to suffix matcher when empty.
	if cur.noSpace.string == "" && e.opts.GetBool("autocomplete") {
		cur.noSpace.Add([]rune{' '}...)
	}

	// When the suffix matcher is a wildcard, that just means
	// it's a noSpace directive: if the currently inserted key
	// is a space, don't remove anything, but keep it for later.
	if cur.noSpace.string == "*" && suffix != inputrc.Space && key == inputrc.Space {
		return
	}

	// Special case when completing paths: if the comp is ended
	// by a slash, only remove this slash if the inserted key is
	// one of the suffix matchers and not a space, otherwise keep it.
	if strings.HasSuffix(comp, "/") && key != inputrc.Space {
		if notMatcher(key, cur.noSpace.string) {
			return
		}
	}

	// Else if the suffix matches a pattern, remove
	if cur.noSpace.Matches(comp) {
		comp = comp[:len(comp)-1]
	}

	return comp
}

// remove any virtually inserted candidate.
func (e *Engine) cutCandidate() {
}

func notMatcher(key rune, matchers string) bool {
	for _, r := range matchers {
		if r == key {
			return false
		}
	}

	return true
}
