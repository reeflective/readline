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

// clearHelpers will go the end of the input line, and remove:
// - Any right prompt printed next to the last line of the buffer.
// - The suggested history line if any.
// - All hints.
// - All completions.
func (rl *Instance) clearHelpers() {
	rl.moveToLineEnd()
	print(seqClearScreenBelow)
	rl.moveFromLineEndToCursor()
}

// renderHelpers - prints all components (prompt, line, hints & comps)
// and replaces the cursor to its current position. This function never
// computes or refreshes any value, except from inside the echo function.
func (rl *Instance) renderHelpers() {
	if rl.config.HistoryAutosuggest {
		rl.autosuggestHistory(rl.lineCompleted())
	}
	rl.linePrint()

	// Go at beginning of the last line of input
	rl.moveToHintStart()

	// Print hints, check for any confirmation hint current.
	// (do not overwrite the confirmation question hint)
	rl.writeHintText()
	moveCursorBackwards(GetTermWidth())

	// Print completions and go back
	// to beginning of this line
	rl.printCompletions()

	// And move back to the last line of input, then to the cursor.
	rl.moveFromHelpersEndToHintStart()
	rl.moveFromLineEndToCursor()
}
