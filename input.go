package readline

// InputMode - The shell input mode
type InputMode int

const (
	// Vim - Vim editing mode
	Vim = iota
	// Emacs - Emacs (classic) editing mode
	Emacs
)

func errorCtrlC(rl *Instance, b []byte, i int, r []rune) (read, ret bool, err error) {
	err = CtrlC

	if rl.modeTabCompletion {
		rl.resetVirtualComp(true)
		rl.resetHelpers()
		rl.renderHelpers()

		read = true
		return
	}
	rl.clearHelpers()

	ret = true
	return
}

func errorEOF(rl *Instance, b []byte, i int, r []rune) (read, ret bool, err error) {
	rl.clearHelpers()
	ret = true
	err = EOF

	return
}

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

// Escape key generally aborts most completion/prompt helpers.
func inputEsc(rl *Instance, b []byte, i int, r []rune) (read, ret bool, err error) {
	// If we were waiting for completion confirm, abort
	if rl.compConfirmWait {
		rl.compConfirmWait = false
		rl.renderHelpers()
	}

	// We always refresh the completion candidates, except if we are currently
	// cycling through them, because then it would just append the candidate.
	if rl.modeTabCompletion {
		if string(r[:i]) != seqShiftTab &&
			string(r[:i]) != seqArrowRight && string(r[:i]) != seqArrowLeft &&
			string(r[:i]) != seqArrowUp && string(r[:i]) != seqArrowDown {
			rl.resetVirtualComp(false)
		}
	}

	// Once helpers of all sorts are cleared, we can process
	// the change of input modes, etc.
	rl.escapeSeq(r[:i])

	return
}

// inputEditor is an unexported function used to determine what mode of text
// entry readline is currently configured for and then update the line entries
// accordingly.
func (rl *Instance) inputEditor(r []rune) {
	switch rl.modeViMode {
	case vimKeys:
		rl.vi(r[0])
		rl.refreshVimStatus()

	case vimDelete:
		rl.viDelete(r[0])
		rl.refreshVimStatus()

	case vimReplaceOnce:
		rl.modeViMode = vimKeys
		rl.deleteX()
		rl.insert([]rune{r[0]})
		rl.refreshVimStatus()

	case vimReplaceMany:
		for _, char := range r {
			rl.deleteX()
			rl.insert([]rune{char})
		}
		rl.refreshVimStatus()

	default:
		// For some reason Ctrl+k messes with the input line, so ignore it.
		if r[0] == 11 {
			return
		}
		// We reset the history nav counter each time we come here:
		// We don't need it when inserting text.
		rl.histNavIdx = 0
		rl.insert(r)
	}

	if len(rl.multilineSplit) == 0 {
		rl.syntaxCompletion()
	}
}

func (rl *Instance) escapeSeq(r []rune) {
	// Test input movements
	if moved := rl.inputLineMove(r); moved {
		return
	}

	// Movement keys while not being inserting the stroke in a buffer.
	// Test input movements
	if moved := rl.inputMenuMove(r); moved {
		return
	}

	switch string(r) {
	case string(charEscape):
		if skip := rl.inputEscAll(r); skip {
			return
		}
		rl.viUndoSkipAppend = true

	case seqAltQuote:
		if rl.inputRegisters() {
			return
		}
	default:
		rl.inputInsertKey(r)
	}
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
	if (rl.modeAutoFind || rl.modeTabFind) && rl.searchMode != RegisterFind {
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

// func (rl *Instance) inputBackspace() (done bool) {
// 	// When currently in history completion, we refresh and automatically
// 	// insert the first (filtered) candidate, virtually
// 	if rl.modeAutoFind && rl.searchMode == HistoryFind {
// 		rl.resetVirtualComp(true)
// 		rl.backspaceTabFind()
//
// 		// Then update the printing, with the new candidate
// 		rl.updateVirtualComp()
// 		rl.renderHelpers()
// 		rl.viUndoSkipAppend = true
// 		return true
// 	}
//
// 	// Normal completion search does only refresh the search pattern and the comps
// 	if rl.modeTabFind || rl.modeAutoFind && rl.searchMode != RegisterFind {
// 		rl.backspaceTabFind()
// 		rl.viUndoSkipAppend = true
// 	} else {
// 		// Always cancel any virtual completion
// 		rl.resetVirtualComp(false)
//
// 		// Vim mode has different behaviors
// 		if rl.InputMode == Vim {
// 			if rl.modeViMode == vimInsert {
// 				rl.backspace()
// 			} else {
// 				rl.pos--
// 			}
// 			rl.renderHelpers()
// 			return true
// 		}
//
// 		// Else emacs deletes a character
// 		rl.backspace()
// 		rl.renderHelpers()
// 	}
//
// 	return
// }
