package readline

import (
	"fmt"
	"strconv"
)

// initMap - Map display details. Called each time we want to be sure to have
// a working completion group either immediately, or later on. Generally defered.
func (g *CompletionGroup) initMap(rl *Instance) {
	g.grouped, g.columnsWidth, g.rows = g.groupValues()

	// Compute size of each completion item box. Group independent
	g.tcMaxLength = 1
	for _, val := range g.Values {
		if len(val.Description) > g.tcMaxLength {
			g.tcMaxLength = len(val.Description)
		}
	}

	g.tcPosX = 0
	g.tcPosY = 0
	g.tcOffset = 0

	// Number of lines allowed to be printed for group
	if len(g.Values) > g.MaxLength {
		g.tcMaxY = g.MaxLength
	} else {
		g.tcMaxY = len(g.Values)
	}
}

// moveTabMapHighlight - Moves the highlighting for currently selected completion item (map display)
func (g *CompletionGroup) moveTabMapHighlight(rl *Instance, x, y int) (done bool, next bool) {
	g.tcPosY += x
	// g.tcPosY += y

	// Lines
	if g.tcPosY < 1 {
		if x < 0 || y < 0 {
			if g.tcOffset > 0 {
				g.tcPosY = 1
				g.tcOffset--
			} else {
				return true, false
			}
		}
	}
	if g.tcPosY > g.tcMaxY {
		g.tcPosY--
		g.tcOffset++
	}

	if g.tcOffset+g.tcPosY < 1 && len(g.Values) > 0 {
		g.tcPosY = g.tcMaxY
		g.tcOffset = len(g.Values) - g.tcMaxY
	}
	if g.tcOffset < 0 {
		g.tcOffset = 0
	}

	if g.tcOffset+g.tcPosY > len(g.Values) {
		g.tcOffset--
		return true, true
	}
	return false, false
}

// writeMap - A map or list completion string
func (g *CompletionGroup) writeMap(rl *Instance) (comp string) {
	if g.Name != "" {
		// Print group title (changes with line returns depending on type)
		comp += fmt.Sprintf("%s%s%s %s\n", seqBold, seqFgYellow, g.Name, seqReset)
		rl.tcUsedY++
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

	// Generate the aggregated completions block as a string.
	comps, usedY := g.buildMap(rl, maxLength, maxDescWidth)
	comp += comps
	rl.tcUsedY += usedY

	// Add the equivalent of this group's size to final screen clearing
	if len(g.Values) > g.MaxLength {
		rl.tcUsedY += g.MaxLength
	}

	return
}

func (g *CompletionGroup) buildMap(rl *Instance, maxLen, maxDescLen int) (comp string, y int) {
	cellWidth := strconv.Itoa(maxLen)
	itemWidth := strconv.Itoa(maxDescLen)

	for i := g.tcOffset; i < len(g.Values); i++ {
		y++ // Consider new item
		if y > g.tcMaxY {
			break
		}

		val := g.Values[i]
		item := val.Display

		if len(item) > maxDescLen {
			item = item[:maxDescLen-3] + "..."
		}

		styling := g.highlight(val.Style, y, g.tcPosX)
		item = g.isearchHighlight(rl, item, styling, y, g.tcPosX)

		description := val.Description
		if len(description) > maxLen {
			description = description[:maxLen-3] + "..."
		}

		comp += fmt.Sprintf("\r%-"+cellWidth+"s %s %-"+itemWidth+"s %s\n",
			description, styling, item, seqReset)
	}

	return
}
