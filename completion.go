package readline

func (rl *Instance) completionWidgets() baseWidgets {
	return map[string]func(){
		"expand-or-complete":         rl.expandOrComplete, // DONE
		"expand-or-complete-prefix":  rl.expandOrCompletePrefix,
		"menu-complete":              rl.menuComplete, // DONE
		"complete-word":              rl.completeWord, // DONE
		"menu-expand-or-complete":    rl.menuExpandOrComplete,
		"reverse-menu-complete":      rl.reverseMenuComplete,   // DONE
		"accept-and-menu-complete":   rl.acceptAndMenuComplete, // DONE
		"expand-word":                rl.expandWord,
		"list-choices":               rl.listChoices,         // DONE
		"vi-registers-complete":      rl.viRegistersComplete, // DONE
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

func (rl *Instance) noCompletions() bool {
	for _, group := range rl.tcGroups {
		if len(group.Values) > 0 {
			return false
		}
	}

	return true
}

// initializeCompletions lets each group compute its completion strings,
// and compute its various coordinates/limits according to what it contains.
// Once done, adjust some start coordinates for some groups.
func (rl *Instance) initializeCompletions() {
	for i, group := range rl.tcGroups {
		// Let the group compute all its coordinates.
		group.init(rl)

		if i > 0 {
			group.tcPosY = 1

			if group.DisplayType == TabDisplayGrid {
				group.tcPosX = 1
			}
		}
	}
}

func (rl *Instance) resetTabCompletion() {
	if rl.local == menuselect {
		rl.local = ""
	}

	// rl.modeTabCompletion = false
	// rl.tabCompletionSelect = false
	rl.compConfirmWait = false

	rl.tcUsedY = 0
	// rl.modeTabFind = false
	// rl.modeAutoFind = false
	rl.tfLine = []rune{}

	// Reset tab highlighting
	if len(rl.tcGroups) > 0 {
		for _, g := range rl.tcGroups {
			g.isCurrent = false
		}
		rl.tcGroups[0].isCurrent = true
	}
}

// Check if we have a single completion candidate
func (rl *Instance) hasUniqueCandidate() bool {
	switch len(rl.tcGroups) {
	case 0:
		return false

	case 1:
		cur := rl.getCurrentGroup()
		if cur == nil {
			return false
		}

		return len(cur.Values) == 1
	default:
		var count int

	GROUPS:
		for _, group := range rl.tcGroups {
			for range group.Values {
				count++
				if count > 1 {
					break GROUPS
				}
			}
		}

		return count == 1
	}
}
