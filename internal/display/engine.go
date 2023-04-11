package display

import (
	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/history"
	"github.com/reeflective/readline/internal/term"
	"github.com/reeflective/readline/internal/ui"
)

// Engine handles all display operations: it refreshes the terminal
// interface and stores the necessary offsets of each components.
type Engine struct {
	// Operating parameters
	highlighter func(line []rune) string
	startAt     int
	lineCol     int
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
}

// Refresh recomputes and redisplays the entire readline interface, except
// the first lines of the primary prompt when the latter is a multiline one.
func (e *Engine) Refresh() {
	// Recompute completions and related hints if autocompletion is on.
	e.completer.Autocomplete()

	// Go back to the end of the prompt.
	e.CursorBelowLine()
	print(term.ClearScreenBelow)
	term.MoveCursorUp(1)
	e.CursorToPos()
	print(term.HideCursor)
	e.CursorToLineStart()

	var line *core.Line

	// Get desired input line and auto-suggested one.
	line, e.cursor = e.completer.Line()
	suggested := e.histories.Suggest(line)

	// Get all positions required for the redisplay to come:
	// prompt end (thus indentation), cursor positions, etc.
	e.computeCoordinates(suggested)

	// Print the line, and adjust the cursor if the line fits
	// exactly in the terminal width, then print the right prompt.
	e.displayLine(*line, suggested)
	if e.lineCol == 0 && e.cursorCol == 0 && e.cursorRow > 1 {
		term.MoveCursorDown(1)
	}
	e.prompt.RightPrint(e.startAt+e.lineCol, true)

	// Display the hint section and completions.
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorDown(1)
	print(term.ClearScreenBelow)
	e.hint.Display()
	e.displayCompletions()
	term.MoveCursorUp(e.hint.Coordinates())

	// Go back to the start of the line, then to cursor.
	term.MoveCursorUp(e.lineRows)
	e.LineStartToCursorPos()
	print(term.ShowCursor)
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
	e.completer.Cancel(true, false)
}

// AcceptLine redraws the current UI when the line has been accepted
// and returned to the caller. After clearing various things such as
// hints, completions and some right prompts, the shell will put the
// display at the start of the line immediately following the line.
func (e *Engine) AcceptLine() {
	e.CursorToLineStart()

	line, _ := e.completer.Line()
	e.computeCoordinates(*line)

	// Go back to the end of the non-suggested line.
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorDown(e.lineRows)
	term.MoveCursorForwards(e.lineCol)
	print(term.ClearScreenBelow)

	// Reprint the right-side prompt if it's not a tooltip one.
	e.prompt.RightPrint(e.lineCol, false)

	// Go below this non-suggested line and clear everything.
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorDown(1)
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
	term.MoveCursorForwards(e.cursorCol)

	term.MoveCursorUp(e.lineRows)
	term.MoveCursorDown(e.cursorRow)
}

// LineStartToCursorPos can be used if the cursor is currently
// at the very start of the input line, that is just after the
// last character of the prompt.
func (e *Engine) LineStartToCursorPos() {
	term.MoveCursorDown(e.cursorRow)
	term.MoveCursorBackwards(term.GetWidth())
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
	term.MoveCursorBackwards(e.cursorCol)
	term.MoveCursorUp(e.cursorRow)
	term.MoveCursorForwards(e.startAt)
}

func (e *Engine) computeCoordinates(suggested core.Line) {
	e.startAt = e.prompt.LastUsed()
	e.lineCol, e.lineRows = suggested.Coordinates(e.startAt)
	e.cursorCol, e.cursorRow = e.cursor.Coordinates(e.startAt)
}

func (e *Engine) displayLine(input, suggested core.Line) {
	var line string

	// Apply user-defined highlighter to the input line.
	if e.highlighter != nil {
		line = e.highlighter([]rune(input))
	} else {
		line = string(input)
	}

	// Apply visual selections highlighting if any.
	line = ui.Highlight([]rune(line), *e.selection)

	// Get the subset of the suggested line to print.
	if len(suggested) > len(input) {
		line += color.FgBlackBright + string(suggested[len(input):]) + color.Reset
	}

	// And display the line.
	suggested.Set([]rune(line)...)
	suggested.Display(e.startAt)
}

func (e *Engine) displayCompletions() {
	e.completer.Display()
	e.compRows = e.completer.Coordinates()

	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorUp(e.compRows)
}
