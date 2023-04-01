package display

import (
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/history"
	"github.com/reeflective/readline/internal/term"
	"github.com/reeflective/readline/internal/ui"
	"github.com/xo/inputrc"
)

// Engine handles all display operations: it refreshes the terminal
// interface and stores the necessary offsets of each components.
type Engine struct {
	// Operating parameters
	highlighter func(line []rune) string
	startAt     int
	lineRows    int
	cursorRow   int
	cursorCol   int
	hintRows    int
	compRows    int

	// UI components
	cursor    *core.Cursor
	selection *core.Selection
	histories *history.Sources
	prompt    *ui.Prompt
	hint      *ui.Hint
	completer *completion.Engine
	opts      *inputrc.Config
}

// NewEngine is a required constructor for the display engine.
func NewEngine(s *core.Selection, h *history.Sources, p *ui.Prompt, i *ui.Hint, c *completion.Engine, opts *inputrc.Config) *Engine {
	return &Engine{
		selection: s,
		histories: h,
		prompt:    p,
		hint:      i,
		completer: c,
		opts:      opts,
	}
}

// Init computes some base coordinates needed before displaying the line and helpers.
// The shell syntax highlighter is also provided here, since any consumer library will
// have bound it after instantiating a new shell instance.
func (e *Engine) Init(highlighter func([]rune) string) {
	e.highlighter = highlighter

	var line *core.Line

	// Some coordinates must be available before all else.
	line, e.cursor = e.completer.Line()
	suggested := e.histories.Suggest(line)
	e.computeCoordinates(suggested)

	// rl.resetHintText()
	// rl.resetCompletion()
	// rl.completer = nil
	// rl.getHintText()
}

// Refresh recomputes and redisplays the entire readline interface,
// except the the first lines of the primary prompt when the latter
// is a multiline one.
func (e *Engine) Refresh() {
	e.CursorToLineStart()
	print(term.ClearScreenBelow)

	var line *core.Line

	// Use the coordinates computed during the last refresh
	line, e.cursor = e.completer.Line()
	suggested := e.histories.Suggest(line)

	// Get all positions required for the redisplay to come:
	// prompt end (thus indentation), cursor positions, etc.
	e.computeCoordinates(suggested)

	// Display line and go to cursor, and right prompt if any.
	e.displayLine(suggested)
	term.MoveCursorUp(e.lineRows - e.cursorRow)
	e.prompt.RightPrint(line, e.cursor)

	// Go to the last row of the line, and display hints.
	e.CursorBelowLine()
	print(term.ClearScreenBelow)
	e.hint.Display()
	e.hintRows = e.hint.Coordinates()

	// Display completions.
	e.displayCompletions()

	// Go back to cursor position.
	term.MoveCursorUp(e.hintRows + 1)
	term.MoveCursorUp(e.lineRows - e.cursorRow)
	term.MoveCursorForwards(e.cursorCol)
}

// ClearHelpers clears and resets the hint and completion sections.
func (e *Engine) ClearHelpers() {
	cursorCols, cursorRows := e.cursor.Coordinates(e.startAt)
	term.MoveCursorDown((e.lineRows - cursorRows) + 1)

	print(term.ClearScreenBelow)

	term.MoveCursorUp((e.lineRows - cursorRows) + 1)
	term.MoveCursorForwards(cursorCols)
}

// ResetHelpers cancels all active hints and completions.
func (e *Engine) ResetHelpers() {
	e.hint.Reset()
	e.completer.Reset(true, false)
}

// AcceptLine redraws the current UI when the line has been accepted
// and returned to the caller. After clearing various things such as
// hints, completions and some right prompts, the shell will put the
// display at the start of the line immediately following the line.
func (e *Engine) AcceptLine() {
	e.ClearHelpers()
	e.prompt.RightClear(false)
	e.CursorBelowLine()
	print(term.ClearScreenBelow)
}

// CursorBelowLine moves the cursor to the leftmost column
// of the first row after the last line of input.
func (e *Engine) CursorBelowLine() {
	term.MoveCursorDown((e.lineRows - e.cursorRow) + 1)
	term.MoveCursorBackwards(term.GetWidth())
}

// CursorToPos moves the cursor back to
// where the cursor is on the input line.
func (e *Engine) CursorToPos() {
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorUp(e.lineRows - e.cursorRow)
	term.MoveCursorForwards(e.cursorCol)
}

// CursorBelowHint moves the cursor to the leftmost
// column of the first line after the hint section.
func (e *Engine) CursorBelowHint() {
	term.MoveCursorDown(e.lineRows - e.cursorRow)
	term.MoveCursorBackwards(term.GetWidth())
}

// CursorToLineStart moves the cursor just after the primary prompt.
func (e *Engine) CursorToLineStart() {
	term.MoveCursorBackwards(e.cursorCol - e.startAt)
	term.MoveCursorUp(e.cursorRow)
}

func (e *Engine) computeCoordinates(suggested core.Line) {
	e.startAt = e.prompt.LastUsed()
	_, e.lineRows = suggested.Coordinates(e.startAt)
	e.cursorCol, e.cursorRow = e.cursor.Coordinates(e.startAt)
}

func (e *Engine) displayLine(suggested core.Line) {
	var highlighted string

	// Apply user-defined highlighter if any
	if e.highlighter != nil {
		highlighted = e.highlighter([]rune(suggested))
	} else {
		highlighted = string(suggested)
	}

	// Apply visual selections highlighting if any.
	highlighted = ui.Highlight([]rune(highlighted), *e.selection)

	// And display the line.
	suggested.Set([]rune(highlighted)...)
	suggested.Display(e.startAt)
}

func (e *Engine) displayCompletions() {
	// TODO: Here autocomplete call
	// rl.autoComplete()
	e.completer.Display()
	e.compRows = e.completer.Coordinates()

	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorUp(e.compRows)
}
