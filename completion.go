package readline

import (
	"fmt"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/keymap"
)

func (rl *Shell) completionCommands() commands {
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

	rl.completer.Cancel(false, false)
	rl.keymaps.SetLocal(keymap.MenuSelect)
	rl.completer.GenerateWith(rl.commandCompletion)
}

func (rl *Shell) insertCompletions() {}

func (rl *Shell) menuComplete() {
	rl.undo.SkipSave()

	// No completions are being printed yet, so simply generate the completions
	// as if we just request them without immediately selecting a candidate.
	if !rl.completer.IsActive() {
		rl.startMenuComplete(rl.commandCompletion)
	} else {
		rl.keymaps.SetLocal(keymap.MenuSelect)
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

	rl.keymaps.SetLocal(keymap.MenuSelect)
	rl.completer.Select(-1, 0)
}

func (rl *Shell) menuCompleteNextTag() {
	rl.undo.SkipSave()

	if !rl.completer.IsActive() {
		return
	}

	rl.keymaps.SetLocal(keymap.MenuSelect)
	rl.completer.SelectTag(true)
}

func (rl *Shell) menuCompletePrevTag() {
	rl.undo.SkipSave()

	if !rl.completer.IsActive() {
		return
	}

	rl.keymaps.SetLocal(keymap.MenuSelect)
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

	if !rl.completer.IsActive() {
		rl.completer.GenerateWith(rl.commandCompletion)
	}

	rl.completer.IsearchStart("completions")
}

//
// Utilities --------------------------------------------------------------------------
//

// startMenuComplete generates a completion menu with completions
// generated from a given completer, without selecting a candidate.
func (rl *Shell) startMenuComplete(completer completion.Completer) {
	rl.undo.SkipSave()

	rl.keymaps.SetLocal(keymap.MenuSelect)
	rl.completer.GenerateWith(rl.commandCompletion)
}

func (rl *Shell) commandCompletion() completion.Values {
	if rl.Completer == nil {
		return completion.Values{}
	}

	line, cursor := rl.completer.Line()
	comps := rl.Completer(*line, cursor.Pos())

	return comps.convert()
}

func (rl *Shell) historyCompletion(forward, filterLine bool) {
	switch rl.keymaps.Local() {
	case keymap.MenuSelect, keymap.Isearch:
		// If we are currently completing the last
		// history source, cancel history completion.
		if rl.histories.OnLastSource() {
			rl.histories.Cycle(true)
			rl.completer.Cancel(true, true)
			rl.completer.Drop(true)
			rl.hint.Reset()
			rl.completer.IsearchStop()

			return
		}

		// Else complete the next history source.
		rl.histories.Cycle(true)

		fallthrough

	default:
		// Notify if we don't have history sources at all.
		if rl.histories.Current() == nil {
			rl.hint.Set(fmt.Sprintf("%s%s%s %s", color.Dim, color.FgRed, "No command history source", color.Reset))
			return
		}

		// Generate the completions with specified behavior.
		completer := func() completion.Values {
			return rl.histories.Complete(forward, filterLine)
		}

		rl.completer.GenerateWith(completer)
		rl.completer.IsearchStart(rl.histories.Name())
	}
}
