package display

import (
	"fmt"
	"regexp"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/history"
	"github.com/reeflective/readline/internal/strutil"
	"github.com/reeflective/readline/internal/term"
	"github.com/reeflective/readline/internal/ui"
)

var (
	oneThirdTerminalHeight = 3
	halfTerminalHeight     = 2
)

// Engine handles all display operations: it refreshes the terminal
// interface and stores the necessary offsets of each components.
type Engine struct {
	// Operating parameters
	highlighter    func(line []rune) string
	startCols      int
	startRows      int
	lineCol        int
	lineRows       int
	cursorRow      int
	cursorCol      int
	hintRows       int
	compRows       int
	primaryPrinted bool

	// commentRegex is the compiled comment-highlight pattern. It is rebuilt
	// only when the comment-begin option changes, not on every refresh.
	commentToken string
	commentRegex *regexp.Regexp

	// UI components
	keys      *core.Keys
	line      *core.Line
	suggested core.Line
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
func Init(e *Engine, highlighter func([]rune) string) {
	e.highlighter = highlighter
}

// PrintPrimaryPrompt redraws the primary prompt.
// There are relatively few cases where you want to use this.
// It is currently only used when using clear-screen commands.
func (e *Engine) PrintPrimaryPrompt() {
	e.prompt.PrimaryPrint()
	e.primaryPrinted = true
}

// ClearHelpers clears the hint and completion sections below the line.
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

	e.computeCoordinates(false)

	// Go back to the end of the non-suggested line.
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorDown(e.lineRows)
	term.MoveCursorForwards(e.lineCol)
	fmt.Print(term.ClearScreenBelow)

	// Reprint the right-side prompt if it's not a tooltip one.
	e.prompt.RightPrint(e.lineCol, false)

	// Go below this non-suggested line and clear everything.
	term.MoveCursorBackwards(term.GetWidth())
	fmt.Print(term.NewlineReturn)
}

// RefreshTransient goes back to the first line of the input buffer
// and displays the transient prompt, then redisplays the input line.
func (e *Engine) RefreshTransient() {
	if !e.opts.GetBool("prompt-transient") {
		return
	}

	// Go to the beginning of the primary prompt.
	e.CursorToLineStart()
	term.MoveCursorUp(e.prompt.PrimaryUsed())

	// And redisplay the transient/primary/line.
	e.prompt.TransientPrint()
	e.displayLine()
	fmt.Print(term.NewlineReturn)
}

// CursorToLineStart moves the cursor just after the primary prompt.
// This function should only be called when the cursor is on its
// "cursor" position on the input line.
func (e *Engine) CursorToLineStart() {
	term.MoveCursorBackwards(e.cursorCol)
	term.MoveCursorUp(e.cursorRow)
	term.MoveCursorForwards(e.startCols)
}

// CursorBelowLine moves the cursor to the leftmost
// column of the first row after the last line of input.
// This function should only be called when the cursor
// is on its "cursor" position on the input line.
func (e *Engine) CursorBelowLine() {
	term.MoveCursorUp(e.cursorRow)
	term.MoveCursorDown(e.lineRows)
	fmt.Print(term.NewlineReturn)
}

func (e *Engine) computeCoordinates(suggested bool) {
	// Get the new input line and auto-suggested one.
	e.line, e.cursor = e.completer.Line()
	if e.completer.IsInserting() {
		e.suggested = *e.line
	} else {
		e.suggested = e.histories.Suggest(e.line)
	}

	// Recompute the passive provider hint from the current line, so it tracks
	// the input as it changes. Runs every refresh, on the main loop goroutine.
	e.hint.UpdateProvided([]rune(*e.line), e.cursor.Pos())

	// Get the position of the line's beginning by querying the terminal for the
	// cursor position. Some environments (PTY test harnesses, minimal emulators,
	// constrained CI) don't reliably answer the "ESC[6n" query, so consumers can
	// turn the cursor-position-probe option off; we then fall back to a position
	// derived from the printed prompt width.
	if e.opts.GetBool("cursor-position-probe") {
		e.startCols, e.startRows = e.keys.GetCursorPos()
	} else {
		e.startCols, e.startRows = -1, -1
	}

	if e.startCols > 0 {
		e.startCols--
	}

	// Cursor column might be misleading if invalid (negative), or unavailable
	// because probing is disabled: fall back to the printed prompt width. This
	// is exact whenever the input line starts at column 0 (the common case).
	if e.startCols == -1 {
		e.startCols = e.prompt.LastUsed()
	}

	e.cursorCol, e.cursorRow = core.CoordinatesCursor(e.cursor, e.startCols)

	// Get the number of rows used by the line, and the end line X pos.
	if e.opts.GetBool("history-autosuggest") && suggested {
		e.lineCol, e.lineRows = core.CoordinatesLine(&e.suggested, e.startCols)
	} else {
		e.lineCol, e.lineRows = core.CoordinatesLine(e.line, e.startCols)
	}

	e.primaryPrinted = false
}

func (e *Engine) displayLine() {
	var line string

	// Apply user-defined highlighter to the input line.
	if e.highlighter != nil {
		line = e.highlighter(*e.line)
	} else {
		line = string(*e.line)
	}

	// Highlight matching parenthesis
	if e.opts.GetBool("blink-matching-paren") {
		core.HighlightMatchers(e.selection)
		defer core.ResetMatchers(e.selection)
	}

	// Apply visual selections highlighting if any
	line = e.highlightLine([]rune(line), *e.selection)

	// Get the subset of the suggested line to print.
	if len(e.suggested) > e.line.Len() && e.opts.GetBool("history-autosuggest") {
		line += color.Dim + color.Fmt(color.Fg+"242") + string(e.suggested[e.line.Len():]) + color.Reset
	}

	// Format tabs as spaces, for consistent display
	line = strutil.FormatTabs(line) + term.ClearLineAfter

	// And display the line.
	e.suggested.Set([]rune(line)...)
	core.DisplayLine(&e.suggested, e.startCols)

	// Adjust the cursor if the line fits exactly in the terminal width.
	if e.lineCol == 0 {
		fmt.Print(term.NewlineReturn)
		fmt.Print(term.ClearLineAfter)
	}
}

// AvailableHelperLines returns the number of lines available below the hint section.
// It returns half the terminal space if we currently have less than 1/3rd of it below.
func (e *Engine) AvailableHelperLines() int {
	termHeight := term.GetLength()
	compLines := termHeight - e.startRows - e.lineRows - e.hintRows

	if compLines < (termHeight / oneThirdTerminalHeight) {
		compLines = (termHeight / halfTerminalHeight)
	}

	return compLines
}
