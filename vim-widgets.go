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
	"vi-end-of-line":            viEndOfLine,
	"vi-set-buffer":             viSetBuffer,
}

var viinsWidgets = map[string]keyHandler{
	"visual-mode":                   viVisualMode,
	"vi-digit-or-beginning-of-line": viDigitOrBeginningOfLine,
}

// vimEditorWidgets maps Vim widget names (named almost identically to ZSH ones)
// to their function implementation. All widgets should be mapped in here.
var vimEditorWidgets = viWidgets{
	"vi-delete":               viDelete,      // d
	"down-line-or-history":    viHistoryNext, // j
	"up-line-or-history":      viHistoryPrev, // k
	"vi-put-before":           viPasteP,      // P
	"vi-replace-chars":        viReplace,     // r
	"vi-replace":              viReplaceR,    // R
	"vi-yank":                 viYank,        // y "vi-yank": dummyHandler, // y
	"vi-yank-whole-line":      viYankY,       // Y
	"vi-move-around-surround": viJumpBracket, // %

	// Non-standard
	"vi-jump-previous-brace": viJumpPreviousBrace,
	"vi-jump-next-brace":     viJumpNextBrace,
}

func viInsertMode(rl *Instance) {
	rl.main = viins

	rl.viIteration = ""
	rl.viUndoSkipAppend = true
	rl.mark = -1

	// Change the cursor
	print(cursorBlinkingBeam)

	rl.refreshVimStatus()
}

func viCommandMode(rl *Instance) {
	rl.viIteration = ""
	rl.viUndoSkipAppend = true
	rl.mark = -1

	// Only go back if not in insert mode
	if rl.main == viins && len(rl.line) > 0 && rl.pos > 0 {
		rl.pos--
	}

	rl.local = ""
	rl.main = vicmd

	// Change the cursor
	print(cursorBlinkingBlock)

	rl.refreshVimStatus()
}

func viVisualMode(rl *Instance, _ []byte, i int, r []rune) (read, ret bool, val string, err error) {
	lastMode := rl.local
	wasVisualLine := rl.visualLine

	rl.viIteration = ""
	rl.viUndoSkipAppend = true

	switch string(r[:i]) {
	case "V":
		rl.local = visual
		rl.visualLine = true
		rl.mark = 0 // start at the beginning of the line.
	case "v":
		rl.local = visual
		rl.visualLine = false
		rl.mark = rl.pos // Or rl.posX ? combined with rl.posY ?
	default:
		rl.local = lastMode
	}

	// We don't do anything else if the mode did not change.
	if lastMode == rl.local && wasVisualLine == rl.visualLine {
		return
	}

	print(cursorBlock)

	return
}

func viInsertBol(rl *Instance) {
	rl.main = viins

	rl.viIteration = ""
	rl.viUndoSkipAppend = true

	rl.pos = 0

	// Change the cursor
	print(cursorBlinkingBeam)

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
	if rl.viIsYanking {
		vii := rl.getViIterations()
		rl.saveToRegisterTokenize(tokeniseLine, rl.viJumpB, vii)
		rl.viIsYanking = false
		return
	}
	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	}
}

func viBackwardBlankWord(rl *Instance) {
	if rl.viIsYanking {
		vii := rl.getViIterations()
		rl.saveToRegisterTokenize(tokeniseSplitSpaces, rl.viJumpB, vii)
		rl.viIsYanking = false
		return
	}
	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))
	}
}

func viDelete(rl *Instance) {
	rl.modeViMode = vimDelete
	rl.viUndoSkipAppend = true
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
	if rl.viIsYanking {
		vii := rl.getViIterations()
		rl.saveToRegisterTokenize(tokeniseLine, rl.viJumpE, vii)
		rl.viIsYanking = false
		return
	}

	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
	}
}

func viForwardBlankWordEnd(rl *Instance) {
	if rl.viIsYanking {
		vii := rl.getViIterations()
		rl.saveToRegisterTokenize(tokeniseSplitSpaces, rl.viJumpE, vii)
		rl.viIsYanking = false
		return
	}

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

func viHistoryNext(rl *Instance) {
	rl.mainHist = true
	rl.walkHistory(-1)
}

func viHistoryPrev(rl *Instance) {
	rl.mainHist = true
	rl.walkHistory(1)
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

func viPasteP(rl *Instance) {
	// paste before
	rl.viUndoSkipAppend = true
	buffer := rl.pasteFromRegister()
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.insert(buffer)
	}
}

func viReplace(rl *Instance) {
	rl.modeViMode = vimReplaceOnce
	rl.viIteration = ""
	rl.viUndoSkipAppend = true
}

func viReplaceR(rl *Instance) {
	rl.modeViMode = vimReplaceMany
	rl.viIteration = ""
	rl.viUndoSkipAppend = true
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
	// If we were not yanking
	rl.viUndoSkipAppend = true
	// If the input line is empty, we don't do anything
	if rl.pos == 0 && len(rl.line) == 0 {
		return
	}

	// If we were yanking, we forge the new yank buffer
	// and return without moving the cursor.
	if rl.viIsYanking {
		vii := rl.getViIterations()
		rl.saveToRegisterTokenize(tokeniseLine, rl.viJumpW, vii)
		rl.viIsYanking = false
		return
	}

	// Else get iterations and move
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpW(tokeniseLine))
	}
}

func viForwardBlankWord(rl *Instance) {
	// If the input line is empty, we don't do anything
	if rl.pos == 0 && len(rl.line) == 0 {
		return
	}
	rl.viUndoSkipAppend = true

	if rl.viIsYanking {
		vii := rl.getViIterations()
		rl.saveToRegisterTokenize(tokeniseSplitSpaces, rl.viJumpW, vii)
		rl.viIsYanking = false
		return
	}
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
	if rl.viIsYanking {
		rl.saveBufToRegister(rl.line)
		rl.viIsYanking = false
	}
	rl.viIsYanking = true
	rl.viUndoSkipAppend = true
}

func viYankY(rl *Instance) {
	rl.saveBufToRegister(rl.line)
	rl.viUndoSkipAppend = true
}

func viJumpPreviousBrace(rl *Instance) {
	if rl.viIsYanking {
		rl.saveToRegister(rl.viJumpPreviousBrace())
		rl.viIsYanking = false
		return
	}
	rl.viUndoSkipAppend = true
	rl.moveCursorByAdjust(rl.viJumpPreviousBrace())
}

func viJumpNextBrace(rl *Instance) {
	if rl.viIsYanking {
		rl.saveToRegister(rl.viJumpNextBrace())
		rl.viIsYanking = false
		return
	}
	rl.viUndoSkipAppend = true
	rl.moveCursorByAdjust(rl.viJumpNextBrace())
}

func viEndOfLine(rl *Instance) {
	if rl.viIsYanking {
		rl.saveBufToRegister(rl.line[rl.pos:])
		rl.viIsYanking = false
		return
	}
	rl.pos = len(rl.line)
	rl.viUndoSkipAppend = true
}

func viJumpBracket(rl *Instance) {
	if rl.viIsYanking {
		rl.saveToRegister(rl.viJumpBracket())
		rl.viIsYanking = false
		return
	}
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
	rl.registers.registerSelectWait = true
}
