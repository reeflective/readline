package readline

import (
	"bufio"
	"context"
	"fmt"
	"strings"
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

	// Always ensure we have a current group.
	rl.currentGroup()

	// When there is only candidate, automatically insert it
	// and exit the completion mode, except in history completion.
	if rl.hasUniqueCandidate() && len(rl.histHint) == 0 {
		rl.undoSkipAppend = false
		rl.insertCandidate()
		rl.resetTabCompletion()
	}
}

// generateCompletions - Calls the completion engine/function to yield a list of 0 or more completion groups,
// sets up a delayed tab context and passes it on to the tab completion engine function, and ensure no
// nil groups/items will pass through. This function is called by different comp search/nav modes.
func (rl *Instance) generateCompletions() {
	if rl.Completer == nil {
		return
	}

	rl.tcGroups = make([]*comps, 0)

	// Cancel any existing tab context first.
	if rl.delayedTabContext.cancel != nil {
		rl.delayedTabContext.cancel()
	}

	// Recreate a new context
	rl.delayedTabContext = DelayedTabContext{rl: rl}
	rl.delayedTabContext.Context, rl.delayedTabContext.cancel = context.WithCancel(context.Background())

	// Get the correct line to be completed, and the current cursor position
	compLine, compPos := rl.getCompletionLine()

	// Generate the completions, setup the prefix and group the results.
	comps := rl.Completer(compLine, compPos, rl.delayedTabContext)
	rl.groupCompletions(comps)
	rl.setCompletionPrefix(comps)
}

func (rl *Instance) groupCompletions(comps Completions) {
	// TODO: Set up the hints with our messages/usage strings.

	// Nothing else to do if no completions
	if len(comps.values) == 0 {
		return
	}

	comps.values.eachTag(func(tag string, values rawValues) {
		// Separate the completions that have a description and
		// those which don't, and devise if there are aliases.
		vals, noDescVals, aliased := groupValues(values)

		// Create a "first" group with the "first" grouped values
		rl.newGroup(tag, vals, aliased, comps.noSpace)

		// If we have a remaining group of values without descriptions,
		// we will print and use them in a separate, anonymous group.
		if len(noDescVals) > 0 {
			rl.newGroup("", noDescVals, false, comps.noSpace)
		}
	})
}

// groupValues separates values based on whether they have descriptions, or are aliases of each other.
func groupValues(values rawValues) (vals, noDescVals rawValues, aliased bool) {
	var descriptions []string

	for _, val := range values {
		// Grid completions
		if val.Description == "" {
			noDescVals = append(noDescVals, val)
			continue
		}

		// List/map completions.
		if stringInSlice(val.Description, descriptions) {
			aliased = true
		}
		descriptions = append(descriptions, val.Description)
		vals = append(vals, val)
	}

	// if no candidates have a description, swap
	if len(vals) == 0 {
		vals = noDescVals
		noDescVals = make(rawValues, 0)
	}

	return
}

func (rl *Instance) setCompletionPrefix(comps Completions) {
	switch comps.PREFIX {
	case "":
		// When no prefix has been specified, use
		// the current word up to the cursor position.
		lineWords, _, _ := tokeniseSplitSpaces(rl.line, rl.pos)
		if len(lineWords) > 0 {
			last := lineWords[len(lineWords)-1]
			if last[len(last)-1] != ' ' {
				rl.tcPrefix = lineWords[len(lineWords)-1]
			}
		}

	default:
		// When the prefix has been overriden, add it to all
		// completions AND as a line prefix, for correct candidate insertion.
		rl.tcPrefix = comps.PREFIX
	}
}

func (rl *Instance) updateSelector(x, y int) {
	grp := rl.currentGroup()

	// If there is no current group, we
	// leave any current completion mode.
	if grp == nil || len(grp.values) == 0 {
		return
	}

	done, next := grp.moveSelector(rl, x, y)
	if !done {
		return
	}

	var newGrp *comps

	if next {
		rl.cycleNextGroup()
		newGrp = rl.currentGroup()
		newGrp.firstCell()

	} else {
		rl.cyclePreviousGroup()
		newGrp = rl.currentGroup()
		newGrp.lastCell()
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
		completions += group.writeComps(rl)
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
		_, _, adjusted := rl.completionCount()
		remain := adjusted - offset
		if remain == 0 {
			return cropped, true
		}
		hint := fmt.Sprintf(seqDim+seqFgYellow+" %d more completions... (scroll down to show)"+seqReset+"\n", remain)
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

// this is called once and only if the local keymap has not
// matched a given input key: that means no completion menu
// helpers were used, so we need to update our completion
// menu before actually editing/moving around the line.
func (rl *Instance) updateCompletionState() {
	rl.resetVirtualComp(false)
	rl.resetTabCompletion()
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
		cur := rl.currentGroup()
		if cur == nil {
			return false
		}

		if len(cur.values) == 1 {
			return len(cur.values[0]) == 1
		}

		return len(cur.values) == 1

	default:
		var count int

	GROUPS:
		for _, group := range rl.tcGroups {
			for _, row := range group.values {
				count++
				for range row {
					count++
				}
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

	isCorrectMenu := rl.main != vicmd && rl.local != isearch

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
}
