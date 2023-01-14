package readline

import (
	"regexp"
	"strings"
)

func (rl *Instance) standardWidgets() lineWidgets {
	widgets := map[string]widget{
		"clear-screen":            rl.clearScreen,
		"self-insert":             rl.selfInsert,
		"accept-line":             rl.acceptLine,
		"accept-and-hold":         rl.acceptAndHold,
		"beginning-of-line":       rl.beginningOfLine,
		"end-of-line":             rl.endOfLine,
		"up-line":                 rl.upLine,
		"down-line":               rl.downLine,
		"kill-line":               rl.killLine,
		"backward-kill-line":      rl.backwardKillLine,
		"kill-whole-line":         rl.killWholeLine,
		"kill-buffer":             rl.killBuffer,
		"backward-kill-word":      rl.backwardKillWord,
		"kill-word":               rl.killWord,
		"yank":                    rl.yank,
		"backward-delete-char":    rl.backwardDeleteChar,
		"delete-char-or-list":     rl.deleteCharOrList,
		"delete-char":             rl.deleteChar,
		"delete-word":             rl.deleteWord,
		"forward-char":            rl.forwardChar,
		"backward-char":           rl.backwardChar,
		"forward-word":            rl.forwardWord,
		"backward-word":           rl.backwardWord,
		"digit-argument":          rl.digitArgument,
		"undo":                    rl.undo,
		"overwrite-mode":          rl.overwriteMode,
		"set-mark-command":        rl.setMarkCommand,
		"exchange-point-and-mark": rl.exchangePointAndMark,
		"quote-region":            rl.quoteRegion,
		"quote-line":              rl.quoteLine,
		"neg-argument":            rl.negArgument,
		"capitalize-word":         rl.capitalizeWord,
		"down-case-word":          rl.downCaseWord,
		"up-case-word":            rl.upCaseWord,
		"transpose-words":         rl.transposeWords,
		"transpose-chars":         rl.transposeChars,
		"copy-region-as-kill":     rl.copyRegionAsKill,
		"copy-prev-word":          rl.copyPrevWord,
		"copy-prev-shell-word":    rl.copyPrevShellWord,
		"kill-region":             rl.killRegion,
		"redo":                    rl.redo,
		"switch-keyword":          rl.switchKeyword,
	}

	return widgets
}

// selfInsert inserts the given rune into the input line at the current cursor position.
func (rl *Instance) selfInsert() {
	rl.skipUndoAppend()

	// If we just inserted a completion candidate, we still have the
	// corresponding suffix matcher. Remove the last rune if needed,
	// and forget the matcher.
	rl.removeSuffixInserted()

	buf := []rune(rl.keys)

	for {
		// I don't really understand why `0` is creaping in at the end of the
		// array but it only happens with unicode characters.
		if len(buf) > 1 && buf[len(buf)-1] == 0 {
			buf = buf[:len(buf)-1]

			continue
		}
		break
	}

	// When the key is a control character, translate it to caret notation.
	if len(buf) == 1 && charCtrlA < byte(buf[0]) && byte(buf[0]) < charCtrlUnderscore {
		caret := byte(buf[0]) ^ caretMultiplier
		buf = append([]rune{'^'}, rune(caret))
	}

	// If the key is a bracket or quote, and that the next character
	// in the line is not its matcher, insert the pair altogether.
	// backOne := false
	// if len(buf) == 1 {
	// 	toInsert := buf[0]
	// 	isSurround := isBracket(toInsert) || toInsert == '\'' || toInsert == '"'
	// 	_, echar := rl.matchSurround(toInsert)
	// 	matcher := false
	// 	if len(rl.line) > rl.pos {
	// 		matcher = rl.matches(toInsert, rl.line[rl.pos])
	// 	}
	// 	if isSurround && !matcher {
	// 		buf = append(buf, echar)
	// 		backOne = true
	// 	}
	// }

	switch {
	// The line is empty
	case len(rl.line) == 0:
		rl.line = buf

	// We are inserting somewhere in the middle
	case rl.pos < len(rl.line):
		forwardLine := append(buf, rl.line[rl.pos:]...)
		rl.line = append(rl.line[:rl.pos], forwardLine...)

	// We are at the end of the input line
	case rl.pos == len(rl.line):
		rl.line = append(rl.line, buf...)
	}

	rl.pos += len(buf)
	// if backOne {
	// 	rl.pos--
	// }
}

