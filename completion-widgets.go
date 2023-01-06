package readline

func (rl *Instance) completionWidgets() lineWidgets {
	return map[string]widget{
		"expand-or-complete":        rl.expandOrComplete,
		"menu-complete":             rl.menuComplete,
		"complete-word":             rl.completeWord,
		"menu-expand-or-complete":   rl.menuExpandOrComplete,
		"reverse-menu-complete":     rl.reverseMenuComplete,
		"menu-complete-next-tag":    rl.menuCompleteNextTag,
		"menu-complete-prev-tag":    rl.menuCompletePrevTag,
		"accept-and-menu-complete":  rl.acceptAndMenuComplete,
		"list-choices":              rl.listChoices,
		"vi-registers-complete":     rl.viRegistersComplete,
		"menu-incremental-search":   rl.menuIncrementalSearch,
		"accept-completion-or-line": rl.acceptCompletionOrLine,
	}
}

func (rl *Instance) expandOrComplete() {
	switch rl.local {
	case menuselect, isearch:
		rl.menuComplete()
	default:
		if rl.completer != nil {
			rl.startMenuComplete(rl.completer)
		} else {
			rl.startMenuComplete(rl.normalCompletions)
		}

		// In autocomplete mode, we already have completions
		// printed, so we automatically move to the first comp.
		if rl.isAutoCompleting() && rl.local == menuselect {
			rl.menuComplete()
		}
	}
}

func (rl *Instance) completeWord() {
	switch rl.local {
	case menuselect, isearch:
		rl.menuComplete()
	default:
		rl.startMenuComplete(rl.normalCompletions)

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
	if rl.local != menuselect && rl.local != isearch && len(rl.histHint) == 0 {
		rl.startMenuComplete(rl.normalCompletions)
	}

	// Some of the actions taken in the above switch might have exited
	// completions, and if that is the case, we should not do anything.
	if rl.local != menuselect && rl.local != isearch && len(rl.histHint) == 0 {
		return
	}

	// Override the default move depending on the group
	cur := rl.currentGroup()
	if cur == nil {
		return
	}

	tabX, tabY := 1, 0

	if cur.aliased && rl.keys != seqArrowRight && rl.keys != seqArrowDown {
		tabX, tabY = 0, 1
	} else if rl.keys == seqArrowDown {
		tabX, tabY = 0, 1
	}

	// Else, select the next candidate.
	switch rl.keys {
	case seqArrowRight:
		rl.updateSelector(tabX, tabY)
	case seqArrowDown:
		rl.updateSelector(tabX, tabY)
	default:
		rl.updateSelector(tabX, tabY)
	}

	rl.updateVirtualComp()
}

func (rl *Instance) menuExpandOrComplete() {
	switch rl.local {
	case menuselect, isearch:
		rl.menuComplete()
	default:
		if rl.completer != nil {
			rl.startMenuComplete(rl.completer)
		} else {
			rl.startMenuComplete(rl.normalCompletions)
		}

		// In autocomplete mode, we already have completions
		// printed, so we automatically move to the first comp.
		if rl.isAutoCompleting() && rl.local == menuselect {
			rl.menuComplete()
		}
	}
}

func (rl *Instance) reverseMenuComplete() {
	rl.skipUndoAppend()

	// We don't do anything when not already completing.
	if rl.local != menuselect && rl.local != isearch {
		return
	}

	tabX, tabY := -1, 0

	// Override the default move depending on the group
	cur := rl.currentGroup()
	if cur.aliased && rl.keys != seqArrowLeft && rl.keys != seqArrowUp {
		tabX, tabY = 0, -1
	} else if rl.keys == seqArrowUp {
		tabX, tabY = 0, -1
	}

	// Else, select the previous candidate.
	switch rl.keys {
	case seqArrowLeft:
		rl.updateSelector(tabX, tabY)
	case seqArrowUp:
		rl.updateSelector(tabX, tabY)
	default:
		rl.updateSelector(tabX, tabY)
	}

	rl.updateVirtualComp()
}

func (rl *Instance) menuCompleteNextTag() {
	rl.skipUndoAppend()

	// We don't do anything when not already completing.
	if rl.local != menuselect && rl.local != isearch {
		return
	} else if len(rl.tcGroups) <= 1 {
		return
	}

	rl.cycleNextGroup()
	newGrp := rl.currentGroup()
	newGrp.firstCell()
}

func (rl *Instance) menuCompletePrevTag() {
	rl.skipUndoAppend()

	if rl.local != menuselect && rl.local != isearch {
		return
	} else if len(rl.tcGroups) <= 1 {
		return
	}

	rl.cyclePreviousGroup()
	newGrp := rl.currentGroup()
	newGrp.firstCell()
}

func (rl *Instance) acceptAndMenuComplete() {
	rl.skipUndoAppend()

	// We don't do anything when not already completing.
	if rl.local != menuselect && rl.local != isearch {
		return
	}

	// Also return if no candidate
	if rl.currentCandidate() == "" {
		return
	}

	// First insert the current candidate
	rl.resetVirtualComp(false)

	// And cycle to the next one, without quiting our mode
	rl.updateSelector(1, 0)
	rl.updateVirtualComp()
}

func (rl *Instance) listChoices() {
	rl.skipUndoAppend()

	switch rl.local {
	case menuselect, isearch:
		rl.resetVirtualComp(false)
	}

	rl.local = menuselect

	// Call the completer to produce
	// all possible completions.
	rl.normalCompletions()

	// Cancel completion mode if
	// we don't have any candidates.
	if rl.noCompletions() {
		rl.resetCompletion()
		return
	}
}

func (rl *Instance) viRegistersComplete() {
	rl.skipUndoAppend()

	switch rl.local {
	case isearch:
	default:
		rl.startMenuComplete(rl.registerCompletion)
	}
}

func (rl *Instance) menuIncrementalSearch() {
	rl.skipUndoAppend()

	switch rl.local {
	case isearch:
		fallthrough
	default:
		// First initialize completions.
		if rl.completer != nil {
			rl.startMenuComplete(rl.completer)
		}

		// Then enter the isearch mode, which updates
		// the hint line, and initializes other things.
		if rl.local == menuselect {
			rl.enterIsearchMode()
		}
	}
}

func (rl *Instance) acceptCompletionOrLine() {
	switch rl.local {
	case menuselect, isearch:
		// If we have a completion, simply accept this candidate
		comp := rl.currentCandidate()
		if comp != "" {
			rl.resetVirtualComp(false)
			rl.resetCompletion()

			return
		}

		// Or accept the line.
		fallthrough
	default:
		rl.carriageReturn()
		rl.accepted = true
	}
}
