package readline

import (
	"regexp"
)

//
// Main selection helpers ------------------------------------------------------ //
//

type selection struct {
	regionType string
	active     bool
	bpos       int
	epos       int
	fg         string // Highlighting foreground
	bg         string // Highlighting background
}

// markSelection starts a visual selection at the specified mark position.
func (rl *Instance) markSelection(mark int) {
	rl.markSelectionRange(mark, -1)
}

func (rl *Instance) markSelectionRange(bpos, epos int) {
	vsel := rl.visualSelection()

	if vsel != nil {
		sel := rl.visualSelection()
		sel.bpos = bpos
		sel.epos = epos

		return
	}

	sel := &selection{
		bpos:       bpos,
		epos:       epos,
		active:     false,
		regionType: "visual",
		fg:         "",
		bg:         seqBgBlueDark,
	}

	rl.marks = append(rl.marks, sel)
}

func (rl *Instance) markSelectionSurround(bpos, epos int) {
	sel := &selection{
		bpos:       bpos,
		epos:       epos,
		active:     false,
		regionType: "surround",
		fg:         "",
		bg:         seqBgRed,
	}

	rl.marks = append(rl.marks, sel)
}

func (rl *Instance) activeSelection() bool {
	if rl.local == visual {
		return true
	}

	// We might have a visual selection used by another widget
	// (generally in Vi operator pending mode), in which case
	// the selection is deemed active if:
	if selection := rl.visualSelection(); selection != nil {
		// - there is not defined end, and that the beginning is not the cursor.
		if selection.epos == -1 && rl.pos != selection.bpos {
			return true
		}
		// - there is a defined range
		if selection.epos != -1 {
			return true
		}

		return false
	}

	return false
}

func (rl *Instance) visualSelection() *selection {
	for _, sel := range rl.marks {
		if sel.regionType == "visual" {
			return sel
		}
	}
	return nil
}

// Compute begin and end of visual selection.
func (rl *Instance) calcSelection() (bpos, epos, cpos int) {
	sel := rl.visualSelection()
	if sel == nil || len(rl.line) == 0 {
		return -1, -1, -1
	}

	bpos = sel.bpos
	epos = sel.epos

	// If the visual selection has one of its end
	// as the cursor, actualize this value.
	if sel.epos == -1 {
		bpos, epos = rl.selectionCursor(bpos)
	}

	// In visual mode, we include the cursor
	if rl.local == visual {
		epos++
	}

	// Ensure nothing is out of bounds
	if epos > len(rl.line) {
		epos = len(rl.line)
	}
	if bpos < 0 {
		bpos = 0
	}

	// Compute the desired cursor position.
	cpos = rl.selectionCursorPos(bpos, epos)

	return bpos, epos, cpos
}

func (rl *Instance) selectionCursor(bpos int) (int, int) {
	var epos int

	// The cursor might be now before its original mark,
	// in which case we invert before doing any move.
	if rl.pos < bpos {
		bpos, epos = rl.pos, bpos
	} else {
		epos = rl.pos
	}

	if rl.visualLine {
		for bpos--; bpos >= 0; bpos-- {
			if rl.line[bpos] == '\n' {
				bpos++
				break
			}
		}
		for ; epos < len(rl.line); epos++ {
			if epos == -1 {
				epos = 0
			}
			if rl.line[epos] == '\n' {
				break
			}
		}
	}

	// Check again in case the visual line inverted both.
	if bpos > epos {
		bpos, epos = epos, bpos
	}

	return bpos, epos
}

func (rl *Instance) selectionCursorPos(bpos, epos int) (cpos int) {
	cpos = bpos
	var indent int

	if rl.local == visual && rl.visualLine {

		// Get the indent of the cursor line.
		for cpos = rl.pos - 1; cpos >= 0; cpos-- {
			if rl.line[cpos] == '\n' {
				break
			}
		}
		indent = rl.pos - cpos - 1

		// If the selection includes the last line,
		// the cursor will move up the above line.
		var hpos, rpos int

		if epos < len(rl.line) {
			hpos = epos + 1
			rpos = bpos
		} else {
			for hpos = bpos - 2; hpos >= 0; hpos-- {
				if rl.line[hpos] == '\n' {
					break
				}
			}
			if hpos < -1 {
				hpos = -1
			}
			hpos++
			rpos = hpos
		}

		// Now calculate the cursor position, the indent
		// must be less than the line characters.
		for cpos = hpos; cpos < len(rl.line); cpos++ {
			if rl.line[cpos] == '\n' {
				break
			}
			if hpos+indent <= cpos {
				break
			}
		}

		// That cursor position might be bigger than the line itself:
		// it should be controlled when the line is redisplayed.
		cpos = rpos + cpos - hpos
	}

	return
}

// popSelection returns the active region and resets it.
func (rl *Instance) popSelection() (s string, bpos, epos, cpos int) {
	if len(rl.marks) == 0 {
		return
	}

	defer rl.resetSelection()

	bpos, epos, cpos = rl.calcSelection()
	if bpos == -1 || epos == -1 {
		return
	}

	s = string(rl.line[bpos:epos])

	return
}

// insertBlock inserts a given string at the specified indexes, with
// an optional string to surround the block with. Everything before
// bpos and after epos is retained from the line, word inserted in the middle.
func (rl *Instance) insertBlock(bpos, epos int, word, surround string) {
	if surround != "" {
		word = surround + word + surround
	}

	begin := string(rl.line[:bpos])
	end := string(rl.line[epos:])

	newLine := append([]rune(begin), []rune(word)...)
	newLine = append(newLine, []rune(end)...)
	rl.line = newLine
}

