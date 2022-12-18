package readline

import (
	"bufio"
	"context"
	"fmt"
	"strings"
)

// TabDisplayType defines how the autocomplete suggestions display
type TabDisplayType int

const (
	// TabDisplayGrid is the default. It's where the screen below the prompt is
	// divided into a grid with each suggestion occupying an individual cell.
	TabDisplayGrid = iota

	// TabDisplayList is where suggestions are displayed as a list with a
	// description. The suggestion gets highlighted but both are searchable (ctrl+f)
	TabDisplayList

	// TabDisplayMap is where suggestions are displayed as a list with a
	// description however the description is what gets highlighted and only
	// that is searchable (ctrl+f). The benefit of TabDisplayMap is when your
	// autocomplete suggestions are IDs rather than human terms.
	TabDisplayMap
)

// startMenuComplete generates a completion menu with completions
// generated from a given completer, without selecting a candidate.
func (rl *Instance) startMenuComplete(completer func()) {
	rl.local = menuselect
	rl.compConfirmWait = false
	rl.skipUndoAppend()

	// Call the provided completer function
	// to produce all possible completions.
	completer()

	// And store it if it's going to be used by autocomplete.
	rl.completer = completer

	// Cancel completion mode if we don't have any candidates.
	if rl.noCompletions() {
		rl.hintNoMatches()
		rl.resetTabCompletion()
		return
	}

	// Let all groups compute their display/candidate strings
	// and coordinates, and do some adjustments where needed.
	rl.initializeCompletions()

	// Always ensure we have a current group.
	rl.getCurrentGroup()

	// When there is only candidate, automatically insert it
	// and exit the completion mode, except in history completion.
	if rl.hasUniqueCandidate() && len(rl.histHint) == 0 {
		rl.undoSkipAppend = false
		rl.insertCandidate()
		rl.resetTabCompletion()
	}
}

// getTabSearchCompletion - Populates and sets up completion for completion search.
func (rl *Instance) getTabSearchCompletion() {
	// Get completions from the engine, and make sure there is a current group.
	rl.generateCompletions()
	if len(rl.tcGroups) == 0 {
		return
	}
	rl.getCurrentGroup()

	// Set the hint for this completion mode
	rl.hintText = append([]rune("Completion search: "), rl.tfLine...)

	for _, g := range rl.tcGroups {
		g.updateTabFind(rl)
	}

	// If total number of matches is zero, we directly change the hint, and return
	if comps, _, _ := rl.getCompletionCount(); comps == 0 {
		rl.hintText = append(rl.hintText, []rune(DIM+RED+" ! no matches (Ctrl-G/Esc to cancel)"+RESET)...)
	}
}

// generateCompletions - Calls the completion engine/function to yield a list of 0 or more completion groups,
// sets up a delayed tab context and passes it on to the tab completion engine function, and ensure no
// nil groups/items will pass through. This function is called by different comp search/nav modes.
func (rl *Instance) generateCompletions() {
	if rl.TabCompleter == nil {
		return
	}

	// Cancel any existing tab context first.
	if rl.delayedTabContext.cancel != nil {
		rl.delayedTabContext.cancel()
	}

	// Recreate a new context
	rl.delayedTabContext = DelayedTabContext{rl: rl}
	rl.delayedTabContext.Context, rl.delayedTabContext.cancel = context.WithCancel(context.Background())

	// Get the correct line to be completed, and the current cursor position
	compLine, compPos := rl.getCompletionLine()

	prefix, groups := rl.TabCompleter(compLine, compPos, rl.delayedTabContext)

	rl.tcPrefix = prefix
	rl.tcGroups = checkNilItems(groups)
}

// moveCompletionSelection - This function is in charge of
// computing the new position in the current completions liste.
func (rl *Instance) moveCompletionSelection(x, y int) {
	g := rl.getCurrentGroup()

	// If there is no current group, we leave any current completion mode.
	if g == nil || len(g.Values) == 0 {
		return
	}

	// done means we need to find the next/previous group.
	// next determines if we need to get the next OR previous group.
	var done, next bool

	// Depending on the display, we only keep track of x or (x and y)
	switch g.DisplayType {
	case TabDisplayGrid:
		done, next = g.moveTabGridHighlight(rl, x, y)
	case TabDisplayList:
		done, next = g.moveTabListHighlight(rl, x, y)
	case TabDisplayMap:
		done, next = g.moveTabMapHighlight(rl, x, y)
	}

	// Cycle to next/previous group, if done with current one.
	if done {
		g.selected = CompletionValue{}

		if next {
			rl.cycleNextGroup()
			nextGroup := rl.getCurrentGroup()
			nextGroup.goFirstCell()

			nextGroup.selected = nextGroup.grouped[0][0]
		} else {
			rl.cyclePreviousGroup()
			prevGroup := rl.getCurrentGroup()
			prevGroup.goLastCell()

			lastRow := g.grouped[len(g.grouped)-1]
			g.selected = lastRow[len(lastRow)-1]
		}
	}
}

