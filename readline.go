package readline

import (
	"os"
	"regexp"
)

var rxMultiline = regexp.MustCompile(`[\r\n]+`)

// Readline displays the readline prompt.
// It will return a string (user entered data) or an error.
func (rl *Instance) Readline() (string, error) {
	fd := int(os.Stdin.Fd())
	state, err := MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer Restore(fd, state)

	rl.initInput()
	rl.initPrompt()
	rl.initLine()
	rl.initHelpers()
	rl.initHistory()

	// Multisplit
	if len(rl.multilineSplit) > 0 {
		return rl.initMultiline()
	}

	// Finally, print any hints or completions
	// if the TabCompletion engines so desires
	rl.renderHelpers()

	// Start handling keystrokes. Classified by subject for most.
	for {
		// Read the input from stdin if any, and upon successfull
		// read, convert the input into runes for better scanning.
		b, i, readErr := rl.readInput()
		if readErr != nil {
			return "", err
		}

		r := []rune(string(b))

		// If the last input is a carriage return, process
		// according to configured multiline behavior.
		if isMultiline(r[:i]) || len(rl.multilineBuffer) > 0 {
			done, ret, val, err := rl.processMultiline(r, b, i)
			if ret {
				return val, err
			} else if done {
				continue
			}
		}

		// If we caught a key press for which a
		// handler is registered, execute it.
		s := string(r[:i])
		if rl.evtKeyPress[s] != nil {
			done, ret, val, err := rl.handleKeyPress(s)
			if ret {
				return val, err
			} else if done {
				continue
			}
		}

		// Ensure the completion system is in a sane
		// state before processing an input key.
		rl.ensureCompState()

		switch b[0] {
		// Root keypresses. ---------------------------------------------------
		case charEscape:
			rl.inputEsc(r, b, i)
		case charCtrlL:
			rl.clearScreen()

			// Error sequences ------------------------------------------------
		case charCtrlC:
			done, ret := rl.errorCtrlC()
			if ret {
				return "", CtrlC
			} else if done {
				continue
			}
		case charEOF:
			rl.clearHelpers()
			return "", EOF

		// Emacs bindings -----------------------------------------------------
		case charCtrlU:
			rl.deleteLine()
		case charCtrlW:
			if done := rl.deleteWord(); done {
				continue
			}
		case charCtrlY:
			rl.pasteDefaultRegister()
		case charCtrlE:
			if done := rl.goToInputEnd(); done {
				continue
			}
		case charCtrlA:
			if done := rl.goToInputBegin(); done {
				continue
			}

		// Special non-nil characters -----------------------------------------
		case '\r':
			fallthrough
		case '\n':
			done, ret, val, err := rl.inputEnter()
			if ret {
				return val, err
			} else if done {
				continue
			}

		case charBackspace, charBackspace2:
			if done := rl.inputBackspace(); done {
				continue
			}

		// Completion and history/menu helpers. -------------------------------
		case charCtrlR:
			rl.inputCompletionHelper(b, i)
		case charTab:
			done, ret, val, err := rl.inputCompletionTab(b, i)
			if ret {
				return val, err
			} else if done {
				continue
			}
		case charCtrlF:
			rl.inputCompletionFind()
		case charCtrlG:
			done, ret, val, err := rl.inputCompletionReset()
			if ret {
				return val, err
			} else if done {
				continue
			}
		default:
			done, ret, val, err := rl.inputDispatch(r, i)
			if ret {
				return val, err
			} else if done {
				continue
			}
		}

		// If no core helper has not caugth on the provided key,
		// neither completions or editors for inputing it in the
		// line, we store the key in our Undo history (Vim mode)
		rl.undoAppendHistory()
	}
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
