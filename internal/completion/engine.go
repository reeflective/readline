package completion

import (
	"regexp"

	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/term"
	"github.com/reeflective/readline/internal/ui"
	"github.com/xo/inputrc"
)

// Engine is responsible for all completion tasks: generating, computing,
// displaying and updating completion values and inserted candidates.
type Engine struct {
	opts   *inputrc.Config // The inputrc contains options relative to completion.
	cached Completer       // A cached completer function to use when updating.
	hint   *ui.Hint        // The completions can feed hint/usage messages

	// Completion parameters
	groups []*group      // All of our suggestions tree is in here
	prefix string        // The current tab completion prefix  against which to build candidates
	usedY  int           // Comprehensive offset of the currently built completions
	comp   []rune        // The currently selected item, not yet a real part of the input line.
	suffix SuffixMatcher // The suffix matcher is kept for removal after actually inserting the candidate.

	// Line parameters
	keys       *core.Keys   // The input keys reader
	line       *core.Line   // The real input line of the shell.
	cursor     *core.Cursor // The cursor of the shell.
	completed  *core.Line   // A line that might include a virtually inserted candidate.
	compCursor *core.Cursor // The adjusted cursor.

	// Incremental search
	isearchBuf *core.Line   // The isearch minibuffer
	isearch    *regexp.Regexp // Holds the current search regex match
}

// NewEngine initializes a new completion engine with the shell operating parameters.
func NewEngine(k *core.Keys, l *core.Line, c *core.Cursor, h *ui.Hint, o *inputrc.Config) *Engine {
	return &Engine{
		opts:   o,
		keys:   k,
		line:   l,
		cursor: c,
		hint:   h,
	}
}

// Generate uses a list of completions to group/order
// and prepare completions before printing them.
func (e *Engine) Generate(completions Values) {
}

// Cache accepts a completer function so that the next time
// it has to update its results, it can reuse this completer.
func (e *Engine) Cache(completer Completer) {
	e.cached = completer
}

// Select moves the completion selector by some X or Y value.
func (e *Engine) Select(key string, column, row int) {
	grp := e.currentGroup()

	// If there is no current group, we
	// leave any current completion mode.
	if grp == nil || len(grp.values) == 0 {
		return
	}

	done, next := grp.moveSelector(column, row)
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

// SelectTag allows to select the first value of the next tag
// (if next is true), or the last value of the previous tag
// (if next is false). If there is only one tag, nothing is done.
func (e *Engine) SelectTag(next bool) {
	if len(e.groups) <= 1 {
		return
	}

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
	if len(e.comp) > 0 {
		return e.completed, e.compCursor
	}

	return e.line, e.cursor
}

// Update refreshes the completion list according
// to the current settings and line constraints.
func (e *Engine) Update() {
}

// IsearchStart starts incremental search (fuzzy-finding)
// with values matching the isearch minibuffer as a regexp.
func (e *Engine) IsearchStart(buffer *core.Line) {
}

// IsearchStop exists the incremental search mode.
func (e *Engine) IsearchStop() {
}

// TrimSuffix removes the last inserted completion's suffix if the required constraints
// are satisfied, among which the index position, the suffix matching patterns, etc.
func (e *Engine) TrimSuffix() {
}

// Reset exits the current completions, and either inserts
// the currently selected candidate (if any) into the input
// line (if drop is false), or drops it (if drop is true).
func (e *Engine) Reset(drop bool) {
	if len(e.comp) == 0 {
		return
	}

	// Get the current candidate and its group.
	cur := e.currentGroup()
	if cur == nil {
		return
	}

	defer e.clearVirtualComp()

	completion := cur.selected().Value

	if completion == "" {
		return
	}

	if drop {
		e.completed.Set(*e.line...)
		e.compCursor.Set(e.cursor.Pos())

		return
	}

	completion = e.removeSuffixInserted()
	e.insertCandidate([]rune(completion[len(e.prefix):]))
}

// Display prints the current completion list to the screen,
// respecting the current display and completion settings.
func (e *Engine) Display() {
	e.usedY = 0

	// The final completions string to print.
	var completions string

	for _, group := range e.groups {
		completions += group.writeComps(e)
	}

	// Crop the completions so that it fits within our terminal
	completions, e.usedY = e.cropCompletions(completions)

	if completions != "" {
		print("\n")
		e.usedY++

		print(term.ClearScreenBelow)
	}

	print(completions)
}

// Coordinates returns the number of terminal rows used
// when displaying the completions with Display().
func (e *Engine) Coordinates() int {
	return e.usedY
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
