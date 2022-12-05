package readline

type lineWidgets map[string]keyHandler

// standardLineWidgets either need access to the input key,
// or need to return specific instructions and values.
func (rl *Instance) initStandardLineWidgets() lineWidgets {
	widgets := map[string]keyHandler{
		"accept-line":    rl.acceptLine,
		"self-insert":    rl.selfInsert,
		"digit-argument": rl.digitArgument,
	}

	return widgets
}

// standardWidgets don't need access to the input key.
func (rl *Instance) initStandardWidgets() baseWidgets {
	widgets := map[string]func(){
		"clear-screen":         rl.clearScreen,
		"beginning-of-line":    rl.beginningOfLine,
		"end-of-line":          rl.endOfLine,
		"kill-line":            rl.killLine,
		"kill-whole-line":      rl.killWholeLine,
		"backward-kill-word":   rl.backwardKillWord,
		"kill-word":            rl.killWord,
		"yank":                 rl.yank,
		"backward-delete-char": rl.backwardDeleteChar,
		"delete-char":          rl.deleteChar,
		"forward-char":         rl.forwardChar,
		"backward-char":        rl.backwardChar,
		"forward-word":         rl.forwardWord,
		"backward-word":        rl.backwardWord,
		"undo":                 rl.undo,
		"down-line-or-history": rl.historyNext,
		"up-line-or-history":   rl.historyPrev,
		"down-history":         rl.downHistory,
		"up-history":           rl.upHistory,
	}

	return widgets
}

func (rl *Instance) selfInsert(r []rune) (read, ret bool, val string, err error) {
	rl.viUndoSkipAppend = true

	// Prepare the line
	// line, err := rl.mainHistory.GetLine(rl.mainHistory.Len() - 1)
	// if err != nil {
	// 	return
	// }
	// if !rl.mainHist {
	// 	line, err = rl.altHistory.GetLine(rl.altHistory.Len() - 1)
	// 	if err != nil {
	// 		return
	// 	}
	// }
	//
	// tokens, _, _ := tokeniseSplitSpaces([]rune(line), 0)
	// pos := int(r[1]) - 48 // convert ASCII to integer
	// if pos > len(tokens) {
	// 	return
	// }
	//
	// // The line is prepared and the actual runes to insert are as well.
	// r = []rune(tokens[pos-1])

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

	// This should also update the rl.pos
	rl.updateHelpers()

	return
}

// acceptLine returns the line to the readline caller for being executed/evaluated.
func (rl *Instance) acceptLine(_ []rune) (read, ret bool, val string, err error) {
	if rl.modeTabCompletion {
		cur := rl.getCurrentGroup()

		// Check that there is a group indeed, as we might have no completions.
		if cur == nil {
			rl.clearHelpers()
			rl.resetTabCompletion()
			rl.renderHelpers()
			read = true

			return
		}

		// IF we have a prefix and completions printed, but no candidate
		// (in which case the completion is ""), we immediately return.
		completion := cur.getCurrentCell(rl)
		prefix := len(rl.tcPrefix)
		if prefix > len(completion.Value) {
			rl.carriageReturn()

			val = string(rl.line)
			ret = true

			return
		}

		// Else, we insert the completion candidate in the real input line.
		// By default we add a space, unless completion group asks otherwise.
		rl.compAddSpace = true
		rl.resetVirtualComp(false)

		// If we were in history completion, immediately execute the line.
		if rl.modeAutoFind && rl.searchMode == HistoryFind {
			rl.carriageReturn()

			val = string(rl.line)
			ret = true

			return
		}

		// Reset completions and update input line
		rl.clearHelpers()
		rl.resetTabCompletion()
		rl.renderHelpers()

		read = true
		return
	}

	rl.carriageReturn()

	val = string(rl.line)
	ret = true

	return
}

func (rl *Instance) clearScreen() {
	print(seqClearScreen)
	print(seqCursorTopLeft)

	// Print the prompt, all or part of it.
	print(rl.Prompt.getPrimary())
	print(seqClearScreenBelow)

	rl.resetHintText()
	rl.getHintText()
	rl.renderHelpers()

	return
}

