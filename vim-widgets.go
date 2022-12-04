package readline

import (
	"fmt"
)

// vimHandlers maps keys to Vim actions.
type viWidgets map[string]func(rl *Instance)

var standardViWidgets = viWidgets{
	"vi-cmd-mode":               viCommandMode,
	"vi-insert-mode":            viInsertMode,
	"vi-insert-bol":             viInsertBol,
	"vi-backward-char":          viBackwardChar,
	"vi-forward-char":           viForwardChar,
	"vi-delete-char":            viDeleteChar,
	"vi-backward-delete-char":   viBackwardDeleteChar,
	"vi-forward-word":           viForwardWord,
	"vi-forward-blank-word":     viForwardBlankWord,
	"vi-forward-word-end":       viForwardWordEnd,
	"vi-forward-blank-word-end": viForwardBlankWordEnd,
	"vi-backward-word":          viBackwardWord,
	"vi-backward-blank-word":    viBackwardBlankWord, // TODO vi-backward-blank-word-end/vi-backward-word-end (ge / gE)
	"vi-kill-eol":               viKillEol,
	"vi-change-eol":             viChangeEol,
	"vi-edit-command-line":      viEditCommandLine,
	"vi-add-eol":                viAddEol,
	"vi-add-next":               viAddNext,
	"vi-put-after":              viPutAfter,
	"vi-put-before":             viPutBefore,
	"vi-end-of-line":            viEndOfLine,
	"vi-set-buffer":             viSetBuffer,
	"vi-yank":                   viYank,
	"vi-yank-whole-line":        viYankWholeLine,
	"vi-find-next-char":         viFindNextChar,
	"vi-find-next-char-skip":    viFindNextCharSkip,
	"vi-find-prev-char":         viFindPrevChar,
	"vi-find-prev-char-skip":    viFindPrevCharSkip,
	"vi-delete":                 viDelete,
	"vi-replace-chars":          viReplaceChars,
	"vi-replace":                viReplace,
}

var viinsWidgets = map[string]keyHandler{
	"visual-mode":                   viVisualMode,
	"vi-digit-or-beginning-of-line": viDigitOrBeginningOfLine,
}

// vimEditorWidgets maps Vim widget names (named almost identically to ZSH ones)
// to their function implementation. All widgets should be mapped in here.
var vimEditorWidgets = viWidgets{
	"vi-move-around-surround": viJumpBracket, // %

	// Non-standard
	"vi-jump-previous-brace": viJumpPreviousBrace,
	"vi-jump-next-brace":     viJumpNextBrace,
}

func viInsertMode(rl *Instance) {
	rl.main = viins

	rl.viIteration = ""
	rl.mark = -1
	rl.activeRegion = false
	rl.visualLine = false

	rl.updateCursor()

	rl.refreshVimStatus()
}

func viCommandMode(rl *Instance) {
	rl.viIteration = ""
	rl.viUndoSkipAppend = true
	rl.mark = -1
	rl.activeRegion = false
	rl.visualLine = false

	// Only go back if not in insert mode
	if rl.main == viins && len(rl.line) > 0 && rl.pos > 0 {
		rl.pos--
	}

	rl.local = ""
	rl.main = vicmd

	rl.updateCursor()

	rl.refreshVimStatus()
}

func viVisualMode(rl *Instance, _ []byte, i int, r []rune) (read, ret bool, val string, err error) {
	lastMode := rl.local
	wasVisualLine := rl.visualLine

	rl.viIteration = ""
	rl.viUndoSkipAppend = true

	switch string(r[:i]) {
	case "V":
		rl.enterVisualLineMode()
	case "v":
		rl.enterVisualMode()
	default:
		rl.local = lastMode
	}

	// We don't do anything else if the mode did not change.
	if lastMode == rl.local && wasVisualLine == rl.visualLine {
		return
	}

	rl.updateCursor()

	return
}

func viInsertBol(rl *Instance) {
	rl.main = viins

	rl.viIteration = ""
	rl.viUndoSkipAppend = true

	rl.pos = 0

	rl.updateCursor()

	rl.refreshVimStatus()
}

func viDigitOrBeginningOfLine(rl *Instance, b []byte, i int, r []rune) (read, ret bool, val string, err error) {
	// If the last command was a digit argument,
	// then our Vi iterations' length is not 0
	if len(rl.viIteration) > 0 {
		return digitArgument(rl, b, i, r)
	}

	// Else we go the beginning of line.
	read, ret, err = beginningOfLine(rl)

	return
}

func viAddNext(rl *Instance) {
	if len(rl.line) > 0 {
		rl.pos++
	}

	viInsertMode(rl)
}

func viAddEol(rl *Instance) {
	if len(rl.line) > 0 {
		rl.pos = len(rl.line)
	}

	viInsertMode(rl)
}

func viBackwardWord(rl *Instance) {
	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	}
}

func viBackwardBlankWord(rl *Instance) {
	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))
	}
}

func viKillEol(rl *Instance) {
	rl.saveBufToRegister(rl.line[rl.pos-1:])
	rl.line = rl.line[:rl.pos]
	// Only go back if there is an input
	if len(rl.line) > 0 {
		rl.pos--
	}
	rl.viIteration = ""
	rl.resetHelpers()
	rl.updateHelpers()
}

func viChangeEol(rl *Instance) {
	rl.saveBufToRegister(rl.line[rl.pos-1:])
	rl.line = rl.line[:rl.pos]
	rl.viIteration = ""

	rl.resetHelpers()

	viInsertMode(rl)
}

func viForwardWordEnd(rl *Instance) {
	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
	}
}

func viForwardBlankWordEnd(rl *Instance) {
	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpE(tokeniseSplitSpaces))
	}
}