func (rl *Instance) acceptLine() {
	rl.lineCarriageReturn()
}

func (rl *Instance) acceptAndHold() {
	rl.inferLine = true
	rl.histPos = -1
	rl.acceptLine()
}

func (rl *Instance) clearScreen() {
	rl.skipUndoAppend()

	print(seqClearScreen)
	print(seqCursorTopLeft)

	rl.Prompt.init(rl)
	rl.renderHelpers()
}

func (rl *Instance) beginningOfLine() {
	rl.skipUndoAppend()
	switch {
	case rl.numLines() > 1:
		rl.findAndMoveCursor("\n", 1, false, true)
	default:
		rl.pos = 0
	}
}

func (rl *Instance) endOfLine() {
	rl.skipUndoAppend()

	switch {
	case rl.numLines() > 1:
		rl.findAndMoveCursor("\n", 1, true, false)
	default:
		rl.pos = len(rl.line)
	}
}

func (rl *Instance) upLine() {
	var cpos int // The vertical position of the cursor on the current line.
	var bpos int // The beginning of the current line.

	// Get the index of each newline in the buffer.
	line := append(rl.lineCompleted(), '\n')
	nl := regexp.MustCompile("\n")
	newlinesIdx := nl.FindAllStringIndex(string(line), -1)

	for line := len(newlinesIdx) - 1; line >= 0; line-- {
		epos := newlinesIdx[line][0]
		if line > rl.hpos {
			continue
		}

		// Get the beginning of the previous line.
		if line > 0 {
			bpos = newlinesIdx[line-1][0]
		} else {
			bpos = 0
		}

		// If we are on the current line,
		// go at the beginning of the previous one.
		if line == rl.hpos {
			cpos = rl.pos - bpos
			if cpos == 1 && line == 1 {
				cpos--
			}
			continue
		}

		// And either go at the end of the line
		// or to the previous cursor X coordinate.
		if epos-bpos > cpos {
			rl.pos = bpos + cpos
		} else {
			rl.pos = bpos - 1
		}

		break
	}
}

func (rl *Instance) downLine() {
	var cpos int // The vertical position of the cursor on the current line.
	var bpos int // The beginning of the next line.

	// Get the index of each newline in the buffer.
	line := append(rl.lineCompleted(), '\n')
	nl := regexp.MustCompile("\n")
	newlinesIdx := nl.FindAllStringIndex(string(line), -1)

	for line := 0; line < len(newlinesIdx); line++ {
		epos := newlinesIdx[line][0]
		if line < rl.hpos {
			bpos = epos
			continue
		}

		// If we are on the current line,
		// go at the end of it
		if line == rl.hpos {
			cpos = rl.pos - bpos
			bpos = epos
			rl.pos = bpos + 1
			continue
		}

		// Adjust for newline rounded.
		if cpos == 0 {
			cpos++
		}

		// And either go at the end of the line
		// or to the previous cursor X coordinate.
		if epos-bpos > cpos {
			rl.pos = bpos + cpos
		} else {
			rl.pos = epos - 1
		}

		break
	}
}

func (rl *Instance) killLine() {
	rl.undoHistoryAppend()

	rl.saveBufToRegister(rl.line[rl.pos:])
	rl.line = rl.line[:rl.pos]
	rl.resetHelpers()
	rl.addIteration("")
}

func (rl *Instance) backwardKillLine() {
	rl.undoHistoryAppend()

	rl.saveBufToRegister(rl.line[:rl.pos])
	rl.line = rl.line[rl.pos:]
	rl.resetHelpers()
	rl.addIteration("")
}

func (rl *Instance) killWholeLine() {
	rl.undoHistoryAppend()

	if len(rl.line) == 0 {
		return
	}

	rl.saveBufToRegister(rl.line)
	rl.lineClear()
}

