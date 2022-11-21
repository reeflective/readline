package readline

import (
	"fmt"
	"os"
	"strings"
)

func (rl *Instance) inputMenuMove(r []rune) (ret bool) {
	switch string(r) {

	case seqShiftTab:
		if rl.modeTabCompletion && !rl.compConfirmWait {
			rl.tabCompletionReverse = true
			rl.moveTabCompletionHighlight(-1, 0)
			rl.updateVirtualComp()
			rl.tabCompletionReverse = false
			rl.renderHelpers()
			rl.viUndoSkipAppend = true
			return true
		}

	case seqUp:
		if rl.modeTabCompletion {
			rl.tabCompletionSelect = true
			rl.tabCompletionReverse = true
			rl.moveTabCompletionHighlight(-1, 0)
			rl.updateVirtualComp()
			rl.tabCompletionReverse = false
			rl.renderHelpers()
			return true
		}
		rl.mainHist = true
		rl.walkHistory(1)

	case seqDown:
		if rl.modeTabCompletion {
			rl.tabCompletionSelect = true
			rl.moveTabCompletionHighlight(1, 0)
			rl.updateVirtualComp()
			rl.renderHelpers()
			return true
		}
		rl.mainHist = true
		rl.walkHistory(-1)

	case seqForwards:
		if rl.modeTabCompletion {
			rl.tabCompletionSelect = true
			rl.moveTabCompletionHighlight(1, 0)
			rl.updateVirtualComp()
			rl.renderHelpers()
			return true
		}
		if (rl.modeViMode == vimInsert && rl.pos < len(rl.line)) ||
			(rl.modeViMode != vimInsert && rl.pos < len(rl.line)-1) {
			moveCursorForwards(1)
			rl.pos++
		}
		rl.updateHelpers()
		rl.viUndoSkipAppend = true

	case seqBackwards:
		if rl.modeTabCompletion {
			rl.tabCompletionSelect = true
			rl.tabCompletionReverse = true
			rl.moveTabCompletionHighlight(-1, 0)
			rl.updateVirtualComp()
			rl.tabCompletionReverse = false
			rl.renderHelpers()
			return true
		}
		if rl.pos > 0 {
			moveCursorBackwards(1)
			rl.pos--
		}
		rl.viUndoSkipAppend = true
		rl.updateHelpers()
	}

	return
}

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
		if rl.InputMode == Vim {
			rl.viEscape(r)
			return true
		}

		// This refreshed and actually prints the new Vim status
		// if we have indeed change the Vim mode.
		rl.clearHelpers()
		rl.renderHelpers()
	}

	return
}

func (rl *Instance) inputLineMove(r []rune) (ret bool) {
	switch string(r) {
	case seqCtrlLeftArrow:
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
		rl.updateHelpers()
		return true
	case seqCtrlRightArrow:
		rl.moveCursorByAdjust(rl.viJumpW(tokeniseLine))
		rl.updateHelpers()
		return true

	case seqDelete:
		if rl.modeTabFind {
			rl.backspaceTabFind()
		} else {
			rl.deleteBackspace()
		}
	case seqHome, seqHomeSc:
		if rl.modeTabCompletion {
			return true
		}
		moveCursorBackwards(rl.pos)
		rl.pos = 0
		rl.viUndoSkipAppend = true

	case seqEnd, seqEndSc:
		if rl.modeTabCompletion {
			return true
		}
		moveCursorForwards(len(rl.line) - rl.pos)
		rl.pos = len(rl.line)
		rl.viUndoSkipAppend = true

	case seqAltR:
		// TODO: Same here, that is a completion helper, should not be here.
		rl.resetVirtualComp(false)
		// For some modes only, if we are in vim Keys mode,
		// we toogle back to insert mode. For others, we return
		// without getting the completions.
		if rl.modeViMode != vimInsert {
			rl.modeViMode = vimInsert
		}

		rl.mainHist = false // true before
		rl.searchMode = HistoryFind
		rl.modeAutoFind = true
		rl.modeTabCompletion = true

		rl.modeTabFind = true
		rl.updateTabFind([]rune{})
		rl.viUndoSkipAppend = true
	}

	return
}

// inputInsertKey is the last helper that can be caught in the key dispatcher
// process, and it will either use this key as an action modifier (Vim) or input
// it into the current shell line.
func (rl *Instance) inputInsertKey(r []rune) {
	if rl.modeTabFind {
		return
	}

	// alt+numeric append / delete
	if len(r) == 2 && '1' <= r[1] && r[1] <= '9' {
		if rl.modeViMode == vimDelete {
			rl.viDelete(r[1])
			return
		}

		line, err := rl.mainHistory.GetLine(rl.mainHistory.Len() - 1)
		if err != nil {
			return
		}
		if !rl.mainHist {
			line, err = rl.altHistory.GetLine(rl.altHistory.Len() - 1)
			if err != nil {
				return
			}
		}

		tokens, _, _ := tokeniseSplitSpaces([]rune(line), 0)
		pos := int(r[1]) - 48 // convert ASCII to integer
		if pos > len(tokens) {
			return
		}
		rl.insert([]rune(tokens[pos-1]))

		return
	}

	// The character has been inserted as a buffer, or caught
	// as an action modifier, so we don't add it to our undo buffer.
	rl.viUndoSkipAppend = true
}

func (rl *Instance) carriageReturn() {
	rl.clearHelpers()
	print("\r\n")
	if rl.HistoryAutoWrite {
		var err error

		// Main history
		if rl.mainHistory != nil {
			rl.histPos, err = rl.mainHistory.Write(string(rl.line))
			if err != nil {
				print(err.Error() + "\r\n")
			}
		}
		// Alternative history
		if rl.altHistory != nil {
			rl.histPos, err = rl.altHistory.Write(string(rl.line))
			if err != nil {
				print(err.Error() + "\r\n")
			}
		}
	}
}

