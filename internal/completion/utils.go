package completion

import (
	"os"
	"strings"

	"github.com/reeflective/readline/internal/term"
)

var (
	maxValuesAreaRatio = 0.5 // Maximum ratio of the screen that described values can have.
	maxRowsRatio       = 2   // Maximu ratio of the screen rows that we can use by default.
	minRowsSpaceBelow  = 15  // Minimum acceptable space below cursor to use.
)

var sanitizer = strings.NewReplacer(
	"\n", ``,
	"\r", ``,
	"\t", ``,
)

func (e *Engine) setPrefix(comps Values) {
	switch comps.PREFIX {
	case "":
		// When no prefix has been specified, use
		// the current word up to the cursor position.
		lineWords, _, _ := e.line.TokenizeSpace(e.cursor.Pos())
		if len(lineWords) > 0 {
			last := lineWords[len(lineWords)-1]
			if last[len(last)-1] != ' ' {
				e.prefix = lineWords[len(lineWords)-1]
			}
		}

		// Newlines should not be accounted for, as they are
		// not printable: whether or not the completions provider
		// replaced them with spaces or not, we must not count them
		// as part of the prefix length, so not as part of the prefix.
		if strings.Contains(e.prefix, "\n") {
			e.prefix = strings.ReplaceAll(e.prefix, "\n", "")
		}

	default:
		// When the prefix has been overridden, add it to all
		// completions AND as a line prefix, for correct candidate insertion.
		e.prefix = comps.PREFIX
	}
}

func (rl *Engine) clearVirtualComp() {
	// Merge lines.
	rl.line.Set(*rl.completed...)
	rl.cursor.Set(rl.compCursor.Pos())

	// And no virtual candidate anymore.
	rl.comp = make([]rune, 0)
}

func (e *Engine) currentGroup() (grp *group) {
	for _, g := range e.groups {
		if g.isCurrent {
			return g
		}
	}
	// We might, for whatever reason, not find one.
	// If there are groups but no current, make first one the king.
	if len(e.groups) > 0 {
		for _, g := range e.groups {
			if len(g.values) > 0 {
				g.isCurrent = true
				return g
			}
		}
	}

	return
}

func (e *Engine) adjustCycleKeys(row, column int) (int, int) {
	cur := e.currentGroup()

	keyRunes, _ := e.keys.PeekAll()
	keys := string(keyRunes)

	if row > 0 {
		if cur.aliased && keys != term.ArrowRight && keys != term.ArrowDown {
			row, column = 0, row
		}
	} else {
		if cur.aliased && keys != term.ArrowLeft && keys != term.ArrowUp {
			row, column = 0, -1*row
		}
	}

	return row, column
}

// cycleNextGroup - Finds either the first non-empty group,
// or the next non-empty group after the current one.
func (e *Engine) cycleNextGroup() {
	for pos, g := range e.groups {
		if g.isCurrent {
			g.isCurrent = false

			if pos == len(e.groups)-1 {
				e.groups[0].isCurrent = true
			} else {
				e.groups[pos+1].isCurrent = true
				next := e.currentGroup()
				if len(next.values) == 0 {
					e.cycleNextGroup()
				}
			}

			break
		}
	}
}

// cyclePreviousGroup - Same as cycleNextGroup but reverse.
func (e *Engine) cyclePreviousGroup() {
	for pos, g := range e.groups {
		if g.isCurrent {
			g.isCurrent = false

			if pos == 0 {
				e.groups[len(e.groups)-1].isCurrent = true
			} else {
				e.groups[pos-1].isCurrent = true
				prev := e.currentGroup()
				if len(prev.values) == 0 {
					e.cyclePreviousGroup()
				}
			}

			break
		}
	}
}

