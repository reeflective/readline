package readline

import (
	"fmt"
	"strconv"
	"strings"
)

// initList - List display details. Because of the way alternative completions
// are handled, MaxLength cannot be set when there are alternative completions.
func (g *CompletionGroup) initList(rl *Instance) {
	// Get the number of columns in which to print candidates/aliases,
	// and the max pad for any given row (sum of all columns + spaces)
	g.grouped, g.columnsWidth, g.rows = g.getPaddings()

	for _, col := range g.columnsWidth {
		g.tcMaxLength += col + 1 // +1 for spacing // NOTE: Should that +1 be added in groupCompletions() ?
	}

	g.tcMaxX = len(g.columnsWidth)

	if g.rows > g.MaxLength {
		g.tcMaxY = g.MaxLength
	} else {
		g.tcMaxY = g.rows
	}

	g.tcPosX = 0
	g.tcPosY = 0
	g.tcOffset = 0
}

// moveTabListHighlight - Moves the highlighting for currently selected completion item (list display)
// We don't care about the x, because only can have 2 columns of selectable choices (--long and -s)
func (g *CompletionGroup) moveTabListHighlight(rl *Instance, x, y int) (done bool, next bool) {
	// We dont' pass to x, because not managed by callers
	g.tcPosY += x
	g.tcPosY += y

	// Lines
	if g.tcPosY < 1 {
		if rl.tabCompletionReverse {
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

	// Once we get to the end of choices: check which column we were selecting.
	if g.tcOffset+g.tcPosY > g.rows {
		done, next := g.goNextLineColumn()
		if done {
			return done, next
		}
	}

	// Get the row and subrow of the current candidate
	row, inRow := g.getCurrentRowValues()

	if rl.tabCompletionReverse {
		newY, found := g.getPreviousCandidate(row, inRow)
		if found {
			g.tcPosY -= newY
			return false, false
		} else {
			// HERE GO TO last candidate of PREVIOUS COLUMN
			g.tcPosX = 0
			g.tcPosY = g.tcMaxY
			// return true, false
		}
	} else {
		newY, found := g.getNextCandidate(row, inRow)
		if !found {
			return true, true
		}
		g.tcPosY += newY
		// TODO HERE: return ?
	}

	// Setup offset if needs to be.
	if g.tcOffset+g.tcPosY < 1 && g.rows > 0 {
		g.tcPosY = g.tcMaxY
		g.tcOffset = g.rows - g.tcMaxY
	}
	if g.tcOffset < 0 {
		g.tcOffset = 0
	}

	return false, false
}

func (g *CompletionGroup) getNextCandidate(i int, inRow int) (newY int, found bool) {
	remaining := g.grouped[i:] // NOTE: maybe i-1

	for i, row := range remaining {

		// Adjust the first row if it has multiple subrows
		if i == 0 && inRow > 0 {
			colCounter := 0
			rowCounter := 0

			for range row {
				if colCounter == len(g.columnsWidth) {
					rowCounter++
					colCounter = 0
				}

				if rowCounter == inRow {
					break
				}

				colCounter = +1
			}

			row = row[colCounter-1:]
		}

		// Skip if its does not have enough columns
		if len(row)-1 < g.tcPosX {
			newY++
			continue
		}

		// Else we have a candidate for the given column,
		// just break since our posY has been updated.
		g.selected = row[g.tcPosX]

		found = true
		break
	}

	return
}

func (g *CompletionGroup) getPreviousCandidate(i int, inRow int) (newY int, done bool) {
	remaining := g.grouped[:i] // NOTE: maybe i-1

	for i := len(remaining); i > 0; i-- {
		row := remaining[i-1]

		// Adjust the first row if it has multiple subrows
		if i == len(remaining) && inRow > 0 {
			row = row[:(inRow * len(g.columnsWidth))]
		}

		// Skip if its does not have enough columns
		if len(row) < g.tcPosX {
			newY++
			continue
		}

		// Else we have candidate for the given column,
		// just break since our posY has been updated.
		newY++
		done = true
		break
	}

	return
}

func (g *CompletionGroup) getCurrentRowValues() (rowIndex, inRow int) {
	y := 0

	for i, row := range g.grouped {
		y++
		rowIndex = i
		if y == g.tcPosY {
			break
		}

		colCounter := 0
		for range row {
			if colCounter == len(g.columnsWidth) {
				y++
				inRow++
				colCounter = 0
			}

			if y == g.tcPosY {
				break
			}

			colCounter = +1
		}
	}

	return
}

func (g *CompletionGroup) goNextLineColumn() (done bool, next bool) {
	// If we have completion aliases and that we are not yet
	// completing them, start on top of the next column.
	if g.tcPosX < len(g.columnsWidth)-1 {
		g.tcPosX++
		g.tcPosY = 1
		g.tcOffset = 0
		return false, false
	}

	// Else no alternatives, return for next group.
	// Reset all values, in case we pass on them again.
	g.tcPosX = 0 // First column
	g.tcPosY = 1 // first row
	g.tcOffset = 0
	return true, true
}

// writeList - A list completion string
func (g *CompletionGroup) writeList(rl *Instance) (comp string) {
	// Print group title and adjust offset if there is one.
	if g.Name != "" {
		comp += fmt.Sprintf("%s%s%s %s\n", BOLD, YELLOW, g.Name, RESET)
		rl.tcUsedY++
	}

	termWidth := GetTermWidth()
	if termWidth < 20 {
		// terminal too small. Probably better we do nothing instead of crash
		// We are more conservative than lmorg, and push it to 20 instead of 10
		return
	}

	// Suggestion cells dimensions
	maxLength := g.tcMaxLength
	if maxLength > termWidth-9 {
		maxLength = termWidth - 9
	}

	// Dimensions for description cells, uses the overall completion candidates pad.
	maxDescWidth := termWidth - maxLength - 4

	// Generate the aggregated completions block as a string.
	comps, usedY := g.buildCompList(maxLength, maxDescWidth)
	comp += comps
	rl.tcUsedY += usedY

	// Add the equivalent of this group's size to final screen clearing
	// Can be set and used only if no alterative completions have been given.
	// TODO: Don't use uniqueValues anymore, not reliable mesure.
	// if g.uniqueValues > g.MaxLength {
	// 	rl.tcUsedY += g.MaxLength
	// } else {
	// 	rl.tcUsedY += g.uniqueValues
	// }

	return
}

func (g *CompletionGroup) buildCompList(maxLength, maxDescWidth int) (comp string, y int) {
	// Our values are grouped under the same description in here,
	// including those that have no description.

	for i := g.tcOffset; i < len(g.grouped); i++ {
		y++ // Consider next item
		if y > g.tcMaxY {
			return
		}

		colCounter := 0

		// If the number of values will span a number of lines that
		// will overflow on tcMaxY, we cut the list to what is possible.
		for _, val := range g.grouped[i] {
			if colCounter == len(g.columnsWidth) {
				y++
				colCounter = 0
			}

			// If we have reached our max, we will append the description and return
			if y-1 == g.tcMaxY {
				break
			}

			// Else, good to print the candidate.
			colCounter += 1

			// NOTE: This might have to be removed
			item := val.Value
			if len(item) > maxLength {
				item = item[:maxLength-3] + "..."
			}

			columnPad := strconv.Itoa(g.columnsWidth[colCounter-1])

			item = fmt.Sprintf("%s%-"+columnPad+"s", g.highlight(val.Style, y, colCounter-1), item)
			comp += item + seqReset
		}

		// Here we must add the description for this(ose) candidates,
		// and potentially add the remaining padding needed before it.
		comp += strings.Repeat(" ", sum(g.columnsWidth[colCounter:]))

		// And add the description
		desc := g.grouped[i][0].Description
		if desc != "" {
			if len(desc) > maxDescWidth {
				desc = g.ListSeparator + " " + desc[:maxDescWidth-3] + "..." + RESET // TODO: here change with seqReset ?
			} else {
				desc = g.ListSeparator + " " + desc + RESET
			}
		}
		comp += desc + "\n"
	}

	return
}

// getMaxColumns computes the maximum number of completion candidate
// columns we'll have to use, if any of them have one or more aliases,
// computes the padding for each of these columns and the total one.
func (g *CompletionGroup) getPaddings() (values [][]CompletionValue, columns []int, actualY int) {
	// We have at least one column
	columns = append(columns, 0)
	g.rowsStartAt = make(map[int]int)

NEXT_VALUE:
	for _, value := range g.Values {

		valLen := len([]rune(value.Value))

		// If there is an existing group row for this description.
		for i, aliased := range values {
			if len(aliased) > 0 && aliased[0].Description == value.Description {
				aliased = append(aliased, value)
				values[i] = aliased

				// If the total space taken by columns is greater than half the terminal,
				// we find the column under which this value will be shown, and update pad.
				if (sum(columns) + valLen) > (GetTermWidth() / 2) {
					columnX := len(aliased) % len(columns)

					if columns[columnX] < valLen {
						columns[columnX] = valLen
					}
				} else if len(aliased) > len(columns) {
					columns = append(columns, valLen)
				} else if columns[len(aliased)-1] < valLen {
					columns[len(aliased)-1] = valLen
				}

				continue NEXT_VALUE
			}
		}

		// Else create a new row, and update the row pad.
		values = append(values, []CompletionValue{value})
		if columns[0] < valLen {
			columns[0] = valLen
		}
	}

	// Compute the actual number of lines for this group
	for _, vals := range values {
		actualY += len(vals) / len(columns)
		if (len(vals) % len(columns)) > 0 {
			actualY++
		}
	}

	return
}

func sum(vals []int) (sum int) {
	for _, val := range vals {
		sum += val
	}

	return
}