// insertSelection works the same as insertBlock, but directly uses the
// current active region as a block to insert. Resets the selection once done.
// Returns the computed cursor position after insert.
func (rl *Instance) insertSelection(bchar, echar string) (wlen, cpos int) {
	if len(rl.marks) == 0 {
		return
	}

	defer rl.resetSelection()

	bpos, epos, cpos := rl.calcSelection()
	if bpos == -1 || epos == -1 {
		return
	}
	selection := string(rl.line[bpos:epos])

	selection = bchar + selection + echar

	begin := string(rl.line[:bpos])
	end := string(rl.line[epos:])

	newLine := append([]rune(begin), []rune(selection)...)
	newLine = append(newLine, []rune(end)...)
	rl.line = newLine

	return len(selection), cpos
}

// replaceSelectionRune replaces all runes in the current visual selection.
func (rl *Instance) replaceSelectionRune(replacer func(r rune) rune) {
	bpos, epos, _ := rl.calcSelection()
	if bpos == -1 || epos == -1 {
		return
	}
	rl.resetSelection()
	rl.pos = bpos

	for range rl.line[bpos:epos] {
		char := rl.line[rl.pos]
		char = replacer(char)
		rl.line[rl.pos] = char
		rl.pos++
	}

	rl.viCommandMode()
	rl.updateCursor()
}

// yankSelection copies the active selection in the active/default register.
func (rl *Instance) yankSelection() {
	if len(rl.marks) == 0 {
		return
	}

	bpos, epos, cpos := rl.calcSelection()
	if bpos == -1 || epos == -1 {
		return
	}

	selection := string(rl.line[bpos:epos])

	// Visual adjustmeents when we are on the last line.
	if rl.local == visual && rl.visualLine && epos == len(rl.line) {
		selection += "\n"
	}

	// And copy to active register
	rl.saveBufToRegister([]rune(selection))

	// and reset the cursor position if not in visual mode
	if !rl.visualLine {
		rl.pos = cpos
	}
}

// yankSelection deletes the active selection.
func (rl *Instance) deleteSelection() {
	if len(rl.marks) == 0 {
		return
	}
	defer rl.resetSelection()

	var newline []rune

	// Get the selection.
	bpos, epos, cpos := rl.calcSelection()
	if bpos == -1 || epos == -1 {
		return
	}

	selection := string(rl.line[bpos:epos])

	// Visual adjustmeents when we are on the last line.
	if rl.local == visual && rl.visualLine && epos == len(rl.line) {
		selection += "\n"
	}

	// Save it and update the line
	rl.saveBufToRegister([]rune(selection))
	newline = append(rl.line[:bpos], rl.line[epos:]...)
	rl.line = newline

	rl.pos = cpos
}

// resetSelection unmarks the mark position and deactivates the region.
func (rl *Instance) resetSelection() {
	for i, reg := range rl.marks {
		if reg.regionType == "visual" {
			if len(rl.marks) > i {
				rl.marks = append(rl.marks[:i], rl.marks[i+1:]...)
			}
		}
	}
}

//
// Selection search/modification helpers ----------------------------------------- //
//

// selectInWord returns the entire non-blank word around specified cursor position.
func (rl *Instance) selectInWord(cpos int) (bpos, epos int) {
	pattern := "[0-9a-zA-Z_]"
	bpos, epos = cpos, cpos

	if match, _ := regexp.MatchString(pattern, string(rl.line[cpos])); !match {
		pattern = "[^0-9a-zA-Z_ ]"
	}

	// To first space found backward
	for ; bpos >= 0; bpos-- {
		if match, _ := regexp.MatchString(pattern, string(rl.line[bpos])); !match {
			break
		}
	}

	// And to first space found forward
	for ; epos < len(rl.line); epos++ {
		if match, _ := regexp.MatchString(pattern, string(rl.line[epos])); !match {
			break
		}
	}

	bpos++

	// Ending position must be greater than 0
	if epos > 0 {
		epos--
	}

	return
}

// searchSurround returns the index of the enclosing rune (either matching signs
// or the rune itself) of the input line, as well as each enclosing char.
func (rl *Instance) searchSurround(r rune) (bpos, epos int, bchar, echar rune) {
	posInit := rl.pos

	bchar, echar = rl.matchSurround(r)

	bpos = rl.substrPos(bchar, false)
	epos = rl.substrPos(echar, true)

	if bpos == epos {
		rl.pos++
		epos = rl.substrPos(echar, true)
		if epos == -1 {
			rl.pos--
			epos = rl.substrPos(echar, false)
			if epos != -1 {
				tmp := epos
				epos = bpos
				bpos = tmp
			}
		}
	}

	rl.pos = posInit

	return
}

// adjustSurroundQuotes returns the correct mark and cursor positions when
// we want to know where a shell word enclosed with quotes (and potentially
// having inner ones) starts and ends.
func adjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos int) (mark, cpos int) {
	mark = -1
	cpos = -1

	if (sBpos == -1 || sEpos == -1) && (dBpos == -1 || dEpos == -1) {
		return
	}

	doubleFirstAndValid := (dBpos < sBpos && // Outtermost
		dBpos >= 0 && // Double found
		sBpos >= 0 && // compared with a found single
		dEpos > sEpos) // ensuring that we are not comparing unfound

	singleFirstAndValid := (sBpos < dBpos &&
		sBpos >= 0 &&
		dBpos >= 0 &&
		sEpos > dEpos)

	if (sBpos == -1 || sEpos == -1) || doubleFirstAndValid {
		mark = dBpos
		cpos = dEpos
	} else if (dBpos == -1 || dEpos == -1) || singleFirstAndValid {
		mark = sBpos
		cpos = sEpos
	}

	return
}
