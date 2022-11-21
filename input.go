package readline

// NOTE: To this is where we must change things like ErrorCtrlC
// Note also that this function will most of the time return an error, or will probably trigger
// a new key read loop, aborting any other editor/completion components to use the signal key.
func (rl *Instance) inputErrorKeys(b []byte, i int) (done, ret bool, val string, err error) {
	switch b[0] {
	case charCtrlC:
		if rl.modeTabCompletion {
			rl.resetVirtualComp(true)
			rl.resetHelpers()
			rl.renderHelpers()
			done = true

			return
		}
		rl.clearHelpers()

		return done, true, val, CtrlC

	case charEOF:
		rl.clearHelpers()
		ret = true

		return done, true, val, EOF
	}

	return
}

// Escape key generally aborts most completion/prompt helpers.
func (rl *Instance) inputEsc(r []rune, b []byte, i int) {
	// If we were waiting for completion confirm, abort
	if rl.compConfirmWait {
		rl.compConfirmWait = false
		rl.renderHelpers()
	}

	// We always refresh the completion candidates, except if we are currently
	// cycling through them, because then it would just append the candidate.
	if rl.modeTabCompletion {
		if string(r[:i]) != seqShiftTab &&
			string(r[:i]) != seqForwards && string(r[:i]) != seqBackwards &&
			string(r[:i]) != seqUp && string(r[:i]) != seqDown {
			rl.resetVirtualComp(false)
		}
	}

	// Once helpers of all sorts are cleared, we can process
	// the change of input modes, etc.
	rl.escapeSeq(r[:i])
}

func (rl *Instance) inputEnter() (done, ret bool, val string, err error) {
	if rl.modeTabCompletion {
		cur := rl.getCurrentGroup()

		// Check that there is a group indeed, as we might have no completions.
		if cur == nil {
			rl.clearHelpers()
			rl.resetTabCompletion()
			rl.renderHelpers()
			done = true

			return
		}

		// IF we have a prefix and completions printed, but no candidate
		// (in which case the completion is ""), we immediately return.
		completion := cur.getCurrentCell(rl)
		prefix := len(rl.tcPrefix)
		if prefix > len(completion) {
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

		done = true
		return
	}

	rl.carriageReturn()

	val = string(rl.line)
	ret = true

	return
}

// inputEmacs runs the provided keystroke if its a line modifier in Emacs mode.
func (rl *Instance) inputEmacs(b []byte, i int) (done, ret bool, val string, err error) {
	switch b[0] {
	case charCtrlU:
		if rl.modeTabCompletion {
			rl.resetVirtualComp(true)
		}
		// Delete everything from the beginning of the line to the cursor position
		rl.saveBufToRegister(rl.line[:rl.pos])
		rl.deleteToBeginning()
		rl.resetHelpers()
		rl.updateHelpers()

	case charCtrlW:
		if rl.modeTabCompletion {
			rl.resetVirtualComp(false)
		}
		// This is only available in Insert mode
		if rl.modeViMode != vimInsert {
			done = true
			return
		}
		rl.saveToRegister(rl.viJumpB(tokeniseLine))
		rl.viDeleteByAdjust(rl.viJumpB(tokeniseLine))
		rl.updateHelpers()

	case charCtrlY:
		if rl.modeTabCompletion {
			rl.resetVirtualComp(false)
		}
		// paste after the cursor position
		rl.viUndoSkipAppend = true
		buffer := rl.pasteFromRegister()
		rl.insert(buffer)
		rl.updateHelpers()

	case charCtrlE:
		if rl.modeTabCompletion {
			rl.resetVirtualComp(false)
		}
		// This is only available in Insert mode
		if rl.modeViMode != vimInsert {
			done = true
			return
		}
		if len(rl.line) > 0 {
			rl.pos = len(rl.line)
		}
		rl.viUndoSkipAppend = true
		rl.updateHelpers()

	case charCtrlA:
		if rl.modeTabCompletion {
			rl.resetVirtualComp(false)
		}
		// This is only available in Insert mode
		if rl.modeViMode != vimInsert {
			done = true
			return
		}
		rl.viUndoSkipAppend = true
		rl.pos = 0
		rl.updateHelpers()
	}

	return
}

// inputDispatch handles any key that is not a key press not bound to a core action.
// This means grossly not error codes/signals, completion/menu keys, and editing mode changes.
func (rl *Instance) inputDispatch(r []rune, i int) (done, ret bool, val string, err error) {
	// If we were waiting for completion confirm, abort it and go on with our input.
	if rl.compConfirmWait {
		rl.resetVirtualComp(false)
		rl.compConfirmWait = false
		rl.renderHelpers()
	}

	// Completion modes hijack the text input, and cuts
	// the editor from using/interpreting the key.
	if rl.modeAutoFind && rl.searchMode == HistoryFind {
		rl.resetVirtualComp(true)
		rl.updateTabFind(r[:i])
		rl.updateVirtualComp()
		rl.renderHelpers()
		rl.viUndoSkipAppend = true

		done = true

		return
	}

	// Not sure that CompletionFind is useful, nor one of the other two
	if rl.modeAutoFind || rl.modeTabFind {
		rl.resetVirtualComp(false)
		rl.updateTabFind(r[:i])
		rl.viUndoSkipAppend = true
	} else {
		rl.resetVirtualComp(false)
		rl.inputEditor(r[:i])
		if len(rl.multiline) > 0 && rl.modeViMode == vimKeys {
			rl.skipStdinRead = true
		}
	}

	// Notice we don't return done = true, since any action independent of our
	// while could still have to run while us not knowing it, so just shut up.
	return
}
