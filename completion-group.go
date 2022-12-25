package readline

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	ansi "github.com/acarl005/stripansi"
)

// comps is used to structure different types of completions with different
// display types, autosuffix removal matchers, under their tag heading.
type comps struct {
	tag           string         // Printed on top of the group's completions
	values        [][]Completion // Values are grouped by aliases/rows, with computed paddings.
	noSpace       suffixMatcher  // Suffixes to remove if a space or non-nil character is entered after the completion.
	columnsWidth  []int          // Computed width for each column of completions, when aliases
	listSeparator string         // This is used to separate completion candidates from their descriptions.
	list          bool           // Force completions to be listed instead of grided
	noSort        bool           // Don't sort completions
	aliased       bool           // Are their aliased completions
	isCurrent     bool           // Currently cycling through this group, for highlighting choice
	maxLength     int            // Each group can be limited in the number of comps offered
	tcMaxLength   int            // Used when display is map/list, for determining message width
	maxDescWidth  int
	maxCellLength int
	tcPosX        int
	tcPosY        int
	tcMaxX        int
	tcMaxY        int
	tcOffset      int
}

//
// Initialization-time functions ----------------------------------------------------------------------------
//

func (rl *Instance) newGroup(c Completions, tag string, vals rawValues, aliased bool) {
	grp := &comps{
		tag:           tag,
		noSpace:       c.noSpace,
		listSeparator: "--",
		tcPosX:        -1,
		tcPosY:        -1,
		tcOffset:      0,
		aliased:       aliased,
		columnsWidth:  []int{0},
	}

	// Check that all comps have a display value,
	// and begin computing some parameters.
	vals = grp.checkDisplays(vals)

	// Override grid/list displays
	_, grp.list = c.listLong[tag]
	if _, all := c.listLong["*"]; all && len(c.listLong) == 1 {
		grp.list = true
	}

	// Override sorting
	_, grp.noSort = c.noSort[tag]
	if _, all := c.noSort["*"]; all && len(c.noSort) == 1 {
		grp.noSort = true
	}

	// Sort completions by default
	if !grp.noSort {
		sort.Slice(vals, func(i, j int) bool {
			return vals[i].Display < vals[j].Display
		})
	}

	// Keep computing/devising some parameters and constraints.
	// This does not do much when we have aliased completions.
	grp.computeCells(vals)

	// Rearrange all candidates into a matrix of completions,
	// based on most parameters computed above.
	grp.makeMatrix(vals)

	rl.tcGroups = append(rl.tcGroups, grp)
}

func (g *comps) checkDisplays(vals rawValues) rawValues {
	for index, val := range vals {
		if val.Display == "" {
			vals[index].Display = val.Value
		}

		// If we have aliases, the padding will be computed later.
		// Don't concatenate the description to the value as display.
		if g.aliased {
			continue
		}

		// Otherwise update the size of the longest candidate
		valLen := utf8.RuneCountInString(val.Display)
		if valLen > g.columnsWidth[0] {
			g.columnsWidth[0] = valLen
		}
	}

	return vals
}

func (g *comps) makeMatrix(vals rawValues) {
NEXT_VALUE:
	for _, val := range vals {
		valLen := utf8.RuneCountInString(val.Display)

		// If we have an alias, and we must get the right
		// column and the right padding for this column.
		if g.aliased {
			for i, row := range g.values {
				if row[0].Description == val.Description {
					g.values[i] = append(row, val)
					g.columnsWidth = getColumnPad(g.columnsWidth, valLen, len(g.values[i]))

					continue NEXT_VALUE
				}
			}
		}

		// Else, either add it to the current row if there is still room
		// on it for this candidate, or add a new one. We only do that when
		// we know we don't have aliases, or when we don't have to display list.
		if !g.aliased && g.canFitInRow(val) && !g.list {
			g.values[len(g.values)-1] = append(g.values[len(g.values)-1], val)
		} else {
			// Else create a new row, and update the row pad.
			g.values = append(g.values, []Completion{val})
			if g.columnsWidth[0] < valLen+1 {
				g.columnsWidth[0] = valLen + 1
			}
		}
	}

	if g.aliased {
		g.tcMaxX = len(g.columnsWidth)
		g.tcMaxLength = sum(g.columnsWidth) + len(g.columnsWidth)
	}

	g.tcMaxY = len(g.values)
	if g.tcMaxY > g.maxLength && g.maxLength != 0 {
		g.tcMaxY = g.maxLength
	}
}

