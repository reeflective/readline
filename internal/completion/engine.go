package completion

import (
	"regexp"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/keymap"
	"github.com/reeflective/readline/internal/ui"
)

// Engine is responsible for all completion tasks: generating, computing,
// displaying and updating completion values and inserted candidates.
type Engine struct {
	opts   *inputrc.Config // The inputrc contains options relative to completion.
	cached Completer       // A cached completer function to use when updating.
	defaut Completer       // Completer used by things like autocomplete
	hint   *ui.Hint        // The completions can feed hint/usage messages

	// Completion parameters
	groups    []*group      // All of our suggestions tree is in here
	sm        SuffixMatcher // The suffix matcher is kept for removal after actually inserting the candidate.
	selected  Candidate     // The currently selected item, not yet a real part of the input line.
	prefix    string        // The current tab completion prefix against which to build candidates
	suffix    string        // The current word suffix
	inserted  []rune        // The selected candidate (inserted in line) without prefix or suffix.
	usedY     int           // Comprehensive offset of the currently built completions
	auto      bool          // Is the engine autocompleting ?
	autoForce bool

	// Line parameters
	keys       *core.Keys      // The input keys reader
	line       *core.Line      // The real input line of the shell.
	cursor     *core.Cursor    // The cursor of the shell.
	selection  *core.Selection // The line selection
	completed  *core.Line      // A line that might include a virtually inserted candidate.
	compCursor *core.Cursor    // The adjusted cursor.
	keymaps    *keymap.Modes   // The main/local keymaps of the shell

	// Incremental search
	isearchBuf    *core.Line     // The isearch minibuffer
	isearchCur    *core.Cursor   // Cursor position in the minibuffer.
	isearch       *regexp.Regexp // Holds the current search regex match
	isearchName   string         // What is being incrementally searched for.
	isearchInsert bool           // Whether to insert the first match in the line
}

// NewEngine initializes a new completion engine with the shell operating parameters.
func NewEngine(k *core.Keys, l *core.Line, c *core.Cursor, s *core.Selection, h *ui.Hint, km *keymap.Modes, o *inputrc.Config) *Engine {
	return &Engine{
		opts:       o,
		keys:       k,
		line:       l,
		cursor:     c,
		selection:  s,
		completed:  l,
		compCursor: c,
		hint:       h,
		keymaps:    km,
	}
}

// Generate uses a list of completions to group/order and prepares completions before printing them.
// If either no completions or only one is available after all constraints are applied, the engine
// will automatically insert/accept and/or reset itself.
func (e *Engine) Generate(completions Values) {
	e.prepare(completions)

	if e.noCompletions() {
		e.ClearMenu(true)
	}

	if e.hasUniqueCandidate() {
		e.acceptCandidate()
		e.ClearMenu(true)
	}
}

// GenerateWith generates completions with a completer function, itself cached
// so that the next time it must update its results, it can reuse this completer.
func (e *Engine) GenerateWith(completer Completer) {
	e.cached = completer
	if e.cached == nil {
		return
	}

	// Call the provided/cached completer
	// and use the completions as normal
	e.Generate(e.cached())
}

// Select moves the completion selector by some X or Y value,
// and updates the inserted candidate in the input line.
func (e *Engine) Select(row, column int) {
	grp := e.currentGroup()

	if grp == nil || len(grp.values) == 0 {
		return
	}

	// Ensure the completion keymaps are set.
	e.adjustSelectKeymap()

	// Some keys used to move around completions
	// will influence the coordinates' offsets.
	row, column = e.adjustCycleKeys(row, column)

	// If we already have an inserted candidate
	// remove it before inserting the new one.
	if len(e.selected.Value) > 0 {
		e.cancelCompletedLine()
	}

	defer e.refreshLine()

	// Move the selector
	done, next := grp.moveSelector(row, column)
	if !done {
		return
	}

	var newGrp *group

	if next {
		e.cycleNextGroup()
		newGrp = e.currentGroup()
		newGrp.firstCell()
	} else {
		e.cyclePreviousGroup()
		newGrp = e.currentGroup()
		newGrp.lastCell()
	}
}

// SelectTag allows to select the first value of the next tag (next=true),
// or the last value of the previous tag (next=false).
func (e *Engine) SelectTag(next bool) {
	// Ensure the completion keymaps are set.
	e.adjustSelectKeymap()

	if len(e.groups) <= 1 {
		return
	}

	// If the completion candidate is not empty,
	// it's also inserted in the line, so remove it.
	if len(e.selected.Value) > 0 {
		e.cancelCompletedLine()
	}

	// In the end we will update the line with the
	// newly/currently selected completion candidate.
	defer e.refreshLine()

	if next {
		e.cycleNextGroup()
		newGrp := e.currentGroup()
		newGrp.firstCell()
	} else {
		e.cyclePreviousGroup()
		newGrp := e.currentGroup()
		newGrp.firstCell()
	}
}

// Update should be called only once in between the two shell keymaps
// (local/main), to either drop or confirm a virtually inserted candidate.
func (e *Engine) Update() {
	// If the user currently has a completion selected, any change
	// in the input line will drop the current completion list, in
	// effect deactivating the completion engine.
	// This is so that when a user asks for the list of choices, but
	// then deletes or types something in the input line, the list
	// is still displayed to the user, otherwise it's removed.
	// This does not apply when autocomplete is on.
	choices := len(e.selected.Value) != 0
	if !e.auto {
		defer e.ClearMenu(choices)
	}

	// If autocomplete is on, we also drop the list of generated
	// completions, because it will be recomputed shortly after.
	// Do the same when using incremental search, except if the
	// last key typed is an escape, in which case the user wants
	// to quit incremental search but keeping any selected comp.
	inserted := e.removeInserted()

	e.Cancel(inserted, false)

	if choices && e.autoForce && len(e.selected.Value) == 0 {
		e.Reset()
	}
}

