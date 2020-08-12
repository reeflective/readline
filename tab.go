package readline

import (
	"fmt"

	"github.com/evilsocket/islazy/tui"
)

// This file gathers all alterative tab completion functions, therefore is not separated in files like
// tabgrid.go, tabmap.go, etc., because in this new setup such a structure and distinction is now irrelevant.

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

// getTabCompletion - This root function sets up all completion items and engines,
// dealing with all search and completion modes. It also sets/checks various values.
func (rl *Instance) getTabCompletion() {
	rl.tcOffset = 0

	if rl.TabCompleter == nil {
		return // No completions to offer
	}

	// Populate for History search if in this mode
	if rl.modeAutoFind && rl.searchMode == HistoryFind {
		rl.getHistorySearchCompletion()
	}

	// Populate for completion search if in this mode
	if rl.searchMode == CompletionFind {
		rl.getTabSearchCompletion()
	}

	// Not in either search mode, just yield completions
	if !rl.modeAutoFind {
		rl.getNormalCompletion()
	}

	// If no completions available, return
	if len(rl.tcGroups) == 0 {
		return
	}
	rl.tcGroups = checkNilItems(rl.tcGroups) // Avoid nil maps in groups
	rl.tcGroups[0].isCurrent = true          // Init this for starting comp select somewhere

	// Init/Setup all groups with their priting details
	for _, group := range rl.tcGroups {
		group.init(rl)
	}

}

// writeTabCompletion - Prints all completion groups and their items
func (rl *Instance) writeTabCompletion() {

	// We adjust for a supplementary prompt line
	// It NEEDS to precede the following line, because its effect is immediate
	if rl.Multiline {
		rl.tcUsedY++
	}
	// This stablizes the completion printing just beyond the input line
	rl.tcUsedY -= rl.tcUsedY

	if !rl.modeTabCompletion {
		return
	}

	// Each group produces its own string, added to the main one
	var completions string
	for _, group := range rl.tcGroups {
		completions += group.writeCompletion(rl)
	}

	// Then we print all of them.
	fmt.Printf(completions)
}

// getTabSearchCompletion - Populates and sets up completion for completion search
func (rl *Instance) getTabSearchCompletion() {

	rl.tcPrefix, rl.tcGroups = rl.TabCompleter(rl.line, rl.pos)

	// Handle empty list
	if len(rl.tcGroups) == 0 {
		return
	}

	for _, g := range rl.tcGroups {
		g.updateTabFind(rl)
	}
}

// getHistorySearchCompletion - Populates and sets up completion for command history search
func (rl *Instance) getHistorySearchCompletion() {
	rl.tcGroups = rl.completeHistory() // Refresh full list each time

	// Handle empty list
	if len(rl.tcGroups) == 0 {
		return
	}
	rl.tcGroups[0].DisplayType = TabDisplayMap // History is always shown as map

	if len(rl.tcGroups[0].Suggestions) == 0 {
		rl.hintText = []rune(fmt.Sprintf("%s%s%s %s", tui.DIM, tui.RED, "No command history source, or empty", tui.RESET))
	}

	if rl.regexSearch.String() != "(?i)" {
		rl.tcGroups[0].updateTabFind(rl) // Refresh filtered candidates
	}
}

// getNormalCompletion - Populates and sets up completion for normal comp mode
func (rl *Instance) getNormalCompletion() {
	rl.tcPrefix, rl.tcGroups = rl.TabCompleter(rl.line, rl.pos)

	// Handle empty list
	if len(rl.tcGroups) == 0 {
		return
	}
	rl.tcGroups[0].isCurrent = true
}

func (rl *Instance) getCurrentGroup() (group *CompletionGroup) {
	for _, g := range rl.tcGroups {
		if g.isCurrent {
			return g
		}
	}
	return
}

// getScreenCleanSize - not used
func (rl *Instance) getScreenCleanSize() (size int) {
	for _, g := range rl.tcGroups {
		size++ // Group title
		size += g.tcPosY
	}
	return
}

func (rl *Instance) resetTabCompletion() {
	rl.modeTabCompletion = false
	rl.tcOffset = 0
	rl.tcUsedY = 0
	rl.modeTabFind = false
	rl.modeAutoFind = false
	rl.tfLine = []rune{}

	// Reset tab highlighting
	if len(rl.tcGroups) > 0 {
		for _, g := range rl.tcGroups {
			g.isCurrent = false
		}
		rl.tcGroups[0].isCurrent = true

	}
}

// checkNilItems - For each completion group we avoid nil maps and possibly other items
func checkNilItems(groups []*CompletionGroup) (checked []*CompletionGroup) {

	for _, grp := range groups {
		if grp.Descriptions == nil || len(grp.Descriptions) == 0 {
			grp.Descriptions = make(map[string]string)
		}
		checked = append(checked, grp)
	}

	return
}
