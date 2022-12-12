package readline

import (
	"fmt"
	"unicode"
)

// baseWidgets maps widget names to their implementation.
type baseWidgets map[string]func()

// standardViWidgets don't need access to the input key.
func (rl *Instance) viWidgets() baseWidgets {
	return map[string]func(){
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
		"vi-backward-blank-word":        rl.viBackwardBlankWord, // TODO vi-backward-blank-word-end/vi-backward-word-end (ge / gE)
		"vi-kill-eol":                   rl.viKillEol,
		"vi-change-eol":                 rl.viChangeEol,
		"vi-edit-command-line":          rl.viEditCommandLine,
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
	}
}

func (rl *Instance) viInsertMode() {
	rl.exitVisualMode()

	rl.local = ""
	rl.main = viins

	rl.addIteration("")
	rl.activeRegion = false

	rl.updateCursor()
}

func (rl *Instance) viCommandMode() {
	rl.exitVisualMode()

	rl.addIteration("")
	rl.viUndoSkipAppend = true
	rl.activeRegion = false

	// Only go back if not in insert mode
	if rl.main == viins && len(rl.line) > 0 && rl.pos > 0 {
		rl.pos--
	}

	rl.local = ""
	rl.main = vicmd

	rl.updateCursor()
}

func (rl *Instance) viVisualMode() {
	lastMode := rl.local
	wasVisualLine := rl.visualLine

	rl.addIteration("")
	rl.viUndoSkipAppend = true

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

	rl.addIteration("")
	rl.viUndoSkipAppend = true

	rl.enterVisualLineMode()

	// We don't do anything else if the mode did not change.
	if lastMode == rl.local && wasVisualLine == rl.visualLine {
		return
	}

	rl.updateCursor()
}

func (rl *Instance) viInsertBol() {
	rl.main = viins

	rl.addIteration("")
	rl.viUndoSkipAppend = true

	rl.pos = 0

	rl.updateCursor()

	rl.refreshVimStatus()
}

func (rl *Instance) viAddNext() {
	if len(rl.line) > 0 {
		rl.pos++
	}

	rl.viInsertMode()
}

func (rl *Instance) viAddEol() {
	if len(rl.line) > 0 {
		rl.pos = len(rl.line)
	}

	rl.viInsertMode()
}

func (rl *Instance) viBackwardWord() {
	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	}
}

func (rl *Instance) viBackwardBlankWord() {
	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))
	}
}

func (rl *Instance) viKillEol() {
	pos := rl.pos
	if pos < 0 {
		pos--
	}
	rl.saveBufToRegister(rl.line[pos:])
	rl.line = rl.line[:rl.pos]
	// Only go back if there is an input
	if len(rl.line) > 0 {
		rl.pos--
	}
	rl.addIteration("")
	rl.resetHelpers()
	rl.updateHelpers()
}

func (rl *Instance) viChangeEol() {
	rl.saveBufToRegister(rl.line[rl.pos-1:])
	rl.line = rl.line[:rl.pos]
	rl.addIteration("")

	rl.resetHelpers()

	rl.viInsertMode()
}

func (rl *Instance) viForwardWordEnd() {
	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
	}
}

func (rl *Instance) viForwardBlankWordEnd() {
	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpE(tokeniseSplitSpaces))
	}
}

func (rl *Instance) viForwardChar() {
	rl.viUndoSkipAppend = true

	// In vi-cmd-mode, we don't go further than the
	// last character in the line, hence rl.line-1
	if rl.main != viins && rl.pos < len(rl.line)-1 {
		rl.pos++

		return
	}

	// And we can't go further than the line anyway.
	if rl.main == viins && rl.pos < len(rl.line) {
		rl.pos++

		return
	}
}

func (rl *Instance) viBackwardChar() {
	rl.viUndoSkipAppend = true

	if rl.pos > 0 {
		rl.pos--
	}
}

func (rl *Instance) viPutAfter() {
	// paste after the cursor position
	rl.viUndoSkipAppend = true
	rl.pos++

	buffer := rl.pasteFromRegister()
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.insert(buffer)
	}
	rl.pos--
}

func (rl *Instance) viPutBefore() {
	// paste before
	rl.viUndoSkipAppend = true
	buffer := rl.pasteFromRegister()
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.insert(buffer)
	}
}