func (rl *Instance) clearScreen() {
	print(seqClearScreen)
	print(seqCursorTopLeft)
	if rl.Multiline {
		// TODO: here rander prompt in function correctly, all prompts.
		fmt.Println(rl.mainPrompt)
	}
	print(seqClearScreenBelow)

	rl.resetHintText()
	rl.getHintText()
	rl.renderHelpers()
}

// initLine is ran once at the beginning of an instance start.
func (rl *Instance) initLine() {
	rl.line = []rune{}
	rl.currentComp = []rune{} // No virtual completion yet
	rl.lineComp = []rune{}    // So no virtual line either
	rl.modeViMode = vimInsert
	rl.pos = 0
	rl.posY = 0
}

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

// When the DelayedSyntaxWorker gives us a new line, we need to check if there
// is any processing to be made, that all lines match in terms of content.
func (rl *Instance) updateLine(line []rune) {
	if len(rl.currentComp) > 0 {
	} else {
		rl.line = line
	}

	rl.renderHelpers()
}

// getLine - In many places we need the current line input. We either return the real line,
// or the one that includes the current completion candidate, if there is any.
func (rl *Instance) getLine() []rune {
	if len(rl.currentComp) > 0 {
		return rl.lineComp
	}
	return rl.line
}

// echo - refresh the current input line, either virtually completed or not.
// also renders the current completions and hints. To be noted, the updateReferences()
// function is only ever called once, and after having moved back to prompt position
// and having printed the line: this is so that at any moment, everyone has the good
// values for moving around, synchronized with the update input line.
func (rl *Instance) echo() {
	// Then we print the prompt, and the line,
	switch {
	case rl.PasswordMask != 0:
	case rl.PasswordMask > 0:
		print(strings.Repeat(string(rl.PasswordMask), len(rl.line)) + " ")

	default:
		// Go back to prompt position, and clear everything below
		moveCursorBackwards(GetTermWidth())
		moveCursorUp(rl.posY)
		print(seqClearScreenBelow)

		// Print the prompt
		print(string(rl.realPrompt))

		// Assemble the line, taking virtual completions into account
		var line []rune
		if len(rl.currentComp) > 0 {
			line = rl.lineComp
		} else {
			line = rl.line
		}

		// Print the input line with optional syntax highlighting
		if rl.SyntaxHighlighter != nil {
			print(rl.SyntaxHighlighter(line) + " ")
		} else {
			print(string(line) + " ")
		}
	}

	// Update references with new coordinates only now, because
	// the new line may be longer/shorter than the previous one.
	rl.updateReferences()

	// Go back to the current cursor position, with new coordinates
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.fullY)
	moveCursorDown(rl.posY)
	moveCursorForwards(rl.posX)
}

func (rl *Instance) insert(r []rune) {
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
		r := append(r, rl.line[rl.pos:]...)
		rl.line = append(rl.line[:rl.pos], r...)

	// We are at the end of the input line
	case rl.pos == len(rl.line):
		rl.line = append(rl.line, r...)
	}

	rl.pos += len(r)

	// This should also update the rl.pos
	rl.updateHelpers()
}

func (rl *Instance) deleteX() {
	switch {
	case len(rl.line) == 0:
		return
	case rl.pos == 0:
		rl.line = rl.line[1:]
	case rl.pos > len(rl.line):
		rl.pos = len(rl.line)
	case rl.pos == len(rl.line):
		rl.pos--
		rl.line = rl.line[:rl.pos]
	default:
		rl.line = append(rl.line[:rl.pos], rl.line[rl.pos+1:]...)
	}

	rl.updateHelpers()
}

func (rl *Instance) deleteBackspace() {
	switch {
	case len(rl.line) == 0:
		return
	case rl.pos == 0:
		rl.line = rl.line[1:]
	case rl.pos > len(rl.line):
		rl.backspace() // There is an infite loop going on here...
	case rl.pos == len(rl.line):
		rl.pos--
		rl.line = rl.line[:rl.pos]
	default:
		rl.pos--
		rl.line = append(rl.line[:rl.pos], rl.line[rl.pos+1:]...)
	}

	rl.updateHelpers()
}

func (rl *Instance) clearLine() {
	if len(rl.line) == 0 {
		return
	}

	// We need to go back to prompt
	moveCursorUp(rl.posY)
	moveCursorBackwards(GetTermWidth())
	moveCursorForwards(rl.promptLen)

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
}

func (rl *Instance) deleteToBeginning() {
	rl.resetVirtualComp(false)
	// Keep the line length up until the cursor
	rl.line = rl.line[rl.pos:]
	rl.pos = 0
}

// handleKeyPress is in charge of executing the handler that is register for a given keypress.
func (rl *Instance) handleKeyPress(s string) (done, mustReturn bool, val string, err error) {
	rl.clearHelpers()

	ret := rl.evtKeyPress[s](s, rl.line, rl.pos)

	rl.clearLine()
	rl.line = append(ret.NewLine, []rune{}...)
	rl.updateHelpers() // rl.echo
	rl.pos = ret.NewPos

	if ret.ClearHelpers {
		rl.resetHelpers()
	} else {
		rl.updateHelpers()
	}

	if len(ret.HintText) > 0 {
		rl.hintText = ret.HintText
		rl.clearHelpers()
		rl.renderHelpers()
	}
	if !ret.ForwardKey {
		done = true

		return
	}

	if ret.CloseReadline {
		rl.clearHelpers()
		mustReturn = true
		val = string(rl.line)

		return
	}

	return
}
