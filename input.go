package readline

func (rl *Instance) errorCtrlC() (done, ret bool) {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(true)
		rl.resetHelpers()
		rl.renderHelpers()

		return true, false
	}
	rl.clearHelpers()

	return false, true
}

func (rl *Instance) inputBackspace() (done bool) {
	// When currently in history completion, we refresh and automatically
	// insert the first (filtered) candidate, virtually
	if rl.modeAutoFind && rl.searchMode == HistoryFind {
		rl.resetVirtualComp(true)
		rl.backspaceTabFind()

		// Then update the printing, with the new candidate
		rl.updateVirtualComp()
		rl.renderHelpers()
		rl.viUndoSkipAppend = true
		return true
	}

	// Normal completion search does only refresh the search pattern and the comps
	if rl.modeTabFind || rl.modeAutoFind {
		rl.backspaceTabFind()
		rl.viUndoSkipAppend = true
	} else {
		// Always cancel any virtual completion
		rl.resetVirtualComp(false)

		// Vim mode has different behaviors
		if rl.InputMode == Vim {
			if rl.modeViMode == vimInsert {
				rl.backspace()
			} else {
				rl.pos--
			}
			rl.renderHelpers()
			return true
		}

		// Else emacs deletes a character
		rl.backspace()
		rl.renderHelpers()
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

func (rl *Instance) deleteLine() {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(true)
	}
	// Delete everything from the beginning of the line to the cursor position
	rl.saveBufToRegister(rl.line[:rl.pos])
	rl.deleteToBeginning()
	rl.resetHelpers()
	rl.updateHelpers()
}

func (rl *Instance) deleteWord() (done bool) {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	// This is only available in Insert mode
	if rl.modeViMode != vimInsert {
		return true
	}
	rl.saveToRegister(rl.viJumpB(tokeniseLine))
	rl.viDeleteByAdjust(rl.viJumpB(tokeniseLine))
	rl.updateHelpers()

	return
}

func (rl *Instance) pasteDefaultRegister() {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	// paste after the cursor position
	rl.viUndoSkipAppend = true
	buffer := rl.pasteFromRegister()
	rl.insert(buffer)
	rl.updateHelpers()
}

func (rl *Instance) goToInputEnd() (done bool) {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	// This is only available in Insert mode
	if rl.modeViMode != vimInsert {
		return true
	}
	if len(rl.line) > 0 {
		rl.pos = len(rl.line)
	}
	rl.viUndoSkipAppend = true
	rl.updateHelpers()

	return
}

func (rl *Instance) goToInputBegin() (done bool) {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	// This is only available in Insert mode
	if rl.modeViMode != vimInsert {
		return true
	}
	rl.viUndoSkipAppend = true
	rl.pos = 0
	rl.updateHelpers()

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
		if len(rl.multilineBuffer) > 0 && rl.modeViMode == vimKeys {
			rl.skipStdinRead = true
		}
	}

	// Notice we don't return done = true, since any action independent of our
	// while could still have to run while us not knowing it, so just shut up.
	return
}
