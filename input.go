package readline

import (
	"os"
	"regexp"
)

// InputMode - The shell input mode
type InputMode string

const (
	// Vim - Vim editing mode
	Vim InputMode = "vim"
	// Emacs - Emacs (classic) editing mode
	Emacs InputMode = "emacs"
)

// readInput reads input from stdin and returns the result, length or an error.
func (rl *Instance) readInput() (b []byte, i int, err error) {
	rl.viUndoSkipAppend = false
	b = make([]byte, 1024)

	if !rl.skipStdinRead {
		i, err = os.Stdin.Read(b)
		if err != nil {
			return
		}
	}

	rl.skipStdinRead = false

	return
}

// readArgumentKey reads a key required by some (rare) widgets
// that directly read/need their argument/operator, without
// going though operator pending mode first.
func (rl *Instance) readArgumentKey() (key string, ret bool) {
	b, i, _ := rl.readInput()
	key = string(b[:i])

	// If the last key is a number, add to iterations instead,
	// and read another key input.
	numMatcher, _ := regexp.Compile(`^[1-9][0-9]*$`)
	for numMatcher.MatchString(string(key[len(key)-1])) {
		rl.viIteration += string(key[len(key)-1])

		b, i, _ = rl.readInput()
		key = string(b[:i])
	}

	if b[0] == charEscape {
		ret = true
	}

	return
}

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

// // Escape key generally aborts most completion/prompt helpers.
// func inputEsc(rl *Instance, b []byte, i int, r []rune) (read, ret bool, err error) {
// 	// If we were waiting for completion confirm, abort
// 	if rl.compConfirmWait {
// 		rl.compConfirmWait = false
// 		rl.renderHelpers()
// 	}
//
// 	// We always refresh the completion candidates, except if we are currently
// 	// cycling through them, because then it would just append the candidate.
// 	if rl.modeTabCompletion {
// 		if string(r[:i]) != seqShiftTab &&
// 			string(r[:i]) != seqArrowRight && string(r[:i]) != seqArrowLeft &&
// 			string(r[:i]) != seqArrowUp && string(r[:i]) != seqArrowDown {
// 			rl.resetVirtualComp(false)
// 		}
// 	}
//
// 	// Once helpers of all sorts are cleared, we can process
// 	// the change of input modes, etc.
// 	rl.escapeSeq(r[:i])
//
// 	return
// }

// func (rl *Instance) escapeSeq(r []rune) {
// 	// Test input movements
// 	if moved := rl.inputLineMove(r); moved {
// 		return
// 	}
//
// 	// Movement keys while not being inserting the stroke in a buffer.
// 	// Test input movements
// 	if moved := rl.inputMenuMove(r); moved {
// 		return
// 	}
//
// 	switch string(r) {
// 	case string(charEscape):
// 		if skip := rl.inputEscAll(r); skip {
// 			return
// 		}
// 		rl.viUndoSkipAppend = true
//
// 	case seqAltQuote:
// 		if rl.inputRegisters() {
// 			return
// 		}
// 	default:
// 		// rl.inputInsertKey(r)
// 	}
// }

// inputDispatch handles any key that is not a key press not bound to a core action.
// This means grossly not error codes/signals, completion/menu keys, and editing mode changes.
// func (rl *Instance) inputDispatch(r []rune, i int) (done, ret bool, val string, err error) {
// 	// If we were waiting for completion confirm, abort it and go on with our input.
// 	if rl.compConfirmWait {
// 		rl.resetVirtualComp(false)
// 		rl.compConfirmWait = false
// 		rl.renderHelpers()
// 	}
//
// 	// Completion modes hijack the text input, and cuts
// 	// the editor from using/interpreting the key.
// 	if rl.modeAutoFind && rl.searchMode == HistoryFind {
// 		rl.resetVirtualComp(true)
// 		rl.updateTabFind(r[:i])
// 		rl.updateVirtualComp()
// 		rl.renderHelpers()
// 		rl.viUndoSkipAppend = true
//
// 		done = true
//
// 		return
// 	}
//
// 	// Not sure that CompletionFind is useful, nor one of the other two
// 	if (rl.modeAutoFind || rl.modeTabFind) && rl.searchMode != RegisterFind {
// 		rl.resetVirtualComp(false)
// 		rl.updateTabFind(r[:i])
// 		rl.viUndoSkipAppend = true
// 	} else {
// 		rl.resetVirtualComp(false)
// 		rl.inputEditor(r[:i])
// 		if len(rl.multilineBuffer) > 0 && rl.modeViMode == vimKeys {
// 			rl.skipStdinRead = true
// 		}
// 	}
//
// 	// Notice we don't return done = true, since any action independent of our
// 	// while could still have to run while us not knowing it, so just shut up.
// 	return
// }

