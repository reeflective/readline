package readline

import (
	"fmt"
	"strconv"

	"github.com/evilsocket/islazy/tui"
)

// writeCompletion - This function produces a formatted string containing all appropriate items
// and according to display settings. This string is then appended to the main completion string.
func (g *CompletionGroup) writeCompletion(rl *Instance) (comp string) {

	// Avoids empty groups in suggestions
	if len(g.Suggestions) == 0 {
		return
	}

	// Depending on display type we produce the approriate string
	switch g.DisplayType {

	case TabDisplayGrid:
		comp += g.writeGrid(rl)
	case TabDisplayMap:
		comp += g.writeMap(rl)
	case TabDisplayList:
		comp += g.writeList(rl)
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

	// Print group title
	comp += fmt.Sprintf("\n%s%s%s %s\n", tui.BOLD, tui.YELLOW, g.Name, tui.RESET)

	// print(seqClearScreenBelow + "\r\n") // might need to be conditional, like first group only

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
			}
		}

		// For having a highlighted choice, we might need to set a flag like:
		// "we are currently selecting an option in this one, so highlight according to x & y"
		// We use tcPosX and tcPosY because we don't care about the other groups, they don't
		// print anything important to us right now.
		if (x == g.tcPosX && y == g.tcPosY) && (g.isCurrent) {
			comp += seqBgWhite + seqFgBlack
		}
		comp += fmt.Sprintf(" %-"+cellWidth+"s %s", rl.tcPrefix+g.Suggestions[i], seqReset)
	}

	// Add the equivalent of this group's size to final screen clearing
	rl.tcUsedY += y + 1 // + 1 for title

	return
}

// writeList - A list completion string
func (g *CompletionGroup) writeList(rl *Instance) (comp string) {

	// Print group title (changes with line returns depending on type)
	comp += fmt.Sprintf("\n%s%s%s %s", tui.BOLD, tui.YELLOW, g.Name, tui.RESET)

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

	// print(seqClearScreenBelow) // might need to be conditional, like first group only

	// Highlighting function
	highlight := func(y int) string {
		if y == g.tcPosY && g.isCurrent {
			return seqBgWhite + seqFgBlack
		}
		return ""
	}

	var item, description string
	for i := g.tcOffset; i < len(g.Suggestions); i++ {
		y++
		if y > g.tcMaxY {
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

		comp += fmt.Sprintf("\r\n%s %-"+cellWidth+"s %s %s",
			highlight(y), item, seqReset, description)
	}

	// Add the equivalent of this group's size to final screen clearing
	if len(g.Suggestions) < g.tcMaxX {
		rl.tcUsedY += len(g.Suggestions) + 1 // + 1 for title
	} else {
		rl.tcUsedY += g.tcMaxY + 1 // + 1 for title
	}

	return
}

// writeMap - A map or list completion string
func (g *CompletionGroup) writeMap(rl *Instance) (comp string) {

	// Title is not printed for history
	if rl.modeAutoFind && rl.modeTabFind && rl.searchMode == HistoryFind {
		if len(g.Suggestions) == 0 {
			rl.hintText = []rune(fmt.Sprintf("\n%s%s%s %s", tui.DIM, tui.RED, "No command history source, or empty", tui.RESET))
		}
	} else {
		// Print group title (changes with line returns depending on type)
		comp += fmt.Sprintf("\n%s%s%s %s", tui.BOLD, tui.YELLOW, g.Name, tui.RESET)
	}

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

	// print(seqClearScreenBelow) // might need to be conditional, like first group only

	// Highlighting function
	highlight := func(y int) string {
		if y == g.tcPosY && g.isCurrent {
			return seqBgWhite + seqFgBlack
		}
		return ""
	}

	// String formating
	var item, description string
	for i := g.tcOffset; i < len(g.Suggestions); i++ {
		y++
		if y > g.tcMaxY {
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

		comp += fmt.Sprintf("\r\n %-"+cellWidth+"s %s %-"+itemWidth+"s %s",
			description, highlight(y), item, seqReset)
	}

	// Add the equivalent of this group's size to final screen clearing
	if len(g.Suggestions) < g.tcMaxX {
		if rl.modeAutoFind && rl.modeTabFind && rl.searchMode == HistoryFind {
			rl.tcUsedY += len(g.Suggestions)
		} else {
			rl.tcUsedY += g.tcMaxY + 1 // + 1 for title
		}
	} else {
		if rl.modeAutoFind && rl.modeTabFind && rl.searchMode == HistoryFind {
			rl.tcUsedY += len(g.Suggestions)
		} else {
			rl.tcUsedY += g.tcMaxY + 1 // + 1 for title
		}
	}

	return
}
