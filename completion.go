package readline

import (
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/keymap"
)

func (rl *Shell) completionWidgets() lineWidgets {
	return map[string]func(){
		"complete":               rl.completeWord,        // complete-word
		"possible-completions":   rl.possibleCompletions, // list-choices
		"insert-completions":     rl.insertCompletions,
		"menu-complete":          rl.menuComplete,
		"menu-complete-backward": rl.menuCompleteBackward, // reverse-menu-complete
		"delete-char-or-list":    rl.deleteCharOrList,

		"expand-or-complete":        rl.expandOrComplete,
		"menu-expand-or-complete":   rl.menuExpandOrComplete,
		"menu-complete-next-tag":    rl.menuCompleteNextTag,
		"menu-complete-prev-tag":    rl.menuCompletePrevTag,
		"accept-and-menu-complete":  rl.acceptAndMenuComplete,
		"accept-completion-or-line": rl.acceptCompletionOrLine,
		"vi-registers-complete":     rl.viRegistersComplete,
		"menu-incremental-search":   rl.menuIncrementalSearch,
	}
}

//
// Commands ---------------------------------------------------------------------------
//

func (rl *Shell) completeWord() {
	rl.undo.SkipSave()

	// This completion function should attempt to insert the first
	// valid completion found, without printing the actual list.

	// switch rl.local {
	// case menuselect, isearch:
	// 	rl.menuComplete()
	// default:
	// 	// rl.
	// 	// rl.startMenuComplete(rl.normalCompletions)
	//
	// 	// In autocomplete mode, we already have completions
	// 	// printed, so we automatically move to the first comp.
	// 	if rl.isAutoCompleting() && rl.local == menuselect {
	// 		rl.menuComplete()
	// 	}
	// }
}

func (rl *Shell) possibleCompletions() {
	rl.undo.SkipSave()

	// switch rl.local {
	// case menuselect, isearch:
	// 	rl.resetVirtualComp(false)
	// }
	//
	// rl.local = menuselect
	//
	// // Call the completer to produce
	// // all possible completions.
	// rl.normalCompletions()
	//
	// // Cancel completion mode if
	// // we don't have any candidates.
	// if rl.noCompletions() {
	// 	rl.resetCompletion()
	// 	return
	// }

	rl.completer.Reset(false, false)
	rl.completer.GenerateWith(rl.normalCompletions)
}

func (rl *Shell) insertCompletions() {}

func (rl *Shell) menuComplete() {
	rl.undo.SkipSave()

	// No completions are being printed yet, so simply generate the completions
	// as if we just request them without immediately selecting a candidate.
	if !rl.completer.IsActive() {
		rl.completer.GenerateWith(rl.normalCompletions)
	} else {
		rl.completer.Select(1, 0)
	}

	// No completions are being printed yet, so simply generate the completions
	// as if we just request them without immediately selecting a candidate.
	// if rl.local != menuselect && rl.local != isearch && len(rl.histHint) == 0 {
	// 	rl.startMenuComplete(rl.normalCompletions)
	// }

	// Some of the actions taken in the above switch might have exited
	// completions, and if that is the case, we should not do anything.
	// if rl.local != menuselect && rl.local != isearch && len(rl.histHint) == 0 {
	// 	return
	// }
}

func (rl *Shell) deleteCharOrList() {
	switch {
	case rl.cursor.Pos() < rl.line.Len():
		rl.line.CutRune(rl.cursor.Pos())
	default:
		rl.possibleCompletions()
	}
}

func (rl *Shell) expandOrComplete() {
	// switch rl.local {
	// case menuselect, isearch:
	// 	rl.menuComplete()
	// default:
	// 	if rl.completer != nil {
	// 		rl.startMenuComplete(rl.completer)
	// 	} else {
	// 		rl.startMenuComplete(rl.normalCompletions)
	// 	}
	//
	// 	// In autocomplete mode, we already have completions
	// 	// printed, so we automatically move to the first comp.
	// 	if rl.isAutoCompleting() && rl.local == menuselect {
	// 		rl.menuComplete()
	// 	}
	// }
}

func (rl *Shell) menuExpandOrComplete() {
	// switch rl.local {
	// case menuselect, isearch:
	// 	rl.menuComplete()
	// default:
	// 	if rl.completer != nil {
	// 		rl.startMenuComplete(rl.completer)
	// 	} else {
	// 		rl.startMenuComplete(rl.normalCompletions)
	// 	}
	//
	// 	// In autocomplete mode, we already have completions
	// 	// printed, so we automatically move to the first comp.
	// 	if rl.isAutoCompleting() && rl.local == menuselect {
	// 		rl.menuComplete()
	// 	}
	// }
}

