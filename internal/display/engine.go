package display

import (
	"fmt"
	"os"

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
	startCols   int
	startRows   int
	lineCol     int
	lineRows    int
	cursorRow   int
	cursorCol   int
	compRows    int

	// UI components
	keys      *core.Keys
	line      *core.Line
	cursor    *core.Cursor
	selection *core.Selection
	histories *history.Sources
	prompt    *ui.Prompt
	hint      *ui.Hint
	completer *completion.Engine
	opts      *inputrc.Config
}

// NewEngine is a required constructor for the display engine.
func NewEngine(k *core.Keys, s *core.Selection, h *history.Sources, p *ui.Prompt, i *ui.Hint, c *completion.Engine, opts *inputrc.Config) *Engine {
	return &Engine{
		keys:      k,
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
	// Clear everything below the current input line.
	fmt.Print(term.HideCursor)
	e.CursorBelowLine()
	e.CursorHintToLineStart()

	// Get the new input line and auto-suggested one.
	e.line, e.cursor = e.completer.Line()
	suggested := e.histories.Suggest(e.line)

	// Get all positions required for the redisplay to come:
	// prompt end (thus indentation), cursor positions, etc.
	e.computeCoordinates(suggested)

	// term.MoveCursorBackwards(e.startCols)
	// e.prompt.LastPrint()
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorForwards(e.startCols)

	// Print the line, right prompt, hints and completions.
	e.displayLine(suggested)
	e.prompt.RightPrint(e.lineCol, true)
	e.displayHelpers()

	// Go back to the start of the line, then to cursor.
	e.CursorHintToLineStart()
	e.LineStartToCursorPos()
	fmt.Print(term.ShowCursor)
}

// ClearHelpers clears and resets the hint and completion sections.
func (e *Engine) ClearHelpers() {
	e.CursorBelowLine()
	fmt.Print(term.ClearScreenBelow)

	term.MoveCursorUp(1)
	term.MoveCursorUp(e.lineRows)
	term.MoveCursorDown(e.cursorRow)
	term.MoveCursorForwards(e.cursorCol)
}

// ResetHelpers cancels all active hints and completions.
func (e *Engine) ResetHelpers() {
	e.hint.Reset()
	e.completer.ClearMenu(true)
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
	fmt.Print(term.ClearScreenBelow)

	// Reprint the right-side prompt if it's not a tooltip one.
	e.prompt.RightPrint(e.lineCol, false)

	// Go below this non-suggested line and clear everything.
	fmt.Println()
	// term.MoveCursorBackwards(term.GetWidth())
	// term.MoveCursorDown(1)
}

// CursorBelowLine moves the cursor to the leftmost column
// of the first row after the last line of input.
func (e *Engine) CursorBelowLine() {
	term.MoveCursorUp(e.cursorRow)
	term.MoveCursorDown(e.lineRows)
	term.MoveCursorDown(1)
}

// CursorToPos moves the cursor back to where the cursor is on the input line.
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
	term.MoveCursorForwards(e.startCols - 1)
}

// cursor is on the line below the last line of input.
func (e *Engine) CursorHintToLineStart() {
	term.MoveCursorUp(1)
	term.MoveCursorUp(e.lineRows - e.cursorRow)
	e.CursorToLineStart()
}

func (e *Engine) computeCoordinates(suggested core.Line) {
	// Get the cursor position through terminal query:
	// we have printed our prompt, and we are at the start pos.
	e.startCols, e.startRows = e.keys.GetCursorPos()

	if e.startCols == -1 {
		e.startCols = e.prompt.LastUsed()
	}

	e.cursorCol, e.cursorRow = e.cursor.Coordinates(e.startCols)

	if e.opts.GetBool("history-autosuggest") {
		e.lineCol, e.lineRows = suggested.Coordinates(e.startCols)
	} else {
		e.lineCol, e.lineRows = e.line.Coordinates(e.startCols)
	}
}

func (e *Engine) displayLine(suggested core.Line) {
	var line string

	// Apply user-defined highlighter to the input line.
	if e.highlighter != nil {
		line = e.highlighter(*e.line)
	} else {
		line = string(*e.line)
	}

	// Apply visual selections highlighting if any.
	line = ui.Highlight([]rune(line), *e.selection)

	// Get the subset of the suggested line to print.
	if len(suggested) > e.line.Len() && e.opts.GetBool("history-autosuggest") {
		line += color.FgBlackBright + string(suggested[e.line.Len():]) + color.Reset
	}

	// And display the line.
	suggested.Set([]rune(line)...)
	suggested.Display(e.startCols)

	// Adjust the cursor if the line fits exactly in the terminal width.
	if e.lineCol == 0 && e.cursorCol == 0 && e.cursorRow > 1 {
		term.MoveCursorDown(1)
	}
}

// displayHelpers renders the hint and completion sections.
// It assumes that the cursor is on the last line of input,
// and goes back to this same line after displaying this.
func (e *Engine) displayHelpers() {
	// Clear everything below the input line.
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorDown(1)

	// Recompute completions and hints if autocompletion is on.
	e.completer.Autocomplete()

	// Compute the number of available lines we have for displaying completions.
	_, termHeight, _ := term.GetSize(int(os.Stdin.Fd()))
	compLines := termHeight - e.startRows - e.lineRows - e.hint.Coordinates() - 1

	// Display hint and completions.
	e.hint.Display()
	e.completer.Display(compLines)
	e.compRows = e.completer.Coordinates()

	// Go back to the first line below the input line.
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorUp(e.compRows)
	term.MoveCursorUp(e.hint.Coordinates())
}
