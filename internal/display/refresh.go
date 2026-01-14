package display

import (
	"fmt"
	"strconv"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/strutil"
	"github.com/reeflective/readline/internal/term"
	"github.com/reeflective/readline/internal/ui"
)

// Refresh recomputes and redisplays the entire readline interface, except
// the first lines of the primary prompt when the latter is a multiline one.
func (e *Engine) Refresh() {
	// 1. Preparation & Coordinates
	fmt.Print(term.HideCursor)
	// Go back to the first column, and if the primary prompt
	// was not printed yet, back up to the line's beginning row.
	term.MoveCursorBackwards(term.GetWidth())

	if !e.primaryPrinted {
		term.MoveCursorUp(e.cursorRow)
	}
	// 2. Primary Prompt
	e.prompt.LastPrint()
	// Compute Coordinates: StartPos, LineHeight, CursorPos (row/col).
	e.computeCoordinates(true)

	// Determine the width of the multiline indicator.
	// We need to ensure that the indentation of the input line is at least
	// as wide as the indicator, otherwise the indicator will overwrite the text
	// on subsequent lines.
	var indicatorWidth int
	if e.opts.GetBool("multiline-column-numbered") {
		indicatorWidth = len(strconv.Itoa(1)) + 1
	} else {
		indicatorWidth = 2
	}

	// Adjust indentation if the primary prompt is empty,
	// because we will print a column indicator on the first line.
	if e.prompt.LastUsed() == 0 && e.line.Lines() > 0 {
		var indicator string
		if e.opts.GetBool("multiline-column-numbered") {
			indicator = fmt.Sprintf(color.FgBlackBright+"%d"+color.Reset+" ", 1)
		} else {
			indicator = ui.DefaultMultilineColumn
		}

		e.startCols += indicatorWidth
		// Print the indicator on the first line.
		fmt.Print(indicator)
	} else if e.line.Lines() > 0 && e.startCols < indicatorWidth {
		// If the prompt is shorter than the indicator, pad with spaces
		// to ensure the input text starts aligned with subsequent lines
		// and isn't overwritten by the indicator.
		padding := indicatorWidth - e.startCols
		fmt.Print(fmt.Sprintf("%*s", padding, ""))

		e.startCols = indicatorWidth
	}

	// Recompute coordinates with the new indentation/cursor position.
	if e.line.Lines() > 0 {
		e.cursorCol, e.cursorRow = core.CoordinatesCursor(e.cursor, e.startCols)
		e.lineCol, e.lineRows = core.CoordinatesLine(e.line, e.startCols)
	}

	// Ensure that we have enough space to print the line.
	// We probe the terminal to verify that we are not at the bottom of the screen.
	// If we are, we scroll the screen to make space for the line.
	if e.lineRows > 1 {
		// 1. Probe the terminal height.
		// We move the cursor down to the last line of the input line,
		// and check if the cursor is at the expected position.
		term.MoveCursorDown(e.lineRows - 1)
		_, actualRow := e.keys.GetCursorPos()
		term.MoveCursorUp(e.lineRows - 1)

		// 2. Calculate the overshoot.
		expectedRow := e.startRows + e.lineRows - 1
		overshoot := expectedRow - actualRow

		// 3. Scroll the screen if needed.
		if overshoot > 0 {
			// Move to the bottom of the terminal.
			term.MoveCursorDown(actualRow - e.startRows)

			// Scroll the screen by printing newlines.
			for i := 0; i < overshoot; i++ {
				fmt.Print("\n")
			}

			// Update the start row to reflect the scrolling.
			e.startRows -= overshoot

			// Move the cursor back up to the new start position.
			term.MoveCursorUp(e.lineRows - 1)
			term.MoveCursorForwards(e.startCols)
		}
	}

	// 3. Input Area Rendering
	e.renderInputArea()

	// 4. Helpers Rendering
	// We clear everything below the input area to ensure that no artifacts
	// from previous renders (like longer lines or helpers) remain visible.
	term.MoveCursorDown(1)
	term.MoveCursorBackwards(term.GetWidth())
	fmt.Print(term.ClearScreenBelow)
	term.MoveCursorUp(1)
	term.MoveCursorForwards(e.lineCol)

	e.renderHelpers()

	// 5. Final Cursor Positioning
	// The cursor is currently at the end of the input line (lineRows, lineCol).
	// We need to move it to the actual cursor position (cursorRow, cursorCol).
	if e.lineRows > e.cursorRow {
		term.MoveCursorUp(e.lineRows - e.cursorRow)
	}

	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorForwards(e.cursorCol)

	fmt.Print(term.ShowCursor)
}

