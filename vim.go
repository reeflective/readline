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
	rl.mark = rl.pos // Or rl.posX ? combined with rl.posY ?
}

func (rl *Instance) enterVisualLineMode() {
	rl.local = visual
	rl.visualLine = true
	rl.mark = 0 // start at the beginning of the line.
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
	rl.pos = cpos
	if rl.pos == len(newline) && len(newline) > 0 {
		rl.pos--
	}
}

func (rl *Instance) resetSelection() {
	rl.activeRegion = false
	rl.mark = -1
}

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
