package readline

import (
	"unicode"
)

// lineWidgets maps widget names to their implementation.
type lineWidgets map[string]widget

// standardViWidgets don't need access to the input key.
func (rl *Instance) viWidgets() lineWidgets {
	return map[string]widget{
		"vi-insert-mode":                rl.viInsertMode,
		"vi-cmd-mode":                   rl.viCommandMode,
		"visual-mode":                   rl.viVisualMode,
		"visual-line-mode":              rl.viVisualLineMode,
		"vi-insert-bol":                 rl.viInsertBol,
		"vi-backward-char":              rl.viBackwardChar,
		"vi-forward-char":               rl.viForwardChar,
		"vi-delete-char":                rl.viDeleteChar,
		"vi-backward-delete-char":       rl.viBackwardDeleteChar,
		"vi-forward-word":               rl.viForwardWord,
		"vi-forward-blank-word":         rl.viForwardBlankWord,
		"vi-forward-word-end":           rl.viForwardWordEnd,
		"vi-forward-blank-word-end":     rl.viForwardBlankWordEnd,
		"vi-backward-word":              rl.viBackwardWord,
		"vi-backward-blank-word":        rl.viBackwardBlankWord,
		"vi-backward-word-end":          rl.viBackwardWordEnd,
		"vi-backward-blank-word-end":    rl.viBackwardBlankWordEnd,
		"vi-kill-eol":                   rl.viKillEol,
		"vi-kill-line":                  rl.viKillLine,
		"vi-change-eol":                 rl.viChangeEol,
		"vi-add-eol":                    rl.viAddEol,
		"vi-add-next":                   rl.viAddNext,
		"vi-put-after":                  rl.viPutAfter,
		"vi-put-before":                 rl.viPutBefore,
		"vi-end-of-line":                rl.viEndOfLine,
		"vi-set-buffer":                 rl.viSetBuffer,
		"vi-yank":                       rl.viYank,
		"vi-yank-whole-line":            rl.viYankWholeLine,
		"vi-find-next-char":             rl.viFindNextChar,
		"vi-find-next-char-skip":        rl.viFindNextCharSkip,
		"vi-find-prev-char":             rl.viFindPrevChar,
		"vi-find-prev-char-skip":        rl.viFindPrevCharSkip,
		"vi-delete":                     rl.viDelete,
		"vi-replace-chars":              rl.viReplaceChars,
		"vi-replace":                    rl.viReplace,
		"vi-match-bracket":              rl.viMatchBracket,
		"select-a-blank-word":           rl.viSelectABlankWord,
		"select-a-shell-word":           rl.viSelectAShellWord,
		"select-a-word":                 rl.viSelectAWord,
		"select-in-blank-word":          rl.viSelectInBlankWord,
		"select-in-shell-word":          rl.viSelectInShellWord,
		"select-in-word":                rl.viSelectInWord,
		"vi-digit-or-beginning-of-line": rl.viDigitOrBeginningOfLine,
		"vi-goto-column":                rl.viGotoColumn,
		"vi-swap-case":                  rl.viSwapCase,
		"vi-oper-swap-case":             rl.viOperSwapCase,
		"vi-first-non-blank":            rl.viFirstNonBlank,
		"vi-substitute":                 rl.viSubstitute,
		"vi-change":                     rl.viChange,
		"vi-add-surround":               rl.viAddSurround,
		"vi-change-surround":            rl.viChangeSurround,
		"vi-select-surround":            rl.viSelectSurround,
		"vi-set-mark":                   rl.viSetMark,
		"vi-open-line-above":            rl.viOpenLineAbove,
		"vi-open-line-below":            rl.viOpenLineBelow,
	}
}

func (rl *Instance) viInsertMode() {
	rl.exitVisualMode()

	rl.local = ""
	rl.main = viins
	rl.viinsEnterPos = rl.pos

	rl.addIteration("")

	rl.updateCursor()
}

