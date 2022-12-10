package readline

type undoItem struct {
	line string
	pos  int
}

func (rl *Instance) undoAppendHistory() {
	defer func() {
		rl.viUndoSkipAppend = false
		if !rl.viIsUndoing {
			rl.viUndoPos = 0
		}

		rl.viIsUndoing = false
	}()

	if rl.viUndoSkipAppend {
		return
	}

	// When we add an item to the undo history, the history
	// is cut from the current undo hist position onwards.
	rl.viUndoHistory = rl.viUndoHistory[:len(rl.viUndoHistory)-rl.viUndoPos]

	rl.viUndoHistory = append(rl.viUndoHistory, undoItem{
		line: string(rl.line),
		pos:  rl.pos,
	})
}

func (rl *Instance) undo() {
	rl.viUndoSkipAppend = true
	rl.viIsUndoing = true
	if len(rl.viUndoHistory) == 0 {
		return
	}

	if rl.viUndoPos == 0 {
		rl.lineBuf = string(rl.line)
	}

	rl.viUndoPos++

	if rl.viUndoPos > len(rl.viUndoHistory) {
		rl.viUndoPos = len(rl.viUndoHistory)
		return
	}

	undo := rl.viUndoHistory[len(rl.viUndoHistory)-rl.viUndoPos]
	rl.line = []rune(undo.line)
	rl.pos = undo.pos

	// TODO: Also related to vim position adjustment after many handlers.
	// if rl.main != viins && len(rl.line) > 0 && rl.pos == len(rl.line) {
	// 	rl.pos--
	// }
}

func (rl *Instance) redo() {
	rl.viUndoSkipAppend = true
	rl.viIsUndoing = true
	if len(rl.viUndoHistory) == 0 {
		return
	}

	rl.viUndoPos--

	if rl.viUndoPos < 1 {
		rl.viUndoPos = 0

		rl.line = []rune(rl.lineBuf)
		rl.pos = len(rl.lineBuf)
		return
	}

	undo := rl.viUndoHistory[len(rl.viUndoHistory)-rl.viUndoPos]
	rl.line = []rune(undo.line)
	rl.pos = undo.pos

	// TODO: Also related to vim position adjustment after many handlers.
	// if rl.main != viins && len(rl.line) > 0 && rl.pos == len(rl.line) {
	// 	rl.pos--
	// }
}
