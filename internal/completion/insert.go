package completion

import (
	"unicode"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/keymap"
)

// UpdateInserted should be called only once in between the two shell keymaps
// (local/main) in the main readline loop, to either drop or confirm a virtually
// inserted candidate.
func UpdateInserted(eng *Engine) {
	// If the user currently has a completion selected, any change
	// in the input line will drop the current completion list, in
	// effect deactivating the completion engine.
	// This is so that when a user asks for the list of choices, but
	// then deletes or types something in the input line, the list
	// is still displayed to the user, otherwise it's removed.
	// This does not apply when autocomplete is on.
	choices := len(eng.selected.Value) != 0
	if !eng.auto {
		defer eng.ClearMenu(choices)
	}

	// If autocomplete is on, we also drop the list of generated
	// completions, because it will be recomputed shortly after.
	// Do the same when using incremental search, except if the
	// last key typed is an escape, in which case the user wants
	// to quit incremental search but keeping any selected comp.
	inserted := eng.mustRemoveInserted()
	cached := eng.keymap.Local() != keymap.Isearch && !eng.autoForce

	eng.Cancel(inserted, cached)

	if choices && eng.autoForce && len(eng.selected.Value) == 0 {
		eng.Reset()
	}
}

// TrimSuffix removes the last inserted completion's suffix if the required constraints
// are satisfied (among which the index position, the suffix matching patterns, etc).
func (e *Engine) TrimSuffix() {
	if e.line.Len() == 0 || e.cursor.Pos() == 0 || len(e.selected.Value) > 0 {
		return
	}

	// If our suffix matcher was registered at a different
	// place in our line, then it's an orphan.
	if e.sm.pos != e.cursor.Pos()-1 || e.sm.string == "" {
		e.sm = SuffixMatcher{}
		return
	}

	suf := (*e.line)[e.cursor.Pos()-1]
	keys := e.keys.Caller()
	key := keys[0]

	// Special case when completing paths: if the comp is ended
	// by a slash, only remove this slash if the inserted key is
	// one of the suffix matchers, otherwise keep it.
	if suf == '/' && key != inputrc.Space && notMatcher(key, e.sm.string) {
		return
	}

	// If the key is a space or matches the suffix matcher, cut the suffix.
	if e.sm.Matches(string(key)) || unicode.IsSpace(key) {
		e.cursor.Dec()
		e.line.CutRune(e.cursor.Pos())
	}

	// But when the key is a space, we also drop the suffix matcher,
	// because the user is done with this precise completion (or group of).
	if unicode.IsSpace(key) {
		e.sm = SuffixMatcher{}
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

	// Incremental search is a special case, because the user may
	// want to keep searching for another match, so we don't drop
	// the completion list and exit the incremental search mode.
	if e.hasUniqueCandidate() && e.keymap.Local() != keymap.Isearch {
		e.acceptCandidate()
		e.ResetForce()
	} else {
		e.insertCandidate()
	}
}

// InsertCommonPrefix extends the input line with the longest prefix shared by
// every current completion candidate, beyond what the user has already typed.
// It implements the insertion half of the GNU readline menu-complete-display-prefix
// option ("display the common prefix of the list of possible completions before
// cycling"); the caller invokes it only when that option is set. When the
// candidates share nothing past the typed prefix it is a no-op, and the menu is
// simply displayed as before.
func (e *Engine) InsertCommonPrefix() {
	common := e.commonPrefix()

	// Only extend the typed word; never shorten or re-case it.
	if len([]rune(common)) <= len([]rune(e.prefix)) {
		return
	}

	// Replace the typed prefix with the longer shared one, mirroring how
	// acceptCandidate swaps the prefix for a full candidate value.
	e.cursor.Move(-1 * len(e.prefix))
	e.line.Cut(e.cursor.Pos(), e.cursor.Pos()+len(e.prefix))
	e.cursor.InsertAt([]rune(common)...)

	// Keep the engine prefix in sync so the menu highlighting and any
	// subsequent candidate insertion stay consistent with the new word.
	e.prefix = common
}

// commonPrefix returns the longest string that prefixes every current candidate's
// inserted value, or "" when they share nothing (or there are no candidates).
// The comparison is case-sensitive: candidates differing only in case stop the
// prefix early, a conservative choice that never inserts characters not shared
// verbatim by all candidates.
func (e *Engine) commonPrefix() string {
	var (
		common string
		seen   bool
	)

	for _, grp := range e.groups {
		for _, row := range grp.rows {
			for _, cand := range row {
				if !seen {
					common = cand.Value
					seen = true

					continue
				}

				common = commonStringPrefix(common, cand.Value)
				if common == "" {
					return ""
				}
			}
		}
	}

	return common
}

// commonStringPrefix returns the longest rune-aligned common prefix of a and b.
func commonStringPrefix(a, b string) string {
	ra, rb := []rune(a), []rune(b)
	bound := min(len(ra), len(rb))

	i := 0
	for i < bound && ra[i] == rb[i] {
		i++
	}

	return string(ra[:i])
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
	e.inserted = []rune(completion)

	// Remove the line prefix and insert the candidate.
	e.cursor.Move(-1 * len(e.prefix))
	e.line.Cut(e.cursor.Pos(), e.cursor.Pos()+len(e.prefix))
	e.cursor.InsertAt(e.inserted...)

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
	e.inserted = []rune(completion)

	// Copy the current (uncompleted) line/cursor.
	completed := core.Line(string(*e.line))
	e.compLine = &completed

	e.compCursor = core.NewCursor(e.compLine)
	e.compCursor.Set(e.cursor.Pos())

	// Remove the line prefix and insert the candidate.
	e.compCursor.Move(-1 * len(e.prefix))
	e.compLine.Cut(e.compCursor.Pos(), e.compCursor.Pos()+len(e.prefix))
	e.compCursor.InsertAt(e.inserted...)
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

	// If we are to even consider removing a suffix, we keep the suffix
	// matcher for later: whatever the decision we take here will be identical
	// to the one we take while removing suffix in "non-virtual comp" mode.
	e.sm = cur.noSpace
	e.sm.pos = e.cursor.Pos() + len(comp) - prefix - 1

	return comp
}

func (e *Engine) cancelCompletedLine() {
	// The completed line includes any currently selected
	// candidate, just overwrite it with the normal line.
	e.compLine.Set(*e.line...)
	e.compCursor.Set(e.cursor.Pos())

	// And no virtual candidate anymore.
	e.selected = Candidate{}
}

func (e *Engine) mustRemoveInserted() bool {
	// All other completion modes do not want
	// the candidate to be removed from the line.
	if e.keymap.Local() != keymap.Isearch {
		return false
	}

	// Normally, we should have a key.
	key, empty := core.PeekKey(e.keys)
	if empty {
		return false
	}

	// Some keys trigger behavior different from the normal one:
	// Ex: if the key is a letter, the isearch buffer is updated
	// and the line-inserted match might be different, so remove.
	// If the key is 'Enter', the line will likely be accepted
	// with the currently inserted candidate.
	switch rune(key) {
	case inputrc.Esc, inputrc.Return:
		return false
	default:
		return true
	}
}

func notMatcher(key rune, matchers string) bool {
	for _, r := range matchers {
		if r == key {
			return false
		}
	}

	return true
}
