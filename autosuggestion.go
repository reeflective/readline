package readline

// autosuggestHistory returns the remainder of a
// history line if one matches the current input line.
func (rl *Instance) autosuggestHistory(line []rune) {
	rl.histSuggested = make([]rune, 0)

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
		if len(histline) <= len(line) {
			continue
		}

		// Or if not fully matching
		match := false
		for i, char := range line {
			if byte(char) == histline[i] {
				match = true
			} else {
				match = false
				break
			}
		}

		// If the line fully matches, we have our suggestion
		if match {
			rl.histSuggested = append(rl.histSuggested, []rune(histline[len(line):])...)
			return
		}
	}
}

// historyAutosuggestInsert inserts the currently autosuggested history line.
// This does nothing if history autosuggestion is not enabled in the config.
func (rl *Instance) historyAutosuggestInsert() {
	if !rl.config.HistoryAutosuggest {
		return
	}

	if len(rl.histSuggested) == 0 {
		return
	}

	rl.undoHistoryAppend()

	rl.line = append(rl.line, rl.histSuggested...)
	rl.pos = len(rl.line) - 1
	rl.histSuggested = []rune{}
}