// Cancel exits the current completions with the following behavior:
// - If inserted is true, any inserted candidate is removed.
// - If cached is true, any cached completer function is dropped.
// This function does not exit the completion keymap,
// so any active completion menu will still be kept.
func (e *Engine) Cancel(inserted, cached bool) {
	if cached {
		e.cached = nil
	}

	if len(e.selected.Value) == 0 && !inserted {
		return
	}

	// In the end, there is no completed line anymore.
	defer e.cancelCompletedLine()

	// Either drop the inserted candidate.
	if inserted {
		cpos := e.cursor.Pos()
		e.cursor.Move(-1 * len(e.inserted))
		e.line.Cut(e.cursor.Pos(), cpos)
		e.line.Insert(e.cursor.Pos(), []rune(e.suffix)...)

		return
	}

	// Or make it part of the real input line.
	e.line.Set(*e.completed...)
	e.cursor.Set(e.compCursor.Pos())
}

// ResetForce drops any currently inserted candidate from the line,
// drops any cached completer function and generated list, and exits
// the incremental-search mode.
// All those steps are performed whether or not the engine is active.
func (e *Engine) ResetForce() {
	e.Cancel(!e.autoForce, true)
	e.ClearMenu(true)
	e.IsearchStop()
	e.autoForce = false
}

// Reset accepts the currently inserted candidate (if any), clears the current
// list of completions and exits the incremental-search mode if active.
// If the completion engine was not active to begin with, nothing will happen.
func (e *Engine) Reset() {
	e.autoForce = false
	if !e.IsActive() {
		return
	}

	e.Cancel(false, true)
	e.ClearMenu(true)
	e.IsearchStop()
}

// ClearMenu exits the current completion keymap (if set) and clears
// the current list of generated completions (if completions is true).
func (e *Engine) ClearMenu(completions bool) {
	e.resetValues(completions, false)

	if e.keymaps.Local() == keymap.MenuSelect {
		e.keymaps.SetLocal("")
	}
}

// IsActive indicates if the engine is currently in possession of a
// non-empty list of generated completions (following all constraints).
func (e *Engine) IsActive() bool {
	return e.keymaps.Local() == keymap.MenuSelect ||
		e.keymaps.Local() == keymap.Isearch ||
		e.auto || e.autoForce
}

// IsInserting returns true if a candidate is currently virtually inserted.
func (e *Engine) IsInserting() bool {
	return e.selected.Value != ""
}

// Matches returns the number of completion candidates
// matching the current line/settings requirements.
func (e *Engine) Matches() int {
	comps, _ := e.completionCount()
	return comps
}

// Line returns the relevant input line at the time this function is called:
// if a candidate is currently selected, the line returned is the one containing
// the candidate. If no candidate is selected, the normal input line is returned.
// When the line returned is the completed one, the corresponding, adjusted cursor.
func (e *Engine) Line() (*core.Line, *core.Cursor) {
	if len(e.selected.Value) > 0 {
		return e.completed, e.compCursor
	}

	return e.line, e.cursor
}

// Autocomplete generates the correct completions in autocomplete mode.
// We don't do it when we are currently in the completion keymap,
// since that means completions have already been computed.
func (e *Engine) Autocomplete() {
	e.auto = e.needsAutoComplete()

	// Clear the current completion list when we are at the
	// beginning of the line, and not currently completing.
	if e.auto || (!e.IsActive() && e.cursor.Pos() == 0) {
		e.resetValues(true, false)
	}

	// We are not auto when either: autocomplete is disabled,
	// incremental-search mode is active, or a completion is
	// currently selected in the menu.
	if !e.auto {
		return
	}

	// Regenerate the completions.
	if e.cached != nil {
		e.prepare(e.cached())
	} else if e.defaut != nil {
		e.prepare(e.defaut())
	}
}

// Enable as-you-type autocomplete on the real input line,
// even if the current cursor position is 0.
func (e *Engine) AutocompleteForce() {
	e.autoForce = true
}

// IsAutoCompleting returns true if the completion engine is an
// autocompletion mode: refreshing results on each input key.
func (e *Engine) IsAutoCompleting() bool {
	return e.autoForce
}

// SetAutocompleter sets the default completer to use in autocomplete mode.
// This completer is different from the one passed to e.GenerateWith()
// This is used so that autocomplete does not use lists such as history,
// Vim registers, etc, instead of just generating normal command completion.
func (e *Engine) SetAutocompleter(comp Completer) {
	e.defaut = comp
}

// CompleteSyntax updates the line with either user-defined syntax completers, or with the builtin ones.
func (e *Engine) CompleteSyntax(completer func([]rune, int) ([]rune, int)) {
	if completer == nil {
		return
	}

	line := []rune(*e.line)
	pos := e.cursor.Pos() - 1

	newLine, newPos := completer(line, pos)
	if string(newLine) == string(line) {
		return
	}

	newPos++

	e.line.Set(newLine...)
	e.cursor.Set(newPos)
}
