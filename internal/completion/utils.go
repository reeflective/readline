package completion

import (
	"strings"
	"unicode"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/keymap"
	"github.com/reeflective/readline/internal/term"
)

var sanitizer = strings.NewReplacer(
	"\n", ``,
	"\r", ``,
	"\t", ``,
)

// prepare builds the list of completions, hint/usage messages
// and prefix/suffix strings, but does not attempt any candidate
// insertion/abortion on the line.
func (e *Engine) prepare(completions Values) {
	e.groups = make([]*group, 0)

	e.setPrefix(completions)
	e.setSuffix(completions)
	e.generate(completions)
}

func (e *Engine) generate(completions Values) {
	// Compute hints once we found either any errors,
	// or if we have no completions but usage strings.
	defer func() {
		e.hintCompletions(completions)
	}()

	// Nothing else to do if no completions
	if len(completions.values) == 0 {
		return
	}

	// Apply the prefix to the completions, and filter out any
	// completions that don't match, optionally ignoring case.
	matchCase := e.config.GetBool("completion-ignore-case")
	completions.values = completions.values.FilterPrefix(e.prefix, !matchCase)

	// Classify, group together and initialize completions.
	completions.values.EachTag(e.generateGroup(completions))
	e.justifyGroups(completions)
}

func (e *Engine) setPrefix(completions Values) {
	switch completions.PREFIX {
	case "":
		// Select the character just before the cursor.
		cpos := e.cursor.Pos() - 1
		if cpos < 0 {
			cpos = 0
		}

		bpos, _ := e.line.SelectBlankWord(cpos)

		// Safety checks and adjustments.
		if bpos > cpos {
			bpos, cpos = cpos, bpos
		}

		if cpos < e.line.Len() {
			cpos++
		}

		e.prefix = strings.TrimSpace(string((*e.line)[bpos:cpos]))

	default:
		e.prefix = completions.PREFIX
	}
}

func (e *Engine) setSuffix(completions Values) {
	switch completions.SUFFIX {
	case "":
		cpos := e.cursor.Pos()
		_, epos := e.line.SelectBlankWord(cpos)

		// Safety checks and adjustments.
		if epos < e.line.Len() {
			epos++
		}

		if epos < cpos {
			epos, cpos = cpos, epos
		}

		// Add back single or double quotes in the character after epos is one of them.
		if epos < e.line.Len() {
			if (*e.line)[epos] == '\'' || (*e.line)[epos] == '"' {
				epos++
			}
		}

		e.suffix = strings.TrimSpace(string((*e.line)[cpos:epos]))

	default:
		e.suffix = completions.SUFFIX
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
			if len(g.rows) > 0 {
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

	for {
		next := e.currentGroup()
		if len(next.rows) == 0 {
			e.cycleNextGroup()
			continue
		}

		return
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

	for {
		prev := e.currentGroup()
		if len(prev.rows) == 0 {
			e.cyclePreviousGroup()
			continue
		}

		return
	}
}

func (e *Engine) justifyGroups(values Values) {
	// commandGroups := make([]*group, 0)
	// maxCellLength := 0

	// for _, group := range e.groups {
	// 	// Skip groups that are not to be justified
	// 	if _, justify := values.Pad[group.tag]; !justify {
	// 		if _, all := values.Pad["*"]; !all {
	// 			continue
	// 		}
	// 	}
	//
	// 	// Skip groups that are aliased or have more than one column
	// 	if group.aliased || len(group.columnsWidth) > 1 {
	// 		continue
	// 	}
	//
	// 	commandGroups = append(commandGroups, group)
	//
	// 	if group.longestValueLen > maxCellLength {
	// 		maxCellLength = group.longestValueLen
	// 	}
	// }
	//
	// for _, group := range commandGroups {
	// 	group.longestValueLen = maxCellLength
	// }
}

func (e *Engine) adjustCycleKeys(row, column int) (int, int) {
	cur := e.currentGroup()

	keyRunes := e.keys.Caller()
	keys := string(keyRunes)

	if row > 0 {
		if cur.aliased && keys != term.ArrowRight && keys != term.ArrowDown {
			row, column = 0, row
		} else if keys == term.ArrowDown {
			row, column = 0, row
		}
	} else {
		if cur.aliased && keys != term.ArrowLeft && keys != term.ArrowUp {
			row, column = 0, 1*row
		} else if keys == term.ArrowUp {
			row, column = 0, 1*row
		}
	}

	return row, column
}

// adjustSelectKeymap is only called when the selector function has been used.
func (e *Engine) adjustSelectKeymap() {
	if e.keymap.Local() != keymap.Isearch {
		e.keymap.SetLocal(keymap.MenuSelect)
	}
}

// completionCount returns the number of completions for a given group,
// as well as the number of real terminal lines it spans on, including
// the group name if there is one.
func (e *Engine) completionCount() (comps int, used int) {
	for _, group := range e.groups {
		// First, agree on the number of comps.
		for _, row := range group.rows {
			comps += len(row)
		}

		// One line for the group name
		if group.tag != "" {
			used++
		}

		if group.maxY > len(group.rows) {
			used += group.maxY
		} else {
			used += len(group.rows)
		}
	}

	return comps, used
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

		if len(cur.rows) == 1 {
			return len(cur.rows[0]) == 1
		}

		return len(cur.rows) == 1

	default:
		var count int

	GROUPS:
		for _, group := range e.groups {
			for _, row := range group.rows {
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
		if len(group.rows) > 0 {
			return false
		}
	}

	return true
}

func (e *Engine) resetValues(comps, cached bool) {
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
	needsComplete := e.config.GetBool("autocomplete") &&
		e.keymap.Local() != keymap.MenuSelect &&
		e.keymap.Local() != keymap.Isearch &&
		e.line.Len() > 0

		// Not possible in Vim command mode either.
	isCorrectMenu := e.keymap.Main() != keymap.ViCommand &&
		e.keymap.Local() != keymap.Isearch

	if needsComplete && isCorrectMenu && len(e.selected.Value) == 0 {
		return true
	}

	if e.keymap.Local() != keymap.MenuSelect && e.autoForce {
		return true
	}

	return false
}

func (e *Engine) getAbsPos() int {
	var prev int
	var foundCurrent bool

	for _, grp := range e.groups {
		groupComps := 0

		for _, row := range grp.rows {
			groupComps += len(row)
		}

		if groupComps == 0 {
			continue
		}

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

func longest(vals []string, trimEscapes bool) int {
	var length int
	for _, val := range vals {
		if trimEscapes {
			val = color.Strip(val)
		}

		if len(val) > length {
			length = len(val)
		}
	}

	return length
}