// inputMenuMove updates helpers when keys have an effect on them,
// in normal (non-insert) editing mode, so most of the time in things
// like completion menus.
// func (rl *Instance) inputMenuMove(r []rune) (ret bool) {
// 	switch string(r) {
//
// 	case seqShiftTab:
// 		if rl.modeTabCompletion && !rl.compConfirmWait {
// 			rl.tabCompletionReverse = true
// 			rl.moveTabCompletionHighlight(-1, 0)
// 			rl.updateVirtualComp()
// 			rl.tabCompletionReverse = false
// 			rl.renderHelpers()
// 			rl.viUndoSkipAppend = true
// 			return true
// 		}
//
// 	case seqArrowUp:
// 		if rl.modeTabCompletion {
// 			rl.tabCompletionSelect = true
// 			rl.tabCompletionReverse = true
// 			rl.moveTabCompletionHighlight(-1, 0)
// 			rl.updateVirtualComp()
// 			rl.tabCompletionReverse = false
// 			rl.renderHelpers()
// 			return true
// 		}
// 		rl.mainHist = true
// 		rl.walkHistory(1)
//
// 	case seqArrowDown:
// 		if rl.modeTabCompletion {
// 			rl.tabCompletionSelect = true
// 			rl.moveTabCompletionHighlight(1, 0)
// 			rl.updateVirtualComp()
// 			rl.renderHelpers()
// 			return true
// 		}
// 		rl.mainHist = true
// 		rl.walkHistory(-1)
//
// 	case seqArrowRight:
// 		if rl.modeTabCompletion {
// 			rl.tabCompletionSelect = true
// 			rl.moveTabCompletionHighlight(1, 0)
// 			rl.updateVirtualComp()
// 			rl.renderHelpers()
// 			return true
// 		}
// 		if (rl.modeViMode == vimInsert && rl.pos < len(rl.line)) ||
// 			(rl.modeViMode != vimInsert && rl.pos < len(rl.line)-1) {
// 			moveCursorForwards(1)
// 			rl.pos++
// 		}
// 		rl.updateHelpers()
// 		rl.viUndoSkipAppend = true
//
// 	case seqArrowLeft:
// 		if rl.modeTabCompletion {
// 			rl.tabCompletionSelect = true
// 			rl.tabCompletionReverse = true
// 			rl.moveTabCompletionHighlight(-1, 0)
// 			rl.updateVirtualComp()
// 			rl.tabCompletionReverse = false
// 			rl.renderHelpers()
// 			return true
// 		}
// 		if rl.pos > 0 {
// 			moveCursorBackwards(1)
// 			rl.pos--
// 		}
// 		rl.viUndoSkipAppend = true
// 		rl.updateHelpers()
// 	}
//
// 	return
// }

