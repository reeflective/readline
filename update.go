package readline

// initHelpers is called once at the very beginning of a readline start.
func (rl *Instance) initHelpers() {
	rl.resetHintText()
	rl.resetTabCompletion()
	rl.completer = nil
	rl.getHintText()
}

// updateHelpers is a key part of the whole refresh process:
// it should coordinate reprinting the input line, any hints and completions
// and manage to get back to the current (computed) cursor coordinates
func (rl *Instance) updateHelpers() {
	rl.autoComplete()

	rl.getHintText()
	rl.clearHelpers()
	rl.renderHelpers()
}

// Update reference should be called only once in a "loop" (not Readline(), but key control loop)
func (rl *Instance) updateReferences() {
	// We always need to work with clean data,
	// since we will have incrementers all around
	rl.posX = 0
	rl.fullX = 0
	rl.posY = 0
	rl.fullY = 0

	if rl.pos < 0 {
		rl.pos = 0
	}

	var fullLine, cPosLine int
	if len(rl.comp) > 0 {
		fullLine = len(rl.compLine)
		cPosLine = len(rl.compLine[:rl.pos])
	} else {
		fullLine = len(rl.line)
		cPosLine = len(rl.line[:rl.pos])
	}

	// Adjust if we have an autosuggested history
	if len(rl.histSuggested) > 0 {
		fullLine = fullLine + len(rl.histSuggested)
	}

	// We need the X offset of the whole line
	toEndLine := rl.Prompt.inputAt + fullLine
	fullOffset := toEndLine / GetTermWidth()
	rl.fullY = fullOffset
	fullRest := toEndLine % GetTermWidth()
	rl.fullX = fullRest

	// Use rl.pos value to get the offset to go TO/FROM the CURRENT POSITION
	lineToCursorPos := rl.Prompt.inputAt + cPosLine
	offsetToCursor := lineToCursorPos / GetTermWidth()
	cPosRest := lineToCursorPos % GetTermWidth()

	// If we are at the end of line
	if fullLine == rl.pos {
		rl.posY = fullOffset

		if fullRest == 0 {
			rl.posX = 0
		} else if fullRest > 0 {
			rl.posX = fullRest
		}
	} else if rl.pos < fullLine {
		// If we are somewhere in the middle of the line
		rl.posY = offsetToCursor

		if cPosRest == 0 {
		} else if cPosRest > 0 {
			rl.posX = cPosRest
		}
	}
}

func (rl *Instance) resetHelpers() {
	rl.resetHintText()
	rl.resetTabCompletion()
}

// clearHelpers - Clears everything: prompt, input, hints & comps,
// and comes back at the prompt.
func (rl *Instance) clearHelpers() {
	// Now go down to the last line of input
	moveCursorDown(rl.fullY - rl.posY)
	moveCursorBackwards(rl.posX)
	moveCursorForwards(rl.fullX)

	// Clear everything below
	print(seqClearScreenBelow)

	// Go back to current cursor position
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.fullY - rl.posY)
	moveCursorForwards(rl.posX)
}

// renderHelpers - prints all components (prompt, line, hints & comps)
// and replaces the cursor to its current position. This function never
// computes or refreshes any value, except from inside the echo function.
func (rl *Instance) renderHelpers() {
	// Optional, because neutral on placement
	rl.printLine()

	// Go at beginning of first line after input remainder
	moveCursorDown(rl.fullY - rl.posY)
	moveCursorBackwards(GetTermWidth())

	// Print hints, check for any confirmation hint current.
	// (do not overwrite the confirmation question hint)
	if !rl.compConfirmWait {
		if len(rl.hintText) > 0 {
			print("\n")
		}
		rl.writeHintText()
		moveCursorBackwards(GetTermWidth())

		// Print completions and go back to beginning of this line
		print("\n")
		rl.printCompletions()
		moveCursorBackwards(GetTermWidth())
		moveCursorUp(rl.tcUsedY)
	}

	// If we are still waiting for the user to confirm too long completions
	// Immediately refresh the hints
	if rl.compConfirmWait {
		print("\n")
		rl.writeHintText()
		rl.getHintText()
		moveCursorBackwards(GetTermWidth())
	}

	// Anyway, compensate for hint printout
	if len(rl.hintText) > 0 {
		moveCursorUp(rl.hintY)
	} else if !rl.compConfirmWait {
		moveCursorUp(1)
	} else if rl.compConfirmWait {
		moveCursorUp(1)
	}

	// Go back to current cursor position
	moveCursorUp(rl.fullY - rl.posY)
	moveCursorForwards(rl.posX)
}
