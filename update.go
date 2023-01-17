package readline

// init gathers all steps to perform at the beginning of readline loop.
func (rl *Instance) init() {
	rl.lineInit()        // Clear the line in most cases
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
		rl.autosuggestHistory(rl.lineCompleted())
	}
	rl.linePrint()

	// Go at beginning of first line after input remainder
	moveCursorDown(rl.fullY - rl.posY)
	moveCursorBackwards(GetTermWidth())

	// Print hints, check for any confirmation hint current.
	// (do not overwrite the confirmation question hint)
	rl.writeHintText()
	moveCursorBackwards(GetTermWidth())

	// Print completions and go back
	// to beginning of this line
	rl.printCompletions()
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.tcUsedY)

	// Compensate for hint printout
	switch {
	case len(rl.hint) > 0:
		moveCursorUp(rl.hintY)
	}

	// Go back to current cursor position
	moveCursorUp(rl.fullY - rl.posY)
	moveCursorForwards(rl.posX)
}
