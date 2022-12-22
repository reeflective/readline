package readline

// We pass a special subset of the current input line, so that
// completions are available no matter where the cursor is.
func (rl *Instance) getCompletionLine() (line []rune, pos int) {
	pos = rl.pos - len(rl.comp)
	if pos < 0 {
		pos = 0
	}

	switch {
	case rl.pos == len(rl.line):
		line = rl.line
	case rl.pos < len(rl.line):
		line = rl.line[:pos]
	default:
		line = rl.line
	}

	return
}

// When the completions are either longer than:
// - The user-specified max completion length
// - The terminal lengh
// we use this function to prompt for confirmation before printing comps.
func (rl *Instance) promptCompletionConfirm(sentence string) {
	rl.hint = []rune(sentence)

	rl.compConfirmWait = true
	rl.undoSkipAppend = true

	rl.renderHelpers()
}

func (rl *Instance) currentGroup() (group *comps) {
	for _, g := range rl.tcGroups {
		if g.isCurrent && len(g.values) > 0 {
			return g
		}
	}
	// We might, for whatever reason, not find one.
	// If there are groups but no current, make first one the king.
	if len(rl.tcGroups) > 0 {
		// Find first group that has list > 0, as another checkup
		for _, g := range rl.tcGroups {
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
func (rl *Instance) cycleNextGroup() {
	for i, g := range rl.tcGroups {
		if g.isCurrent {
			g.isCurrent = false
			if i == len(rl.tcGroups)-1 {
				rl.tcGroups[0].isCurrent = true
			} else {
				rl.tcGroups[i+1].isCurrent = true
				next := rl.currentGroup()
				if len(next.values) == 0 {
					rl.cycleNextGroup()
				}
			}
			break
		}
	}
}

// cyclePreviousGroup - Same as cycleNextGroup but reverse
func (rl *Instance) cyclePreviousGroup() {
	for i, g := range rl.tcGroups {
		if g.isCurrent {
			g.isCurrent = false
			if i == 0 {
				rl.tcGroups[len(rl.tcGroups)-1].isCurrent = true
			} else {
				rl.tcGroups[i-1].isCurrent = true
				prev := rl.currentGroup()
				if len(prev.values) == 0 {
					rl.cyclePreviousGroup()
				}
			}
			break
		}
	}
}

func (rl *Instance) currentCandidate() (comp string) {
	cur := rl.currentGroup()
	if cur == nil {
		return
	}

	return cur.selected().Value
}

func (rl *Instance) noCompletions() bool {
	for _, group := range rl.tcGroups {
		if len(group.values) > 0 {
			return false
		}
	}

	return true
}

func (rl *Instance) completionCount() (comps int, lines int, adjusted int) {
	for _, group := range rl.tcGroups {
		comps += len(group.values)
		adjusted++
		if group.tcMaxY > len(group.values) {
			lines += len(group.values)
			adjusted += len(group.values)
		} else {
			lines += group.tcMaxY
			adjusted += group.tcMaxY
		}
	}
	return
}

func (rl *Instance) getAbsPos() int {
	var prev int
	var foundCurrent bool
	for _, grp := range rl.tcGroups {
		if grp.isCurrent {
			prev += grp.tcPosY + 1 // + 1 for title
			foundCurrent = true
			break
		} else {
			prev += grp.tcMaxY + 1 // + 1 for title
		}
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
	if (sum(columns) + valLen) > (GetTermWidth() / 2) {
		columnX := numAliases % len(columns)

		if columns[columnX] < valLen {
			columns[columnX] = valLen
		}
	} else if numAliases > len(columns) {
		columns = append(columns, valLen)
	} else if columns[numAliases-1] < valLen {
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