func (rl *Instance) viCommandMode() {
	rl.exitVisualMode()

	rl.addIteration("")
	rl.skipUndoAppend()

	// Only go back if not in insert mode
	if rl.main == viins && rl.pos > 0 && !rl.cursorAtBeginningOfLine() {
		rl.pos--
	}

	rl.local = ""
	rl.main = vicmd
	rl.viinsEnterPos = 0

	rl.updateCursor()
}

func (rl *Instance) viVisualMode() {
	lastMode := rl.local
	wasVisualLine := rl.visualLine
	rl.visualLine = false

	rl.addIteration("")
	rl.skipUndoAppend()

	rl.enterVisualMode()

	// We don't do anything else if the mode did not change.
	if lastMode == rl.local && wasVisualLine == rl.visualLine {
		return
	}

	rl.updateCursor()
}

func (rl *Instance) viVisualLineMode() {
	lastMode := rl.local
	wasVisualLine := rl.visualLine
	rl.visualLine = true

	rl.addIteration("")
	rl.skipUndoAppend()

	rl.enterVisualMode()

	// We don't do anything else if the mode did not change.
	if lastMode == rl.local && wasVisualLine == rl.visualLine {
		return
	}

	rl.updateCursor()
}

func (rl *Instance) viInsertBol() {
	rl.addIteration("")
	rl.beginningOfLine()
	rl.viInsertMode()
}

func (rl *Instance) viAddNext() {
	if len(rl.line) > 0 {
		rl.pos++
	}

	rl.viInsertMode()
}

func (rl *Instance) viAddEol() {
	rl.addIteration("")
	rl.endOfLine()
	if rl.pos < len(rl.line)-1 {
		rl.pos++
	}
	rl.viInsertMode()
}

func (rl *Instance) viBackwardWord() {
	rl.skipUndoAppend()
	vii := rl.getIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	}
}

func (rl *Instance) viBackwardBlankWord() {
	rl.skipUndoAppend()
	vii := rl.getIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))
	}
}

func (rl *Instance) viBackwardWordEnd() {
	rl.skipUndoAppend()

	vii := rl.getIterations()

	for i := 1; i <= vii; i++ {
		if len(rl.line) == 0 || rl.pos == 0 {
			return
		}

		rl.pos++
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))

		// Then move forward, adjusting if we are on a punctuation.
		if isPunctuation(rl.line[rl.pos]) {
			rl.pos--
		}

		rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
	}
}

func (rl *Instance) viBackwardBlankWordEnd() {
	rl.skipUndoAppend()

	vii := rl.getIterations()

	for i := 1; i <= vii; i++ {
		if len(rl.line) == 0 || rl.pos == 0 {
			return
		}

		rl.pos++
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))

		rl.moveCursorByAdjust(rl.viJumpE(tokeniseSplitSpaces))
	}
}

func (rl *Instance) viKillEol() {
	rl.skipUndoAppend()
	rl.undoHistoryAppendPos(rl.pos)

	pos := rl.pos
	rl.markSelection(rl.pos)

	for ; rl.pos < len(rl.line); rl.pos++ {
		if rl.line[rl.pos] == '\n' {
			break
		}
	}

	rl.deleteSelection()

	rl.pos = pos
	if rl.pos > 0 {
		rl.pos--
	}

	rl.addIteration("")
	rl.resetHelpers()
}

func (rl *Instance) viKillLine() {
	if rl.pos <= rl.viinsEnterPos || rl.pos == 0 {
		return
	}

	rl.undoHistoryAppend()
	rl.skipUndoAppend()

	rl.markSelectionRange(rl.viinsEnterPos, len(rl.line))
	rl.pos--
	rl.deleteSelection()
}

func (rl *Instance) viChangeEol() {
	rl.undoHistoryAppend()
	rl.skipUndoAppend()

	pos := rl.pos
	rl.markSelection(rl.pos)

	switch {
	case rl.numLines() > 1:
		for ; rl.pos < len(rl.line); rl.pos++ {
			if rl.line[rl.pos] == '\n' {
				break
			}
		}
	default:
		rl.pos = len(rl.line)
	}

	rl.deleteSelection()

	rl.pos = pos

	rl.addIteration("")
	rl.resetHelpers()

	rl.viInsertMode()
}

