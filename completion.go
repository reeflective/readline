package readline

import (
	"fmt"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/history"
	"github.com/reeflective/readline/internal/keymap"
)

func (rl *Shell) completionCommands() commands {
	return map[string]func(){
		"complete":               rl.completeWord,
		"possible-completions":   rl.possibleCompletions,
		"insert-completions":     rl.insertCompletions,
		"menu-complete":          rl.menuComplete,
		"menu-complete-backward": rl.menuCompleteBackward,
		"delete-char-or-list":    rl.deleteCharOrList,

		"menu-complete-next-tag":   rl.menuCompleteNextTag,
		"menu-complete-prev-tag":   rl.menuCompletePrevTag,
		"accept-and-menu-complete": rl.acceptAndMenuComplete,
		"vi-registers-complete":    rl.viRegistersComplete,
		"menu-incremental-search":  rl.menuIncrementalSearch,
	}
}

//
// Commands ---------------------------------------------------------------------------
//

// Attempt completion on the current word.
// Currently identitical to menu-complete.
func (rl *Shell) completeWord() {
	rl.History.SkipSave()

	// This completion function should attempt to insert the first
	// valid completion found, without printing the actual list.
	if !rl.completer.IsActive() {
		rl.startMenuComplete(rl.commandCompletion)

		if rl.Config.GetBool("menu-complete-display-prefix") {
			// Insert the prefix shared by all candidates, then display the
			// menu without selecting one (GNU menu-complete-display-prefix).
			rl.completer.InsertCommonPrefix()
			return
		}
	}

	rl.completer.Select(1, 0)
	rl.completer.SkipDisplay()
}

// List possible completions for the current word.
func (rl *Shell) possibleCompletions() {
	rl.History.SkipSave()

	rl.startMenuComplete(rl.commandCompletion)
}

// Insert all completions for the current word into the line.
func (rl *Shell) insertCompletions() {
	rl.History.Save()

	// Generate all possible completions
	if !rl.completer.IsActive() {
		rl.startMenuComplete(rl.commandCompletion)
	}

	// Insert each match, cancel insertion with preserving
	// the candidate just inserted in the line, for each. The match count is
	// invariant here: Select only moves the highlight and Cancel(false,false)
	// only restores the line buffer, neither alters the candidate groups.
	for range rl.completer.Matches() {
		rl.completer.Select(1, 0)
		rl.completer.Cancel(false, false)
	}

	// Clear the completion menu.
	rl.completer.Cancel(false, false)
	rl.completer.ClearMenu(true)
}

// Like complete-word, except that menu completion is used.
func (rl *Shell) menuComplete() {
	rl.History.SkipSave()

	// No completions are being printed yet, so simply generate the completions
	// as if we just request them without immediately selecting a candidate.
	if !rl.completer.IsActive() {
		rl.startMenuComplete(rl.commandCompletion)

		// Immediately select only if not asked to display first.
		if rl.Config.GetBool("menu-complete-display-prefix") {
			// Insert the prefix shared by all candidates before displaying.
			rl.completer.InsertCommonPrefix()
			return
		}
	}

	rl.completer.Select(1, 0)
}

// Deletes the character under the cursor if not at the
// beginning or end of the line (like delete-char).
// If at the end of the line, behaves identically to
// possible-completions.
func (rl *Shell) deleteCharOrList() {
	switch {
	case rl.cursor.Pos() < rl.line.Len():
		rl.line.CutRune(rl.cursor.Pos())
	default:
		rl.possibleCompletions()
	}
}

// Identical to menu-complete, but moves backward through the
// list of possible completions, as if menu-complete had been
// given a negative argument.
func (rl *Shell) menuCompleteBackward() {
	rl.History.SkipSave()

	// We don't do anything when not already completing.
	if !rl.completer.IsActive() {
		rl.startMenuComplete(rl.commandCompletion)
	}

	rl.completer.Select(-1, 0)
}

// In a menu completion, if there are several tags
// of completions, go to the first result of the next tag.
func (rl *Shell) menuCompleteNextTag() {
	rl.History.SkipSave()

	if !rl.completer.IsActive() {
		return
	}

	rl.completer.SelectTag(true)
}

// In a menu completion, if there are several tags of
// completions, go to the first result of the previous tag.
func (rl *Shell) menuCompletePrevTag() {
	rl.History.SkipSave()

	if !rl.completer.IsActive() {
		return
	}

	rl.completer.SelectTag(false)
}

