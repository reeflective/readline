package readline

func (rl *Instance) completionWidgets() baseWidgets {
	return map[string]func(){
		"expand-or-complete":         rl.expandOrComplete,
		"expand-or-complete-prefix":  rl.expandOrCompletePrefix,
		"menu-complete":              rl.menuComplete,
		"complete-word":              rl.completeWord,
		"menu-expand-or-complete":    rl.menuExpandOrComplete,
		"reverse-menu-complete":      rl.reverseMenuComplete,
		"accept-and-menu-complete":   rl.acceptAndMenuComplete,
		"expand-word":                rl.expandWord,
		"list-choices":               rl.listChoices,
		"vi-registers-complete":      rl.viRegistersComplete,
		"incremental-search-history": rl.incrementalSearchHistory,
	}
}

func (rl *Instance) expandOrComplete() {
	switch rl.local {
	case isearch:
	case menuselect:
		rl.menuComplete()
	default:
		rl.startMenuComplete(rl.generateCompletions)

		// In autocomplete mode, we already have completions
		// printed, so we automatically move to the first comp.
		if rl.isAutoCompleting() && rl.local == menuselect {
			rl.menuComplete()
		}
	}
	// If too many completions and no yet confirmed, ask user for completion
	// comps, lines := rl.getCompletionCount()
	// if ((lines > GetTermLength()) || (lines > rl.MaxTabCompleterRows)) && !rl.compConfirmWait {
	//         sentence := fmt.Sprintf("%s show all %d completions (%d lines) ? tab to confirm",
	//                 FOREWHITE, comps, lines)
	//         rl.promptCompletionConfirm(sentence)
	//         continue
	// }
}

func (rl *Instance) expandOrCompletePrefix() {
}

func (rl *Instance) completeWord() {
	switch rl.local {
	case isearch:
	case menuselect:
		rl.menuComplete()
	default:
		rl.startMenuComplete(rl.generateCompletions)

		// In autocomplete mode, we already have completions
		// printed, so we automatically move to the first comp.
		if rl.isAutoCompleting() && rl.local == menuselect {
			rl.menuComplete()
		}
	}
}

func (rl *Instance) menuComplete() {
	rl.skipUndoAppend()

	// No completions are being printed yet, so simply generate the completions
	// as if we just request them without immediately selecting a candidate.
	if rl.local != menuselect && rl.local != isearch {
		rl.startMenuComplete(rl.generateCompletions)
	}

	// Some of the actions taken in the above switch might have exited
	// completions, and if that is the case, we should not do anything.
	if rl.local != menuselect && rl.local != isearch {
		return
	}

	// Else, select the next candidate.
	rl.moveCompletionSelection(1, 0)
	rl.updateVirtualComp()
}

func (rl *Instance) menuExpandOrComplete() {
}

func (rl *Instance) reverseMenuComplete() {
	rl.skipUndoAppend()

	// We don't do anything when not already completing.
	if rl.local != menuselect && rl.local != isearch {
		return
	}

	// Else, select the previous candidate.
	rl.moveCompletionSelection(-1, 0)
	rl.updateVirtualComp()
}

func (rl *Instance) acceptAndMenuComplete() {
	rl.skipUndoAppend()

	// We don't do anything when not already completing.
	if rl.local != menuselect && rl.local != isearch {
		return
	}

	// Also return if no candidate
	if rl.getCurrentCandidate() == "" {
		return
	}

	// First insert the current candidate
	rl.resetVirtualComp(false)

	// And cycle to the next one, without quiting our mode
	rl.moveCompletionSelection(1, 0)
	rl.updateVirtualComp()
}

func (rl *Instance) expandWord() {
}

func (rl *Instance) listChoices() {
	switch rl.local {
	case isearch:
	case menuselect:
		rl.resetVirtualComp(false)
	}

	rl.local = menuselect
	rl.compConfirmWait = false
	rl.skipUndoAppend()

	// Call the completer to produce
	// all possible completions.
	rl.generateCompletions()

	// Cancel completion mode if
	// we don't have any candidates.
	if rl.noCompletions() {
		rl.resetTabCompletion()
		return
	}

	// Let all groups compute their display/candidate strings
	// and coordinates, and do some adjustments where needed.
	rl.initializeCompletions()
}

func (rl *Instance) viRegistersComplete() {
	switch rl.local {
	case isearch:
	default:
		registerCompletion := func() {
			rl.tcGroups = rl.completeRegisters()
			if len(rl.tcGroups) == 0 {
				return
			}

			var groups []CompletionGroup
			for _, group := range rl.tcGroups {
				groups = append(groups, *group)
			}
			rl.tcGroups = checkNilItems(groups)
		}

		rl.startMenuComplete(registerCompletion)
	}
}

func (rl *Instance) incrementalSearchHistory() {
}

// startMenuComplete generates a completion menu with completions
// generated from a given completer, without selecting a candidate.
func (rl *Instance) startMenuComplete(completer func()) {
	rl.local = menuselect
	rl.compConfirmWait = false
	rl.skipUndoAppend()

	// Call the provided completer function
	// to produce all possible completions.
	completer()

	// Cancel completion mode if we don't have any candidates.
	if rl.noCompletions() {
		rl.resetTabCompletion()
		return
	}

	// Let all groups compute their display/candidate strings
	// and coordinates, and do some adjustments where needed.
	rl.initializeCompletions()

	// When there is only candidate, automatically insert it
	// and exit the completion mode.
	if rl.hasUniqueCandidate() {
		rl.undoSkipAppend = false
		rl.insertCandidate()
		rl.resetTabCompletion()
	}
}