func (rl *Instance) beginningOfLine() {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}

	rl.viUndoSkipAppend = true
	rl.pos = 0
	rl.updateHelpers()

	return
}

func (rl *Instance) endOfLine() {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}

	if len(rl.line) > 0 {
		rl.pos = len(rl.line)
	}

	rl.viUndoSkipAppend = true
	rl.updateHelpers()

	return
}

func (rl *Instance) killLine() {
	rl.saveBufToRegister(rl.line[rl.pos-1:])
	rl.line = rl.line[:rl.pos]
	rl.resetHelpers()
	rl.updateHelpers()
	rl.viIteration = ""
	return
}

func (rl *Instance) killWholeLine() {
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
	rl.lineComp = []rune{}
	rl.pos = 0
	rl.posX = 0
	rl.fullX = 0
	rl.posY = 0
	rl.fullY = 0

	// Completions are also reset
	rl.clearVirtualComp()
	return
}

func (rl *Instance) backwardKillWord() {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}

	rl.saveToRegister(rl.viJumpB(tokeniseLine))
	rl.viDeleteByAdjust(rl.viJumpB(tokeniseLine))
	rl.updateHelpers()

	return
}

func (rl *Instance) killWord() {
	rl.saveToRegisterTokenize(tokeniseLine, rl.viJumpE, 1)
	rl.viDeleteByAdjust(rl.viJumpE(tokeniseLine) + 1)

	return
}

func (rl *Instance) yank() {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	// paste after the cursor position
	rl.viUndoSkipAppend = true
	buffer := rl.pasteFromRegister()
	rl.insert(buffer)
	rl.updateHelpers()

	return
}

func (rl *Instance) backwardDeleteChar() {
	vii := rl.getViIterations()

	// We might be on an active register, but not yanking...
	rl.saveToRegister(vii)

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {
		rl.deleteX()
	}

	return
}

func (rl *Instance) deleteChar() {
	vii := rl.getViIterations()

	// We might be on an active register, but not yanking...
	rl.saveToRegister(vii)

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {
		rl.deletex()
	}
	if rl.pos == len(rl.line) && len(rl.line) > 0 {
		rl.pos--
	}

	return
}

func (rl *Instance) forwardChar() {
	if rl.pos < len(rl.line) {
		rl.pos++
	}

	return
}

func (rl *Instance) backwardChar() {
	if rl.pos > 0 {
		rl.pos--
	}
	rl.viUndoSkipAppend = true

	return
}

func (rl *Instance) forwardWord() {
	// If we were not yanking
	rl.viUndoSkipAppend = true

	// If the input line is empty, we don't do anything
	if rl.pos == 0 && len(rl.line) == 0 {
		return
	}

	// Get iterations and move
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpW(tokeniseLine))
	}

	return
}

func (rl *Instance) backwardWord() {
	rl.viUndoSkipAppend = true

	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	}

	return
}

// TODO: Probably should be including undoLast() code without Vim stuff ?
func (rl *Instance) undo() {
	rl.undoLast()
	rl.viUndoSkipAppend = true

	return
}

func (rl *Instance) downHistory() {
	rl.mainHist = true
	rl.walkHistory(-1)

	return
}

func (rl *Instance) upHistory() {
	rl.mainHist = true
	rl.walkHistory(1)

	return
}

// digitArgument is used both in Emacs and Vim modes,
// but strips the Alt modifier used in Emacs mode.
func (rl *Instance) digitArgument(r []rune) (read, ret bool, val string, err error) {
	if len(r) > 1 {
		// The first rune is the alt modifier.
		rl.viIteration += string(r[1:])
	} else {
		rl.viIteration += string(r)
	}

	rl.viUndoSkipAppend = true

	return
}

func (rl *Instance) historyNext() {
	rl.viUndoSkipAppend = true
	rl.mainHist = true
	rl.walkHistory(-1)

	return
}

func (rl *Instance) historyPrev() {
	rl.viUndoSkipAppend = true
	rl.mainHist = true
	rl.walkHistory(1)

	return
}
