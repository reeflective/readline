package readline

import (
	"regexp"
)

//
// Vim Modes ------------------------------------------------------ //
//

type viMode int

const (
	vimInsert viMode = iota
	vimReplaceOnce
	vimReplaceMany
	vimDelete
	vimKeys
)

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
	rl.oppendMode = true

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
	rl.oppendMode = false
}

// var viModes = map[viMode]string{
// 	vimInsert: "viins",
// 	vimKeys:   "vicmd",
//
// 	// TODO how to map "viopp" => operator pending ?
// 	vimReplaceOnce: "viopp",
// 	vimReplaceMany: "viopp",
// }

// viEditorHandler dispatches some runes to their corresponding
// actions, when the command-line is in editor mode (non insert)
// var vimEditorHandlers = map[viMode]func(rl *Instance, r []rune){
// 	vimKeys:        inputViKeys,
// 	vimDelete:      inputViDelete,
// 	vimReplaceOnce: inputViReplaceOnce,
// 	vimReplaceMany: inputViReplaceMany,
// }

// func inputViKeys(rl *Instance, r []rune) {
// 	rl.vi(r[0])
// 	rl.refreshVimStatus()
// }

func inputViDelete(rl *Instance, r []rune) {
	// rl.viDelete(r[0])
	rl.refreshVimStatus()
}

func inputViReplaceOnce(rl *Instance, r []rune) {
	rl.modeViMode = vimKeys
	rl.deleteX()
	rl.insert([]rune{r[0]})
	rl.refreshVimStatus()
}

func inputViReplaceMany(rl *Instance, r []rune) {
	for _, char := range r {
		rl.deleteX()
		rl.insert([]rune{char})
	}
	rl.refreshVimStatus()
}

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

// viEscape - In case th user is using Vim input, and the escape sequence has not
// been handled by other cases, we dispatch it to Vim and handle a few cases here.
// func (rl *Instance) viEscape(r []rune) {
// 	// Sometimes the escape sequence is interleaved with another one,
// 	// but key strokes might be in the wrong order, so we double check
// 	// and escape the Insert mode only if needed.
// 	if rl.modeViMode == vimInsert && len(r) == 1 && r[0] == 27 {
// 		if len(rl.line) > 0 && rl.pos > 0 {
// 			rl.pos--
// 		}
// 		rl.modeViMode = vimKeys
// 		rl.viIteration = ""
// 		rl.refreshVimStatus()
// 		return
// 	}
// }

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

// func viYankAlt(rl *Instance) {
// 	key := 'y'
//
// 	// We always try to read further keys for a matching widget:
// 	// In some modes we will get a different one, while in others (like visual)
// 	// we will just fallback on this current widget (vi-yank), which will be executed
// 	// as is, since we won't get any remaining key.
// 	_, keys, key := rl.readkeysForWidget(key, rl.mainKeymap)
//
// 	// If we got a remaining key with the widget, we
// 	// first check for special keys such as Escape.
// 	if byte(key) == charEscape {
// 		return
// 	}
//
// 	// If the widget we found is also returned with some remaining keys,
// 	// (such as Vi iterations, range keys, etc) we must keep reading them
// 	// with a range handler before coming back here.
// 	rl.readRangeKeys(string(keys), rl.mainKeymap)
//
// 	// If we don't have an active range, we don't yank anything
// 	// if !rl.activeRegion {
// 	// 	return
// 	// }
//
// 	// All handlers have caught and ran, and we are now ready
// 	// to perform yanking itself, either on a visual range or not.
// 	println("Cursor: " + strconv.Itoa(rl.pos))
// 	println("Mark: " + strconv.Itoa(rl.mark))
//
// 	bpos, epos, _ := rl.getSelection()
//
// 	// println(rl.line[bpos:epos])
// 	rl.saveBufToRegister(rl.line[bpos:epos])
// }

// readkeysForWidget recursively reads for input keys and tries to match any widget against them.
// It will return either a widget if matched, or none if not, along with any residual keys not matched.
// func (rl *Instance) readkeysForWidget(key rune, keymap keyMap) (widget string, keys []rune, retkey rune) {
// 	// The compounded keys that we read, including the caller key, (like yiW => y + iW)
// 	keys = []rune{key}
//
// 	for {
// 		// Append the key to all the keys
// 		keys = append(keys, key)
//
// 		// We already have a "root" key (like y, c, d)
// 		// In the provided keymap, find all widgets which keybinding
// 		// has the key as prefix (like ya, yiW, for key y)
// 		widgets := findBindkeyWidget(key, keymap)
//
// 		// If we have a single widget, or none, we are done reading keys. Break
// 		if len(widgets) <= 1 {
// 			_, widget = getWidget(widgets)
// 			break
// 		}
//
// 		// Now we must read a new key.
// 		// CHECK WHY AND HOW zsh-vi-mode matches the default widget to save if fully matching here.
// 		// If not matching, note that are entering operator pending mode here.
// 		key = rune(0)
// 		if widget = getWidgetMatch(key, keymap); widget != "" {
// 			b, _, _ := rl.readInput()
// 			key = rune(b[0])
// 		} else {
// 			viEnterOppendMode(rl)
// 			b, _, _ := rl.readInput()
// 			key = rune(b[0])
// 		}
// 	}
//
// 	// We have either a widget or none of them.
// 	// First exit operator pending mode.
// 	if rl.oppendMode {
// 		viExitOppendMode(rl)
// 	}
//
// 	// If the last key entered is not empty but we don't have a match, we return this key
// 	// to be used another way. Example: yb => b is not matched, but will actually trigger
// 	// everything back to the previous word to be yanked.
// 	keys = keys[:len(keys)-1]
// 	if len(keys) > 1 && key != rune(0) {
// 		retkey = key
// 	}
//
// 	return
// }