func (e *Engine) renderInputArea() {
	e.displayLineRefactored()
	e.renderMultilineIndicators()
	e.renderRightPrompt()
}

func (e *Engine) renderHelpers() {
	e.completer.Autocomplete()

	// 1. Check if we have anything to print.
	hintRows := ui.CoordinatesHint(e.hint)
	compMatches := e.completer.Matches()
	compSkip := e.completer.DisplaySkipped()

	// 2. Clear below the input line to remove artifacts,
	// unless we are at the bottom of the screen.
	termHeight := term.GetLength()
	if (e.startRows + e.lineRows) < termHeight {
		term.MoveCursorDown(1)
		term.MoveCursorBackwards(term.GetWidth())
		fmt.Print(term.ClearScreenBelow)
		term.MoveCursorUp(1)
		term.MoveCursorForwards(e.lineCol)
	}

	if hintRows == 0 && (compMatches == 0 || compSkip) {
		e.hintRows = 0
		e.compRows = 0

		return
	}

	fmt.Print(term.NewlineReturn)

	// 3. Display Hints
	ui.DisplayHint(e.hint)
	e.hintRows = ui.CoordinatesHint(e.hint)

	// 4. Display Completions
	if compMatches > 0 && !compSkip {
		completion.Display(e.completer, e.AvailableHelperLines())
		e.compRows = completion.Coordinates(e.completer)
	} else {
		e.completer.ResetUsedRows()
		e.compRows = 0
	}

	// 5. Restore Cursor to the "bottom of input area"
	// The cursor is currently at the bottom of the helpers.
	// We need to move it back up to the line just below the input text.
	term.MoveCursorUp(e.compRows)
	term.MoveCursorUp(e.hintRows)
	term.MoveCursorUp(1)

	// We are now on the same row as the end of the input line,
	// but at column 0. We need to move to e.lineCol.
	term.MoveCursorForwards(e.lineCol)
}

func (e *Engine) renderRightPrompt() {
	e.prompt.RightPrint(e.lineCol, true)

	// Restore cursor to the end of the input line.
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorForwards(e.lineCol)
}

func (e *Engine) displayLineRefactored() {
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
}

func (e *Engine) renderMultilineIndicators() {
	// Check if we have multiple lines to manage.
	if e.line.Lines() == 0 {
		return
	}

	// 1. Determine if we need to print columns.
	columns := e.opts.GetBool("multiline-column") ||
		e.opts.GetBool("multiline-column-numbered") ||
		e.opts.GetString("multiline-column-custom") != ""
	promptEmpty := e.prompt.LastUsed() == 0

	// If no columns are requested and the prompt is not empty, we have nothing to do.
	if !columns && !promptEmpty {
		return
	}

	// 2. Move to the top of the input area (first line).
	term.MoveCursorUp(e.lineRows)
	term.MoveCursorBackwards(term.GetWidth())

	// 3. Print the indicators for subsequent lines (1..N).
	printedLines := 0
	numbered := e.opts.GetBool("multiline-column-numbered")

	// Indicators
	pipe := ui.DefaultMultilineColumn

	for i := 1; i <= e.line.Lines(); i++ {
		fmt.Print("\n")

		if numbered {
			fmt.Print(fmt.Sprintf(color.FgBlackBright+"%d"+color.Reset+" ", i+1))
		} else if i == e.line.Lines() {
			e.prompt.SecondaryPrint()
		} else {
			fmt.Print(pipe)
		}

		printedLines++
	}

	// 4. Return cursor to the bottom of the input area.
	correction := e.lineRows - printedLines
	if correction > 0 {
		term.MoveCursorDown(correction)
	}

	// 5. Restore horizontal position to the end of the input text.
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorForwards(e.lineCol)
}
