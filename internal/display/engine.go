package display

import (
	"github.com/reeflective/readline/internal/common"
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/history"
	"github.com/reeflective/readline/internal/term"
	"github.com/reeflective/readline/internal/ui"
	"github.com/xo/inputrc"
)

// Engine handles all display operations: it refreshes the terminal
// interface and stores the necessary offsets of each components.
type Engine struct {
	// Operating parameters
	startAt     int
	lineRows    int
	cursorRow   int
	highlighter func(line []rune) string

	// UI components
	cursor    *common.Cursor
	selection *common.Selection
	histories *history.Sources
	prompt    *ui.Prompt
	hint      *ui.Hint
	completer *completion.Engine
	opts      *inputrc.Config
}

// NewEngine is a required constructor for the display engine.
func NewEngine(s *common.Selection, h *history.Sources, p *ui.Prompt, i *ui.Hint, c *completion.Engine) *Engine {
	return &Engine{
		selection: s,
		histories: h,
		prompt:    p,
		hint:      i,
		completer: c,
	}
}

func (e *Engine) Init() {
	// rl.resetHintText()
	// rl.resetCompletion()
	// rl.completer = nil
	// rl.getHintText()
}

// Refresh recomputes and redisplays the entire readline interface,
// except the the first lines of the primary prompt when the latter
// is a multiline one.
func (e *Engine) Refresh() {
	var line *common.Line

	// Get the completed line (if completions are active),
	// and the corresponding cursor, and find any suggested line.
	line, e.cursor = e.completer.Line()
	suggested := e.histories.Suggest(line)

	// Go back at start of the prompt.
	promptCols := e.prompt.LastUsed()
	_, e.cursorRow = e.cursor.Coordinates(promptCols)
	term.MoveCursorBackwards(e.startAt)
	// term.MoveCursorBackwards(cursorCols - promptCols - 1)
	term.MoveCursorUp(e.cursorRow)

	// Apply user-defined highlighting, then apply visual on top.
	highlighted := e.highlighter([]rune(suggested))
	highlighted = ui.Highlight([]rune(highlighted), *e.selection)
	suggested.Set([]rune(highlighted)...)
	suggested.Display(promptCols)

	// Go back to the cursor position and print any right prompt.
	_, e.lineRows = suggested.Coordinates(promptCols)
	cursorCols, cursorRows := e.cursor.Coordinates(promptCols)
	term.MoveCursorUp(e.lineRows - cursorRows)
	e.prompt.RightPrint(line, e.cursor)

	// Go to the last row of the line, and display hints.
	term.MoveCursorDown(e.lineRows - cursorRows)
	hintRows := e.hint.Coordinates()
	e.hint.Display()

	// Display completions.
	// TODO: Here autocomplete call
	// rl.autoComplete()
	compRows := e.completer.Coordinates()
	e.completer.Display()

	// Go back to cursor position.
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorUp(compRows + hintRows)
	term.MoveCursorUp(e.lineRows - cursorRows)
	term.MoveCursorForwards(cursorCols)

	e.startAt = cursorCols - promptCols
}

// ClearHelpers clears and resets the hint and completion sections.
func (e *Engine) ClearHelpers() {
	cursorCols, cursorRows := e.cursor.Coordinates(e.startAt)
	term.MoveCursorDown(e.lineRows - cursorRows)

	print(term.ClearScreenBelow)

	term.MoveCursorUp(e.lineRows - cursorRows)
	term.MoveCursorForwards(cursorCols)
}

// ResetHelpers cancels all active hints and completions.
func (e *Engine) ResetHelpers() {
	e.hint.Reset()
	e.completer.Reset(true)
}

// CursorBelowLine moves the cursor to the leftmost column
// of the first row after the last line of input.
func (e *Engine) CursorBelowLine() {
	term.MoveCursorDown(e.lineRows - e.cursorRow)
	term.MoveCursorBackwards(term.GetWidth())
	// term.MoveCursorForwards(rl.endLineX())
}

// CursorToLineCursor moves the cursor back
// to where the cursor is on the input line.
func (e *Engine) CursorToLineCursor() {
	// moveCursorBackwards(GetTermWidth())
	// moveCursorUp(rl.fullY)
	// moveCursorDown(rl.posY)
	// moveCursorForwards(rl.posX)
}

// CursorBelowHint moves the cursor to the leftmost
// column of the first line after the hint section.
func (e *Engine) CursorBelowHint() {
	term.MoveCursorDown(e.lineRows - e.cursorRow)
	term.MoveCursorBackwards(term.GetWidth())
}

// CursorToLineStart moves the cursor just after the primary prompt.
func (e *Engine) CursorToLineStart() {
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorUp(e.cursorRow)
	term.MoveCursorForwards(e.startAt)
}

func (e *Engine) moveFromHelpersEndToHintStart() {
	// moveCursorBackwards(GetTermWidth())
	// moveCursorUp(rl.tcUsedY)
	// if len(rl.hint) > 0 {
	// 	moveCursorUp(rl.hintY)
	// }
}

func (e *Engine) displayCompletions() {
	// Display completions.
	// TODO: Here autocomplete call
	// rl.autoComplete()
	// compRows := rl.completer.Coordinates()
	// rl.completer.Display()

	// term.MoveCursorUp(compRows)
}

// renderHelpers - prints all components (prompt, line, hints & comps)
// and replaces the cursor to its current position. This function never
// computes or refreshes any value, except from inside the echo function.
// func (rl *Instance) renderHelpers() {
// if rl.config.HistoryAutosuggest {
// 	rl.autosuggestHistory(rl.lineCompleted())
// }
// rl.linePrint()

// Go at beginning of the last line of input
// rl.moveToHintStart()

// Print hints, check for any confirmation hint current.
// (do not overwrite the confirmation question hint)
// rl.writeHintText()
// moveCursorBackwards(GetTermWidth())

// Print completions and go back
// to beginning of this line
// rl.printCompletions()

// And move back to the last line of input, then to the cursor.
// rl.moveFromHelpersEndToHintStart()
// rl.moveFromLineEndToCursor()
// }