// printCompletions - Prints all completion groups and their items
func (rl *Instance) printCompletions() {
	rl.tcUsedY = 0

	// The final completions string to print.
	var completions string

	// Safecheck
	if rl.local != menuselect && rl.local != isearch && !rl.needsAutoComplete() {
		return
	}

	// In any case, we write the completions strings, trimmed for redundant
	// newline occurences that have been put at the end of each group.
	for _, group := range rl.tcGroups {
		completions += group.writeCompletion(rl)
	}

	// Because some completion groups might have more suggestions
	// than what their MaxLength allows them to, cycling sometimes occur,
	// but does not fully clears itself: some descriptions are messed up with.
	// We always clear the screen as a result, between writings.
	print(seqClearScreenBelow)

	// Crop the completions so that it fits within our MaxTabCompleterRows
	completions, rl.tcUsedY = rl.cropCompletions(completions)

	// Then we print all of them.
	print(completions)
}

// cropCompletions - When the user cycles through a completion list longer
// than the console MaxTabCompleterRows value, we crop the completions string
// so that "global" cycling (across all groups) is printed correctly.
func (rl *Instance) cropCompletions(comps string) (cropped string, usedY int) {
	// If we actually fit into the MaxTabCompleterRows, return the comps
	if rl.tcUsedY < rl.config.MaxTabCompleterRows {
		return comps, rl.tcUsedY
	}

	// Else we go on, but we have more comps than what allowed:
	// we will add a line to the end of the comps, giving the actualized
	// number of completions remaining and not printed
	moreComps := func(cropped string, offset int) (hinted string, noHint bool) {
		_, _, adjusted := rl.getCompletionCount()
		remain := adjusted - offset
		if remain == 0 {
			return cropped, true
		}
		hint := fmt.Sprintf(DIM+YELLOW+" %d more completions... (scroll down to show)"+RESET+"\n", remain)
		hinted = cropped + hint
		return hinted, false
	}

	// Get the current absolute candidate position (prev groups x suggestions + curGroup.tcPosY)
	absPos := rl.getAbsPos()

	// Get absPos - MaxTabCompleterRows for having the number of lines to cut at the top
	// If the number is negative, that means we don't need to cut anything at the top yet.
	maxLines := absPos - rl.config.MaxTabCompleterRows
	if maxLines < 0 {
		maxLines = 0
	}

	// Scan the completions for cutting them at newlines
	scanner := bufio.NewScanner(strings.NewReader(comps))

	// If absPos < MaxTabCompleterRows, cut below MaxTabCompleterRows and return
	if absPos <= rl.config.MaxTabCompleterRows {
		var count int
		for scanner.Scan() {
			line := scanner.Text()
			if count < rl.config.MaxTabCompleterRows {
				cropped += line + "\n"
				count++
			} else {
				count++
				break
			}
		}

		cropped, _ = moreComps(cropped, count)

		return cropped, count
	}

	// If absolute > MaxTabCompleterRows, cut above and below and return
	//      -> This includes de facto when we tabCompletionReverse
	if absPos > rl.config.MaxTabCompleterRows {
		cutAbove := absPos - rl.config.MaxTabCompleterRows
		var count int
		for scanner.Scan() {
			line := scanner.Text()
			if count < cutAbove {
				count++
				continue
			}
			if count >= cutAbove && count < absPos {
				cropped += line + "\n"
				count++
			} else {
				count++
				break
			}
		}

		cropped, _ = moreComps(cropped, rl.config.MaxTabCompleterRows+cutAbove)

		return cropped, count - cutAbove
	}

	return
}

func (rl *Instance) getAbsPos() int {
	var prev int
	var foundCurrent bool
	for _, grp := range rl.tcGroups {
		if grp.isCurrent {
			prev += grp.tcPosY + 1 // + 1 for title
			foundCurrent = true
			break
		} else {
			prev += grp.tcMaxY + 1 // + 1 for title
		}
	}

	// If there was no current group, it means
	// we showed completions but there is no
	// candidate selected yet, return 0
	if !foundCurrent {
		return 0
	}
	return prev
}

// We pass a special subset of the current input line, so that
// completions are available no matter where the cursor is.
func (rl *Instance) getCompletionLine() (line []rune, pos int) {
	pos = rl.pos - len(rl.comp)
	if pos < 0 {
		pos = 0
	}

	switch {
	case rl.pos == len(rl.line):
		line = rl.line
	case rl.pos < len(rl.line):
		line = rl.line[:pos]
	default:
		line = rl.line
	}

	return
}

