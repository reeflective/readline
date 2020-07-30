package readline

import (
	"fmt"
	"strconv"
)

// CompletionGroup - A group of items offered to completion, by category.
// The output, if there are multiple groups available for a given completion input,
// will look like ZSH's completion system.
// The type is exported, because it will be easier to populate groups from the outside,
// then gather them and pass them as parameters to the TabCompleter function.
type CompletionGroup struct {
	Name        string
	Description string

	// Same as readline old system
	Suggestions  []string
	Descriptions map[string]string // Items descriptions
	DisplayType  TabDisplayType    // Map, list or normal
	MaxLength    int               // Each group can be limited in the number of comps offered

	// Values used by the shell
	tcPosX      int
	tcPosY      int
	tcMaxX      int
	tcMaxY      int
	tcOffset    int
	tcMaxLength int // Used when display is map/list, for determining message width

	allowCycle bool // This is true if we want to cycle through suggestions because they overflow MaxLength
	// This is set by the shell when it has detected this group is alone in the suggestions.
	// Might be the case of things like remote processes .

	current bool // This is to say we are currently cycling through this group, for highlighting choice
}

// Because the group might have different display types, we have to init and setup for the one desired
func (g *CompletionGroup) init(rl *Instance) {

	// Details common to all displays
	g.checkCycle(rl) // Based on the number of groups given to the shell, allows cycling or not

	// Details specific to tab display modes
	switch g.DisplayType {

	case TabDisplayGrid:
		g.initGrid(rl)

	case TabDisplayMap:
		g.initMap(rl)
	}

	// Here, handle all things for completion search functions
}

// initGrid - Grid display details
func (g *CompletionGroup) initGrid(rl *Instance) {

	// Max number of suggestions per line, for this group
	tcMaxLength := 1
	for i := range g.Suggestions {
		if len(rl.tcPrefix+g.Suggestions[i]) > tcMaxLength {
			tcMaxLength = len([]rune(rl.tcPrefix + g.Suggestions[i]))
		}
	}

	g.tcPosX = 1
	g.tcPosY = 1
	g.tcOffset = 0

	g.tcMaxX = GetTermWidth() / (tcMaxLength + 2)
	if rl.tcMaxX < 1 {
		rl.tcMaxX = 1 // avoid a divide by zero error
	}
	if g.MaxLength == 0 {
		g.MaxLength = 10 // Handle default value if not set
	}
	g.tcMaxY = g.MaxLength
}

// initMap - Map display details
func (g *CompletionGroup) initMap(rl *Instance) {

	// Max number of suggestions per line, for this group
	// Here, we have decided that tcMaxLength is managed by group, and not rl
	// Therefore we might have made a mistake. Keep that in mind
	g.tcMaxLength = 1
	for i := range g.Suggestions {
		if g.DisplayType == TabDisplayList {
			if len(rl.tcPrefix+g.Suggestions[i]) > g.tcMaxLength {
				g.tcMaxLength = len([]rune(rl.tcPrefix + g.Suggestions[i]))
			}

		} else {
			if len(g.Descriptions[g.Suggestions[i]]) > g.tcMaxLength {
				g.tcMaxLength = len(g.Descriptions[g.Suggestions[i]])
			}
		}
	}

	g.tcPosX = 1
	g.tcPosY = 1
	g.tcOffset = 0

	g.tcMaxX = 1
	if len(g.Suggestions) > g.MaxLength {
		// if len(suggestions) > rl.MaxTabCompleterRows {
		g.tcMaxY = g.MaxLength
		// rl.tcMaxY = rl.MaxTabCompleterRows
	} else {
		g.tcMaxY = len(g.Suggestions)
		// rl.tcMaxY = len(suggestions)
	}
}

// writeCompletion - This function produces a formatted string containing all appropriate items
// and according to display settings. This string is then appended to the main completion string.
func (g *CompletionGroup) writeCompletion(rl *Instance) (comp string) {

	// Depending on display type we produce the approriate string
	switch g.DisplayType {

	case TabDisplayGrid:
		comp = g.writeGrid(rl)
	case TabDisplayMap:
		comp = g.writeMap(rl)
	case TabDisplayList:
		comp = g.writeList(rl)
	}

	// If at the end, for whatever reason, we have a string consisting
	// only of the group's name/description, we don't append it to
	// completions and therefore return ""
	if comp == "" {
		return ""
	}
	return
}