func (rl *Instance) viReplaceChars() {
	rl.viUndoSkipAppend = true

	// We read a character to use first.
	rl.enterVioppMode("")
	rl.updateCursor()

	key, esc := rl.readArgumentKey()
	if esc {
		rl.exitVioppMode()
		rl.updateCursor()
		return
	}
	rl.exitVioppMode()
	rl.updateCursor()

	// In visual mode, we replace all chars of the selection
	if rl.activeRegion || rl.local == visual {
		bpos, epos, _ := rl.getSelection()
		for i := bpos; i < epos; i++ {
			rl.line[i] = []rune(key)[0]
		}
		rl.pos = bpos

		rl.viCommandMode()

		return
	}

	// Or simply the character under the cursor.
	rl.deletex()
	rl.insert([]rune(key))
	rl.pos--
}

func (rl *Instance) viReplace() {
	// We store the current line as an undo item first, but will not
	// store any intermediate changes (in the loop below) as undo items.
	rl.undoAppendHistory()
	rl.viUndoSkipAppend = true

	// All replaced characters are stored, to be used with backspace
	cache := make([]rune, 0)

	// The replace mode is quite special in that it does escape back
	// to the main readline loop: it keeps reading characters and inserts
	// them as long as the escape key is not pressed.
	for {
		rl.enterVioppMode("")
		rl.updateCursor()

		// Read a new key
		keys, esc := rl.readArgumentKey()
		if esc {
			rl.exitVioppMode()
			rl.updateCursor()
			break
		}
		key := rune(keys[0])

		// If the key is a backspace, we go back one character
		if key == charBackspace || key == charBackspace2 {
			if rl.pos > 0 {
				rl.pos--
			}

			// And recover the last replaced character
			if len(cache) > 0 {
				key = cache[len(cache)-1]
				cache = cache[:len(cache)-1]
				rl.line[rl.pos] = key
			}
		} else {
			// If the cursor is at the end of the line,
			// we insert the character instead of replacing.
			if len(rl.line)-1 < rl.pos {
				cache = append(cache, rune(0))
				rl.line = append(rl.line, key)
			} else {
				cache = append(cache, rl.line[rl.pos])
				rl.line[rl.pos] = key
			}

			rl.pos++
		}

		// Update the line
		rl.updateHelpers()
	}

	// When exiting the replace mode, move the cursor back
	rl.pos--

	rl.exitVioppMode()
	rl.updateCursor()
}

func (rl *Instance) viEditCommandLine() {
	rl.clearHelpers()
	var multiline []rune
	if rl.GetMultiLine == nil {
		multiline = rl.line
	} else {
		multiline = rl.GetMultiLine(rl.line)
	}

	// Keep the previous cursor position
	prev := rl.pos

	new, err := rl.StartEditorWithBuffer(multiline, "")
	if err != nil || len(new) == 0 {
		fmt.Println(err)
		rl.viUndoSkipAppend = true
		return
	}

	// Clean the shell and put the new buffer, with adjusted pos if needed.
	rl.clearLine()
	rl.line = new
	if prev > len(rl.line) {
		rl.pos = len(rl.line) - 1
	} else {
		rl.pos = prev
	}
}

func (rl *Instance) viForwardWord() {
	rl.viUndoSkipAppend = true

	// If the input line is empty, we don't do anything
	if rl.pos == 0 && len(rl.line) == 0 {
		return
	}

	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpW(tokeniseLine))
	}

	// We make an adjustment to the mark if we are currently
	// yanking, and this widget is the argument action.
	if rl.local == viopp && rl.activeRegion {
		rl.pos--
	}
}

func (rl *Instance) viForwardBlankWord() {
	// If the input line is empty, we don't do anything
	if rl.pos == 0 && len(rl.line) == 0 {
		return
	}

	rl.viUndoSkipAppend = true

	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpW(tokeniseSplitSpaces))
	}
}

func (rl *Instance) viDeleteChar() {
	vii := rl.getViIterations()

	// We might be on an active register, but not yanking...
	rl.saveToRegister(vii)

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {
		rl.deletex()
	}

	// TODO: This should probably be used after any keymap
	// has been run, when we detect in command mode that our
	// cursor position if off-line.
	// On the other hand, this is the difference between
	// classic backwardDeleteChar and this function(rl *Instance).
	//
	// if rl.pos == len(rl.line) && len(rl.line) > 0 {
	// 	rl.pos--
	// }
}