func (rl *Instance) getCurrentGroup() (group *CompletionGroup) {
	for _, g := range rl.tcGroups {
		if g.isCurrent && len(g.Values) > 0 {
			return g
		}
	}
	// We might, for whatever reason, not find one.
	// If there are groups but no current, make first one the king.
	if len(rl.tcGroups) > 0 {
		// Find first group that has list > 0, as another checkup
		for _, g := range rl.tcGroups {
			if len(g.Values) > 0 {
				g.isCurrent = true
				return g
			}
		}
	}
	return
}

// cycleNextGroup - Finds either the first non-empty group,
// or the next non-empty group after the current one.
func (rl *Instance) cycleNextGroup() {
	for i, g := range rl.tcGroups {
		if g.isCurrent {
			g.isCurrent = false
			if i == len(rl.tcGroups)-1 {
				rl.tcGroups[0].isCurrent = true
			} else {
				rl.tcGroups[i+1].isCurrent = true
				// Here, we check if the cycled group is not empty.
				// If yes, cycle to next one now.
				next := rl.getCurrentGroup()
				if len(next.Values) == 0 {
					rl.cycleNextGroup()
				}
			}
			break
		}
	}
}

// cyclePreviousGroup - Same as cycleNextGroup but reverse
func (rl *Instance) cyclePreviousGroup() {
	for i, g := range rl.tcGroups {
		if g.isCurrent {
			g.isCurrent = false
			if i == 0 {
				rl.tcGroups[len(rl.tcGroups)-1].isCurrent = true
			} else {
				rl.tcGroups[i-1].isCurrent = true
				prev := rl.getCurrentGroup()
				if len(prev.Values) == 0 {
					rl.cyclePreviousGroup()
				}
			}
			break
		}
	}
}

// When the completions are either longer than:
// - The user-specified max completion length
// - The terminal lengh
// we use this function to prompt for confirmation before printing comps.
func (rl *Instance) promptCompletionConfirm(sentence string) {
	rl.hintText = []rune(sentence)

	rl.compConfirmWait = true
	rl.undoSkipAppend = true

	rl.renderHelpers()
}

func (rl *Instance) getCompletionCount() (comps int, lines int, adjusted int) {
	for _, group := range rl.tcGroups {
		comps += group.rows
		// if group.Name != "" {
		adjusted++ // Title
		// }
		if group.tcMaxY > group.rows {
			lines += group.rows
			adjusted += group.rows
		} else {
			lines += group.tcMaxY
			adjusted += group.tcMaxY
		}
	}
	return
}

func (rl *Instance) getCurrentCandidate() (comp string) {
	cur := rl.getCurrentGroup()
	if cur == nil {
		return
	}

	return cur.getCurrentCell(rl).Value
}

// this is called once and only if the local keymap has not
// matched a given input key: that means no completion menu
// helpers were used, so we need to update our completion
// menu before actually editing/moving around the line.
func (rl *Instance) updateCompletionState() {
	rl.resetVirtualComp(false)
	rl.resetTabCompletion()
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
	// When we have a history hint, that means we are
	// currently completing history, potentially in
	// autocomplete: don't exit the current menu.
	if rl.local == menuselect && len(rl.histHint) == 0 {
		rl.local = ""
	}

	rl.tcPrefix = ""
	rl.compConfirmWait = false
	rl.tcUsedY = 0

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

func (rl *Instance) needsAutoComplete() bool {
	needsComplete := rl.config.AutoComplete &&
		len(rl.line) > 0 &&
		rl.local != menuselect &&
		rl.local != isearch

	isCorrectMenu := rl.main != vicmd

	// We might be at the beginning of line,
	// but currently proposing history completion.
	completingHistory := len(rl.histHint) > 0

	// We always refresh history, except when
	// currently having a candidate selection.
	if completingHistory && isCorrectMenu && len(rl.comp) == 0 {
		return true
	}

	if needsComplete && isCorrectMenu {
		return true
	}

	return false
}

func (rl *Instance) isAutoCompleting() bool {
	if rl.config.AutoComplete &&
		len(rl.line) > 0 {
		return true
	}

	return false
}

// autoComplete generates the correct completions in autocomplete mode.
// We don't do it when we are currently in the completion keymap,
// since that means completions have already been computed.
func (rl *Instance) autoComplete() {
	if !rl.needsAutoComplete() {
		return
	}

	rl.resetTabCompletion()

	if rl.completer != nil {
		rl.completer()
	} else {
		rl.generateCompletions()
	}

	rl.initializeCompletions()
}
