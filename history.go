package readline

import "github.com/reeflective/readline/internal/term"

// historyWidgets returns all history commands.
// Under each comment are gathered all commands related to the comment's
// subject. When there are two subgroups separated by an empty line, the
// second one comprises commands that are not legacy readline commands.
func (rl *Shell) historyWidgets() lineWidgets {
	widgets := map[string]func(){
		"accept-line":                            rl.acceptLine,
		"next-history":                           rl.downHistory, // down-history
		"previous-history":                       rl.upHistory,   // up-history
		"beginning-of-history":                   rl.beginningOfHistory,
		"end-of-history":                         rl.endOfHistory,
		"operate-and-get-next":                   rl.acceptLineAndDownHistory, // accept-line-and-down-history
		"fetch-history":                          rl.fetchHistory,
		"reverse-search-history":                 rl.historyIncrementalSearchForward,  // history-incremental-search-forward
		"forward-search-history":                 rl.historyIncrementalSearchBackward, // history-incremental-search-backward
		"non-incremental-forward-search-history": rl.nonIncrementalForwardSearchHistory,
		"non-incremental-reverse-search-history": rl.nonIncrementalReverseSearchHistory,
		"history-search-forward":                 rl.historySearchForward,
		"history-search-backward":                rl.historySearchBackward,
		"history-substring-search-forward":       rl.historySubstringSearchForward,
		"history-substring-search-backward":      rl.historySubstringSearchBackward,
		"yank-last-arg":                          rl.yankLastArg,
		"yank-nth-arg":                           rl.yankNthArg,

		"accept-and-hold":                   rl.acceptAndHold,
		"accept-and-infer-next-history":     rl.acceptAndInferNextHistory,
		"down-line-or-history":              rl.downLineOrHistory,
		"up-line-or-history":                rl.upLineOrHistory,
		"up-line-or-search":                 rl.upLineOrSearch,
		"down-line-or-search":               rl.downLineOrSearch,
		"infer-next-history":                rl.inferNextHistory,
		"beginning-of-buffer-or-history":    rl.beginningOfBufferOrHistory,
		"beginning-history-search-forward":  rl.beginningHistorySearchForward,
		"beginning-history-search-backward": rl.beginningHistorySearchBackward,
		"end-of-buffer-or-history":          rl.endOfBufferOrHistory,
		// "history-autosuggest-insert":          rl.historyAutosuggestInsert,
		"beginning-of-line-hist": rl.beginningOfLineHist,
		"end-of-line-hist":       rl.endOfLineHist,
	}

	return widgets
}

//
// Standard ----------------------------------------------------------------
//

func (rl *Shell) acceptLine() {
	// Without multiline support, we always return the line.
	if rl.AcceptMultiline == nil {
		rl.histories.Accept(false, false, nil)
		return
	}

	// Ask the caller if the line should be accepted as is.
	if rl.AcceptMultiline(*rl.line) {
		// Clear the tooltip prompt and clear hints/completions.
		rl.display.ClearHelpers()
		rl.prompt.RightClear(false)
		rl.display.CursorBelowLine()
		print(term.ClearScreenBelow)

		// Save the command line and accept it.
		rl.histories.Accept(false, false, nil)
		return
	}

	// If not, we should start editing another line,
	// and insert a newline where our cursor value is.
	// This has the nice advantage of being able to work
	// in multiline mode even in the middle of the buffer.
	rl.line.Insert(rl.cursor.Pos(), '\n')
}

func (rl *Shell) downHistory() {
	rl.undo.SkipSave()
	rl.histories.Walk(-1)
}

func (rl *Shell) upHistory() {
	rl.undo.SkipSave()
	rl.histories.Walk(1)
}

func (rl *Shell) beginningOfHistory() {
	rl.undo.SkipSave()

	history := rl.histories.Current()
	if history == nil {
		return
	}

	rl.histories.Walk(history.Len())
}

func (rl *Shell) endOfHistory() {
	history := rl.histories.Current()

	if history == nil {
		return
	}

	rl.histories.Walk(-history.Len() + 1)
}

func (rl *Shell) acceptLineAndDownHistory() {
	// rl.inferLine = true // The next loop will retrieve a line by histPos.
	// rl.acceptLine()
}

func (rl *Shell) fetchHistory() {}

func (rl *Shell) historyIncrementalSearchForward() {
	rl.undo.SkipSave()

	// Start history completion without matching against the current line.
	// rl.historyCompletion(true, false)
	//
	// // And only enter isearch mode when we have some completions: if we
	// // don't, we either exhausted our history sources, or don't have comps.
	// if rl.local == menuselect {
	// 	rl.enterIsearchMode()
	// }
}

