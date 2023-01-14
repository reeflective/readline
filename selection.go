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
	rl.markSelectionRange("visual", mark, -1)
}

func (rl *Instance) markSelectionRange(rType string, bpos, epos int) {
	visual := rl.visualSelection()

	if rType == "visual" && visual != nil {
		sel := rl.visualSelection()
		sel.bpos = bpos
		sel.epos = epos

		return
	}

	sel := &selection{
		bpos:       bpos,
		epos:       epos,
		active:     false,
		regionType: rType,
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
	if sel == nil {
		return -1, -1, -1
	}

	bpos = sel.bpos
	epos = sel.epos

	// If the visual selection has one of its end
	// as the cursor, actualize this value.
	if sel.epos == -1 {
		switch {
		case rl.visualLine:
			bpos = rl.substrPos('\n', false)
			if bpos == rl.pos {
				bpos = 0
			} else {
				bpos++
			}

			epos = rl.substrPos('\n', true)
			if epos == rl.pos {
				epos = len(rl.line) - 1
			}

			// The cursor position is the first preceding
			// newline.
			for cpos = rl.pos; cpos >= 0; cpos-- {
				if rl.line[cpos] == '\n' {
					break
				}
			}

		case sel.bpos <= rl.pos:
			bpos = sel.bpos
			epos = rl.pos
			if bpos < 0 {
				bpos = 0
			}
			cpos = bpos

		default:
			bpos = rl.pos
			epos = sel.bpos
			if bpos < 0 {
				bpos = 0
			}
			cpos = bpos
		}
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

	return
}

// popSelection returns the active region and resets it.
func (rl *Instance) popSelection() (s string, bpos, epos, cpos int) {
	if len(rl.marks) == 0 {
		return
	}

	bpos, epos, cpos = rl.calcSelection()
	s = string(rl.line[bpos:epos])

	rl.resetSelection()

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

	bpos, epos, cpos := rl.calcSelection()
	selection := string(rl.line[bpos:epos])

	selection = bchar + selection + echar

	begin := string(rl.line[:bpos])
	end := string(rl.line[epos:])

	newLine := append([]rune(begin), []rune(selection)...)
	newLine = append(newLine, []rune(end)...)
	rl.line = newLine

	rl.resetSelection()

	return len(selection), cpos
}

// yankSelection copies the active selection in the active/default register.
func (rl *Instance) yankSelection() {
	if len(rl.marks) == 0 {
		return
	}

	bpos, epos, cpos := rl.calcSelection()
	selection := string(rl.line[bpos:epos])

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

	var newline []rune

	// Get the selection.
	bpos, epos, cpos := rl.calcSelection()
	selection := string(rl.line[bpos:epos])

	// Save it and update the line
	rl.saveBufToRegister([]rune(selection))
	newline = append(rl.line[:bpos], rl.line[epos:]...)
	rl.line = newline

	rl.pos = cpos

	// Reset the selection since it does not exist anymore.
	rl.resetSelection()
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
