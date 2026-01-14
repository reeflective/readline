package display

import (
	"fmt"
	"strconv"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/strutil"
	"github.com/reeflective/readline/internal/term"
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
				indicator = fmt.Sprintf("\x1b[1;30m%d\x1b[0m ", 1)
			} else {
				indicator = "\x1b[1;30m\U00002502 \x1b[0m"
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
	
		// 3. Input Area Rendering
	
	e.renderInputArea()
	// 4. Helpers Rendering
	// 5. Final Cursor Positioning
	fmt.Print(term.ShowCursor)
}

func (e *Engine) renderInputArea() {
	e.displayLineRefactored()
	e.renderMultilineIndicators()
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
	pipe := "\x1b[1;30m\U00002502 \x1b[0m"
	angle := "\x1b[1;30m\U00002514 \x1b[0m"

	for i := 1; i <= e.line.Lines(); i++ {
		var indicator string
		if numbered {
			indicator = fmt.Sprintf("\x1b[1;30m%d\x1b[0m ", i+1)
		} else if i == e.line.Lines() {
			indicator = angle
		} else {
			indicator = pipe
		}

		fmt.Print("\n" + indicator)

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