func (rl *Shell) historyIncrementalSearchBackward() {
	rl.undo.SkipSave()

	// Start history completion without matching against the current line.
	// rl.historyCompletion(false, false)
	//
	// // And only enter isearch mode when we have some completions: if we
	// // don't, we either exhausted our history sources, or don't have comps.
	// if rl.local == menuselect {
	// 	rl.enterIsearchMode()
	// }
}

func (rl *Shell) nonIncrementalForwardSearchHistory() {}
func (rl *Shell) nonIncrementalReverseSearchHistory() {}

func (rl *Shell) historySearchForward() {
	rl.undo.SkipSave()

	// And either roll to the next history source, or
	// directly generate completions for the target history.
	//
	// Set the tab completion prefix as a filtering
	// mechanism here: will be updated by the comps anyway.
	// rl.historyCompletion(true, true)
}

func (rl *Shell) historySearchBackward() {
	rl.undo.SkipSave()

	// And either roll to the next history source, or
	// directly generate completions for the target history.
	//
	// Set the tab completion prefix as a filtering
	// mechanism here: will be updated by the comps anyway.
	// rl.historyCompletion(false, true)
}

func (rl *Shell) historySubstringSearchForward()  {}
func (rl *Shell) historySubstringSearchBackward() {}

func (rl *Shell) yankLastArg() {}
func (rl *Shell) yankNthArg()  {}

//
// Added -------------------------------------------------------------------
//

func (rl *Shell) acceptAndHold() {}

func (rl *Shell) acceptAndInferNextHistory() {
	// rl.inferLine = true // The next loop will retrieve a line.
	// rl.histPos = 0      // And will find it by trying to match one.
	// rl.acceptLine()
}

func (rl *Shell) downLineOrHistory() {
	rl.undo.SkipSave()
	switch {
	// case rl.hpos < rl.numLines()-1:
	// 	rl.downLine()
	default:
		rl.histories.Walk(-1)
	}
}

func (rl *Shell) upLineOrHistory() {
	rl.undo.SkipSave()
	switch {
	// case rl.hpos > 0:
	// 	rl.upLine()
	default:
		rl.histories.Walk(1)
	}
}

func (rl *Shell) upLineOrSearch() {
	rl.undo.SkipSave()
	switch {
	// case rl.hpos > 0:
	// 	rl.upLine()
	default:
		rl.historySearchBackward()
	}
}

func (rl *Shell) downLineOrSearch() {
	rl.undo.SkipSave()
	switch {
	// case rl.hpos < rl.numLines()-1:
	// 	rl.upLine()
	default:
		rl.historySearchForward()
	}
}

func (rl *Shell) inferNextHistory() {
	rl.undo.SkipSave()
	rl.histories.InferNext()
}

func (rl *Shell) beginningOfBufferOrHistory() {
	rl.undo.SkipSave()

	if rl.cursor.Pos() > 0 {
		rl.cursor.Set(0)
		return
	}

	rl.beginningOfHistory()
}

func (rl *Shell) endOfBufferOrHistory() {
	rl.undo.SkipSave()

	if rl.cursor.Pos() < rl.line.Len()-1 {
		rl.cursor.Set(rl.line.Len())
		return
	}

	rl.endOfHistory()
}

func (rl *Shell) beginningOfLineHist() {
	rl.undo.SkipSave()

	switch {
	// case rl.pos <= 0:
	// 	rl.beginningOfLine()
	default:
		rl.histories.Walk(1)
	}
}

func (rl *Shell) endOfLineHist() {
	rl.undo.SkipSave()

	switch {
	// case rl.cursor.Pos() < len(rl.line)-1:
	// 	rl.endOfLine()
	default:
		rl.histories.Walk(-1)
	}
}

func (rl *Shell) beginningHistorySearchBackward() {
	// rl.historySearchLine(false)
}

func (rl *Shell) beginningHistorySearchForward() {
	// rl.historySearchLine(true)
}

func (rl *Shell) lineCarriageReturn() {
	// rl.histSuggested = []rune{}
	//
	// // Ask the caller if the line should be accepted as is.
	// if rl.AcceptMultiline(rl.lineCompleted()) {
	// 	// Clear the tooltip prompt if any,
	// 	// then go down and clear hints/completions.
	// 	rl.moveToLineEnd()
	// 	rl.Prompt.clearRprompt(rl, false)
	// 	print("\r\n")
	// 	print(seqClearScreenBelow)
	//
	// 	// Save the command line and accept it.
	// 	rl.writeHistoryLine()
	// 	rl.accepted = true
	// 	return
	// }
	//
	// // If not, we should start editing another line,
	// // and insert a newline where our cursor value is.
	// // This has the nice advantage of being able to work
	// // in multiline mode even in the middle of the buffer.
	// rl.lineInsert([]rune{'\n'})
}
