package readline

import "fmt"

// This file gathers all alterative tab completion functions, therefore is not separated in files like
// tabgrid.go, tabmap.go, etc., because in this new setup such a structure and distinction is now irrelevant.

// THINGS TO TAKE CARE OF MANDATORILY !!!!!!!!!
//
// 1) The completion search function also has a unique list of items, which are populated at some point.
//    We have to modify it so that filtering is still done per-group, and ideally it should not modify
//    they individual display types

// agetTabCompletion - This root function sets up all completion items and engines.
func (rl *Instance) agetTabCompletion() {
	rl.tcOffset = 0
	if rl.TabCompleter == nil {
		return
	}

	rl.tcPrefix, rl.atcGroups = rl.Completer(rl.line, rl.pos)

	// If no completions available, return
	if len(rl.atcGroups) == 0 {
		return
	}
	// Add here, if all groups are empty, don't display them

	// Avoid nil maps in groups
	rl.atcGroups = checkNilItems(rl.atcGroups)

	// Init/Setup all groups
	// This tells what each group is able to do and what not, etc...
	for _, group := range rl.atcGroups {
		group.init(rl)
	}
}

// amoveTabCompletionHighlight - This function is in charge of highlighting the current completion item.
func (rl *Instance) amoveTabCompletionHighlight(x, y int) {

	// This function works differently from

}

// awriteTabCompletion - Prints all completion groups and their items
func (rl *Instance) awriteTabCompletion() {

	if !rl.modeTabCompletion {
		return
	}

	// This is the final string, with all completions of all groups, to be printed
	var completions string

	// Each group produces its own string.
	for _, group := range rl.atcGroups {
		completions += group.writeCompletion(rl) // We add it to the big completion string
	}

	// We modify a few things if needed

	// Then we print it
	fmt.Printf(completions)
}

func (rl *Instance) aResetTabCompletion() {
	rl.modeTabCompletion = false
	rl.tcOffset = 0
	rl.tcUsedY = 0
	rl.modeTabFind = false
	rl.tfLine = []rune{}
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