// In a menu completion, insert the current completion
// into the buffer, and advance to the next possible completion.
func (rl *Shell) acceptAndMenuComplete() {
	rl.History.SkipSave()

	// We don't do anything when not already completing.
	if !rl.completer.IsActive() {
		return
	}

	// Also return if no candidate
	if !rl.completer.IsInserting() {
		return
	}

	// First insert the current candidate.
	rl.completer.Cancel(false, false)

	// And cycle to the next one.
	rl.completer.Select(1, 0)
}

// Open a completion menu (similar to menu-complete) with all currently populated Vim registers.
func (rl *Shell) viRegistersComplete() {
	rl.History.SkipSave()
	rl.startMenuComplete(rl.Buffers.Complete)
}

// In a menu completion (whether a candidate is selected or not), start incremental-search
// (fuzzy search) on the results. Search backward incrementally for a specified string.
// The search is case-insensitive if the search string does not have uppercase letters
// and no numeric argument was given. The string may begin with ‘^’ to anchor the search
// to the beginning of the line. A restricted set of editing functions is available in the
// mini-buffer. Keys are looked up in the special isearch keymap, On each change in the
// mini-buffer, any currently selected candidate is dropped from the line and the menu.
// An interrupt signal, as defined by the stty setting, will stop the search and go back to the original line.
func (rl *Shell) menuIncrementalSearch() {
	rl.History.SkipSave()

	// Always regenerate the list of completions.
	rl.completer.GenerateWith(rl.commandCompletion)
	rl.completer.IsearchStart("completions", false, false)
}

// RefreshCompletions regenerates the currently active completion menu from the
// cached completer and repaints, so completions produced asynchronously (for
// instance by a background producer that updates a cache the completer reads)
// can be shown in place without the user pressing a key.
//
// It is safe to call from any goroutine. If no completion menu is active it is
// a clean no-op. The regeneration runs on the Readline goroutine (rendering
// stays single-writer). If the user has a candidate selected, that selection is
// preserved across the refresh: the same candidate (matched by tag+value) is
// re-selected in the rebuilt menu, or, if it no longer exists in the new
// results, the menu is left active with no selection.
func (rl *Shell) RefreshCompletions() {
	rl.completer.RequestRegen()
	rl.Keys.RequestRefresh()
}

//
// Utilities --------------------------------------------------------------------------
//

// startMenuComplete generates a completion menu with completions
// generated from a given completer, without selecting a candidate.
func (rl *Shell) startMenuComplete(completer completion.Completer) {
	rl.History.SkipSave()

	rl.Keymap.SetLocal(keymap.MenuSelect)
	rl.completer.GenerateWith(completer)
}

// commandCompletion generates the completions for commands/args/flags.
//
// The application-provided completer is user code that readline calls on nearly
// every keystroke (menu completion and as-you-type autocomplete both funnel
// through here). A panic in it — a bad completer or transient command-tree
// state — is recovered and surfaced as a completion message, so a single faulty
// completion degrades into a visible error instead of crashing the shell.
func (rl *Shell) commandCompletion() (values completion.Values) {
	if rl.Completer == nil {
		return completion.Values{}
	}

	defer func() {
		if r := recover(); r != nil {
			msg := CompleteMessage("completion error: %v", r)
			values = msg.convert()
		}
	}()

	line, cursor := rl.completer.Line()
	comps := rl.Completer(*line, cursor.Pos())

	return comps.convert()
}

// historyCompletion manages the various completion/isearch modes related
// to history control. It can start the history completions, stop them, cycle
// through sources if more than one, and adjust the completion/isearch behavior.
func (rl *Shell) historyCompletion(forward, filterLine, substring bool) {
	switch {
	case rl.Keymap.Local() == keymap.MenuSelect || rl.Keymap.Local() == keymap.Isearch || rl.completer.AutoCompleting():
		// If we are currently completing the last
		// history source, cancel history completion.
		if rl.History.OnLastSource() {
			rl.History.Cycle(true)
			rl.completer.ResetForce()
			rl.Hint.Reset()

			return
		}

		// Else complete the next history source.
		rl.History.Cycle(true)

		fallthrough

	default:
		// Notify if we don't have history sources at all.
		if rl.History.Current() == nil {
			rl.Hint.SetTemporary(fmt.Sprintf("%s%s%s %s", color.Dim, color.FgRed, "No command history source", color.Reset))
			return
		}

		// Generate the completions with specified behavior.
		completer := func() completion.Values {
			maxLines := rl.Display.AvailableHelperLines()
			return history.Complete(rl.History, forward, filterLine, maxLines, rl.completer.IsearchRegex)
		}

		if substring {
			rl.completer.GenerateWith(completer)
			rl.completer.IsearchStart(rl.History.Name(), true, true)
		} else {
			rl.startMenuComplete(completer)
			rl.completer.AutocompleteForce()
		}
	}
}
