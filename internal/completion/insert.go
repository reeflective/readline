package completion

import (
	"strings"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/core"
)

// TrimSuffix removes the last inserted completion's suffix if the required constraints
// are satisfied, among which the index position, the suffix matching patterns, etc.
func (e *Engine) TrimSuffix() {
	if e.line.Len() == 0 || e.cursor.Pos() == 0 || len(e.comp) > 0 {
		return
	}

	// If our suffix matcher was registered at a different
	// place in our line, then it's an orphan.
	if e.suffix.pos != e.cursor.Pos()-1 {
		e.suffix = SuffixMatcher{}
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
			e.suffix = SuffixMatcher{}
		}
	}
}

// acceptCandidate inserts the currently selected candidate into the real input line.
func (e *Engine) acceptCandidate() {
	cur := e.currentGroup()
	if cur == nil {
		return
	}

	// Prepare the completion candidate, remove the
	// prefix part and save its sufffixes for later.
	completion := e.prepareSuffix()
	inserted := []rune(completion[len(e.prefix):])

	// Insert it in the line.
	e.line.Insert(e.cursor.Pos(), inserted...)
	e.cursor.Move(len(inserted))
}

// insertCandidate inserts a completion candidate into the virtual (completed) line.
func (e *Engine) insertCandidate() {
	grp := e.currentGroup()
	if grp == nil {
		return
	}

	if len(grp.selected().Value) < len(e.prefix) {
		return
	}

	// Prepare the completion candidate, remove the
	// prefix part and save its sufffixes for later.
	completion := e.prepareSuffix()
	e.comp = []rune(completion[len(e.prefix):])

	// Copy the current (uncompleted) line/cursor.
	completed := core.Line(string(*e.line))
	e.completed = &completed

	e.compCursor = core.NewCursor(e.completed)
	e.compCursor.Set(e.cursor.Pos())

	// And insert it in the completed line.
	e.completed.Insert(e.compCursor.Pos(), e.comp...)
	e.compCursor.Move(len(e.comp))
}

// prepareSuffix caches any suffix matcher associated with the completion candidate
// to be inserted/accepted into the input line, and trims it if required at this point.
func (e *Engine) prepareSuffix() (comp string) {
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
	e.suffix.pos = e.cursor.Pos() + len(comp) - prefix - 1

	// Add a space to suffix matcher when empty.
	if cur.noSpace.string == "" && !e.opts.GetBool("autocomplete") {
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
	if len(e.comp) == 0 {
		return
	}

	bpos := e.cursor.Pos() - len(e.comp)
	epos := e.cursor.Pos()

	e.line.Cut(bpos, epos)
	e.cursor.Set(bpos)
	// e.cursor.Move(-1 * len(e.comp))

	// e.completed.Cut(bpos, epos)
	// e.compCursor.Move(-1 * len(e.comp))
}

func notMatcher(key rune, matchers string) bool {
	for _, r := range matchers {
		if r == key {
			return false
		}
	}

	return true
}