func (rl *Instance) viForwardWordEnd() {
	rl.skipUndoAppend()
	vii := rl.getIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
	}
}

func (rl *Instance) viForwardBlankWordEnd() {
	rl.skipUndoAppend()
	vii := rl.getIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpE(tokeniseSplitSpaces))
	}
}

func (rl *Instance) viForwardChar() {
	// Only exception where we actually don't forward a character.
	if rl.config.HistoryAutosuggest && len(rl.histSuggested) > 0 && rl.pos == len(rl.line)-1 {
		rl.historyAutosuggestInsert()
		return
	}

	rl.skipUndoAppend()

	// In vi-cmd-mode, we don't go further than the
	// last character in the line, hence rl.line-1
	if rl.main != viins && rl.pos < len(rl.line)-1 {
		if rl.line[rl.pos+1] != '\n' {
			rl.pos++
		}

		return
	}

	// And we can't go further than the line anyway.
	if rl.main == viins && rl.pos < len(rl.line) {
		rl.pos++

		return
	}
}

func (rl *Instance) viBackwardChar() {
	rl.skipUndoAppend()

	if rl.pos > 0 {
		if rl.line[rl.pos-1] != '\n' {
			rl.pos--
		}
	}
}

func (rl *Instance) viPutAfter() {
	rl.undoHistoryAppend()

	// posStart := rl.pos
	buffer := rl.pasteFromRegister()

	if buffer[len(buffer)-1] == '\n' {

		// Find the end of the current line.
		for cpos := rl.pos; cpos < len(rl.line); cpos++ {
			rl.pos = cpos
			if rl.line[cpos] == '\n' {
				break
			}
		}

		if rl.pos == len(rl.line)-1 {
			buffer = append([]rune{'\n'}, buffer[:len(buffer)-2]...)
		}
	}

	// Special handling on empty lines and end of buffer
	if !rl.isEmptyLine() && rl.pos < len(rl.line) {
		rl.pos++
	}

	vii := rl.getIterations()
	for i := 1; i <= vii; i++ {
		rl.lineInsert(buffer)
	}
}

func (rl *Instance) viPutBefore() {
	rl.undoHistoryAppend()

	buffer := rl.pasteFromRegister()

	if buffer[len(buffer)-1] == '\n' {
		// Find the end of the current line.
		for cpos := rl.pos; cpos >= 0; cpos-- {
			if rl.line[cpos] == '\n' {
				break
			}
			rl.pos = cpos
		}

		if rl.isEmptyLine() {
			buffer = append(buffer, '\n')
			rl.pos--
		}
	}

	vii := rl.getIterations()
	for i := 1; i <= vii; i++ {
		rl.lineInsert(buffer)
	}
}

func (rl *Instance) viReplaceChars() {
	// We read a character to use first.
	key, esc := rl.readOperator(true)
	if esc {
		rl.skipUndoAppend()
		return
	}

	// In visual mode, we replace all chars of the selection
	if rl.activeSelection() {
		bpos, epos, _ := rl.calcSelection()
		for i := bpos; i < epos; i++ {
			rl.line[i] = []rune(key)[0]
		}
		rl.pos = bpos

		rl.exitVisualMode()
		rl.updateCursor()

		return
	}

	// Or simply the character under the cursor.
	rl.deleteRune(false)
	rl.lineInsert([]rune(key))
	rl.pos--
}

func (rl *Instance) viReplace() {
	// We store the current line as an undo item first, but will not
	// store any intermediate changes (in the loop below) as undo items.
	rl.undoHistoryAppend()

	// All replaced characters are stored, to be used with backspace
	cache := make([]rune, 0)

	// Don't use the delete cache past the end of the line
	lineStart := len(rl.line)

	// The replace mode is quite special in that it does escape back
	// to the main readline loop: it keeps reading characters and inserts
	// them as long as the escape key is not pressed.
	for {
		// We read a character to use first.
		keys, esc := rl.readOperator(true)
		if esc {
			break
		}

		key := rune(keys[0])

		// If the key is a backspace, we go back one character
		if key == charBackspace || key == charBackspace2 {
			if rl.pos > lineStart {
				rl.backwardDeleteChar()
			} else if rl.pos > 0 {
				rl.pos--
			}

			// And recover the last replaced character
			if len(cache) > 0 && rl.pos < lineStart {
				key = cache[len(cache)-1]
				cache = cache[:len(cache)-1]
				rl.line[rl.pos] = key
			}
		} else {
			// If the cursor is at the end of the line,
			// we insert the character instead of replacing.
			if len(rl.line) == rl.pos {
				rl.line = append(rl.line, key)
			} else {
				cache = append(cache, rl.line[rl.pos])
				rl.line[rl.pos] = key
			}

			rl.pos++
		}

		// Update the line
		rl.redisplay()
	}

	// When exiting the replace mode, move the cursor back
	rl.pos--
}