// writeGrid - A grid completion string
func (g *CompletionGroup) writeGrid(rl *Instance) (comp string) {

	print(seqClearScreenBelow + "\r\n") // might need to be conditional, like first group only

	cellWidth := strconv.Itoa((GetTermWidth() / g.tcMaxX) - 2)
	x := 0
	y := 1

	for i := range g.Suggestions {
		x++
		if x > g.tcMaxX {
			x = 1
			y++
			if y > g.tcMaxY {
				y--
				break
			} else {
				comp += "\r\n"
				// print("\r\n")
			}
		}

		// For having a highlighted choice, we might need to set a flag like:
		// "we are currently selecting an option in this one, so highlight according to x & y"
		if g.current {
			// We use tcPosX and tcPosY because we don't care about the other groups, they don't
			// print anything important to us right now.
			if x == g.tcPosX && y == g.tcPosY {
				comp += seqBgWhite + seqFgBlack
				// print(seqBgWhite + seqFgBlack)
			}
			comp += fmt.Sprintf(" %-"+cellWidth+"s %s", rl.tcPrefix+g.Suggestions[i], seqReset)
			// printf(" %-"+cellWidth+"s %s", rl.tcPrefix+suggestions[i], seqReset)
		}
	}

	// Devise what to do with this.
	rl.tcUsedY = y

	return
}

// writeList - A list completion string
func (g *CompletionGroup) writeList(rl *Instance) (comp string) {

	termWidth := GetTermWidth()
	if termWidth < 20 {
		// terminal too small. Probably better we do nothing instead of crash
		// We are more conservative than lmorg, and push it to 20 instead of 10
		return
	}

	// Set all necessary dimensions
	maxLength := g.tcMaxLength
	if maxLength > termWidth-9 {
		maxLength = termWidth - 9
	}
	maxDescWidth := termWidth - maxLength - 4

	cellWidth := strconv.Itoa(maxLength)
	y := 0

	print(seqClearScreenBelow) // might need to be conditional, like first group only

	// Highlighting function
	highlight := func(y int) string {
		if y == g.tcPosY {
			return seqBgWhite + seqFgBlack
		}
		return ""
	}

	var item, description string
	for i := g.tcOffset; i < len(g.Suggestions); i++ {
		y++
		if y > rl.tcMaxY {
			break
		}

		item = rl.tcPrefix + g.Suggestions[i]

		if len(item) > maxLength {
			item = item[:maxLength-3] + "..."
		}

		description = g.Descriptions[g.Suggestions[i]]
		if len(description) > maxDescWidth {
			description = description[:maxDescWidth-3] + "..."
		}

		comp = fmt.Sprintf("\r\n%s %-"+cellWidth+"s %s %s",
			highlight(y), item, seqReset, description)
	}

	// Devise what to do with this.
	// We are using the Instance coordinates. Check this
	if len(g.Suggestions) < g.tcMaxX {
		rl.tcUsedY = len(g.Suggestions)
	} else {
		rl.tcUsedY = g.tcMaxY
	}

	return
}

// writeMap - A map or list completion string
func (g *CompletionGroup) writeMap(rl *Instance) (comp string) {

	termWidth := GetTermWidth()
	if termWidth < 20 {
		// terminal too small. Probably better we do nothing instead of crash
		// We are more conservative than lmorg, and push it to 20 instead of 10
		return
	}

	// Set all necessary dimensions
	maxLength := g.tcMaxLength
	if maxLength > termWidth-9 {
		maxLength = termWidth - 9
	}
	maxDescWidth := termWidth - maxLength - 4

	cellWidth := strconv.Itoa(maxLength)
	itemWidth := strconv.Itoa(maxDescWidth)
	y := 0

	print(seqClearScreenBelow) // might need to be conditional, like first group only

	// Highlighting function
	highlight := func(y int) string {
		if y == g.tcPosY {
			return seqBgWhite + seqFgBlack
		}
		return ""
	}

	// String formating
	var item, description string
	for i := g.tcOffset; i < len(g.Suggestions); i++ {
		y++
		if y > rl.tcMaxY {
			break
		}

		item = rl.tcPrefix + g.Suggestions[i]

		if len(item) > maxDescWidth {
			item = item[:maxDescWidth-3] + "..."
		}

		description = g.Descriptions[g.Suggestions[i]]
		if len(description) > maxLength {
			description = description[:maxLength-3] + "..."
		}

		comp = fmt.Sprintf("\r\n %-"+cellWidth+"s %s %-"+itemWidth+"s %s",
			description, highlight(y), item, seqReset)
	}

	// Devise what to do with this.
	// We are using the Instance coordinates. Check this
	if len(g.Suggestions) < g.tcMaxX {
		rl.tcUsedY = len(g.Suggestions)
	} else {
		rl.tcUsedY = g.tcMaxY
	}
	return
}

// checkCycle - Based on the number of groups given to the shell, allows cycling or not
func (g *CompletionGroup) checkCycle(rl *Instance) {

	if len(rl.atcGroups) == 1 {
		g.allowCycle = true
	}

	// 5 different groups might be a good but conservative beginning.
	if len(rl.atcGroups) >= 5 {
		g.allowCycle = false
	}

}