func (rl *Instance) killBuffer() {
	rl.undoHistoryAppend()

	if len(rl.line) == 0 {
		return
	}
	rl.saveBufToRegister(rl.line)
	rl.lineClear()
}

func (rl *Instance) backwardKillWord() {
	rl.undoHistoryAppend()
	rl.skipUndoAppend()

	rl.saveToRegister(rl.viJumpB(tokeniseLine))
	rl.viDeleteByAdjust(rl.viJumpB(tokeniseLine))
}

func (rl *Instance) killWord() {
	rl.undoHistoryAppend()

	rl.saveToRegisterTokenize(tokeniseLine, rl.viJumpE, 1)
	rl.viDeleteByAdjust(rl.viJumpE(tokeniseLine) + 1)
}

func (rl *Instance) yank() {
	buffer := rl.pasteFromRegister()
	rl.lineInsert(buffer)
}

func (rl *Instance) backwardDeleteChar() {
	rl.undoHistoryAppend()

	vii := rl.getIterations()

	// We might be on an active register, but not yanking...
	rl.saveToRegister(vii)

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {

		var toDelete rune
		var isSurround, matcher bool
		if rl.pos > 0 && len(rl.line) > rl.pos {
			toDelete = rl.line[rl.pos-1]
			isSurround = isBracket(toDelete) || toDelete == '\'' || toDelete == '"'
			matcher = rl.matches(toDelete, rl.line[rl.pos])
		}

		// Delete the character
		rl.deleteRune(true)

		// When the next character was identified
		// as a surround, delete as well.
		if isSurround && matcher {
			rl.pos++
			rl.deleteRune(true)
		}
	}

	if rl.main == viins || rl.main == emacs {
		rl.skipUndoAppend()
	}
}

func (rl *Instance) deleteChar() {
	rl.undoHistoryAppend()

	vii := rl.getIterations()

	// We might be on an active register, but not yanking...
	rl.saveToRegister(vii)

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {
		rl.deleteRune(false)
	}
}

func (rl *Instance) deleteWord() {
	rl.undoHistoryAppend()

	rl.markSelection(rl.pos)
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
	rl.deleteSelection()
}

func (rl *Instance) deleteCharOrList() {
	switch {
	case rl.pos < len(rl.line):
		rl.deleteRune(false)
	default:
		rl.expandOrComplete()
	}
}

func (rl *Instance) forwardChar() {
	rl.skipUndoAppend()
	if rl.pos < len(rl.line) {
		rl.pos++
	}
}

func (rl *Instance) backwardChar() {
	rl.skipUndoAppend()
	if rl.pos > 0 {
		rl.pos--
	}
}

func (rl *Instance) forwardWord() {
	rl.skipUndoAppend()

	// If the input line is empty, we don't do anything
	if rl.pos == 0 && len(rl.line) == 0 {
		return
	}

	// Get iterations and move
	vii := rl.getIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpW(tokeniseLine))
	}
}

func (rl *Instance) backwardWord() {
	rl.skipUndoAppend()

	vii := rl.getIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	}
}

// digitArgument is used both in Emacs and Vim modes,
// but strips the Alt modifier used in Emacs mode.
func (rl *Instance) digitArgument() {
	rl.skipUndoAppend()

	// If we were called in the middle of a pending
	// operation, we should not yet trigger the caller.
	// This boolean is recomputed at the next key read:
	// This just postpones running the caller a little.
	rl.isViopp = false

	if len(rl.keys) > 1 {
		// The first rune is the alt modifier.
		rl.addIteration(rl.keys[1:])
	} else {
		rl.addIteration(string(rl.keys[0]))
	}
}

func (rl *Instance) overwriteMode() {
	// We store the current line as an undo item first, but will not
	// store any intermediate changes (in the loop below) as undo items.
	rl.undoHistoryAppend()

	// The replace mode is quite special in that it does escape back
	// to the main readline loop: it keeps reading characters and inserts
	// them as long as the escape key is not pressed.
	for {
		// Read a new key
		keys, esc := rl.readOperator(true)
		if esc {
			return
		}
		key := rune(keys[0])

		// If the key is a backspace, we go back one character
		if key == charBackspace || key == charBackspace2 {
			rl.backwardDeleteChar()
		} else {
			// If the cursor is at the end of the line,
			// we insert the character instead of replacing.
			if len(rl.line)-1 < rl.pos {
				rl.line = append(rl.line, key)
			} else {
				rl.line[rl.pos] = key
			}

			rl.pos++
		}

		rl.redisplay()
	}
}

