package readline

import (
	"bufio"
	"fmt"
	"strings"
)

// startMenuComplete generates a completion menu with completions
// generated from a given completer, without selecting a candidate.
func (rl *Instance) startMenuComplete(completer func()) {
	rl.local = menuselect
	rl.skipUndoAppend()

	// Call the completer function to produce completions,
	// and store it if it's going to be used by autocomplete.
	completer()
	rl.completer = completer

	// Cancel completion mode if we don't have any candidates.
	// The hint string will be provided in a separate step.
	if rl.noCompletions() {
		rl.resetCompletion()
		rl.completer = nil

		return
	}

	rl.currentGroup()

	// When there is only candidate, automatically insert it
	// and exit the completion mode, except in history completion.
	if rl.hasUniqueCandidate() && len(rl.histHint) == 0 {
		rl.undoSkipAppend = false
		rl.insertCandidate()
		rl.resetCompletion()
		rl.resetHintText()
	}
}

func (rl *Instance) normalCompletions() {
	if rl.Completer == nil {
		return
	}

	rl.tcGroups = make([]*comps, 0)

	// Get the correct line to be completed, and the current cursor position
	compLine, compPos := rl.getCompletionLine()

	// Generate the completions, setup the prefix and group the results.
	comps := rl.Completer(compLine, compPos)
	rl.groupCompletions(comps)
	rl.setCompletionPrefix(comps)
}

func (rl *Instance) historyCompletion(forward, filterLine bool) {
	switch rl.local {
	case menuselect, isearch:
		// If we are currently completing the last
		// history source, cancel history completion.
		if rl.historySourcePos == len(rl.histories)-1 {
			rl.histHint = []rune{}
			rl.nextHistorySource()
			rl.resetCompletion()
			rl.completer = nil
			rl.local = ""
			rl.resetHintText()

			return
		}

		// Else complete the next history source.
		rl.nextHistorySource()

		fallthrough

	default:
		// Notify if we don't have history sources at all.
		if rl.currentHistory() == nil {
			noHistory := fmt.Sprintf("%s%s%s %s", seqDim, seqFgRed, "No command history source", seqReset)
			rl.histHint = []rune(noHistory)

			return
		}

		// Generate the completions with specified behavior.
		historyCompletion := func() {
			rl.tcGroups = make([]*comps, 0)

			// Either against the current line or not.
			if filterLine {
				rl.tcPrefix = string(rl.line)
			}

			comps := rl.completeHistory(forward)
			comps = comps.DisplayList()
			rl.groupCompletions(comps)
			rl.setCompletionPrefix(comps)
		}

		// Else, generate the completions.
		rl.startMenuComplete(historyCompletion)
	}
}

func (rl *Instance) registerCompletion() {
	rl.registersComplete = true
	rl.tcGroups = make([]*comps, 0)
	comps := rl.completeRegisters()
	comps = comps.DisplayList("*")
	rl.groupCompletions(comps)
	rl.setCompletionPrefix(comps)
}