func (rl *Instance) viForwardWord() {
	rl.skipUndoAppend()

	// If the input line is empty, we don't do anything
	if rl.pos == 0 && len(rl.line) == 0 {
		return
	}

	vii := rl.getIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpW(tokeniseLine))
	}
}

func (rl *Instance) viForwardBlankWord() {
	rl.skipUndoAppend()

	// If the input line is empty, we don't do anything
	if rl.pos == 0 && len(rl.line) == 0 {
		return
	}

	vii := rl.getIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpW(tokeniseSplitSpaces))
	}
}

func (rl *Instance) viDeleteChar() {
	rl.undoHistoryAppend()

	vii := rl.getIterations()

	// We might be on an active register, but not yanking...
	rl.saveToRegister(vii)

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {
		rl.deleteRune(false)
	}
}

func (rl *Instance) viBackwardDeleteChar() {
	if rl.main != viins {
		rl.undoHistoryAppend()
	}

	vii := rl.getIterations()

	// We might be on an active register, but not yanking...
	rl.saveToRegister(vii)

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {
		rl.deleteRune(true)
	}
}

func (rl *Instance) viYank() {
	rl.skipUndoAppend()
	self := "vi-yank"

	switch {
	case rl.activeSelection():
		// In visual mode, or with a non-empty selection, just yank.
		rl.yankSelection()
		rl.resetSelection()

		if rl.local == visual {
			rl.viCommandMode()
			rl.updateCursor()
		}

	case rl.local == viopp:
		// In vi operator pending mode, it's that we've been called
		// twice in a row (eg. `yy`), so copy the entire current line.
		rl.releasePending(self)

		rl.viVisualLineMode()
		defer rl.exitVisualMode()

		bpos, epos, _ := rl.calcSelection()
		if bpos == -1 || epos == -1 {
			return
		}
		buffer := rl.line[bpos:epos]
		rl.saveBufToRegister(buffer)

	default:
		// Else if we are actually starting a yank action.
		rl.enterVioppMode(self)
		rl.updateCursor()
		rl.markSelection(rl.pos)
	}
}

func (rl *Instance) viYankWholeLine() {
	rl.skipUndoAppend()

	// calculate line selection.
	rl.viVisualLineMode()
	bpos, epos, _ := rl.calcSelection()

	// If selection has a new line, remove it.
	if rl.line[epos-1] == '\n' {
		epos--
	}

	// Pass the buffer to register.
	buffer := rl.line[bpos:epos]
	rl.saveBufToRegister(buffer)

	// Done with any selection.
	rl.exitVisualMode()
}

func (rl *Instance) viEndOfLine() {
	rl.skipUndoAppend()

	for ; rl.pos < len(rl.line); rl.pos++ {
		if rl.line[rl.pos] == '\n' {
			break
		}
	}
}

func (rl *Instance) viMatchBracket() {
	rl.skipUndoAppend()

	nextPos := rl.pos
	found := false

	// If we are on a bracket/brace/parenthesis, we just find the matcher
	if !isBracket(rl.line[rl.pos]) {
		// First find the next bracket/brace/parenthesis
		for i := rl.pos + 1; i < len(rl.line); i++ {
			char := rl.line[i]
			if char == '}' || char == ')' || char == ']' {
				nextPos = i - rl.pos
				found = true
				break
			}
		}

		if !found {
			return
		}

		rl.moveCursorByAdjust(nextPos)
	}

	// Move to the match first, and then find the matching bracket.
	rl.moveCursorByAdjust(rl.viJumpBracket())
}

