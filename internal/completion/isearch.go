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
	e.isearch = nil
}

// GetBuffer returns either the current input line when incremental
// search is not running, or the incremental minibuffer.
func (e *Engine) GetBuffer() (*core.Line, *core.Cursor, *core.Selection) {
	// Incremental search buffer
	if e.keymaps.Local() == keymap.Isearch {
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
	if e.keymaps.Local() != keymap.Isearch {
		return
	}

	// If we have a virtually inserted candidate, it's because the
	// last action was a tab-complete selection: we don't need to
	// refresh the list of matches, as the minibuffer did not change,
	// and because it would make our currently selected comp to drop.
	if len(e.selected.Value) > 0 {
		return
	}

	// Otherwise, recompute and refresh the matches.
	var regexStr string
	if hasUpper(*e.isearchBuf) {
		regexStr = string(*e.isearchBuf)
	} else {
		regexStr = "(?i)" + string(*e.isearchBuf)
	}

	var err error
	e.isearch, err = regexp.Compile(regexStr)
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
	isearchHint := color.Bold + color.FgCyan + e.isearchName +
		" (isearch): " + color.Reset + color.Bold + string(*e.isearchBuf)
	e.hint.Set(isearchHint)

	// And update the inserted candidate if autoinsert is enabled.
	if e.isearchInsert && e.Matches() > 0 && e.isearchBuf.Len() > 0 {
		e.Select(1, 0)
	}
}

// In history isearch, insert the first matching candidate.
// This candidate will be cleared/replaced as soon as another
// key/change is operated on the isearch buffer.
// if len(rl.histHint) > 0 && len(rl.tcGroups) > 0 && len(rl.tcGroups[0].values) > 0 {
// 	rl.resetVirtualComp(true)
// 	cur := rl.currentGroup()
// 	cur.tcPosY = 0
// 	cur.tcPosX = 0
// 	rl.updateVirtualComp()
// 	cur.tcPosY = -1
// 	cur.tcPosX = -1
// }