// readNavigationKey tries to match a key against a navigation widget.
// func (rl *Instance) readNavigationKey(key string, keymap keyMap) (hasRange bool) {
// 	// When no keys are provided, we return.
// 	if key == "" {
// 		return false
// 	}
//
// 	var widget keyHandler
// 	count := "-1"
//
// 	forwardChar, _ := regexp.Compile(`^([1-9][0-9]*)?([fFtT].?)$`)
//
// 	// Either find a movement key, or a forward/backward character search movement.
// 	if match := forwardChar.FindStringSubmatch(key); len(match) > 0 {
// 		widget = rl.findForwardChar(key, match)
// 		count = "1"
// 	} else {
// 		count, widget = rl.findMovementKey(key)
// 	}
//
// 	// Return if we have no widget.
// 	// TODO: Not sure need to return true
// 	if widget == nil {
// 		return true
// 	}
//
// 	// Match any count in the action, or set it to one.
// 	if isCount, _ := regexp.MatchString(`^[0-9]+$`, count); !isCount {
// 		count = "1"
// 	}
//
// 	// And run the widget we found 'count' times.
// 	// (At the first loop, check if cursor moved: if not, save time and break the loop)
// 	_, lastCursor := rl.pos, rl.pos
// 	times, _ := strconv.Atoi(count)
//
// 	for i := 0; i < times; i++ {
// 		widget(rl, []byte{}, 0, []rune{}) // TODO: not quite clean
//
// 		if lastCursor == rl.pos {
// 			break
// 		} else {
// 			lastCursor = rl.pos
// 		}
// 	}
//
// 	// Only reset the cursor to its position before the loop if the loop failed.
//
// 	// TODO: Find a way to know how exit code is not 0
// 	// When not 0, return the keys to the caller, which needs them.
//
// 	return
// }

// findForwardChar checks the matches from a backward/forward character search
// movement and transforms these matches into the appropriate widget action.
// func (rl *Instance) findForwardChar(key string, match []string) (widget keyHandler) {
// 	var count string
//
// 	if len(match) > 1 && match[1] != "" {
// 		count = match[0]
// 	}
//
// 	// Catch some special key actions (fFtT) for forwarding to chars after
// 	// looking them up. This should give us also an action/widget to run.
// 	if len(match[1]) < 2 {
// 		// Enter operator pending mode
// 		viEnterOppendMode(rl)
//
// 		b, i, _ := rl.readInput()
// 		key += string(b[:i])
//
// 		// If the key we just read is the Escape key in viopp mode, return
// 		// TODO: Here should match against VimOperatingPendingEscape key.
// 		if key[len(key)-1] == charEscape {
// 			return
// 		}
//
// 		// Exit operator pending mode.
// 		viExitOppendMode(rl)
// 	}
//
// 	forward, skip := true, false
//
// 	if match, _ := regexp.MatchString(`[FT]`, string(key[len(key)-2])); match {
// 		forward = false
// 	}
// 	if match, _ := regexp.MatchString(`[tT]`, string(key[len(key)-2])); match {
// 		skip = true
// 	}
//
// 	// The widget will move the cursor to the target character
// 	widget = func(rl *Instance, _ []byte, _ int, _ []rune) (bool, bool, string, error) {
// 		times, _ := strconv.Atoi(count)
// 		rl.findAndMoveCursor(string(key[len(key)-1]), times, forward, skip)
// 		return false, false, "", nil
// 	}
//
// 	return
// }

// findMovementKey finds the movement widget and the count contained in a key string.
// func (rl *Instance) findMovementKey(key string) (count string, widget keyHandler) {
// 	var name string
// 	count = key[:len(key)-1]
//
// 	switch key[len(key)-1] {
// 	case '^':
// 		name = "vi-first-non-blank"
// 	case '$':
// 		name = "vi-end-of-line"
// 	case ' ':
// 		name = "vi-forward-char"
// 	case '0':
// 		name = "vi-digit-or-beginning-of-line"
// 	case 'h':
// 		name = "vi-backward-char"
// 	case 'j':
// 		name = "down-line-or-history"
// 	case 'k':
// 		name = "up-line-or-history"
// 	case 'l':
// 		name = "vi-forward-char"
// 	case 'w':
// 		name = "vi-forward-word"
// 	case 'W':
// 		name = "vi-forward-blank-word"
// 	case 'e':
// 		name = "vi-forward-word-end"
// 	case 'E':
// 		name = "vi-forward-blank-word-end"
// 	case 'b':
// 		name = "vi-backward-word"
// 	case 'B':
// 		name = "vi-backward-blank-word"
// 	}
//
// 	widget = rl.getWidget(name)
//
// 	return
// }
