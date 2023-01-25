package completion

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/term"
)

// group is used to structure different types of completions with different
// display types, autosuffix removal matchers, under their tag heading.
type group struct {
	tag           string        // Printed on top of the group's completions
	values        [][]Candidate // Values are grouped by aliases/rows, with computed paddings.
	noSpace       suffixMatcher // Suffixes to remove if a space or non-nil character is entered after the completion.
	columnsWidth  []int         // Computed width for each column of completions, when aliases
	listSeparator string        // This is used to separate completion candidates from their descriptions.
	list          bool          // Force completions to be listed instead of grided
	noSort        bool          // Don't sort completions
	aliased       bool          // Are their aliased completions
	isCurrent     bool          // Currently cycling through this group, for highlighting choice
	maxLength     int           // Each group can be limited in the number of comps offered
	tcMaxLength   int           // Used when display is map/list, for determining message width
	maxDescWidth  int
	maxCellLength int
	tcPosX        int
	tcPosY        int
	tcMaxX        int
	tcMaxY        int
}

func (e *Engine) groupCompletions(comps Values) {
	e.hintCompletions(comps)

	// Nothing else to do if no completions
	if len(comps.values) == 0 {
		return
	}

	comps.values.eachTag(func(tag string, values rawValues) {
		// Separate the completions that have a description and
		// those which don't, and devise if there are aliases.
		vals, noDescVals, aliased := groupValues(values)

		// Create a "first" group with the "first" grouped values
		e.newGroup(comps, tag, vals, aliased)

		// If we have a remaining group of values without descriptions,
		// we will print and use them in a separate, anonymous group.
		if len(noDescVals) > 0 {
			e.newGroup(comps, "", noDescVals, false)
		}
	})
}

