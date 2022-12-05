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

// action is represents the action of a widget, the number of times
// this widget needs to be run, and an optional operator argument.
// Most of the time we don't need this operator.
//
// Those actions are mostly used by widgets which make the shell enter
// the Vim operator pending mode, and thus require another key to be read.
type action struct {
	widget     string
	iterations int
	key        string
	operator   string
}

// matchPendingAction processes a key against pending operation, first considering
// the key as an operator to this action. It executes anything that should be done
// here, updating any mode if needed, and notifies the caller if it must keep trying
// to match the key against the other (main) keymaps.
func (rl *Instance) matchPendingAction(key string) (read, ret bool, val string, err error) {
	// If the key is a digit, we add it to the viopp-specific iterations
	if isDigit, _ := regexp.MatchString(`^([1-9]{1})$`, key); isDigit {
		rl.viIteration += key

		read = true
		return
	}

	// Some keys might be caught in vicmd mode, but have a special meaning in
	// operator pending mode. We store it to be used later.
	if isSpecialKey, _ := regexp.MatchString(`[ia]`, key); isSpecialKey {
		rl.navKey += key

		read = true
		return
	}

	// Since we can stack pending actions (like in 'y2Ft', where both y and F need
	// pending operators), if the current key matches any of the main keymap bindings,
	// we don't use it now as an argument, and must match it against the main keymap.
	if widget, found := rl.mainKeymap[key]; found && widget != "" {
		return
	}

	// Else, the key is taken as an argument to the last pending widget.
	rl.runPendingWidget(key)

	return
}

// getPendingWidget returns the last widget pushed onto the pending stack.
func (rl *Instance) getPendingWidget() (act action) {
	if len(rl.pendingActions) > 0 {
		act = rl.pendingActions[len(rl.pendingActions)-1]
		rl.pendingActions = rl.pendingActions[:len(rl.pendingActions)-1]
	}

	// Do we need to exit pending mode if list is empty ?

	return
}

func (rl *Instance) runPendingWidget(key string) {
	action := rl.getPendingWidget()

	if action.widget == "" {
		return
	}

	// Exit the pending operator mode if no more widgets
	// waiting for an argument operator.
	defer func() {
		if len(rl.pendingActions) == 0 {
			rl.exitVioppMode()
			rl.updateCursor()
		}
	}()

	widget := rl.getWidget(action.widget)
	if widget == nil {
		// TODO RESET everything
		return
	}

	defer func() {
		if len(rl.pendingActions) == 0 {
			rl.exitVioppMode()
		}
	}()

	// Permutate viIterations and pending iterations,
	// so that further operator iterations are used
	// within the widgets themselves.
	times := action.iterations

	keys := []rune(rl.navKey + key)

	// Run the widget with all navigation keys
	for i := 0; i < times; i++ {
		widget(rl, []byte{}, len(keys), keys)
		// read, ret, val, err = widget(rl, []byte{}, len(keys), keys)
	}
}

// vi - Apply a key to a Vi action. Note that as in the rest of the code, all cursor movements
// have been moved away, and only the rl.pos is adjusted: when echoing the input line, the shell
// will compute the new cursor pos accordingly.
func (rl *Instance) vi(r rune) {
	// Check if we are in register mode. If yes, and for some characters,
	// we select the register and exit this func immediately.
	if rl.registers.registerSelectWait {
		for _, char := range validRegisterKeys {
			if r == char {
				rl.registers.setActiveRegister(r)
				return
			}
		}
	}

	// If we are on register mode and one is already selected,
	// check if the key stroke to be evaluated is acting on it
	// or not: if not, we cancel the active register now.
	if rl.registers.onRegister {
		for _, char := range registerFreeKeys {
			if char == r {
				rl.registers.resetRegister()
			}
		}
	}

	// TODO: HERE FIND THE KEYMAP WIDGET AND RUN IT.

	// Then evaluate the key.
	switch r {
	default:
		if r <= '9' && '0' <= r {
			rl.viIteration += string(r)
		}
		rl.viUndoSkipAppend = true
	}
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
