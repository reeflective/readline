package readline

type undoItem struct {
	line string
	pos  int
}

// will only skip appending to undo history if not forced
// to append by another pending/running widget.
func (rl *Instance) skipUndoAppend() {
	rl.undoSkipAppend = true
}

func (rl *Instance) undoHistoryAppend() {
	defer rl.resetUndoDirectives()

	if rl.undoSkipAppend {
		return
	}

	// When the line is identical to the previous undo, we skip it.
	if len(rl.undoHistory) > 0 {
		if rl.undoHistory[len(rl.undoHistory)-1].line == string(rl.line) {
			return
		}
	}

	// When we add an item to the undo history, the history
	// is cut from the current undo hist position onwards.
	rl.undoHistory = rl.undoHistory[:len(rl.undoHistory)-rl.undoPos]

	rl.undoHistory = append(rl.undoHistory, undoItem{
		line: string(rl.line),
		pos:  rl.pos,
	})
}

func (rl *Instance) undo() {
	rl.undoSkipAppend = true
	rl.isUndoing = true
	if len(rl.undoHistory) == 0 {
		return
	}

	if rl.undoPos == 0 {
		rl.lineBuf = string(rl.line)
	}

	rl.undoPos++

	if rl.undoPos > len(rl.undoHistory) {
		rl.undoPos = len(rl.undoHistory)
		return
	}

	undo := rl.undoHistory[len(rl.undoHistory)-rl.undoPos]
	rl.line = []rune(undo.line)
	rl.pos = undo.pos
}

func (rl *Instance) redo() {
	rl.undoSkipAppend = true
	rl.isUndoing = true
	if len(rl.undoHistory) == 0 {
		return
	}

	rl.undoPos--

	if rl.undoPos < 1 {
		rl.undoPos = 0

		rl.line = []rune(rl.lineBuf)
		rl.pos = len(rl.lineBuf)
		return
	}

	undo := rl.undoHistory[len(rl.undoHistory)-rl.undoPos]
	rl.line = []rune(undo.line)
	rl.pos = undo.pos
}

func (rl *Instance) resetUndoDirectives() {
	rl.undoSkipAppend = false

	if !rl.isUndoing {
		rl.undoPos = 0
	}

	rl.isUndoing = false
}
