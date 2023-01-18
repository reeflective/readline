package readline

import (
	"regexp"
	"strings"
)

// lineInit is ran once at the beginning of an instance readline run.
func (rl *Instance) lineInit() {
	// We only reset the line when we don't need it
	// to retrieve an matching one/ an history pos index.
	if !rl.inferLine {
		rl.line = []rune{}
		rl.pos = 0
		rl.posY = 0
		rl.hpos = 0
	}

	rl.comp = []rune{}     // No virtual completion yet
	rl.compLine = []rune{} // So no virtual line either

	rl.resetSelection()

	// Highlighting
	rl.resetRegions()
}

// lineCompleted - In many places we need the current line input. We either return the real line,
// or the one that includes the current completion candidate, if there is any.
func (rl *Instance) lineCompleted() []rune {
	if len(rl.comp) > 0 {
		return rl.compLine
	}
	return rl.line
}

func (rl *Instance) lineSuggested() (line []rune, cpos int) {
	rl.checkCursorBounds()

	if len(rl.comp) > 0 {
		line = rl.compLine
		cpos = len(rl.compLine[:rl.pos])
	} else {
		line = rl.line
		cpos = len(rl.line[:rl.pos])
	}

	if len(rl.histSuggested) > 0 {
		line = append(line, rl.histSuggested...)
	}

	return line, cpos
}

func (rl *Instance) lineHighlighted() string {
	line := rl.lineCompleted()
	highlighted := string(line)

	// Use user-defined highlighting if any.
	if rl.SyntaxHighlighter != nil {
		highlighted = rl.SyntaxHighlighter(line)
	}

	// Add visual selection highlight if any, and print
	highlighted = rl.highlightLine([]rune(highlighted))

	if len(rl.histSuggested) > 0 {
		highlighted += seqFgBlack + string(rl.histSuggested) + seqReset
	}

	return highlighted
}

// computeLinePos determines the X and Y coordinates of the cursor.
func (rl *Instance) computeLinePos() {
	// Use the line including any completion or line suggestion,
	// and compute buffer/cursor length. Only add a newline when
	// the current buffer does not end with one.
	line, cpos := rl.lineSuggested()
	line = append(line, '\n')

	// Get the index of each newline in the buffer.
	nl := regexp.MustCompile("\n")
	newlinesIdx := nl.FindAllStringIndex(string(line), -1)

	rl.posY = 0
	rl.fullY = 0
	startLine := 0
	cursorSet := false

	for pos, newline := range newlinesIdx {
		// Compute any adjustment in case this line must be wrapped.
		// Here, compute if line must be wrapped, to adjust posY.
		lineY := rl.realLineLen(line[startLine:newline[0]], pos)

		// All lines add to the global offset.
		rl.fullY += lineY

		switch {
		case newline[0] < cpos:
			// If we are not on the cursor line yet.
			rl.posY += lineY
		case !cursorSet:
			// We are on the cursor line, since we didn't catch
			// the first case, and that our cursor X coordinate
			// has not been set yet.
			rl.computeCursorPos(startLine, cpos, pos)
			cursorSet = true
			rl.hpos = pos
		}

		startLine = newline[1]
	}
}

// computeCursorPos computes the X/Y coordinates of the cursor on a given line.
func (rl *Instance) computeCursorPos(startLine, cpos, lineIdx int) {
	termWidth := GetTermWidth()
	cursorStart := cpos - startLine
	cursorStart += rl.Prompt.inputAt(rl)

	cursorY := cursorStart / termWidth
	cursorX := cursorStart % termWidth

	rl.posY += cursorY
	rl.posX = cursorX

	if lineIdx > 0 {
		rl.posY++
	} else if rl.posX == termWidth {
		rl.posY++
		rl.posX = 0
	}
}

func (rl *Instance) realLineLen(line []rune, idx int) (lineY int) {
	lineLen := getRealLength(string(line))
	termWidth := GetTermWidth()
	lineLen += rl.Prompt.inputAt(rl)

	lineY = lineLen / termWidth
	restY := lineLen % termWidth

	if idx == 0 {
		lineY--
	}

	if restY > 0 {
		lineY++
	}

	// Empty lines are still considered a line.
	if lineY == 0 && idx != 0 {
		lineY++
	}

	return
}

// linePrint - refresh the current input line, either virtually completed or not.
func (rl *Instance) linePrint() {
	switch {
	// Password prompts.
	case rl.PasswordMask != 0:
	case rl.PasswordMask > 0:
		print(strings.Repeat(string(rl.PasswordMask), len(rl.line)) + " ")

	default:
		// Go back to prompt position, and clear everything below.
		// Note that we don't yet compute the line/cursor coordinates,
		// since we are still moving around the old line.
		moveCursorBackwards(GetTermWidth())
		moveCursorUp(rl.posY)
		print(seqClearScreenBelow)

		// We are the very beginning of the line ON WHICH we are
		// going to write the input line, not higher, even if the
		// entire primary+right prompt span several lines.
		// Print the prompt or part of it, and the buffer right after.
		rl.Prompt.printLast(rl)
		rl.printBuffer()

		// Now that we working with an entirely updated line,
		// we must recompute all the offsets for it.
		rl.computeLinePos()

		// Go back to the current cursor position, with new coordinates
		moveCursorBackwards(GetTermWidth())
		moveCursorUp(rl.fullY)
		moveCursorDown(rl.posY)
		moveCursorForwards(rl.posX)

		// Finally, print any right or tooltip prompt.
		rl.Prompt.printRprompt(rl)
	}
}