// inputEscAll is different from inputEsc in that this
// function is triggered when the shell is already in a
// non insert state, which happens in some completion modes,
// and in Vim mode.
func (rl *Instance) inputEscAll(r []rune) (ret bool) {
	switch {
	case rl.modeAutoFind:
		rl.resetTabFind()
		rl.clearHelpers()
		rl.resetTabCompletion()
		rl.resetHelpers()
		rl.renderHelpers()

	case rl.modeTabFind:
		rl.resetTabFind()
		rl.resetTabCompletion()

	case rl.modeTabCompletion:
		rl.clearHelpers()
		rl.resetTabCompletion()
		rl.renderHelpers()

	default:
		// No matter the input mode, we exit
		// any completion confirm if there's one.
		if rl.compConfirmWait {
			rl.compConfirmWait = false
			rl.clearHelpers()
			rl.renderHelpers()
			return true
		}

		// If we are in Vim mode, the escape key has its usage.
		// Otherwise in emacs mode the escape key does nothing.
		// if rl.InputMode == Vim {
		// 	rl.viEscape(r)
		// 	return true
		// }

		// This refreshed and actually prints the new Vim status
		// if we have indeed change the Vim mode.
		rl.clearHelpers()
		rl.renderHelpers()
	}

	return
}

// func (rl *Instance) inputLineMove(r []rune) (ret bool) {
// 	switch string(r) {
// 	case seqCtrlLeftArrow:
// 		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
// 		rl.updateHelpers()
// 		return true
// 	case seqCtrlRightArrow:
// 		rl.moveCursorByAdjust(rl.viJumpW(tokeniseLine))
// 		rl.updateHelpers()
// 		return true
//
// 	case seqDelete:
// 		if rl.modeTabFind {
// 			rl.backspaceTabFind()
// 		} else {
// 			rl.deleteBackspace()
// 		}
// 	case seqHome, seqHomeSc:
// 		if rl.modeTabCompletion {
// 			return true
// 		}
// 		moveCursorBackwards(rl.pos)
// 		rl.pos = 0
// 		rl.viUndoSkipAppend = true
//
// 	case seqEnd, seqEndSc:
// 		if rl.modeTabCompletion {
// 			return true
// 		}
// 		moveCursorForwards(len(rl.line) - rl.pos)
// 		rl.pos = len(rl.line)
// 		rl.viUndoSkipAppend = true
//
// 	case seqAltR:
// 		// TODO: Same here, that is a completion helper, should not be here.
// 		rl.resetVirtualComp(false)
// 		// For some modes only, if we are in vim Keys mode,
// 		// we toogle back to insert mode. For others, we return
// 		// without getting the completions.
// 		if rl.modeViMode != vimInsert {
// 			rl.modeViMode = vimInsert
// 		}
//
// 		rl.mainHist = false // true before
// 		rl.searchMode = HistoryFind
// 		rl.modeAutoFind = true
// 		rl.modeTabCompletion = true
//
// 		rl.modeTabFind = true
// 		rl.updateTabFind([]rune{})
// 		rl.viUndoSkipAppend = true
// 	}
//
// 	return
// }

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

// inputInsertKey is the last helper that can be caught in the key dispatcher
// process, and it will either use this key as an action modifier (Vim) or input
// it into the current shell line.
// func (rl *Instance) inputInsertKey(r []rune) {
// 	if rl.modeTabFind {
// 		return
// 	}
//
// 	// alt+numeric append / delete
// 	if len(r) == 2 && '1' <= r[1] && r[1] <= '9' {
// 		if rl.modeViMode == vimDelete {
// 			rl.viDelete(r[1])
// 			return
// 		}
//
// 		line, err := rl.mainHistory.GetLine(rl.mainHistory.Len() - 1)
// 		if err != nil {
// 			return
// 		}
// 		if !rl.mainHist {
// 			line, err = rl.altHistory.GetLine(rl.altHistory.Len() - 1)
// 			if err != nil {
// 				return
// 			}
// 		}
//
// 		tokens, _, _ := tokeniseSplitSpaces([]rune(line), 0)
// 		pos := int(r[1]) - 48 // convert ASCII to integer
// 		if pos > len(tokens) {
// 			return
// 		}
// 		rl.insert([]rune(tokens[pos-1]))
//
// 		return
// 	}
//
// 	// The character has been inserted as a buffer, or caught
// 	// as an action modifier, so we don't add it to our undo buffer.
// 	rl.viUndoSkipAppend = true
// }