// TODO: Same here
func (rl *Instance) viBackwardDeleteChar() {
	vii := rl.getViIterations()

	// We might be on an active register, but not yanking...
	rl.saveToRegister(vii)

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {
		rl.deleteX()
	}

	// TODO: This should probably be used after any keymap
	// has been run, when we detect in command mode that our
	// cursor position if off-line.
	// On the other hand, this is the difference between
	// classic backwardDeleteChar and this function(rl *Instance).
	//
	if rl.pos == len(rl.line) && len(rl.line) > 0 {
		rl.pos--
	}
}

func (rl *Instance) viYank() {
	// When we are called after a pending operator action, we are a pending
	// usually not in visual mode, but have an active selection.
	// In this case we yank the active region and return.
	if rl.activeRegion || rl.local == visual {
		rl.yankSelection()
		rl.resetSelection()

		if rl.local == visual {
			rl.viCommandMode()
			rl.updateCursor()
		}

		return
	}

	// If we are in operator pending mode, that means the command
	// is 'yy' (optionally with iterations), so we copy the required
	if rl.local == viopp {
		rl.saveBufToRegister(rl.line)
		rl.viUndoSkipAppend = true
		return
	}

	// Else if we are actually starting a yank action. We need an argument:
	// Enter operator pending mode for the next key to be considered this
	// argument (more precisely, the widget to be executed before this argument).
	rl.enterVioppMode("vi-yank")
	rl.updateCursor()

	// We set the initial mark, so that when executing this
	// widget back after the argument, we have a selection.
	// rl.enterVisualMode()
	rl.mark = rl.pos
	rl.activeRegion = true
}

func (rl *Instance) viYankWholeLine() {
	rl.saveBufToRegister(rl.line)
	rl.viUndoSkipAppend = true
}

func (rl *Instance) viEndOfLine() {
	rl.pos = len(rl.line)
	rl.viUndoSkipAppend = true
}