func (rl *Instance) viSetBuffer() {
	rl.skipUndoAppend()

	// We might be on a register already, so reset it,
	// and then wait again for a new register ID.
	if rl.registers.onRegister {
		rl.registers.resetRegister()
	}

	// Then read a key to select the register
	b, _, _ := rl.readInput()
	key := rune(b[0])
	if b[0] == charEscape {
		return
	}

	for _, char := range validRegisterKeys {
		if key == char {
			rl.registers.setActiveRegister(key)
			return
		}
	}
}

func (rl *Instance) viFindNextChar() {
	rl.skipUndoAppend()

	// Read the argument key to use as a pattern to search
	key, esc := rl.readOperator(false)
	if esc {
		return
	}

	forward := true
	skip := false
	times := rl.getIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func (rl *Instance) viFindNextCharSkip() {
	rl.skipUndoAppend()

	// Read the argument key to use as a pattern to search
	key, esc := rl.readOperator(true)
	if esc {
		return
	}

	forward := true
	skip := true
	times := rl.getIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func (rl *Instance) viFindPrevChar() {
	rl.skipUndoAppend()

	// Read the argument key to use as a pattern to search
	key, esc := rl.readOperator(true)
	if esc {
		return
	}

	forward := false
	skip := false
	times := rl.getIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func (rl *Instance) viFindPrevCharSkip() {
	rl.skipUndoAppend()

	// Read the argument key to use as a pattern to search
	key, esc := rl.readOperator(true)
	if esc {
		return
	}

	forward := false
	skip := true
	times := rl.getIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func (rl *Instance) viDelete() {
	self := "vi-delete"

	switch {
	case rl.activeSelection():
		// In visual mode, or with a non-empty selection, just cut it.
		bpos, _, _ := rl.calcSelection()
		rl.undoHistoryAppendPos(bpos)
		rl.skipUndoAppend()

		rl.deleteSelection()

		if rl.local == visual {
			rl.viCommandMode()
			rl.updateCursor()
		}

	case rl.local == viopp:
		// In vi operator pending mode, it's that we've been called
		// twice in a row (eg. `dd`), so delete the entire current line.
		rl.releasePending(self)

		rl.undoHistoryAppend()
		rl.skipUndoAppend()

		rl.viVisualLineMode()
		defer rl.exitVisualMode()

		rl.deleteSelection()

	default:
		// Else if we are actually starting a delete action.
		rl.skipUndoAppend()
		rl.enterVioppMode(self)
		rl.updateCursor()
		rl.markSelection(rl.pos)
	}
}

func (rl *Instance) viDigitOrBeginningOfLine() {
	rl.skipUndoAppend()

	// If the last command was a digit argument,
	// then our Vi iterations' length is not 0
	if len(rl.iterations) > 0 {
		rl.addIteration("0")
		return
	}

	// Else we go the beginning of line.
	rl.beginningOfLine()
}

func (rl *Instance) viSelectABlankWord() {
	rl.skipUndoAppend()

	// Go the beginning of the word and start mark
	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))
	bpos := rl.pos

	// Then go to the end of the blank word
	rl.moveCursorByAdjust(rl.viJumpW(tokeniseSplitSpaces) - 1)
	epos := rl.pos

	// Adjust if we are at the end of the line.
	if rl.pos == len(rl.line)-1 {
		rl.pos = bpos
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))
		rl.moveCursorByAdjust(rl.viJumpE(tokeniseSplitSpaces) + 1)
		bpos = rl.pos
		rl.pos = epos
	}

	rl.markSelectionRange(bpos, epos)
}

func (rl *Instance) viSelectAShellWord() {
	rl.skipUndoAppend()

	sBpos, sEpos, _, _ := rl.searchSurround('\'')
	dBpos, dEpos, _, _ := rl.searchSurround('"')

	mark, cpos := adjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos)

	// If none matched, use blankword
	if mark == -1 && cpos == -1 {
		rl.viSelectABlankWord()

		return
	}

	// Adjust for spaces preceding
	rl.pos = mark
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseSplitSpaces) + 1)
	bpos := rl.pos

	// Else set the region inside those quotes
	rl.pos = cpos
	rl.markSelectionRange(bpos, rl.pos)
}