func (e *Engine) updateCompletedLine() {
	// If no completions, reset ourselves.
	if e.noCompletions() {
		e.Reset(true, true)
		return
	}

	grp := e.currentGroup()
	if grp == nil {
		return
	}

	// If only one completion, insert and reset.
	if e.hasUniqueCandidate() {
		e.Reset(false, true)
	}

	// Else, insert the comp.
	comp := grp.selected().Value

	if len(comp) >= len(e.prefix) {
		diff := []rune(comp[len(e.prefix):])
		e.insertCandidate(diff)
	}
}

func (e *Engine) currentCandidate() (comp string) {
	cur := e.currentGroup()
	if cur == nil {
		return
	}

	return cur.selected().Value
}

func (e *Engine) completionCount() (comps int, used int) {
	for _, group := range e.groups {
		for _, row := range group.values {
			comps += len(row)
		}

		used++

		if group.tcMaxY > len(group.values) {
			used += len(group.values)
		} else {
			used += group.tcMaxY
		}
	}

	return
}

func (e *Engine) hasUniqueCandidate() bool {
	switch len(e.groups) {
	case 0:
		return false

	case 1:
		cur := e.currentGroup()
		if cur == nil {
			return false
		}

		if len(cur.values) == 1 {
			return len(cur.values[0]) == 1
		}

		return len(cur.values) == 1

	default:
		var count int

	GROUPS:
		for _, group := range e.groups {
			for _, row := range group.values {
				count++
				for range row {
					count++
				}
				if count > 1 {
					break GROUPS
				}
			}
		}

		return count == 1
	}
}

func (e *Engine) noCompletions() bool {
	for _, group := range e.groups {
		if len(group.values) > 0 {
			return false
		}
	}

	return true
}

func (e *Engine) resetList(cached bool) {
	e.groups = make([]*group, 0)

	if cached {
		e.cached = nil
	}

	// if len(e.groups) > 0 {
	// 	for _, g := range e.groups {
	// 		g.isCurrent = false
	// 	}
	//
	// 	e.groups[0].isCurrent = true
	// }
}

// returns either the max number of completion rows configured
// or a reasonable amount of rows so as not to bother the user.
func (e *Engine) getCompletionMaxRows() (maxRows int) {
	_, termHeight, _ := term.GetSize(int(os.Stdin.Fd()))

	// Pause the key reading routine and query the terminal
	e.keys.Pause()
	defer e.keys.Resume()

	_, cposY := term.GetCursorPos()
	if cposY == -1 {
		if termHeight != -1 {
			return termHeight / maxRowsRatio
		}

		return maxRows
	}

	spaceBelow := (termHeight - cposY) - 1

	// Only return the space below if it's reasonably large.
	if spaceBelow >= minRowsSpaceBelow {
		return spaceBelow
	}

	// Otherwise return half the terminal.
	return termHeight / maxRowsRatio
}

func (e *Engine) getAbsPos() int {
	var prev int
	var foundCurrent bool

	for _, grp := range e.groups {
		if grp.tag != "" {
			prev++
		}

		if grp.isCurrent {
			prev += grp.tcPosY
			foundCurrent = true

			break
		}

		prev += grp.tcMaxY
	}

	// If there was no current group, it means
	// we showed completions but there is no
	// candidate selected yet, return 0
	if !foundCurrent {
		return 0
	}

	return prev
}

// getColumnPad either updates or adds a new column for an alias.
func getColumnPad(columns []int, valLen int, numAliases int) []int {
	switch {
	case (float64(sum(columns) + valLen)) >
		(float64(term.GetWidth()) * maxValuesAreaRatio):
		columnX := numAliases % len(columns)

		if columns[columnX] < valLen {
			columns[columnX] = valLen
		}
	case numAliases > len(columns):
		columns = append(columns, valLen)
	case columns[numAliases-1] < valLen:
		columns[numAliases-1] = valLen
	}

	return columns
}

func stringInSlice(s string, sl []string) bool {
	for _, str := range sl {
		if s == str {
			return true
		}
	}

	return false
}

func sum(vals []int) (sum int) {
	for _, val := range vals {
		sum += val
	}

	return
}
