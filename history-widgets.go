package readline

func (rl *Instance) historyWidgets() lineWidgets {
	widgets := map[string]widget{
		"down-line-or-history":                rl.downLineOrHistory,
		"up-line-or-history":                  rl.upLineOrHistory,
		"down-history":                        rl.downHistory,
		"up-history":                          rl.upHistory,
		"infer-next-history":                  rl.inferNextHistory,
		"beginning-of-buffer-or-history":      rl.beginningOfBufferOrHistory,
		"beginning-history-search-forward":    rl.historySearchForward,
		"beginning-history-search-backward":   rl.historySearchBackward,
		"end-of-buffer-or-history":            rl.endOfBufferOrHistory,
		"history-autosuggest-insert":          rl.historyAutosuggestInsert,
		"beginning-of-line-hist":              rl.beginningOfLineHist,
		"end-of-line-hist":                    rl.endOfLineHist,
		"beginning-of-history":                rl.beginningOfHistory,
		"end-of-history":                      rl.endOfHistory,
		"accept-and-infer-next-history":       rl.acceptAndInferNextHistory,
		"accept-line-and-down-history":        rl.acceptLineAndDownHistory,
		"history-search-forward":              rl.historySearchForward,
		"history-search-backward":             rl.historySearchBackward,
		"history-incremental-search-forward":  rl.historyIncrementalSearchForward,
		"history-incremental-search-backward": rl.historyIncrementalSearchBackward,
	}

	return widgets
}

func (rl *Instance) downHistory() {
	rl.skipUndoAppend()
	rl.walkHistory(-1)
}

func (rl *Instance) upHistory() {
	rl.skipUndoAppend()
	rl.walkHistory(1)
}

func (rl *Instance) downLineOrHistory() {
	rl.skipUndoAppend()
	switch {
	case rl.hpos < rl.numLines()-1:
		rl.cursorDownLine()
	default:
		rl.walkHistory(-1)
	}
}

func (rl *Instance) upLineOrHistory() {
	rl.skipUndoAppend()
	switch {
	case rl.hpos > 0:
		rl.cursorUpLine()
	default:
		rl.walkHistory(1)
	}
}

func (rl *Instance) inferNextHistory() {
	rl.skipUndoAppend()
	matchIndex := 0
	histSuggested := make([]rune, 0)

	// Work with correct history source
	rl.historySourcePos = 0
	history := rl.currentHistory()

	// Nothing happens if the history is nil or empty.
	if history == nil || history.Len() == 0 {
		return
	}

	for i := 1; i <= history.Len(); i++ {
		histline, err := history.GetLine(history.Len() - i)
		if err != nil {
			return
		}

		// If too short
		if len(histline) < len(rl.line) {
			continue
		}

		// Or if not fully matching
		match := false
		for i, char := range rl.line {
			if byte(char) == histline[i] {
				match = true
			} else {
				match = false
				break
			}
		}

		// If the line fully matches, we have our suggestion
		if match {
			matchIndex = history.Len() - i
			histSuggested = append(histSuggested, []rune(histline)...)
			break
		}
	}

	// If we have no match we return, or check for the next line.
	if (len(histSuggested) == 0 && matchIndex <= 0) || history.Len() <= matchIndex+1 {
		return
	}

	// Get the next history line
	nextLine, err := history.GetLine(matchIndex + 1)
	if err != nil {
		return
	}

	rl.line = []rune(nextLine)
	rl.pos = len(nextLine)
}

func (rl *Instance) beginningOfBufferOrHistory() {
	rl.skipUndoAppend()

	if rl.pos == 0 {
		rl.historySourcePos = 0
		history := rl.currentHistory()

		if history == nil {
			return
		}

		new, err := history.GetLine(0)
		if err != nil {
			rl.resetHelpers()
			print(rl.Prompt.primary)
			return
		}

		rl.lineClear()
		rl.line = []rune(new)
		rl.pos = len(rl.line)

		return
	}

	rl.beginningOfLine()
}

func (rl *Instance) endOfBufferOrHistory() {
	rl.skipUndoAppend()

	if rl.pos == len(rl.line) {
		rl.historySourcePos = 0
		history := rl.currentHistory()

		if history == nil {
			return
		}

		new, err := history.GetLine(history.Len() - 1)
		if err != nil {
			rl.resetHelpers()
			print(rl.Prompt.primary)
			return
		}

		rl.lineClear()
		rl.line = []rune(new)
		rl.pos = len(rl.line)
		return
	}

	rl.endOfLine()
}

func (rl *Instance) beginningOfLineHist() {
	rl.skipUndoAppend()

	switch {
	case rl.pos <= 0:
		rl.beginningOfLine()
	default:
		rl.walkHistory(1)
	}
}

func (rl *Instance) endOfLineHist() {
	rl.skipUndoAppend()

	switch {
	case rl.pos < len(rl.line)-1:
		rl.endOfLine()
	default:
		rl.walkHistory(-1)
	}
}

func (rl *Instance) beginningOfHistory() {
	rl.skipUndoAppend()
	history := rl.currentHistory()

	if history == nil {
		return
	}

	rl.walkHistory(history.Len())
}

func (rl *Instance) endOfHistory() {
	history := rl.currentHistory()

	if history == nil {
		return
	}

	rl.walkHistory(-history.Len() + 1)
}

func (rl *Instance) acceptAndInferNextHistory() {
	rl.inferLine = true // The next loop will retrieve a line.
	rl.histPos = 0      // And will find it by trying to match one.
	rl.acceptLine()
}

func (rl *Instance) acceptLineAndDownHistory() {
	rl.inferLine = true // The next loop will retrieve a line by histPos.
	rl.acceptLine()
}

func (rl *Instance) historySearchForward() {
	rl.skipUndoAppend()

	// And either roll to the next history source, or
	// directly generate completions for the target history.
	//
	// Set the tab completion prefix as a filtering
	// mechanism here: will be updated by the comps anyway.
	rl.historyCompletion(true, true)
}

func (rl *Instance) historySearchBackward() {
	rl.skipUndoAppend()

	// And either roll to the next history source, or
	// directly generate completions for the target history.
	//
	// Set the tab completion prefix as a filtering
	// mechanism here: will be updated by the comps anyway.
	rl.historyCompletion(false, true)
}

func (rl *Instance) historyIncrementalSearchForward() {
	rl.skipUndoAppend()

	// Start history completion without matching against the current line.
	rl.historyCompletion(true, false)

	// And only enter isearch mode when we have some completions: if we
	// don't, we either exhausted our history sources, or don't have comps.
	if rl.local == menuselect {
		rl.enterIsearchMode()
	}
}

func (rl *Instance) historyIncrementalSearchBackward() {
	rl.skipUndoAppend()

	// Start history completion without matching against the current line.
	rl.historyCompletion(false, false)

	// And only enter isearch mode when we have some completions: if we
	// don't, we either exhausted our history sources, or don't have comps.
	if rl.local == menuselect {
		rl.enterIsearchMode()
	}
}

func (rl *Instance) beginningHistorySearchBackward() {
	rl.historySearchLine(false)
}

func (rl *Instance) beginningHistorySearchForward() {
	rl.historySearchLine(true)
}
