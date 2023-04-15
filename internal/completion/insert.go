package completion

import (
	"strings"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/core"
)

// TrimSuffix removes the last inserted completion's suffix if the required constraints
// are satisfied, among which the index position, the suffix matching patterns, etc.
func (e *Engine) TrimSuffix() {
	if e.line.Len() == 0 || e.cursor.Pos() == 0 || len(e.selected.Value) > 0 {
		return
	}

	// If our suffix matcher was registered at a different
	// place in our line, then it's an orphan.
	if e.sm.pos != e.cursor.Pos()-1 {
		e.sm = SuffixMatcher{}
		return
	}

	suf := (*e.line)[e.cursor.Pos()-1]
	key, _ := e.keys.Peek()

	// Special case when completing paths: if the comp is ended
	// by a slash, only remove this slash if the inserted key is
	// one of the suffix matchers, otherwise keep it.
	if suf == '/' && key != inputrc.Space && notMatcher(key, e.sm.string) {
		return
	}

	// Remove the suffix if either:
	switch {
	case e.sm.Matches(string(key)):
		// The key to be inserted matches the suffix matcher.
		e.line.CutRune(e.cursor.Pos())

	case e.sm.Matches(string(suf)) && key == inputrc.Space:
		// The end of the completion matches the suffix and we are inserting a space.
		e.line.CutRune(e.cursor.Pos())
	}
}

// refreshLine - Either insert the only candidate in the real line
// and drop the current completion list, prefix, keymaps, etc, or
// swap the formerly selected candidate with the new one.
func (e *Engine) refreshLine() {
	if e.noCompletions() {
		e.Cancel(true, true)
		return
	}

	if e.currentGroup() == nil {
		return
	}

	if e.hasUniqueCandidate() {
		e.acceptCandidate()
		e.ClearMenu(true)
	} else {
		e.insertCandidate()
	}
}

// acceptCandidate inserts the currently selected candidate into the real input line.
func (e *Engine) acceptCandidate() {
	cur := e.currentGroup()
	if cur == nil {
		return
	}

	e.selected = cur.selected()

	// Prepare the completion candidate, remove the
	// prefix part and save its sufffixes for later.
	completion := e.prepareSuffix()
	e.inserted = []rune(completion[len(e.prefix):])

	// Remove the suffix from the line first.
	e.line.Cut(e.cursor.Pos(), e.cursor.Pos()+len(e.suffix))

	// Insert it in the line.
	e.line.Insert(e.cursor.Pos(), e.inserted...)
	e.cursor.Move(len(e.inserted))

	// And forget about this inserted completion.
	e.inserted = make([]rune, 0)
	e.prefix = ""
	e.suffix = ""
}

// insertCandidate inserts a completion candidate into the virtual (completed) line.
func (e *Engine) insertCandidate() {
	grp := e.currentGroup()
	if grp == nil {
		return
	}

	e.selected = grp.selected()

	if len(e.selected.Value) < len(e.prefix) {
		return
	}

	// Prepare the completion candidate, remove the
	// prefix part and save its sufffixes for later.
	completion := e.prepareSuffix()
	e.inserted = []rune(completion[len(e.prefix):])

	// Copy the current (uncompleted) line/cursor.
	completed := core.Line(string(*e.line))
	e.completed = &completed

	e.compCursor = core.NewCursor(e.completed)
	e.compCursor.Set(e.cursor.Pos())

	// Remove the suffix from the line first.
	e.completed.Cut(e.compCursor.Pos(), e.compCursor.Pos()+len(e.suffix))

	// And insert it in the completed line.
	e.completed.Insert(e.compCursor.Pos(), e.inserted...)
	e.compCursor.Move(len(e.inserted))
}

// prepareSuffix caches any suffix matcher associated with the completion candidate
// to be inserted/accepted into the input line, and trims it if required at this point.
func (e *Engine) prepareSuffix() (comp string) {
	cur := e.currentGroup()
	if cur == nil {
		return
	}

	comp = e.selected.Value
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
	e.sm = cur.noSpace
	e.sm.pos = e.cursor.Pos() + len(comp) - prefix - 1

	// Add a space to suffix matcher when empty and the comp ends with a space.
	// if cur.noSpace.string == "" && !e.opts.GetBool("autocomplete") {
	if cur.noSpace.string == "" && suffix == inputrc.Space {
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
	// if cur.noSpace.Matches(comp) {
	// 	comp = comp[:len(comp)-1]
	// }

	return comp
}

func (e *Engine) cancelCompletedLine() {
	// The completed line includes any currently selected
	// candidate, just overwrite it with the normal line.
	e.completed.Set(*e.line...)
	e.compCursor.Set(e.cursor.Pos())

	// And no virtual candidate anymore.
	e.selected = Candidate{}
}

func notMatcher(key rune, matchers string) bool {
	for _, r := range matchers {
		if r == key {
			return false
		}
	}

	return true
}