func viInsert(rl *Instance) {
	rl.modeViMode = vimInsert
	rl.viIteration = ""
	rl.viUndoSkipAppend = true
	rl.registers.resetRegister()
}

func viInsertI(rl *Instance) {
	rl.modeViMode = vimInsert
	rl.viIteration = ""
	rl.viUndoSkipAppend = true
	rl.pos = 0
}

func viForwardChar(rl *Instance) {
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

func viBackwardChar(rl *Instance) {
	rl.viUndoSkipAppend = true

	if rl.pos > 0 {
		rl.pos--
	}
}

// TODO: If pasting multiple lines, instead of only characters, paste below the current line.
func viPutAfter(rl *Instance) {
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

func viPutBefore(rl *Instance) {
	// paste before
	rl.viUndoSkipAppend = true
	buffer := rl.pasteFromRegister()
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.insert(buffer)
	}
}

func viReplaceChars(rl *Instance) {
	rl.viUndoSkipAppend = true

	// We read a character to use first.
	print(cursorBlinkingUnderline)

	key, esc := rl.readArgumentKey()
	if esc {
		print(cursorBlinkingBlock)
		return
	}
	rl.updateCursor()

	// In visual mode, we replace all chars of the selection
	if rl.activeRegion || rl.local == visual {
		bpos, epos, _ := rl.getSelection()
		for i := bpos; i < epos; i++ {
			rl.line[i] = []rune(key)[0]
		}
		rl.pos = bpos

		viCommandMode(rl)

		return
	}

	// Or simply the character under the cursor.
	rl.deletex()
	rl.insert([]rune(key))
	rl.pos--
}

func viReplace(rl *Instance) {
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
		print(cursorBlinkingUnderline)

		// Read a new key
		keys, esc := rl.readArgumentKey()
		if esc {
			print(cursorBlinkingBlock)
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
	print(cursorBlinkingBlock)
}

func viEditCommandLine(rl *Instance) {
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
	if err != nil || len(new) == 0 || string(new) == string(multiline) {
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

func viForwardWord(rl *Instance) {
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

func viForwardBlankWord(rl *Instance) {
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

// TODO: Either redundant with deleteChar, or has to be modified somehow.
func viDeleteChar(rl *Instance) {
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
	// classic backwardDeleteChar and this function.
	//
	// if rl.pos == len(rl.line) && len(rl.line) > 0 {
	// 	rl.pos--
	// }
}

// TODO: Same here
func viBackwardDeleteChar(rl *Instance) {
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
	// classic backwardDeleteChar and this function.
	//
	if rl.pos == len(rl.line) && len(rl.line) > 0 {
		rl.pos--
	}
}

func viYank(rl *Instance) {
	// When we are called after a pending operator action, we are a pending
	// usually not in visual mode, but have an active selection.
	// In this case we yank the active region and return.
	if rl.activeRegion || rl.local == visual {
		rl.yankSelection()
		rl.resetSelection()

		if rl.local == visual {
			viCommandMode(rl)
			rl.updateCursor()
		}

		return
	}

	// If we are in operator pending mode, that means the command
	// is 'yy' (optionally with iterations), so we copy the required
	if rl.local == viopp {
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

func viYankWholeLine(rl *Instance) {
	rl.saveBufToRegister(rl.line)
	rl.viUndoSkipAppend = true
}

func viJumpPreviousBrace(rl *Instance) {
	rl.viUndoSkipAppend = true
	rl.moveCursorByAdjust(rl.viJumpPreviousBrace())
}

func viJumpNextBrace(rl *Instance) {
	rl.viUndoSkipAppend = true
	rl.moveCursorByAdjust(rl.viJumpNextBrace())
}

func viEndOfLine(rl *Instance) {
	rl.pos = len(rl.line)
	rl.viUndoSkipAppend = true
}

func viJumpBracket(rl *Instance) {
	rl.viUndoSkipAppend = true
	rl.moveCursorByAdjust(rl.viJumpBracket())
}

// TODO: Currently we don't handle the argument in this widget.
func viSetBuffer(rl *Instance) {
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

// TODO: only use a single rune to match against in those widgets
func viFindNextChar(rl *Instance) {
	print(cursorUnderline)

	// Read the argument key to use as a pattern to search
	key, esc := rl.readArgumentKey()
	if esc {
		return
	}
	rl.updateCursor()

	forward := true
	skip := false
	times := rl.getViIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func viFindNextCharSkip(rl *Instance) {
	print(cursorUnderline)

	// Read the argument key to use as a pattern to search
	key, esc := rl.readArgumentKey()
	if esc {
		return
	}
	rl.updateCursor()

	forward := true
	skip := true
	times := rl.getViIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func viFindPrevChar(rl *Instance) {
	print(cursorUnderline)

	// Read the argument key to use as a pattern to search
	key, esc := rl.readArgumentKey()
	if esc {
		return
	}
	rl.updateCursor()

	forward := false
	skip := false
	times := rl.getViIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func viFindPrevCharSkip(rl *Instance) {
	print(cursorUnderline)

	// Read the argument key to use as a pattern to search
	key, esc := rl.readArgumentKey()
	if esc {
		return
	}
	rl.updateCursor()

	forward := false
	skip := true
	times := rl.getViIterations()

	rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
}

func viDelete(rl *Instance) {
	// When we are called after a pending operator action, we are a pending
	// usually not in visual mode, but have an active selection.
	// In this case we yank the active region and return.
	if rl.activeRegion || rl.local == visual {
		rl.deleteSelection()
		rl.resetSelection()

		if rl.local == visual {
			viCommandMode(rl)
			rl.updateCursor()
		}

		return
	}

	// If we are in operator pending mode, that means the command
	// is 'yy' (optionally with iterations), so we copy the required
	if rl.local == viopp {
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
