package readline

import (
	"regexp"
)

//
// Vim Modes ------------------------------------------------------ //
//

func (rl *Instance) enterVisualMode() {
	rl.local = visual
	rl.visualLine = false
	rl.mark = rl.pos
}

func (rl *Instance) enterVisualLineMode() {
	rl.local = visual
	rl.visualLine = true
	rl.mark = 0 // start at the beginning of the line.
}

func (rl *Instance) exitVisualMode() {
	if rl.local == visual {
		rl.local = ""
	}
	rl.visualLine = false
	rl.mark = -1

	for i, reg := range rl.regions {
		if reg.regionType == "visual" {
			if len(rl.regions) > i {
				rl.regions = append(rl.regions[:i], rl.regions[i+1:]...)
			}
		}
	}
}

// enterVioppMode adds a widget to the list of widgets waiting for an operator/action,
// enters the vi operator pending mode and updates the cursor.
func (rl *Instance) enterVioppMode(widget string) {
	rl.local = viopp

	// When the widget is empty, we just want to update the cursor.
	if widget == "" {
		return
	}

	act := action{
		widget:     widget,
		iterations: rl.getViIterations(),
	}

	// Push the widget on the stack of widgets
	if widget != "" {
		rl.pendingActions = append(rl.pendingActions, act)
	}
}

func (rl *Instance) exitVioppMode() {
	rl.local = ""
}

//
// Selection ------------------------------------------------------- //
//

// Compute begin and end of region
func (rl *Instance) selection() (start, end int) {
	if rl.mark < rl.pos {
		start = rl.mark
		end = rl.pos + 1
	} else {
		start = rl.pos
		end = rl.mark
	}

	// Here, compute for visual line mode if needed.
	// We select the whole line.
	if rl.visualLine {
		start = 0

		for i, char := range rl.line[end:] {
			if string(char) == "\n" {
				break
			}
			end = end + i
		}
	}

	// Ensure nothing is out of bounds
	if end > len(rl.line)-1 {
		end = len(rl.line)
	}
	if start < 0 {
		start = 0
	}

	return
}

// TODO: Not needed
func (rl *Instance) getSelection() (bpos, epos, cpos int) {
	// Get current region and save the current cursor position
	bpos, epos = rl.selection()
	cpos = bpos

	// Ensure selection is within bounds
	if bpos < 0 {
		bpos = 0
	}

	return
}

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

// yankSelection copies the active selection in the active/default register.
func (rl *Instance) yankSelection() {
	// Get the selection.
	bpos, epos, cpos := rl.getSelection()
	selection := string(rl.line[bpos:epos])

	// The visual line mode always adds a newline
	if rl.local == visual && rl.visualLine {
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
	var newline []rune

	// Get the selection.
	bpos, epos, cpos := rl.getSelection()
	selection := string(rl.line[bpos:epos])

	// Here, adapt cursor position if visual line

	rl.saveBufToRegister([]rune(selection))
	newline = append(rl.line[:bpos], rl.line[epos:]...)
	rl.line = newline

	// Adapt cursor position when at the end of the
	// TODO: Same refactor pos thing
	rl.pos = cpos
	if rl.pos == len(newline) && len(newline) > 0 {
		rl.pos--
	}
}

// TODO: Not sure this is complete.
func (rl *Instance) resetSelection() {
	rl.activeRegion = false
	rl.mark = -1
}

// lineSlice returns a subset of the current input line.
func (rl *Instance) lineSlice(adjust int) (slice string) {
	switch {
	case rl.pos+adjust > len(rl.line):
		slice = string(rl.line[rl.pos:])
	case adjust < 0:
		if rl.pos+adjust < 0 {
			slice = string(rl.line[:rl.pos])
		} else {
			slice = string(rl.line[rl.pos+adjust : rl.pos])
		}
	default:
		slice = string(rl.line[rl.pos : rl.pos+adjust])
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

// substrPos gets the index pos of a char in the input line, starting
// from cursor, either backward or forward. Returns -1 if not found.
func (rl *Instance) substrPos(r rune, forward bool) (pos int) {
	pos = -1
	initPos := rl.pos

	rl.findAndMoveCursor(string(r), 1, forward, false)

	if rl.pos != initPos {
		pos = rl.pos
		rl.pos = initPos
	}

	return
}