func (rl *Instance) lineClear() {
	if len(rl.line) == 0 {
		return
	}

	// We need to go back to prompt
	moveCursorUp(rl.posY)
	moveCursorBackwards(GetTermWidth())
	moveCursorForwards(rl.Prompt.inputAt(rl))

	// Clear everything after & below the cursor
	print(seqClearScreenBelow)

	// Real input line
	rl.line = []rune{}
	rl.compLine = []rune{}
	rl.pos = 0
	rl.posX = 0
	rl.fullX = 0
	rl.posY = 0
	rl.fullY = 0

	// Completions are also reset
	rl.clearVirtualComp()
}

// When the DelayedSyntaxWorker gives us a new line, we need to check if there
// is any processing to be made, that all lines match in terms of content.
func (rl *Instance) lineUpdate(line []rune) {
	if len(rl.comp) > 0 {
	} else {
		rl.line = line
	}

	rl.renderHelpers()
}

func (rl *Instance) lineInsert(r []rune) {
	for {
		// I don't really understand why `0` is creaping in at the end of the
		// array but it only happens with unicode characters.
		if len(r) > 1 && r[len(r)-1] == 0 {
			r = r[:len(r)-1]
			continue
		}
		break
	}

	// We can ONLY have three fondamentally different cases:
	switch {
	// The line is empty
	case len(rl.line) == 0:
		rl.line = r

	// We are inserting somewhere in the middle
	case rl.pos < len(rl.line):
		forwardLine := append(r, rl.line[rl.pos:]...)
		rl.line = append(rl.line[:rl.pos], forwardLine...)

	// We are at the end of the input line
	case rl.pos == len(rl.line):
		rl.line = append(rl.line, r...)
	}

	rl.pos += len(r)
}

// lineSlice returns a subset of the current input line.
func (rl *Instance) lineSlice(adjust int) (slice string) {
	switch {
	case rl.pos+adjust > len(rl.line):
		slice = string(rl.line[rl.pos:])
	case adjust < 0:
		if rl.pos+adjust < 0 {
			slice = string(rl.line[:rl.pos])
		} else {
			slice = string(rl.line[rl.pos+adjust : rl.pos])
		}
	default:
		slice = string(rl.line[rl.pos : rl.pos+adjust])
	}

	return
}

func (rl *Instance) lineCarriageReturn() {
	// Remove all helpers and line autosuggest,
	// but keep the line and go just below it.
	rl.histSuggested = []rune{}
	rl.clearHelpers()
	rl.linePrint()
	moveCursorDown(rl.fullY - rl.posY)
	print(seqClearScreenBelow)

	// Ask the caller if the line should be accepted as is: if yes, return it.
	// Determine if the line is complete per the caller's standards.
	// We always return the entire buffer, including previous multisplits.
	if rl.AcceptMultiline(rl.lineCompleted()) {
		moveCursorDown(rl.fullY - rl.posY)
		print("\r\n")
		rl.writeHistoryLine()
		rl.accepted = true
		return
	}

	// If not, we should start editing another line,
	// so get back to the cursor line (but don't move
	// the cursor VALUE)
	moveCursorUp(rl.fullY - rl.posY)

	// And insert a newline where our cursor value is.
	// This has the nice advantage of being able to work
	// in multiline mode even in the middle of the buffer.
	rl.lineInsert([]rune{'\n'})
}

func (rl *Instance) numLines() int {
	return len(strings.Split(string(rl.line), "\n"))
}

// Returns the real length of the line on which the cursor currently is.
func (rl *Instance) cursorLineLen() (lineLen int) {
	lines := strings.Split(string(rl.lineCompleted()), "\n")
	if len(lines) == 0 {
		return 0
	}

	lineLen += rl.Prompt.inputAt(rl)
	lineLen += getRealLength(lines[rl.hpos])

	termWidth := GetTermWidth()
	if lineLen > termWidth {
		lines := lineLen / termWidth
		restLen := lineLen % termWidth

		if lines > 0 && lineLen == 0 {
			return termWidth
		}

		return restLen
	}

	return lineLen
}

func (rl *Instance) printBuffer() {
	// Generate the entire line as an highlighted line,
	// and split it at each newline.
	line := rl.lineHighlighted()
	lines := strings.Split(line, "\n")

	if len(line) > 0 && line[len(line)-1] == '\n' {
		lines = append(lines, "")
	}

	for i, line := range lines {
		// Indent according to the prompt.
		if i > 0 {
			moveCursorForwards(rl.Prompt.inputAt(rl))
		}

		if i < len(lines)-1 {
			line += "\n"
		}

		print(line)
	}
}
