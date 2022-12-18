package readline

import (
	"strings"
)

// CompletionValue represents a completion candidate
type CompletionValue struct {
	Value       string
	Display     string
	Description string
	Style       string
}

// CompletionGroup - A group/category of items offered to completion, with its own
// name, descriptions and completion display format/type.
// The output, if there are multiple groups available for a given completion input,
// will look like ZSH's completion system.
type CompletionGroup struct {
	Name          string            // Printed on top of the group's completions
	DisplayType   TabDisplayType    // Map, list or normal
	MaxLength     int               // Each group can be limited in the number of comps offered
	Values        []CompletionValue // All candidates with their styles and descriptions.
	SuffixMatcher []rune            // Suffixes to remove if a space or non-nil character is entered after the completion.
	ListSeparator string            // This is used to separate completion candidates from their descriptions.

	// Internal parameters
	grouped        [][]CompletionValue // Values are grouped by aliases/rows, with computed paddings.
	columnsWidth   []int               // Computed width for each column of completions, when aliases
	selected       CompletionValue     // The currently selected completion in this group
	tcMaxLength    int                 // Used when display is map/list, for determining message width
	tcMaxLengthAlt int                 // Same as tcMaxLength but for SuggestionsAlt.
	allowCycle     bool                // Cycle through suggestions because they overflow MaxLength
	isCurrent      bool                // Currently cycling through this group, for highlighting choice
	minCellLength  int
	maxCellLength  int
	rows           int
	tcPosX         int
	tcPosY         int
	tcMaxX         int
	tcMaxY         int
	tcOffset       int
}

// init - The completion group computes and sets all its values, and is then ready to work.
func (g *CompletionGroup) init(rl *Instance) {
	// Details common to all displays
	g.checkCycle(rl)
	g.checkMaxLength(rl)

	// Details specific to tab display modes
	switch g.DisplayType {
	case TabDisplayGrid:
		g.initGrid(rl)
	case TabDisplayMap:
		g.initMap(rl)
	case TabDisplayList:
		g.initList(rl)
	}
}

// updateTabFind - When searching through all completion groups (whether it be command history or not),
// we ask each of them to filter its own items and return the results to the shell for aggregating them.
// The rx parameter is passed, as the shell already checked that the search pattern is valid.
func (g *CompletionGroup) updateTabFind(rl *Instance) {
	if rl.regexSearch == nil {
		return
	}

	var suggs []CompletionValue

	// We perform filter right here, so we create a new
	// completion group, and populate it with our results.
	for i := range g.Values {
		value := g.Values[i]

		if rl.regexSearch.MatchString(value.Value) {
			suggs = append(suggs, value)
		} else if g.DisplayType == TabDisplayList && rl.regexSearch.MatchString(value.Description) {
			suggs = append(suggs, value)
		}
	}

	// We overwrite the group's items, (will be refreshed
	// as soon as something is typed in the search)
	g.Values = suggs

	// Finally, the group computes its new printing settings
	g.init(rl)
}

// checkCycle - Based on the number of groups given to the shell, allows cycling or not
func (g *CompletionGroup) checkCycle(rl *Instance) {
	if len(rl.tcGroups) == 1 {
		g.allowCycle = true
	}
	if len(rl.tcGroups) >= 10 {
		g.allowCycle = false
	}
}

// checkMaxLength - Based on the number of groups given to the shell, check/set MaxLength defaults
func (g *CompletionGroup) checkMaxLength(rl *Instance) {
	// This means the user forgot to set it
	if g.MaxLength == 0 {
		if len(rl.tcGroups) < 5 {
			g.MaxLength = 20
		}

		if len(rl.tcGroups) >= 5 {
			g.MaxLength = 20
		}

		// Lists that have a alternative completions are not allowed to have
		// MaxLength set, because rolling does not work yet.
		if g.DisplayType == TabDisplayList {
			g.MaxLength = 1000 // Should be enough not to trigger anything related.
		}
	}
}