func (rl *Instance) setMarkCommand() {
	rl.skipUndoAppend()

	vii := rl.getIterations()
	switch {
	case vii < 0:
		rl.resetSelection()
		rl.visualLine = false
	default:
		rl.markSelection(rl.pos)
	}
}

func (rl *Instance) quoteRegion() {
	rl.undoHistoryAppend()

	_, cpos := rl.insertSelection("'", "'")
	rl.pos = cpos + 1
}

func (rl *Instance) quoteLine() {
	newLine := make([]rune, 0)
	newLine = append(newLine, '\'')

	for _, char := range rl.line {
		if char == '\n' {
			break
		}
		if char == '\'' {
			newLine = append(newLine, []rune("\\'")...)
		} else {
			newLine = append(newLine, char)
		}
	}

	newLine = append(newLine, '\'')

	rl.line = newLine
}

func (rl *Instance) negArgument() {
	rl.negativeArg = true
}

func (rl *Instance) capitalizeWord() {
	rl.undoHistoryAppend()

	posInit := rl.pos
	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	letter := rl.line[rl.pos]
	upper := strings.ToUpper(string(letter))
	rl.line[rl.pos] = rune(upper[0])
	rl.pos = posInit
}

func (rl *Instance) downCaseWord() {
	rl.undoHistoryAppend()

	posInit := rl.pos
	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))

	rl.markSelection(rl.pos)
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))

	word, bpos, epos, _ := rl.popSelection()
	word = strings.ToLower(word)
	rl.insertBlock(bpos, epos, word, "")

	rl.pos = posInit
}

func (rl *Instance) upCaseWord() {
	rl.undoHistoryAppend()

	posInit := rl.pos
	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))

	rl.markSelection(rl.pos)
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))

	word, bpos, epos, _ := rl.popSelection()
	word = strings.ToUpper(word)
	rl.insertBlock(bpos, epos, word, "")

	rl.pos = posInit
}

func (rl *Instance) transposeWords() {
	rl.undoHistoryAppend()

	posInit := rl.pos

	// Save the current word
	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))

	rl.markSelection(rl.pos)
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))

	toTranspose, tbpos, tepos, _ := rl.popSelection()

	// First move the number of words
	vii := rl.getIterations()
	for i := 0; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	}

	// Save the word to transpose with
	rl.markSelection(rl.pos)
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))

	transposeWith, wbpos, wepos, _ := rl.popSelection()

	// Assemble the newline
	begin := string(rl.line[:wbpos])
	newLine := append([]rune(begin), []rune(toTranspose)...)
	newLine = append(newLine, rl.line[wepos:tbpos]...)
	newLine = append(newLine, []rune(transposeWith)...)
	newLine = append(newLine, rl.line[tepos:]...)
	rl.line = newLine

	// And replace cursor
	if vii < 0 {
		rl.pos = posInit
	} else {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
		for i := 0; i <= vii; i++ {
			rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
		}
	}
}

func (rl *Instance) copyRegionAsKill() {
	rl.skipUndoAppend()
	rl.yankSelection()
	rl.resetSelection()
}

func (rl *Instance) copyPrevWord() {
	rl.undoHistoryAppend()

	posInit := rl.pos

	rl.markSelection(rl.pos)
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))

	wlen, _ := rl.insertSelection("", "")
	rl.pos = posInit + wlen
}

