package readline

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

func (rl *Instance) viChange(b []byte, i int, r []rune) {
	// key := r[0]
	// We always try to read further keys for a matching widget:
	// In some modes we will get a different one, while in others (like visual)
	// we will just fallback on this current widget (vi-delete), which will be executed
	// as is, since we won't get any remaining key.

	// If we got a remaining key with the widget, we
	// first check for special keys such as Escape.

	// If the widget we found is also returned with some remaining keys,
	// (such as Vi iterations, range keys, etc) we must keep reading them
	// with a range handler before coming back here.

	// All handlers have caught and ran, and we are now ready
	// to perform yanking itself, either on a visual range or not.

	// Reset the repeat commands, instead of doing it in the range handler function

	// And reset the cursor position if not nil (moved)
}