// writeCompletion - This function produces a formatted string containing all appropriate items
// and according to display settings. This string is then appended to the main completion string.
func (g *CompletionGroup) writeCompletion(rl *Instance) (comp string) {
	// Avoids empty groups in suggestions
	if len(g.Values) == 0 {
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
	return
}

// getCurrentCell - The completion groups computes the current cell value,
// depending on its display type and its different parameters
func (g *CompletionGroup) getCurrentCell(rl *Instance) CompletionValue {
	switch g.DisplayType {
	case TabDisplayGrid:
		// x & y coodinates + safety check
		cell := (g.tcMaxX * (g.tcPosY - 1)) + g.tcOffset + g.tcPosX - 1
		if cell < 0 {
			cell = 0
		}

		if cell < len(g.Values) {
			return g.Values[cell]
		}
		return CompletionValue{}

	case TabDisplayMap:
		// x & y coodinates + safety check
		cell := g.tcOffset + g.tcPosY - 1
		if cell < 0 {
			cell = 0
		}

		// TODO: Here we didn't ensure some values have not the same description
		sugg := g.Values[cell]
		return sugg

	case TabDisplayList:
		return g.selected
	}

	// We should never get here
	return CompletionValue{}
}

func (g *CompletionGroup) goFirstCell() {
	switch g.DisplayType {
	case TabDisplayGrid:
		g.tcPosX = 1
		g.tcPosY = 1

	case TabDisplayList:
		g.tcPosX = 0
		g.tcPosY = 1
		g.tcOffset = 0

	case TabDisplayMap:
		g.tcPosX = 0
		g.tcPosY = 1
		g.tcOffset = 0
	}
}

func (g *CompletionGroup) goLastCell() {
	switch g.DisplayType {
	case TabDisplayGrid:
		g.tcPosY = g.tcMaxY

		restX := len(g.Values) % g.tcMaxX
		if restX != 0 {
			g.tcPosX = restX
		} else {
			g.tcPosX = g.tcMaxX
		}

		// We need to adjust the X position depending
		// on the interpretation of the remainder with
		// respect to the group's MaxLength.
		restY := len(g.Values) % g.tcMaxY
		maxY := len(g.Values) / g.tcMaxX
		if restY == 0 && maxY > g.MaxLength {
			g.tcPosX = g.tcMaxX
		}
		if restY != 0 && maxY > g.MaxLength-1 {
			g.tcPosX = g.tcMaxX
		}

	case TabDisplayList:
		// By default, the last item is at maxY
		g.tcPosY = g.tcMaxY - 1
		//
		// // If the max length is smaller than the number
		// // of suggestions, we need to adjust the offset.
		// if g.rows > g.MaxLength {
		// 	g.tcOffset = g.rows - g.tcMaxY
		// }
		//
		// // We do not take into account the alternative suggestions
		// g.tcPosX = 0

		// ALTERNATIVE
		g.tcPosX = len(g.columnsWidth) - 1
		g.lastCellList()

	case TabDisplayMap:
		// By default, the last item is at maxY
		g.tcPosY = g.tcMaxY

		// If the max length is smaller than the number
		// of suggestions, we need to adjust the offset.
		if g.rows > g.MaxLength {
			g.tcOffset = g.rows - g.tcMaxY
		}

		// We do not take into account the alternative suggestions
		g.tcPosX = 0
	}
}

func (g *CompletionGroup) lastCellList() {
	remaining := g.grouped
	y := 0
	found := false

	for i := len(remaining); i > 0; i-- {
		row := remaining[i-1]

		// Adjust the first row if it has multiple subrows
		// if i == len(remaining) && inRow > 0 {
		// 	row = row[:(inRow * len(g.columnsWidth))]
		// }

		// Skip if its does not have enough columns
		if len(row)-1 < g.tcPosX {
			y++
			continue
		}

		// Else we have candidate for the given column,
		// just break since our posY has been updated.
		g.selected = row[g.tcPosX]

		found = true
		break
	}

	// If this column did not yield a candidate, perform
	// the same lookup on the previous column, starting at bottom.
	if !found && g.tcPosX > 0 {
		g.tcPosX--
		g.tcPosY = g.rows
		g.lastCellList()

		return
	}

	g.tcPosY -= y - 1
}

// numValues returns the number of unique completion values, which
// is the number of completions that do NOT have the same description.
func (g *CompletionGroup) numValues() int {
	var unique []string

	for _, value := range g.Values {
		var found bool

		// If the value has no descriptions, it will
		// not be printed along with any other value.
		if value.Description == "" {
			for _, existing := range unique {
				if existing == value.Description {
					found = true
					break
				}
			}
		}

		if !found {
			unique = append(unique, value.Description)
		}
	}

	return len(unique)
}

// checkNilItems - For each completion group we avoid nil maps and possibly other items
func checkNilItems(groups []CompletionGroup) (checked []*CompletionGroup) {
	for i := range groups {
		// if grp.Descriptions == nil || len(grp.Descriptions) == 0 {
		// 	grp.Descriptions = make(map[string]string)
		// }
		// if grp.Aliases == nil || len(grp.Aliases) == 0 {
		// 	grp.Aliases = make(map[string]string)
		// }
		checked = append(checked, &groups[i])
	}

	return
}

func (g *CompletionGroup) matchesSuffix(value string) (yes bool, suf rune) {
	for _, r := range g.SuffixMatcher {
		if r == '*' || strings.HasSuffix(value, string(r)) {
			return true, r
		}
	}
	return
}

// function highlights the cell depending on current selector place.
func (g *CompletionGroup) highlight(style string, y int, x int) string {
	if y == g.tcPosY && x == g.tcPosX && g.isCurrent {
		return seqCtermFg255 + seqFgBlackBright
	}

	return sgrStart + fgColorStart + style + sgrEnd
}

// isearchHighlight applies highlighting to all isearch matches in a completion.
func (g *CompletionGroup) isearchHighlight(rl *Instance, item, reset string, y, x int) string {
	if rl.local != isearch || rl.regexSearch == nil || len(rl.tfLine) == 0 {
		return item
	}

	if y == g.tcPosY && x == g.tcPosX && g.isCurrent {
		return item
	}

	match := rl.regexSearch.FindString(item)
	match = seqBgBlackBright + match + seqReset + reset

	return rl.regexSearch.ReplaceAllLiteralString(item, match)
}

const (
	sgrStart     = "\x1b["
	fgColorStart = "38;05;"
	bgColorStart = "48;05;"
	sgrEnd       = "m"
)