func (rl *Instance) copyPrevShellWord() {
	rl.undoHistoryAppend()

	posInit := rl.pos

	// First go back a single blank word
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))

	// Now try to find enclosing quotes from here.
	sBpos, sEpos, _, _ := rl.searchSurround('\'')
	dBpos, dEpos, _, _ := rl.searchSurround('"')

	mark, cpos := adjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos)
	if mark == -1 && cpos == -1 {
		rl.markSelection(rl.pos)
		rl.moveCursorByAdjust(rl.viJumpE(tokeniseSplitSpaces))
	} else {
		rl.markSelection(mark)
		rl.pos = cpos
	}

	word, _, _, _ := rl.popSelection()

	// Replace the cursor before reassembling the line.
	rl.pos = posInit

	rl.insertBlock(rl.pos, rl.pos, word, "")
	rl.pos += len(word)
}

func (rl *Instance) killRegion() {
	rl.undoHistoryAppend()

	rl.deleteSelection()
}

// Cursor position cases:
//
// 1. Cursor on symbol:
// 2+2   => +
// 2-2   => -
// 2 + 2 => +
// 2 +2  => +2
// 2 -2  => -2
// 2 -a  => -a
//
// 2. Cursor on number or alpha:
// 2+2   => +2
// 2-2   => -2
// 2 + 2 => 2
// 2 +2  => +2
// 2 -2  => -2
// 2 -a  => -a.
func (rl *Instance) switchKeyword() {
	rl.undoHistoryAppend()

	cpos := rl.pos
	increase := rl.keys[0] == charCtrlA

	if match, _ := regexp.MatchString(`[+-][0-9]`, rl.lineSlice(2)); match {
		// If cursor is on the `+` or `-`, we need to check if it is a
		// number with a sign or an operator, only the number needs to
		// forward the cursor.
		digit := regexp.MustCompile(`[^0-9]`)
		if cpos == 0 || digit.MatchString(string(rl.line[cpos-1])) {
			cpos++
		}
	} else if match, _ := regexp.MatchString(`[+-][a-zA-Z]`, rl.lineSlice(2)); match {
		// If cursor is on the `+` or `-`, we need to check if it is a
		// short option, only the short option needs to forward the cursor.
		if cpos == 0 || rl.line[rl.pos-1] == ' ' {
			cpos++
		}
	}

	// Select in word and get the selection positions
	bpos, epos := rl.selectInWord(cpos)
	epos++

	// Move the cursor backward if needed/possible
	if bpos != 0 && (rl.line[bpos-1] == '+' || rl.line[bpos-1] == '-') {
		bpos--
	}

	// Get the selection string
	selection := string(rl.line[bpos:epos])

	// For each of the keyword handlers, run it, which returns
	// false/none if didn't operate, then continue to next handler.
	for _, switcher := range rl.keywordSwitchers() {

		changed, word, obpos, oepos := switcher(selection, increase)
		if !changed {
			continue
		}

		// We are only interested in the end position after all runs
		epos = bpos + oepos
		bpos += obpos
		if cpos < bpos || cpos >= epos {
			continue
		}

		// Update the line and the cursor, and return
		// since we have a handler that has been ran.
		begin := string(rl.line[:bpos])
		end := string(rl.line[epos:])
		newLine := append([]rune(begin), []rune(word)...)
		newLine = append(newLine, []rune(end)...)
		rl.line = newLine
		rl.pos = bpos + len(word) - 1

		return
	}
}

func (rl *Instance) exchangePointAndMark() {
	rl.skipUndoAppend()
	vii := rl.getIterations()

	visual := rl.visualSelection()
	if visual == nil {
		return
	}

	switch {
	case vii < 0:
		rl.pos, visual.bpos = visual.bpos, rl.pos
	case vii > 0:
		rl.pos, visual.bpos = visual.bpos, rl.pos
		visual.active = true
	case vii == 0:
		visual.active = true
	}
}

func (rl *Instance) transposeChars() {
	if rl.pos < 2 || len(rl.line) < 2 {
		rl.skipUndoAppend()

		return
	}

	switch {
	case rl.pos == len(rl.line):
		last := rl.line[rl.pos-1]
		blast := rl.line[rl.pos-2]
		rl.line[rl.pos-2] = last
		rl.line[rl.pos-1] = blast
	default:
		last := rl.line[rl.pos]
		blast := rl.line[rl.pos-1]
		rl.line[rl.pos-1] = last
		rl.line[rl.pos] = blast
	}
}