func (rl *Instance) viSelectAWord() {
	rl.skipUndoAppend()

	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	bpos := rl.pos

	rl.moveCursorByAdjust(rl.viJumpW(tokeniseLine) - 1)
	epos := rl.pos

	// When nothing after, or non-empty, use spaces before word.
	if rl.pos == len(rl.line)-1 {
		rl.pos = bpos
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
		rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
		bpos = rl.pos
		rl.pos = epos
	}

	rl.markSelectionRange(bpos, epos)
}

func (rl *Instance) viSelectInBlankWord() {
	rl.skipUndoAppend()
	var bpos int

	// Go the beginning of the word and start mark
	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))
	if rl.local == visual || rl.local == viopp {
		bpos = rl.pos
	}

	// Then go to the end of the blank word
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseSplitSpaces))
	rl.markSelectionRange(bpos, rl.pos)
}

func (rl *Instance) viSelectInShellWord() {
	rl.skipUndoAppend()

	sBpos, sEpos, _, _ := rl.searchSurround('\'')
	dBpos, dEpos, _, _ := rl.searchSurround('"')

	mark, cpos := adjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos)

	// If none matched, use blankword
	if mark == -1 && cpos == -1 {
		rl.viSelectInBlankWord()

		return
	}

	// Else set the region inside those quotes
	rl.pos = cpos - 1
	rl.markSelectionRange(mark+1, rl.pos)
}

func (rl *Instance) viSelectInWord() {
	rl.skipUndoAppend()
	var bpos int

	// Go the beginning of the word and start mark
	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	if rl.local == visual || rl.local == viopp {
		bpos = rl.pos
	}

	// Then go to the end of the blank word
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
	rl.markSelectionRange(bpos, rl.pos)
}

func (rl *Instance) viGotoColumn() {
	rl.skipUndoAppend()
	iterations := rl.iterations
	column := rl.getIterations()

	if iterations == "" {
		column = 0
	} else if column < 0 {
		column = len(rl.line) + column
	}

	rl.pos = column
}

func (rl *Instance) viSwapCase() {
	switch rl.local {

	case visual:
		posInit := rl.pos

		bpos, epos, _ := rl.calcSelection()
		rl.resetSelection()
		rl.pos = bpos

		for range rl.line[bpos:epos] {
			char := rl.line[rl.pos]
			if unicode.IsLower(char) {
				char = unicode.ToUpper(char)
			} else {
				char = unicode.ToLower(char)
			}

			rl.line[rl.pos] = char

			rl.pos++
		}

		rl.pos = posInit
		rl.viCommandMode()
		rl.updateCursor()

	default:
		char := rl.line[rl.pos]
		if unicode.IsLower(char) {
			char = unicode.ToUpper(char)
		} else {
			char = unicode.ToLower(char)
		}

		rl.line[rl.pos] = char
	}
}

func (rl *Instance) viOperSwapCase() {
	switch {
	case rl.activeSelection():
		posInit := rl.pos

		bpos, epos, cpos := rl.calcSelection()
		rl.resetSelection()
		rl.pos = bpos

		for range rl.line[bpos:epos] {
			rl.viSwapCase()
			rl.pos++
		}
		rl.pos = cpos

		if rl.local == visual {
			rl.pos = posInit
			rl.viCommandMode()
			rl.updateCursor()
		}
	default:
		rl.skipUndoAppend()

		rl.enterVioppMode("vi-oper-swap-case")
		rl.updateCursor()
		rl.markSelection(rl.pos)
	}
}

func (rl *Instance) viFirstNonBlank() {
	rl.skipUndoAppend()
	for i := range rl.line {
		if rl.line[i] == ' ' {
			rl.pos = i
			break
		}
	}
}