// groupValues separates values based on whether they have descriptions, or are aliases of each other.
func groupValues(values rawValues) (vals, noDescVals rawValues, aliased bool) {
	var descriptions []string

	for _, val := range values {
		// Ensure all values have a display string.
		if val.Display == "" {
			val.Display = val.Value
		}

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

	return vals, noDescVals, aliased
}

func (e *Engine) newGroup(c Values, tag string, vals rawValues, aliased bool) {
	grp := &group{
		tag:           tag,
		noSpace:       c.NoSpace,
		listSeparator: "--",
		tcPosX:        -1,
		tcPosY:        -1,
		aliased:       aliased,
		columnsWidth:  []int{0},
	}

	// Check that all comps have a display value,
	// and begin computing some parameters.
	vals = grp.checkDisplays(vals)

	// Override grid/list displays
	_, grp.list = c.ListLong[tag]
	if _, all := c.ListLong["*"]; all && len(c.ListLong) == 1 {
		grp.list = true
	}

	listSep, found := c.ListSep[tag]
	if !found {
		if allSep, found := c.ListSep["*"]; found {
			grp.listSeparator = allSep
		}
	} else {
		grp.listSeparator = listSep
	}

	// Override sorting or sort if needed
	_, grp.noSort = c.NoSort[tag]
	if _, all := c.NoSort["*"]; all && len(c.NoSort) == 1 {
		grp.noSort = true
	}

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

	e.groups = append(e.groups, grp)
}

func (g *group) checkDisplays(vals rawValues) rawValues {
	if g.aliased {
		return vals
	}

	// Otherwise update the size of the longest candidate
	for _, val := range vals {
		valLen := utf8.RuneCountInString(val.Display)
		if valLen > g.columnsWidth[0] {
			g.columnsWidth[0] = valLen
		}
	}

	return vals
}

func (g *group) computeCells(vals rawValues) {
	// Aliases will compute themselves individually, later.
	if g.aliased {
		return
	}

	termWidth := term.GetWidth()
	g.tcMaxLength = g.columnsWidth[0]

	// Each value first computes the total amount of space
	// it is going to take in a row (including the description)
	for _, val := range vals {
		candidate := g.displayTrimmed(color.Strip(val.Display))
		pad := g.tcMaxLength - len(candidate)
		desc := g.descriptionTrimmed(val.Description)
		display := fmt.Sprintf("%s%s%s", candidate, strings.Repeat(" ", pad)+" ", desc)
		valLen := utf8.RuneCountInString(display)

		if valLen > g.maxCellLength {
			g.maxCellLength = valLen
		}
	}

	g.tcMaxX = termWidth / (g.maxCellLength)
	if g.tcMaxX < 1 {
		g.tcMaxX = 1 // avoid a divide by zero error
	}

	if g.tcMaxX > len(vals) {
		g.tcMaxX = len(vals)
	}

	numColumns := termWidth / (g.maxCellLength)
	if numColumns == 0 {
		numColumns = 1
	}

	// We also have the width for each column
	g.columnsWidth = make([]int, numColumns)
	for i := 0; i < g.tcMaxX; i++ {
		g.columnsWidth[i] = g.maxCellLength
	}
}

func (g *group) makeMatrix(vals rawValues) {
nextValue:
	for _, val := range vals {
		valLen := utf8.RuneCountInString(val.Display)

		// If we have an alias, and we must get the right
		// column and the right padding for this column.
		if g.aliased {
			for i, row := range g.values {
				if row[0].Description == val.Description {
					g.values[i] = append(row, val)
					g.columnsWidth = getColumnPad(g.columnsWidth, valLen, len(g.values[i]))

					continue nextValue
				}
			}
		}

		// Else, either add it to the current row if there is still room
		// on it for this candidate, or add a new one. We only do that when
		// we know we don't have aliases, or when we don't have to display list.
		if !g.aliased && g.canFitInRow(term.GetWidth()) && !g.list {
			g.values[len(g.values)-1] = append(g.values[len(g.values)-1], val)
		} else {
			// Else create a new row, and update the row pad.
			g.values = append(g.values, []Candidate{val})
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

func (g *group) canFitInRow(termWidth int) bool {
	if len(g.values) == 0 {
		return false
	}

	if termWidth/(g.maxCellLength)-1 < len(g.values[len(g.values)-1]) {
		return false
	}

	return true
}

//
// Usage-time functions (selecting/writing) 9----------------------------------------------------------------
//

func (g *group) firstCell() {
	g.tcPosX = 0
	g.tcPosY = 0
}

func (g *group) lastCell() {
	g.tcPosY = len(g.values) - 1
	g.tcPosX = len(g.columnsWidth) - 1

	if g.aliased {
		g.findFirstCandidate(0, -1)
	} else {
		g.tcPosX = len(g.values[g.tcPosY]) - 1
	}
}

func (g *group) selected() (comp Candidate) {
	if g.tcPosY == -1 || g.tcPosX == -1 {
		return g.values[0][0]
	}

	return g.values[g.tcPosY][g.tcPosX]
}

func (g *group) moveSelector(x, y int) (done, next bool) {
	// When the group has not yet been used, adjust
	if g.tcPosX == -1 && g.tcPosY == -1 {
		if x != 0 {
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

	// 3) If we are on the last row, we might have to move to
	// the next column, if there is another one.
	if g.tcPosY > g.tcMaxY-1 {
		g.tcPosY = 0
		if g.tcPosX < len(g.values[g.tcPosY])-1 {
			g.tcPosX++
		} else {
			return true, true
		}
	}

	// 4) If we are on the last column, go to next row or next group
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
func (g *group) findFirstCandidate(x, y int) (done, next bool) {
	for g.tcPosX > len(g.values[g.tcPosY])-1 {
		g.tcPosY += y
		g.tcPosY += x

		// Previous column or group
		if g.tcPosY < 0 {
			if g.tcPosX == 0 {
				g.tcPosX = 0
				g.tcPosY = 0

				return true, false
			}

			g.tcPosY = len(g.values) - 1
			g.tcPosX--
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

func (g *group) writeComps(eng *Engine) (comp string) {
	if len(g.values) == 0 {
		return
	}

	if g.tag != "" {
		comp += fmt.Sprintf("%s%s%s %s\n", color.Bold, color.FgYellow, g.tag, color.Reset)
		eng.usedY++
	}

	// Base parameters
	var columns, rows int

	for range g.values {
		// Generate the completion string for this row (comp/aliases
		// and/or descriptions), and apply any styles and isearch
		// highlighting with pattern replacement,
		comp += g.writeRow(eng, columns)

		columns++
		rows++

		if rows > g.tcMaxY {
			break
		}
	}

	// Always add a newline to the group if
	// the end if not punctuated with one.
	if !strings.HasSuffix(comp, "\n") {
		comp += "\n"
	}

	eng.usedY += rows

	return comp
}

func (g *group) writeRow(eng *Engine, row int) (comp string) {
	current := g.values[row]

	writeDesc := func(val Candidate, x, y, pad int) string {
		desc := g.highlightDescription(eng, val, y, x)
		descPad := g.padDescription(current, val, pad)
		desc = fmt.Sprintf("%s%s", desc, strings.Repeat(" ", descPad))

		return desc
	}

	for col, val := range current {
		// Generate the highlighted candidate with its padding
		isSelected := row == g.tcPosY && col == g.tcPosX && g.isCurrent
		cell, pad := g.padCandidate(current, val, col)
		comp += g.highlightCandidate(eng, val, cell, pad, isSelected) + " "

		// And append the description only if at the end of the row,
		// or if there are no aliases, in which case all comps are described.
		if !g.aliased || col == len(current)-1 {
			comp += writeDesc(val, col, row, len(cell)+len(pad))
		}
	}

	comp += "\r\n"

	return
}

func (g *group) highlightCandidate(eng *Engine, val Candidate, cell, pad string, selected bool) (candidate string) {
	reset := color.SGR(val.Style, true)
	candidate = g.displayTrimmed(val.Display) + cell

	if eng.isearch != nil && eng.isearchBuf.Len() > 0 {
		match := eng.isearch.FindString(candidate)
		match = color.BgBlackBright + match + color.Reset + reset
		candidate = eng.isearch.ReplaceAllLiteralString(candidate, match)
	}

	switch {
	// If the comp is currently selected, overwrite any highlighting already applied.
	case selected:
		candidate = color.SGR(strconv.Itoa(255), false) + color.FgBlackBright + g.displayTrimmed(color.Strip(val.Display))
		if g.aliased {
			candidate += cell + color.Reset
		}

	default:
		candidate = reset + candidate + color.Reset
	}

	candidate += pad

	return
}

func (g *group) highlightDescription(eng *Engine, val Candidate, row, col int) (desc string) {
	if val.Description == "" {
		return color.Reset
	}

	desc = g.descriptionTrimmed(val.Description)

	if eng.isearch != nil && eng.isearchBuf.Len() > 0 {
		match := eng.isearch.FindString(desc)
		match = color.BgBlackBright + match + color.Reset + color.Dim
		desc = eng.isearch.ReplaceAllLiteralString(desc, match)
	}

	// If the comp is currently selected, overwrite any highlighting already applied.
	if row == g.tcPosY && col == g.tcPosX && g.isCurrent && !g.aliased {
		desc = color.SGR(strconv.Itoa(255), false) + color.FgBlackBright + g.descriptionTrimmed(val.Description)
	}

	desc = color.Dim + g.listSeparator + " " + desc + color.Reset

	return desc
}

func (g *group) padCandidate(row []Candidate, val Candidate, col int) (cell, pad string) {
	var cellLen, padLen int
	valLen := utf8.RuneCountInString(val.Display)

	if !g.aliased {
		padLen = g.tcMaxLength - valLen
		if padLen < 0 {
			padLen = 0
		}

		return "", strings.Repeat(" ", padLen)
	}

	cellLen = g.columnsWidth[col] - valLen

	if len(row) == 1 {
		padLen = sum(g.columnsWidth) + len(g.columnsWidth) - g.columnsWidth[col] - 1
	}

	return strings.Repeat(" ", cellLen), strings.Repeat(" ", padLen)
}

func (g *group) padDescription(row []Candidate, val Candidate, valPad int) (pad int) {
	if g.aliased {
		return 1
	}

	candidateLen := len(g.displayTrimmed(val.Display)) + valPad + 1
	individualRest := (term.GetWidth() % g.maxCellLength) / (g.tcMaxX + len(row))
	pad = g.maxCellLength - candidateLen - len(g.descriptionTrimmed(val.Description)) + individualRest

	if pad > 1 {
		pad--
	}

	return pad
}

func (g *group) displayTrimmed(val string) string {
	termWidth := term.GetWidth()
	if g.tcMaxLength > termWidth-1 {
		g.tcMaxLength = termWidth - 1
	}

	if len(val) > g.tcMaxLength {
		val = val[:g.tcMaxLength-3] + "..."
	}

	val = sanitizer.Replace(val)

	return val
}

func (g *group) descriptionTrimmed(desc string) string {
	if desc == "" {
		return desc
	}

	termWidth := term.GetWidth()
	if g.tcMaxLength > termWidth {
		g.tcMaxLength = termWidth
	}

	g.maxDescWidth = termWidth - g.tcMaxLength - len(g.listSeparator) - 1

	if len(desc) > g.maxDescWidth {
		desc = desc[:g.maxDescWidth-4] + "..."
	}

	desc = sanitizer.Replace(desc)

	return desc
}
