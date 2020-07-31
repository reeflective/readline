package readline

import "fmt"

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

// getTabCompletion - This root function sets up all completion items and engines.
func (rl *Instance) getTabCompletion() {
	rl.tcOffset = 0

	if rl.Completer == nil {
		return // No completions to offer
	}

	// Populate for History search if in this mode
	if rl.modeAutoFind && rl.modeTabFind && rl.regexpMode == HistoryFind {

		rl.tcGroups = rl.completeHistory()         // Refresh full list each time
		rl.tcGroups[0].DisplayType = TabDisplayMap // History is always shown as map

		if rl.regexSearch.String() != "(?i)" {
			rl.tcGroups[0].updateTabFind(rl) // Refresh filtered candidates
		}
	}

	// Populate for completion search if in this mode
	if rl.modeAutoFind && rl.modeTabFind && rl.regexpMode == CompletionFind {

		// for _, g := range rl.tcGroups {
		//         fmt.Println(g)
		// }

		rl.tcPrefix, rl.tcGroups = rl.Completer(rl.line, rl.pos)

		for _, g := range rl.tcGroups {
			// fmt.Println(g.Suggestions)
			g.updateTabFind(rl)
		}
	}

	// Not in either search mode, just yield completions
	if !rl.modeAutoFind {
		rl.tcPrefix, rl.tcGroups = rl.Completer(rl.line, rl.pos)

		// Here we initialize some values for moving completion selection
		rl.tcGroups[0].isCurrent = true
	}

	// If no completions available, return
	if len(rl.tcGroups) == 0 {
		return
	}
	// Add here, if all groups are empty, don't display them

	// Avoid nil maps in groups
	rl.tcGroups = checkNilItems(rl.tcGroups)

	// Init/Setup all groups
	// This tells what each group is able to do and what not, etc...
	for _, group := range rl.tcGroups {
		group.init(rl)
	}

	// Here we initialize some values for moving completion selection
	rl.tcGroups[0].isCurrent = true
}

// moveTabCompletionHighlight - This function is in charge of highlighting the current completion item.
func (rl *Instance) moveTabCompletionHighlight(x, y int) {

	g := rl.getCurrentGroup()

	// We keep track of these values
	// ty := &g.tcPosY

	// This is triggered when we need to cycle through the next group
	var done bool

	switch g.DisplayType {
	// Depending on the display, we only keep track of x or (x and y)
	case TabDisplayGrid:
		done = g.aMoveTabGridHighlight(rl, x, y)

	case TabDisplayList:
		done = g.aMoveTabMapHighlight(x, y)

	case TabDisplayMap:
		done = g.aMoveTabMapHighlight(x, y)
	}

	// Cycle to next group: we tell them who is the next one to handle highlighting
	if done {
		for i, g := range rl.tcGroups {
			if g.isCurrent {
				g.isCurrent = false
				if i == len(rl.tcGroups)-1 {
					rl.tcGroups[0].isCurrent = true
				} else {
					rl.tcGroups[i+1].isCurrent = true
				}
				break
			}
		}
	}

}

func (rl *Instance) getCurrentGroup() (group *CompletionGroup) {
	for _, g := range rl.tcGroups {
		if g.isCurrent {
			return g
		}
	}
	return
}

// writeTabCompletion - Prints all completion groups and their items
func (rl *Instance) writeTabCompletion() {

	// This stablizes the completion printing just beyond the input line
	rl.tcUsedY -= rl.tcUsedY

	if !rl.modeTabCompletion {
		return
	}

	// This is the final string, with all completions of all groups, to be printed
	var completions string

	// Each group produces its own string, added to the main one
	for _, group := range rl.tcGroups {
		completions += group.writeCompletion(rl)
	}

	// Then we print it
	fmt.Printf(completions)

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