func (rl *Instance) viAddSurround() {
	key, esc := rl.readOperator(true)
	if esc {
		rl.skipUndoAppend()
		return
	}

	bchar, echar := rl.matchSurround(rune(key[0]))

	rl.undoHistoryAppend()

	// Surround the selection
	rl.insertSelection(string(bchar), string(echar))

	// This only has an effect when we are in visual mode.
	rl.exitVisualMode()
}

func (rl *Instance) viSubstitute() {
	switch rl.local {
	case visual:
		rl.deleteSelection()
		rl.viInsertMode()
	default:
		vii := rl.getIterations()
		rl.saveToRegister(vii)

		for i := 1; i <= vii; i++ {
			rl.deleteRune(false)
		}

		rl.viInsertMode()
	}
}

func (rl *Instance) viChange() {
	// In visual mode, we have just have a selection to delete.
	if rl.local == visual {
		rl.undoHistoryAppend()
		rl.skipUndoAppend()

		rl.deleteSelection()
		rl.viInsertMode()

		return
	}

	rl.skipUndoAppend()

	// Otherwise, we have to read first key, which
	// is either a navigation or selection widget.
	key, esc := rl.readOperator(true)
	if esc {
		return
	}

	// Find the widget
	action, found := changeMovements[key]
	if !found {
		return
	}

	widget := rl.getWidget(action)
	if widget == nil {
		return
	}

	rl.undoHistoryAppend()

	// Update the pending keys, with an except for surround widgets.
	// Those widgets also need to adjust their selection offsets
	rl.keys = ""

	isSurround := action == "vi-select-surround"
	if isSurround {
		rl.keys = key
	}

	// Before running the widget, set the mark
	rl.markSelection(rl.pos)

	// Run the widget. We don't care about return values
	widget()

	if isSurround && rl.visualSelection() != nil {
		rl.visualSelection().epos++
	}

	rl.deleteSelection()

	if action != "vi-change-surround" {
		rl.viInsertMode()
	}
}

func (rl *Instance) viChangeSurround() {
	rl.skipUndoAppend()

	// Read a key as a rune to search for
	key, esc := rl.readOperator(true)
	if esc {
		return
	}

	char := rune(key[0])

	// Find the corresponding enclosing chars
	bpos, epos, _, _ := rl.searchSurround(char)
	if bpos == -1 && epos == -1 {
		return
	}

	// Add those two positions to highlighting and update.
	rl.markSelectionSurround(bpos, bpos+1)
	rl.markSelectionSurround(epos, epos+1)
	rl.redisplay()
	defer func() { rl.resetRegions() }()

	// Now read another key.
	key, esc = rl.readOperator(true)
	if esc {
		return
	}

	rl.undoHistoryAppend()

	rchar := rune(key[0])

	// There might be a matching equivalent.
	bchar, echar := rl.matchSurround(rchar)

	rl.line[bpos] = bchar
	rl.line[epos] = echar
}

func (rl *Instance) viSelectSurround() {
	rl.skipUndoAppend()

	var inside bool

	switch rl.keys[0] {
	case 'i':
		inside = true
		rl.keys = rl.keys[1:]
	case 'a':
		rl.keys = rl.keys[1:]
	}

	if len(rl.keys) == 0 {
		// Read a key as a rune to search for
		key, esc := rl.readOperator(true)
		if esc {
			return
		}
		rl.keys += key
	}

	char := rune(rl.keys[0])

	bpos, epos, _, _ := rl.searchSurround(char)
	if bpos == -1 && epos == -1 {
		return
	}

	if inside {
		bpos++
		epos--
	}

	rl.markSelectionRange(bpos, epos)
	rl.pos = epos
}

func (rl *Instance) viSetMark() {
	rl.skipUndoAppend()
	rl.markSelection(rl.pos)
}

func (rl *Instance) viOpenLineAbove() {
	rl.undoHistoryAppend()
	rl.beginningOfLine()
	rl.lineInsert([]rune{'\n'})
	rl.pos--
	rl.viInsertMode()
}

func (rl *Instance) viOpenLineBelow() {
	rl.undoHistoryAppend()
	rl.endOfLine()
	rl.pos++
	rl.lineInsert([]rune{'\n'})
	rl.pos--
	rl.viInsertMode()
}
