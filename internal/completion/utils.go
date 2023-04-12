package completion

import (
	"strings"
	"unicode"

	"github.com/reeflective/readline/internal/keymap"
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

// prepare builds the list of completions but does
// not attempt any insertion/abortion on the line.
func (e *Engine) prepare(completions Values) {
	e.groups = make([]*group, 0)

	e.setPrefix(completions)
	e.group(completions)
}

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
			}

			break
		}
	}

	next := e.currentGroup()
	if len(next.values) == 0 {
		e.cycleNextGroup()
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
			}

			break
		}
	}

	prev := e.currentGroup()
	if len(prev.values) == 0 {
		e.cyclePreviousGroup()
	}
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
			row, column = 0, 1*row
		}
	}

	return row, column
}

// adjustSelectKeymap is only when the selector function has been used.
func (e *Engine) adjustSelectKeymap() {
	if e.keymaps.Local() != keymap.Isearch {
		e.keymaps.SetLocal(keymap.MenuSelect)
	}
}

func (e *Engine) completionCount() (comps int, used int) {
	for _, group := range e.groups {
		for _, row := range group.values {
			comps += len(row)
		}

		used++

		if group.maxY > len(group.values) {
			used += len(group.values)
		} else {
			used += group.maxY
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

func (e *Engine) resetValues(comps, cached bool) {
	e.prefix = ""
	e.selected = Candidate{}

	// Drop the list of already generated/prepared completion candidates.
	if comps {
		e.usedY = 0
		e.groups = make([]*group, 0)
	}

	// Drop the completion generation function.
	if cached {
		e.cached = nil
	}

	// If generated choices were kept, select first choice.
	if len(e.groups) > 0 {
		for _, g := range e.groups {
			g.isCurrent = false
		}

		e.groups[0].isCurrent = true
	}
}

func (e *Engine) needsAutoComplete() bool {
	// Autocomplete is not needed when already completing,
	// or when the input line is empty (would always trigger)
	needsComplete := e.opts.GetBool("autocomplete") &&
		e.keymaps.Local() != keymap.MenuSelect &&
		e.keymaps.Local() != keymap.Isearch &&
		e.line.Len() > 0

		// Not possible in Vim command mode either.
	isCorrectMenu := e.keymaps.Main() != keymap.ViCmd &&
		e.keymaps.Local() != keymap.Isearch

	if needsComplete && isCorrectMenu && len(e.selected.Value) == 0 {
		return true
	}

	return false
}

func (e *Engine) getAbsPos() int {
	var prev int
	var foundCurrent bool

	for _, grp := range e.groups {
		if grp.tag != "" {
			prev++
		}

		if grp.isCurrent {
			prev += grp.posY
			foundCurrent = true

			break
		}

		prev += grp.maxY
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

func hasUpper(line []rune) bool {
	for _, r := range line {
		if unicode.IsUpper(r) {
			return true
		}
	}

	return false
}