func (g *comps) computeCells(vals rawValues) {
	// Aliases will compute themselves individually, later.
	if g.aliased {
		return
	}

	g.tcMaxLength = g.columnsWidth[0]

	// Each value first computes the total amount of space
	// it is going to take in a row (including the description)
	for _, val := range vals {
		candidate := g.displayTrimmed(val.Display)
		pad := g.tcMaxLength - len(candidate)
		desc := g.descriptionTrimmed(val.Description)
		display := fmt.Sprintf("%s%s%s", candidate, strings.Repeat(" ", pad)+" ", desc)
		valLen := utf8.RuneCountInString(display)
		if valLen > g.maxCellLength {
			g.maxCellLength = valLen
		}
	}

	g.tcMaxX = GetTermWidth() / (g.maxCellLength + 2)
	if g.tcMaxX < 1 {
		g.tcMaxX = 1 // avoid a divide by zero error
	}

	if g.tcMaxX > len(vals) {
		g.tcMaxX = len(vals)
	}

	// We also have the width for each column
	g.columnsWidth = make([]int, GetTermWidth()/(g.maxCellLength+2))
	for i := 0; i < g.tcMaxX; i++ {
		g.columnsWidth[i] = g.maxCellLength
	}
}

// checkMaxLength - Based on the number of groups given to the shell, check/set MaxLength defaults
func (g *comps) checkMaxLength(rl *Instance) {
	// This means the user forgot to set it
	if g.maxLength == 0 {
		if len(rl.tcGroups) < 5 {
			g.maxLength = 20
		}

		if len(rl.tcGroups) >= 5 {
			g.maxLength = 20
		}
	}
}

func (g *comps) canFitInRow(val Completion) bool {
	if len(g.values) == 0 {
		return false
	}

	if GetTermWidth()/(g.maxCellLength)-1 < len(g.values[len(g.values)-1]) {
		return false
	}

	return true
}

// updateIsearch - When searching through all completion groups (whether it be command history or not),
// we ask each of them to filter its own items and return the results to the shell for aggregating them.
// The rx parameter is passed, as the shell already checked that the search pattern is valid.
func (g *comps) updateIsearch(rl *Instance) {
	if rl.isearch == nil {
		return
	}

	suggs := make(rawValues, 0)
	for i := range g.values {
		row := g.values[i]

		for _, val := range row {
			if rl.isearch.MatchString(val.Value) {
				suggs = append(suggs, val)
			} else if val.Description != "" && rl.isearch.MatchString(val.Description) {
				suggs = append(suggs, val)
			}
		}
	}

	// Reset the group parameters
	g.values = make([][]Completion, 0)
	g.tcPosX = -1
	g.tcPosY = -1
	g.tcOffset = 0
	g.columnsWidth = []int{0}

	// Assign the filtered values
	vals, _, aliased := groupValues(suggs)
	g.aliased = aliased

	if len(vals) == 0 {
		return
	}

	// And perform the usual initialization routines.
	vals = g.checkDisplays(vals)
	g.computeCells(vals)
	g.makeMatrix(vals)
}

//
// Usage-time functions (selecting/writing) 9----------------------------------------------------------------
//

func (g *comps) firstCell() {
	g.tcPosX = 0
	g.tcPosY = 0
	g.tcOffset = 0
}

func (g *comps) lastCell() {
	g.tcPosY = len(g.values) - 1
	g.tcPosX = len(g.columnsWidth) - 1
	g.tcOffset = 0

	if g.aliased {
		g.findFirstCandidate(0, -1)
	} else {
		g.tcPosX = len(g.values[g.tcPosY]) - 1
	}
}

func (g *comps) selected() (comp Completion) {
	if g.tcPosY == -1 || g.tcPosX == -1 {
		return g.values[0][0]
	}
	return g.values[g.tcPosY][g.tcPosX]
}