func (rl *Instance) groupCompletions(comps Completions) {
	rl.hintCompletions(comps)

	// Nothing else to do if no completions
	if len(comps.values) == 0 {
		return
	}

	comps.values.eachTag(func(tag string, values rawValues) {
		// Separate the completions that have a description and
		// those which don't, and devise if there are aliases.
		vals, noDescVals, aliased := groupValues(values)

		// Create a "first" group with the "first" grouped values
		rl.newGroup(comps, tag, vals, aliased)

		// If we have a remaining group of values without descriptions,
		// we will print and use them in a separate, anonymous group.
		if len(noDescVals) > 0 {
			rl.newGroup(comps, "", noDescVals, false)
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
		// When the prefix has been overridden, add it to all
		// completions AND as a line prefix, for correct candidate insertion.
		rl.tcPrefix = comps.PREFIX
	}
}

func (rl *Instance) updateSelector(tabX, tabY int) {
	grp := rl.currentGroup()

	// If there is no current group, we
	// leave any current completion mode.
	if grp == nil || len(grp.values) == 0 {
		return
	}

	done, next := grp.moveSelector(rl, tabX, tabY)
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

// printCompletions - Prints all completion groups and their items.
func (rl *Instance) printCompletions() {
	rl.tcUsedY = 0

	// The final completions string to print.
	var completions string

	// Safecheck
	if rl.local != menuselect && rl.local != isearch && !rl.needsAutoComplete() {
		return
	}

	// In any case, we write the completions strings, trimmed for redundant
	// newline occurrences that have been put at the end of each group.
	for _, group := range rl.tcGroups {
		completions += group.writeComps(rl)
	}

	// Crop the completions so that it fits within our MaxTabCompleterRows
	completions, rl.tcUsedY = rl.cropCompletions(completions)

	if completions != "" {
		print("\n")
		rl.tcUsedY++

		// Because some completion groups might have more suggestions
		// than what their MaxLength allows them to, cycling sometimes occur,
		// but does not fully clears itself: some descriptions are messed up with.
		// We always clear the screen as a result, between writings.
		print(seqClearScreenBelow)
	}

	// Then we print all of them.
	print(completions)
}

// cropCompletions - When the user cycles through a completion list longer
// than the console MaxTabCompleterRows value, we crop the completions string
// so that "global" cycling (across all groups) is printed correctly.
func (rl *Instance) cropCompletions(comps string) (cropped string, usedY int) {
	maxRows := rl.getCompletionMaxRows()

	// Get the current absolute candidate position
	absPos := rl.getAbsPos()

	// Scan the completions for cutting them at newlines
	scanner := bufio.NewScanner(strings.NewReader(comps))

	// If absPos < MaxTabCompleterRows, cut below MaxTabCompleterRows and return
	if absPos < maxRows {
		return rl.cutCompletionsBelow(scanner, maxRows)
	}

	// If absolute > MaxTabCompleterRows, cut above and below and return
	//      -> This includes de facto when we tabCompletionReverse
	if absPos >= maxRows {
		return rl.cutCompletionsAboveBelow(scanner, maxRows, absPos)
	}

	return
}

// this is called once and only if the local keymap has not
// matched a given input key: that means no completion menu
// helpers were used, so we need to update our completion
// menu before actually editing/moving around the line.
func (rl *Instance) updateCompletion() {
	switch rl.local {
	case isearch:
		rl.resetVirtualComp(true)
	default:
		rl.resetVirtualComp(false)
	}

	rl.resetCompletion()
}

func (rl *Instance) cutCompletionsBelow(scanner *bufio.Scanner, maxRows int) (string, int) {
	var count int
	var cropped string

	for scanner.Scan() {
		line := scanner.Text()
		if count < maxRows {
			cropped += line + "\n"
			count++
		} else {
			break
		}
	}

	cropped, _ = rl.excessCompletionsHint(cropped, maxRows, count-1)

	return cropped, count
}

func (rl *Instance) cutCompletionsAboveBelow(scanner *bufio.Scanner, maxRows, absPos int) (string, int) {
	cutAbove := absPos - maxRows + 1

	var cropped string
	var count int
	var noRemain bool

	for scanner.Scan() {
		line := scanner.Text()

		if count <= cutAbove {
			count++

			continue
		}

		if count > cutAbove && count <= absPos {
			cropped += line + "\n"
			count++
		} else {
			break
		}
	}

	cropped, noRemain = rl.excessCompletionsHint(cropped, maxRows, maxRows+cutAbove+1)
	if noRemain {
		count--
	}

	return cropped, count - cutAbove
}

func (rl *Instance) excessCompletionsHint(cropped string, maxRows, offset int) (string, bool) {
	_, _, adjusted := rl.completionCount()
	remain := adjusted - offset

	if remain <= 0 || offset < maxRows {
		return cropped, true
	}

	hint := fmt.Sprintf(seqDim+seqFgYellow+" %d more completion rows... (scroll down to show)"+seqReset+"\n", remain)

	hinted := cropped + hint

	return hinted, false
}

func (rl *Instance) resetCompletion() {
	// When we have a history hint, that means we are
	// currently completing history, potentially in
	// autocomplete: don't exit the current menu.
	if rl.local == menuselect && len(rl.histHint) == 0 {
		rl.local = ""
	}

	rl.tcPrefix = ""
	rl.tcUsedY = 0

	// Don't persist registers completion.
	if rl.registersComplete {
		rl.completer = nil
		rl.registersComplete = false
	}

	// Reset tab highlighting
	if len(rl.tcGroups) > 0 {
		for _, g := range rl.tcGroups {
			g.isCurrent = false
		}

		rl.tcGroups[0].isCurrent = true
	}
}

// Check if we have a single completion candidate.
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

	rl.resetCompletion()

	// We either have a completer, or we use the normal
	// if we are not currently completing the history.
	if rl.completer != nil {
		rl.completer()
	} else if len(rl.histHint) == 0 {
		rl.normalCompletions()
	}
}