func (rl *Shell) menuCompleteBackward() {
	rl.undo.SkipSave()

	// We don't do anything when not already completing.
	if !rl.completer.IsActive() {
		return
	}

	rl.completer.Select(-1, 0)
}

func (rl *Shell) menuCompleteNextTag() {
	rl.undo.SkipSave()

	if !rl.completer.IsActive() {
		return
	}

	rl.completer.SelectTag(true)
}

func (rl *Shell) menuCompletePrevTag() {
	rl.undo.SkipSave()

	if !rl.completer.IsActive() {
		return
	}

	rl.completer.SelectTag(false)
}

func (rl *Shell) acceptAndMenuComplete() {
	rl.undo.SkipSave()

	// // We don't do anything when not already completing.
	// if rl.local != menuselect && rl.local != isearch {
	// 	return
	// }
	//
	// // Also return if no candidate
	// if rl.currentCandidate() == "" {
	// 	return
	// }
	//
	// // First insert the current candidate
	// rl.resetVirtualComp(false)
	//
	// // And cycle to the next one, without quiting our mode
	// rl.updateSelector(1, 0)
	// rl.updateVirtualComp()
}

func (rl *Shell) acceptCompletionOrLine() {
	// switch rl.local {
	// case menuselect, isearch:
	// 	// If we have a completion, simply accept this candidate
	// 	comp := rl.currentCandidate()
	// 	if comp != "" {
	// 		rl.resetVirtualComp(false)
	// 		rl.resetCompletion()
	//
	// 		return
	// 	}
	//
	// 	// Or accept the line.
	// 	fallthrough
	// default:
	// 	rl.lineCarriageReturn()
	// 	rl.accepted = true
	// }
}

func (rl *Shell) viRegistersComplete() {
	rl.keymaps.SetLocal(keymap.MenuSelect)
	rl.undo.SkipSave()

	registers := rl.buffers.Complete()
	rl.completer.Generate(registers)
}

func (rl *Shell) menuIncrementalSearch() {
	rl.undo.SkipSave()
	//
	// switch rl.local {
	// case isearch:
	// 	fallthrough
	// default:
	// 	// First initialize completions.
	// 	if rl.completer != nil {
	// 		rl.startMenuComplete(rl.completer)
	// 	}
	//
	// 	// Then enter the isearch mode, which updates
	// 	// the hint line, and initializes other things.
	// 	if rl.local == menuselect {
	// 		rl.enterIsearchMode()
	// 	}
	// }
}

//
// Utilities --------------------------------------------------------------------------
//

// startMenuComplete generates a completion menu with completions
// generated from a given completer, without selecting a candidate.
func (rl *Shell) startMenuComplete(completer func()) {
	rl.keymaps.SetLocal(keymap.MenuSelect)
	rl.undo.SkipSave()
	// rl.local = menuselect
	// rl.skipUndoAppend()
}

func (rl *Shell) normalCompletions() completion.Values {
	if rl.Completer == nil {
		return completion.Values{}
	}

	line, cursor := rl.completer.Line()
	comps := rl.Completer(*line, cursor.Pos())

	return comps.convert()
}

func (rl *Shell) historyCompletion(forward, filterLine bool) {
	// switch rl.local {
	// case menuselect, isearch:
	// 	// If we are currently completing the last
	// 	// history source, cancel history completion.
	// 	if rl.historySourcePos == len(rl.histories)-1 {
	// 		rl.histHint = []rune{}
	// 		rl.nextHistorySource()
	// 		rl.resetCompletion()
	// 		rl.completer = nil
	// 		rl.local = ""
	// 		rl.resetHintText()
	//
	// 		return
	// 	}
	//
	// 	// Else complete the next history source.
	// 	rl.nextHistorySource()
	//
	// 	fallthrough
	//
	// default:
	// 	// Notify if we don't have history sources at all.
	// 	if rl.currentHistory() == nil {
	// 		noHistory := fmt.Sprintf("%s%s%s %s", seqDim, seqFgRed, "No command history source", seqReset)
	// 		rl.histHint = []rune(noHistory)
	//
	// 		return
	// 	}
	//
	// 	// Generate the completions with specified behavior.
	// 	historyCompletion := func() {
	// 		rl.tcGroups = make([]*comps, 0)
	//
	// 		// Either against the current line or not.
	// 		if filterLine {
	// 			rl.tcPrefix = string(rl.line)
	// 		}
	//
	// 		comps := rl.completeHistory(forward)
	// 		comps = comps.DisplayList()
	// 		rl.groupCompletions(comps)
	// 		rl.setCompletionPrefix(comps)
	// 	}
	//
	// 	// Else, generate the completions.
	// 	rl.startMenuComplete(historyCompletion)
	// }
}