func (g *comps) writeComps(rl *Instance) (comp string) {
	if g.tag != "" {
		comp += fmt.Sprintf("%s%s%s %s\n", seqBold, seqFgYellow, g.tag, seqReset)
		rl.tcUsedY++
	}

	// Base parameters
	var x, y int

	for range g.values {
		// Generate the completion string for this row (comp/aliases
		// and/or descriptions), and apply any styles and isearch
		// highlighting with pattern replacement,
		comp += g.writeRow(rl, x, y)

		x++
		y++
		if y > g.tcMaxY {
			break
		}
	}

	// Always add a newline to the group if
	// the end if not punctuated with one.
	if !strings.HasSuffix(comp, "\n") {
		comp += "\n"
	}

	rl.tcUsedY += y

	return
}

func (g *comps) moveSelector(rl *Instance, x, y int) (done, next bool) {
	// When the group has not yet been used, adjust
	if g.tcPosX == -1 && g.tcPosY == -1 {
		if x > 0 {
			g.tcPosY++
		} else {
			g.tcPosX++
		}
	}

	g.tcPosX += x
	g.tcPosY += y
	reverse := (x < 0 || y < 0)

	// 1) Ensure columns is minimum one, if not, either
	// go to previous row, or go to previous group.
	if g.tcPosX < 0 {
		if g.tcPosY == 0 && reverse {
			g.tcPosX = 0
			g.tcPosY = 0
			return true, false
		}
		g.tcPosY--
		g.tcPosX = len(g.values[g.tcPosY]) - 1
	}

	// 2) If we are reverse-cycling and currently on the first candidate,
	// we are done with this group. Stay on those coordinates still.
	if g.tcPosY < 0 {
		if g.tcPosX == 0 {
			g.tcPosX = 0
			g.tcPosY = 0
			return true, false
		}
		g.tcPosY = len(g.values) - 1
		g.tcPosX--
	}

	// If we are on the last row, we might have to move to
	// the next column, if there is another one.
	if g.tcPosY > g.tcMaxY-1 {
		g.tcPosY = 0
		if g.tcPosX < len(g.values[g.tcPosY])-1 {
			g.tcPosX++
		} else {
			return true, true
		}
	}

	// If we are on the last column, go to next row or next group
	if g.tcPosX > len(g.values[g.tcPosY])-1 {
		if g.aliased {
			return g.findFirstCandidate(x, y)
		}
		g.tcPosX = 0
		if g.tcPosY < g.tcMaxY-1 {
			g.tcPosY++
		} else {
			return true, true
		}
	}

	// By default, come back to this group for next item.
	return false, false
}

// Check that there is indeed a completion in the column for a given row,
// otherwise loop in the direction wished until one is found, or go next/
// previous column, and so on.
func (g *comps) findFirstCandidate(x, y int) (done, next bool) {
	for g.tcPosX > len(g.values[g.tcPosY])-1 {
		g.tcPosY += y
		g.tcPosY += x

		// Previous column or group
		if g.tcPosY < 0 {
			if g.tcPosX == 0 {
				g.tcPosX = 0
				g.tcPosY = 0
				return true, false
			} else {
				g.tcPosY = len(g.values) - 1
				g.tcPosX--
			}
		}

		// Next column or group
		if g.tcPosY > g.tcMaxY-1 {
			g.tcPosY = 0
			if g.tcPosX < len(g.columnsWidth)-1 {
				g.tcPosX++
			} else {
				return true, true
			}
		}
	}

	return
}

func (g *comps) writeRow(rl *Instance, x, y int) (comp string) {
	current := g.values[y]

	writeDesc := func(val Completion, x, y, pad int) string {
		desc := g.highlightDescription(rl, val, y, x)
		descPad := g.padDescription(val, pad)
		desc = fmt.Sprintf("%s%s", desc, strings.Repeat(" ", descPad))

		return desc
	}

	for i, val := range current {
		// Generate the highlighted candidate with its padding
		isSelected := y == g.tcPosY && i == g.tcPosX && g.isCurrent
		cell, pad := g.padCandidate(current, val, i)
		comp += g.highlightCandidate(rl, val, cell, pad, isSelected) + " "

		// And append the description only if at the end of the row,
		// or if there are no aliases, in which case all comps are described.
		if !g.aliased || i == len(current)-1 {
			comp += writeDesc(val, i, y, len(cell)+len(pad))
		}
	}

	comp += "\r\n"

	return
}

