package readline

// init gathers all steps to perform at the beginning of readline loop.
func (rl *Instance) init() {
	rl.initLine()        // Clear the line in most cases
	rl.initHelpers()     // Prepare hints/completions
	rl.initHistory()     // Reset undo/history indexes in most cases.
	rl.initHistoryLine() // Retrieve a line from history when asked.
	rl.initKeymap()      // Verify key mappings and widget binds

	// The prompt reevaluates itself when its corresponding
	// functions are bound. Some of its components (PS1/RPROMPT)
	// are normally only computed here (until the next Readline loop),
	// but other components (PS2/tips) are computed more than once.
	// Also print the primary prompt (or most of it if multiline).
	rl.Prompt.init(rl)
}

// initHelpers is called once at the very beginning of a readline start.
func (rl *Instance) initHelpers() {
	rl.resetHintText()
	rl.resetCompletion()
	rl.completer = nil
	rl.getHintText()
}

// redisplay is a key part of the whole refresh process:
// it should coordinate reprinting the input line, any hints and completions
// and manage to get back to the current (computed) cursor coordinates.
func (rl *Instance) redisplay() {
	rl.Prompt.update(rl)
	rl.autoComplete()
	rl.getHintText()
	rl.clearHelpers()
	rl.renderHelpers()
}

func (rl *Instance) resetHelpers() {
	rl.resetHintText()
	rl.resetCompletion()
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
	if rl.config.HistoryAutosuggest {
		rl.autosuggestHistory(rl.getLineVirtual())
	}

	rl.printLine()

	// Go at beginning of first line after input remainder
	moveCursorDown(rl.fullY - rl.posY)
	moveCursorBackwards(GetTermWidth())

	// Print hints, check for any confirmation hint current.
	// (do not overwrite the confirmation question hint)
	if !rl.compConfirmWait {
		if len(rl.hint) > 0 {
			print("\n")
		}
		rl.writeHintText()
		moveCursorBackwards(GetTermWidth())

		// Print completions and go back
		// to beginning of this line
		print("\n")
		rl.printCompletions()
		moveCursorBackwards(GetTermWidth())
		moveCursorUp(rl.tcUsedY)
	}

	// If we are still waiting for the user to confirm
	// long completions, immediately refresh the hints.
	if rl.compConfirmWait {
		print("\n")
		rl.writeHintText()
		rl.getHintText()
		moveCursorBackwards(GetTermWidth())
	}

	// Anyway, compensate for hint printout
	switch {
	case len(rl.hint) > 0:
		moveCursorUp(rl.hintY)
	default:
		moveCursorUp(1)
	}

	// Go back to current cursor position
	moveCursorUp(rl.fullY - rl.posY)
	moveCursorForwards(rl.posX)
}

// Update reference should be called only once in a "loop" (not Readline(), but key control loop).
func (rl *Instance) computeCoordinates() {
	// We always need to work with clean data,
	// since we will have increments all around
	rl.posX = 0
	rl.fullX = 0
	rl.posY = 0
	rl.fullY = 0

	if rl.pos < 0 {
		rl.pos = 0
	}

	// 1 - Get line lengths.
	bufLen, cpos, suggLen := rl.computeCoordinatesBuffer()

	// Y coordinates always consider the line with the history hint.
	hinted := rl.Prompt.inputAt + suggLen
	rl.fullY = hinted / GetTermWidth()

	// We need the X offset of the whole line
	endBuf := rl.Prompt.inputAt + bufLen
	posY := endBuf / GetTermWidth()
	restY := endBuf % GetTermWidth()
	rl.fullX = restY

	// Use rl.pos value to get the offset to go TO/FROM the CURRENT POSITION
	endPos := rl.Prompt.inputAt + cpos
	cposY := endPos / GetTermWidth()
	restCposY := endPos % GetTermWidth()

	// If we are at the end of line
	if bufLen == rl.pos {
		rl.posY = posY

		if restY == 0 {
			rl.posX = 0
		} else if restY > 0 {
			rl.posX = restY
		}
	} else if rl.pos < bufLen {
		// If we are somewhere in the middle of the line
		rl.posY = cposY

		if restCposY == 0 {
		} else if restCposY > 0 {
			rl.posX = restCposY
		}
	}
}

// returns len of line, len of line including history hint, and len to cursor.
func (rl *Instance) computeCoordinatesBuffer() (int, int, int) {
	var bufLen, cpos, suggLen int

	if len(rl.histSuggested) > 0 {
		suggLen += len(rl.histSuggested)
	}

	if len(rl.comp) > 0 {
		bufLen = len(rl.compLine)
		cpos = len(rl.compLine[:rl.pos])
	} else {
		bufLen = len(rl.line)
		cpos = len(rl.line[:rl.pos])
	}

	suggLen = bufLen
	if len(rl.histSuggested) > 0 {
		suggLen += len(rl.histSuggested)
	}

	return bufLen, cpos, suggLen
}
