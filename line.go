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
	}

	rl.comp = []rune{}     // No virtual completion yet
	rl.compLine = []rune{} // So no virtual line either

	rl.resetSelection()

	// Highlighting
	rl.resetRegions()
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

// lineCompleted - In many places we need the current line input. We either return the real line,
// or the one that includes the current completion candidate, if there is any.
func (rl *Instance) lineCompleted() []rune {
	if len(rl.comp) > 0 {
		return rl.compLine
	}
	return rl.line
}

// lineBuffer returns the aggregate buffer of the current Readline loop:
// this includes all previous lines (that were returned but not accepted
// as is by the caller), separated by a newline, and the current one.
func (rl *Instance) lineBuffer() (buf []rune) {
	for _, line := range rl.multilineSplit {
		buf = append(buf, []rune(line)...)
		buf = append(buf, '\n')
	}
	buf = append(buf, rl.line...)

	return
}

func (rl *Instance) lineDisplay() (line []rune, cpos int) {
	if len(rl.comp) > 0 {
		line = rl.compLine
		cpos = len(rl.compLine[:rl.pos])
	} else {
		line = rl.line
		cpos = len(rl.line[:rl.pos])
	}

	// Cursor cannot be after the line in vicmd mode.
	if rl.main == vicmd && cpos == len(line) && cpos > 0 {
		cpos--
	}

	if len(rl.histSuggested) > 0 {
		line = append(line, rl.histSuggested...)
	}

	// Add a newline for correct split by newline
	return append(line, '\n'), cpos
}

// computeLinePos determines the X and Y coordinates of the cursor.
func (rl *Instance) computeLinePos() {
	if rl.pos < 0 {
		rl.pos = 0
	}

	// Use the line including any completion or line suggestion,
	// and compute buffer/cursor length. This includes a last newline.
	line, cpos := rl.lineDisplay()

	// Get the index of each newline in the buffer.
	nl, _ := regexp.Compile("\n")
	newlinesIdx := nl.FindAllStringIndex(string(line), -1)

	rl.posX = 0
	rl.posY = 0
	rl.fullY = 0
	startLine := 0
	cursorSet := false

	for i, index := range newlinesIdx {
		// Compute any adjustment in case this line must be wrapped.
		// Here, compute if line must be wrapped, to adjust posY.
		lineY := rl.realLineLen(line[startLine:index[0]], i)

		// All lines add to the global offset.
		rl.fullY += lineY

		switch {
		case index[0] < cpos:
			// If we are not on the cursor line yet.
			rl.posY += lineY
		case !cursorSet:
			// We are on the cursor line, since we didn't catch the first case,
			// and that our cursor X coordinate has not been set yet.
			rl.computeCursorPos(startLine, cpos, i)
			cursorSet = true
			rl.hpos = i
		}

		startLine = index[1]
	}

	// When the X value of the cursor has not been set,
	// it's because we are at the end of the line.
	if rl.posX == 0 && rl.fullY == 0 {
		rl.computeCursorPos(startLine, cpos, 0)
	}
}

// computeCursorPos computes the X/Y coordinates of the cursor on a given line.
func (rl *Instance) computeCursorPos(startLine, cpos, lineIdx int) {
	cursorStart := cpos - startLine
	cursorY := cursorStart / GetTermWidth()
	cursorX := cursorStart % GetTermWidth()

	// Adjustments
	switch {
	case lineIdx == 0:
		// The first line has a prompt to account for
		cursorX += rl.Prompt.inputAt
	case cursorY == 0:
		// Even if empty, the line counts for 1.
		// If rounded, the cursor should be on the next line.
		cursorY++
	case cursorX > 0:
		// If we have a rest, that means we use one more line.
		cursorY++
	}

	rl.posY += cursorY
	rl.posX += cursorX
}

func (rl *Instance) realLineLen(line []rune, idx int) (lineY int) {
	lineLen := getRealLength(string(line))
	if idx == 0 {
		lineLen += rl.Prompt.inputAt
	}

	lineY = lineLen / GetTermWidth()
	restY := lineLen % GetTermWidth()

	if idx == 0 {
		lineY--
	}

	if restY > 0 {
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
		rl.Prompt.printLast(rl)

		// Print the input line with optional syntax highlighting
		line := rl.lineCompleted()
		highlighted := string(line)

		if rl.SyntaxHighlighter != nil {
			highlighted = rl.SyntaxHighlighter(line)
		}

		// Highlight visual selection if any, and print
		highlighted = rl.highlightLine([]rune(highlighted))
		print(highlighted)

		if len(rl.histSuggested) > 0 {
			print(seqDim + string(rl.histSuggested) + seqReset)
		}
	}

	// Update references with new coordinates only now, because
	// the new line may be longer/shorter than the previous one.
	rl.computeLinePos()

	// Go back to the current cursor position, with new coordinates
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.fullY)
	moveCursorDown(rl.posY)
	moveCursorForwards(rl.posX)
}

func (rl *Instance) lineClear() {
	if len(rl.line) == 0 {
		return
	}

	// We need to go back to prompt
	moveCursorUp(rl.posY)
	moveCursorBackwards(GetTermWidth())
	moveCursorForwards(rl.Prompt.inputAt)

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
	print("\r\n")

	// Ask the caller if the line should be accepted as is: if yes, return it.
	// Determine if the line is complete per the caller's standards.
	// We always return the entire buffer, including previous multisplits.
	if rl.IsMultiline(rl.lineBuffer()) {
		rl.writeHistoryLine()
		rl.accepted = true
		return
	}

	// If not, we should start editing another line.

	// We append the current line to rl.multiline, so that any
	// edition in EDITOR will use the entire buffer.
	rl.multilineSplit = append(rl.multilineSplit, string(rl.line))

	// Finally, we reset the current line, and keep readling.
	rl.lineInit()
}

func (rl *Instance) numLines() int {
	return len(strings.Split(string(rl.line), "\n"))
}