func (g *comps) highlightCandidate(rl *Instance, val Completion, cell, pad string, selected bool) (candidate string) {
	reset := sgrStart + val.Style + sgrEnd
	candidate = g.displayTrimmed(val.Display) + cell

	if rl.local == isearch && rl.isearch != nil && len(rl.tfLine) > 0 {
		match := rl.isearch.FindString(candidate)
		match = seqBgBlackBright + match + seqReset + reset
		candidate = rl.isearch.ReplaceAllLiteralString(candidate, match)
	}

	switch {
	// If the comp is currently selected, overwrite any highlighting already applied.
	case selected:
		candidate = seqCtermFg255 + seqFgBlackBright + ansi.Strip(val.Display)
		if g.aliased {
			candidate += cell + seqReset
		}

	default:
		candidate = sgrStart + val.Style + sgrEnd + candidate + seqReset
	}

	candidate += pad

	return
}

func (g *comps) highlightDescription(rl *Instance, val Completion, y, x int) (desc string) {
	if val.Description == "" {
		return seqReset
	}

	desc = g.descriptionTrimmed(val.Description)

	if rl.local == isearch && rl.isearch != nil && len(rl.tfLine) > 0 {
		match := rl.isearch.FindString(desc)
		match = seqBgBlackBright + match + seqReset + seqDim
		desc = rl.isearch.ReplaceAllLiteralString(desc, match)
	}

	switch {
	// If the comp is currently selected, overwrite any highlighting already applied.
	case y == g.tcPosY && x == g.tcPosX && g.isCurrent && !g.aliased:
		desc = seqCtermFg255 + seqFgBlackBright + g.descriptionTrimmed(val.Description)
	}

	desc = seqDim + g.listSeparator + " " + desc + seqReset

	return desc
}

func (g *comps) padCandidate(row []Completion, val Completion, x int) (cell, pad string) {
	cellLen, padLen := 0, 0
	valLen := utf8.RuneCountInString(val.Display)

	if !g.aliased {
		padLen = g.tcMaxLength - valLen
		return "", strings.Repeat(" ", padLen)
	}

	cellLen = g.columnsWidth[x] - valLen

	if len(row) == 1 {
		padLen = sum(g.columnsWidth) + len(g.columnsWidth) - g.columnsWidth[x] - 1
	}

	return strings.Repeat(" ", cellLen), strings.Repeat(" ", padLen)
}

func (g *comps) padDescription(val Completion, valPad int) (pad int) {
	if g.aliased {
		return 1
	}

	candidateLen := len(g.displayTrimmed(val.Display)) + valPad + 1
	individualRest := (GetTermWidth() % g.maxCellLength) / (g.tcMaxX + 3)
	pad = g.maxCellLength - candidateLen - len(g.descriptionTrimmed(val.Description)) + individualRest
	if pad > 1 {
		pad--
	}
	return pad
}

func (g *comps) displayTrimmed(val string) string {
	termWidth := GetTermWidth()
	if g.tcMaxLength > termWidth-9 {
		g.tcMaxLength = termWidth - 9
	}

	if len(val) > g.tcMaxLength {
		val = val[:g.tcMaxLength-3] + "..."
	}

	return val
}

func (g *comps) descriptionTrimmed(desc string) string {
	if desc == "" {
		return desc
	}

	termWidth := GetTermWidth()
	if g.tcMaxLength > termWidth-9 {
		g.tcMaxLength = termWidth - 9
	}
	g.maxDescWidth = termWidth - g.tcMaxLength - 6 // TODO: Replace 4 by length of separator.

	if len(desc) > g.maxDescWidth {
		desc = desc[:g.maxDescWidth-3] + "..."
	}

	return desc
}