func (rl *Instance) viMatchBracket() {
	rl.viUndoSkipAppend = true

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
	rl.enterVioppMode("")
	rl.updateCursor()

	// Read the argument key to use as a pattern to search
	key, esc := rl.readArgumentKey()
	if esc {
		rl.exitVioppMode()
		rl.updateCursor()
		return
	}
	rl.exitVioppMode()
	rl.updateCursor()

	forward := true
	skip := false
	times := rl.getViIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func (rl *Instance) viFindNextCharSkip() {
	rl.enterVioppMode("")
	rl.updateCursor()

	// Read the argument key to use as a pattern to search
	key, esc := rl.readArgumentKey()
	if esc {
		rl.exitVioppMode()
		rl.updateCursor()
		return
	}
	rl.exitVioppMode()
	rl.updateCursor()

	forward := true
	skip := true
	times := rl.getViIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func (rl *Instance) viFindPrevChar() {
	rl.enterVioppMode("")
	rl.updateCursor()

	// Read the argument key to use as a pattern to search
	key, esc := rl.readArgumentKey()
	if esc {
		rl.exitVioppMode()
		rl.updateCursor()
		return
	}
	rl.exitVioppMode()
	rl.updateCursor()

	forward := false
	skip := false
	times := rl.getViIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func (rl *Instance) viFindPrevCharSkip() {
	rl.enterVioppMode("")
	rl.updateCursor()

	// Read the argument key to use as a pattern to search
	key, esc := rl.readArgumentKey()
	if esc {
		rl.exitVioppMode()
		rl.updateCursor()
		return
	}
	rl.exitVioppMode()
	rl.updateCursor()

	forward := false
	skip := true
	times := rl.getViIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func (rl *Instance) viDelete() {
	// When we are called after a pending operator action, we are a pending
	// usually not in visual mode, but have an active selection.
	// In this case we yank the active region and return.
	if rl.activeRegion || rl.local == visual {
		rl.deleteSelection()
		rl.resetSelection()

		if rl.local == visual {
			rl.viCommandMode()
			rl.updateCursor()
		}

		return
	}

	// If we are in operator pending mode, that means the command
	// is 'yy' (optionally with iterations), so we copy the required
	if rl.local == viopp {
		rl.killWholeLine()
		return
	}

	// Else if we are actually starting a yank action. We need an argument:
	// Enter operator pending mode for the next key to be considered this
	// argument (more precisely, the widget to be executed before this argument).
	rl.enterVioppMode("vi-delete")
	rl.updateCursor()

	// We set the initial mark, so that when executing this
	// widget back after the argument, we have a selection.
	// rl.enterVisualMode()
	rl.mark = rl.pos
	rl.activeRegion = true
}

func (rl *Instance) viDigitOrBeginningOfLine() {
	// If the last command was a digit argument,
	// then our Vi iterations' length is not 0
	if len(rl.viIteration) > 0 {
		rl.addIteration("0")
		return
	}

	// Else we go the beginning of line.
	rl.beginningOfLine()
}

func (rl *Instance) viSelectABlankWord() {
	// Go the beginning of the word and start mark
	rl.pos++
	rl.viBackwardBlankWord()
	if rl.local == visual || rl.local == viopp {
		rl.mark = rl.pos
	}

	// Then go to the end of the blank word
	rl.viForwardBlankWord()
	if rl.local == visual || rl.local == viopp {
		rl.pos--
		rl.activeRegion = true
	}
}

func (rl *Instance) viSelectAShellWord() {
	sBpos, sEpos, _, _ := rl.searchSurround('\'')
	dBpos, dEpos, _, _ := rl.searchSurround('"')

	// If none matched, use blankword
	if (sBpos == -1 || sEpos == -1) && (dBpos == -1 || dEpos == -1) {
		rl.viSelectABlankWord()

		return
	}

	// Or, we have one of both.
	if (sBpos == -1 || sEpos == -1) || dBpos < sBpos {
		rl.mark = dBpos
		rl.pos = dEpos
	} else if (dBpos == -1 || dEpos == -1) || sBpos < dBpos {
		rl.mark = sBpos
		rl.pos = sEpos
	}

	rl.activeRegion = true
}

func (rl *Instance) viSelectAWord() {
	// Go the beginning of the word and start mark
	rl.pos++
	rl.viBackwardWord()
	if rl.local == visual || rl.local == viopp {
		rl.mark = rl.pos
	}

	// Then go to the end of the blank word
	rl.viForwardWord()
	if rl.local == visual || rl.local == viopp {
		rl.pos--
		rl.activeRegion = true
	}
}

func (rl *Instance) viSelectInBlankWord() {
	// Go the beginning of the word and start mark
	rl.pos++
	rl.viBackwardBlankWord()
	if rl.local == visual || rl.local == viopp {
		rl.mark = rl.pos
	}

	// Then go to the end of the blank word
	rl.viForwardBlankWordEnd()
	if rl.local == visual || rl.local == viopp {
		rl.activeRegion = true
	}
}

func (rl *Instance) viSelectInShellWord() {
	sBpos, sEpos, _, _ := rl.searchSurround('\'')
	dBpos, dEpos, _, _ := rl.searchSurround('"')

	// If none matched, use blankword
	if (sBpos == -1 || sEpos == -1) && (dBpos == -1 || dEpos == -1) {
		rl.viSelectInBlankWord()

		return
	}

	// Or, we have one of both.
	if (sBpos == -1 || sEpos == -1) || dBpos < sBpos {
		rl.mark = dBpos + 1
		rl.pos = dEpos - 1
	} else if (dBpos == -1 || dEpos == -1) || sBpos < dBpos {
		rl.mark = sBpos + 1
		rl.pos = sEpos - 1
	}

	rl.activeRegion = true
}

func (rl *Instance) viSelectInWord() {
	// Go the beginning of the word and start mark
	rl.pos++
	rl.viBackwardWord()
	if rl.local == visual || rl.local == viopp {
		rl.mark = rl.pos
	}

	// Then go to the end of the blank word
	rl.viForwardWordEnd()
	if rl.local == visual || rl.local == viopp {
		rl.activeRegion = true
	}
}

func (rl *Instance) viGotoColumn() {
	iterations := rl.viIteration
	column := rl.getViIterations()

	if iterations == "" {
		column = 0
	} else if column < 0 {
		column = len(rl.line) + column
	}

	rl.pos = column
}

func (rl *Instance) viSwapCase() {
	if rl.local == visual {
		posInit := rl.pos

		bpos, epos, _ := rl.getSelection()
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

		return
	}

	char := rl.line[rl.pos]
	if unicode.IsLower(char) {
		char = unicode.ToUpper(char)
	} else {
		char = unicode.ToLower(char)
	}

	rl.line[rl.pos] = char
}

func (rl *Instance) viOperSwapCase() {
	if rl.activeRegion || rl.local == visual {
		posInit := rl.pos

		bpos, epos, cpos := rl.getSelection()
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

		return
	}

	// Else if we are actually starting a yank action. We need an argument:
	// Enter operator pending mode for the next key to be considered this
	// argument (more precisely, the widget to be executed before this argument).
	rl.enterVioppMode("vi-oper-swap-case")
	rl.updateCursor()

	// We set the initial mark, so that when executing this
	// widget back after the argument, we have a selection.
	// rl.enterVisualMode()
	rl.mark = rl.pos
	rl.activeRegion = true
}

func (rl *Instance) viFirstNonBlank() {
	for i := range rl.line {
		if rl.line[i] == ' ' {
			rl.pos = i
			break
		}
	}
}

func (rl *Instance) viAddSurround() {
	rl.enterVioppMode("")
	rl.updateCursor()

	key, esc := rl.readArgumentKey()
	if esc {
		rl.exitVioppMode()
		rl.updateCursor()
		return
	}

	rl.exitVioppMode()
	rl.updateCursor()

	// Surround the selection
	bpos, epos, _ := rl.getSelection()
	selection := string(rl.line[bpos:epos])
	selection = key + selection + key
	rl.resetSelection()

	// Assemble
	begin := string(rl.line[:bpos])
	end := string(rl.line[epos:])
	newLine := append([]rune(begin), []rune(selection)...)
	newLine = append(newLine, []rune(end)...)
	rl.line = newLine
}

func (rl *Instance) viSubstitute() {
	if rl.local == visual {
		rl.deleteSelection()
		rl.resetSelection()
		rl.viInsertMode()

		return
	}

	vii := rl.getViIterations()
	rl.saveToRegister(vii)

	for i := 1; i <= vii; i++ {
		rl.deletex()
	}

	rl.viInsertMode()
}

func (rl *Instance) viChange() {
	// In visual mode, we have just have a selection to delete.
	if rl.local == visual {
		rl.deleteSelection()
		rl.resetSelection()
		rl.viInsertMode()

		return
	}

	// Otherwise, we have to read first key, which
	// is either a navigation or selection widget.
	rl.enterVioppMode("")
	rl.updateCursor()

	// Read the argument key to use as a pattern to search
	key, esc := rl.readArgumentKey()
	if esc {
		rl.exitVioppMode()
		rl.updateCursor()
		return
	}
	rl.exitVioppMode()
	rl.updateCursor()

	// Find the widget
	action, found := changeMovements[key]
	if !found {
		return
	}

	widget := rl.getWidget(action)
	if widget == nil {
		return
	}

	// Update the pending keys, with an except for surround widgets.
	rl.keys = ""
	if action == "vi-select-surround" {
		rl.keys = key
	}

	// Before running the widget, set the mark
	rl.mark = rl.pos
	rl.activeRegion = true

	// Run the widget. We don't care about return values
	widget([]rune(key))

	rl.deleteSelection()
	rl.resetSelection()

	if action != "vi-change-surround" {
		rl.viInsertMode()
	}
}

func (rl *Instance) viChangeSurround() {
	rl.enterVioppMode("")
	rl.updateCursor()

	defer func() {
		rl.exitVioppMode()
		rl.updateCursor()
	}()

	// Read a key as a rune to search for
	key, esc := rl.readArgumentKey()
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
	rl.addRegion("surround", bpos, bpos+1, "", seqBgRed)
	rl.addRegion("surround", epos, epos+1, "", seqBgRed)
	rl.updateHelpers()
	defer func() { rl.resetRegions() }()

	// Now read another key.
	key, esc = rl.readArgumentKey()
	if esc {
		return
	}

	rchar := rune(key[0])

	// There might be a matching equivalent.
	bchar, echar := rl.matchSurround(rchar)

	rl.line[bpos] = bchar
	rl.line[epos] = echar
}

func (rl *Instance) viSelectSurround() {
	var inside bool

	switch rl.keys[0] {
	case 'i':
		inside = true
		rl.keys = rl.keys[1:]
	case 'a':
		rl.keys = rl.keys[1:]
	}

	if len(rl.keys) == 0 {
		rl.enterVioppMode("")
		rl.updateCursor()

		// Read a key as a rune to search for
		key, esc := rl.readArgumentKey()
		if esc {
			rl.exitVioppMode()
			rl.updateCursor()
			return
		}
		rl.exitVioppMode()
		rl.updateCursor()
		rl.keys += key
	}

	char := rune(rl.keys[0])

	bpos, epos, _, _ := rl.searchSurround(char)
	if bpos == -1 && epos == -1 {
		return
	}

	if inside {
		bpos++
	} else {
		epos++
	}

	rl.mark = bpos
	rl.pos = epos - 1
	rl.activeRegion = true
}

func (rl *Instance) viSetMark() {
	rl.mark = rl.pos
}
