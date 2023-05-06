package completion

import (
	"regexp"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/keymap"
)

// IsearchStart starts incremental search (fuzzy-finding)
// with values matching the isearch minibuffer as a regexp.
func (e *Engine) IsearchStart(name string, autoinsert bool) {
	e.keymaps.SetLocal(keymap.Isearch)
	e.auto = true
	e.isearchInsert = autoinsert
	e.adaptIsearchInsertMode()

	e.isearchBuf = new(core.Line)
	e.isearchCur = core.NewCursor(e.isearchBuf)

	// Hints
	e.isearchName = name
	e.hint.Set(color.Bold + color.FgCyan + e.isearchName + " (isearch): " + color.Reset + string(*e.isearchBuf))
}

// IsearchStop exists the incremental search mode,
// and drops the currently used regexp matcher.
func (e *Engine) IsearchStop() {
	e.keymaps.SetLocal("")
	e.auto = false
	e.autoForce = false
	e.isearchBuf = nil
	e.isearchRgx = nil
	e.isearchCur = nil

	e.resetIsearchInsertMode()
}

// GetBuffer returns either the current input line when incremental
// search is not running, or the incremental minibuffer.
func (e *Engine) GetBuffer() (*core.Line, *core.Cursor, *core.Selection) {
	// Non/Incremental search buffer
	searching, _, _ := e.NonIncrementallySearching()

	if e.keymaps.Local() == keymap.Isearch || searching {
		selection := core.NewSelection(e.isearchBuf, e.isearchCur)
		return e.isearchBuf, e.isearchCur, selection
	}

	// Completed line (with inserted candidate)
	if len(e.selected.Value) > 0 {
		return e.completed, e.compCursor, e.selection
	}

	// Or completer inactive, normal input line.
	return e.line, e.cursor, e.selection
}

// UpdateIsearch recompiles the isearch as a regex and
// filters matching candidates in the available completions.
func (e *Engine) UpdateIsearch() {
	searching, _, _ := e.NonIncrementallySearching()

	if e.keymaps.Local() != keymap.Isearch && !searching {
		return
	}

	// If we have a virtually inserted candidate, it's because the
	// last action was a tab-complete selection: we don't need to
	// refresh the list of matches, as the minibuffer did not change,
	// and because it would make our currently selected comp to drop.
	if len(e.selected.Value) > 0 {
		return
	}

	// Update helpers depending on the search/minibuffer mode.
	if e.keymaps.Local() == keymap.Isearch {
		e.updateIncrementalSearch()
	} else {
		e.updateNonIncrementalSearch()
	}
}

// NonIsearchStart starts a non-incremental, fake search mode:
// it does not produce or tries to match against completions,
// but uses a minibuffer similarly to incremental search mode.
func (e *Engine) NonIsearchStart(name string, repeat, forward, substring bool) {
	if repeat {
		e.isearchBuf = new(core.Line)
		e.isearchBuf.Set([]rune(e.searchLast)...)
	} else {
		e.isearchBuf = new(core.Line)
	}

	e.isearchCur = core.NewCursor(e.isearchBuf)
	e.isearchCur.Set(e.isearchBuf.Len())

	e.isearchName = name
	e.isearchForward = forward
	e.isearchSubstring = substring

	e.keymaps.NonIncrementalSearchStart()
	e.adaptIsearchInsertMode()
}

// NonIsearchStop exits the non-incremental search mode.
func (e *Engine) NonIsearchStop() {
	e.searchLast = string(*e.isearchBuf)
	e.isearchBuf = nil
	e.isearchRgx = nil
	e.isearchCur = nil
	e.isearchForward = false
	e.isearchSubstring = false

	// Reset keymap and helpers
	e.keymaps.NonIncrementalSearchStop()
	e.resetIsearchInsertMode()
	e.hint.Reset()
}

// NonIncrementallySearching returns true if the completion engine
// is currently using a minibuffer for non-incremental search mode.
func (e *Engine) NonIncrementallySearching() (searching, forward, substring bool) {
	searching = e.isearchCur != nil && e.keymaps.Local() != keymap.Isearch
	forward = e.isearchForward
	substring = e.isearchSubstring

	return
}

func (e *Engine) updateIncrementalSearch() {
	var regexStr string
	if hasUpper(*e.isearchBuf) {
		regexStr = string(*e.isearchBuf)
	} else {
		regexStr = "(?i)" + string(*e.isearchBuf)
	}

	var err error
	e.isearchRgx, err = regexp.Compile(regexStr)

	if err != nil {
		e.hint.Set(color.FgRed + "Failed to compile i-search regexp")
	}

	// Refresh completions with the current minibuffer as a filter.
	e.GenerateWith(e.cached)

	// And filter out the completions.
	for _, g := range e.groups {
		g.updateIsearch(e)
	}

	// Update the hint section.
	isearchHint := color.Bold + color.FgCyan + e.isearchName + " (inc-search)"

	if e.Matches() == 0 {
		isearchHint += color.Reset + color.Bold + color.FgRed + " (no matches)"
	}

	isearchHint += ": " + color.Reset + color.Bold + string(*e.isearchBuf) + color.Reset + "_"

	e.hint.Set(isearchHint)

	// And update the inserted candidate if autoinsert is enabled.
	if e.isearchInsert && e.Matches() > 0 && e.isearchBuf.Len() > 0 {
		e.Select(1, 0)
	}
}

func (e *Engine) updateNonIncrementalSearch() {
	isearchHint := color.Bold + color.FgCyan + e.isearchName +
		" (non-inc-search): " + color.Reset + color.Bold + string(*e.isearchBuf) + color.Reset + "_"
	e.hint.Set(isearchHint)
}

func (e *Engine) adaptIsearchInsertMode() {
	e.isearchModeExit = e.keymaps.Main()

	if !e.keymaps.IsEmacs() && e.keymaps.Main() != keymap.ViInsert {
		e.keymaps.SetMain(keymap.ViInsert)
	}
}

func (e *Engine) resetIsearchInsertMode() {
	if e.isearchModeExit == "" {
		return
	}

	if e.keymaps.Main() != e.isearchModeExit {
		e.keymaps.SetMain(e.isearchModeExit)
		e.isearchModeExit = ""
	}

	if e.keymaps.Main() == keymap.ViCommand {
		e.cursor.CheckCommand()
	}
}
