package readline

type viMode int

const (
	vimInsert viMode = iota
	vimReplaceOnce
	vimReplaceMany
	vimDelete
	vimKeys
)

var viModes = map[viMode]string{
	vimInsert: "viins",
	vimKeys:   "vicmd",

	// TODO how to map "viopp" => operator pending ?
	vimReplaceOnce: "viopp",
	vimReplaceMany: "viopp",
}

// viEditorHandler dispatches some runes to their corresponding
// actions, when the command-line is in editor mode (non insert)
var vimEditorHandlers = map[viMode]func(rl *Instance, r []rune){
	vimKeys:        inputViKeys,
	vimDelete:      inputViDelete,
	vimReplaceOnce: inputViReplaceOnce,
	vimReplaceMany: inputViReplaceMany,
}

func inputViKeys(rl *Instance, r []rune) {
	rl.vi(r[0])
	rl.refreshVimStatus()
}

func inputViDelete(rl *Instance, r []rune) {
	rl.viDelete(r[0])
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
func (rl *Instance) viEscape(r []rune) {
	// Sometimes the escape sequence is interleaved with another one,
	// but key strokes might be in the wrong order, so we double check
	// and escape the Insert mode only if needed.
	if rl.modeViMode == vimInsert && len(r) == 1 && r[0] == 27 {
		if len(rl.line) > 0 && rl.pos > 0 {
			rl.pos--
		}
		rl.modeViMode = vimKeys
		rl.viIteration = ""
		rl.refreshVimStatus()
		return
	}
}

// readkeysForWidget recursively reads for input keys and tries to match any widget against them.
// It will return either a widget if matched, or none if not, along with any residual keys not matched.
func (rl *Instance) readkeysForWidget(key string, keymap keyMap) (widget string, keys []rune) {
	// The compounded keys that we read, including the caller key, (like yiW => y + iW)
	keys = []rune(key)

	for {
		// We already have a "root" key (like y, c, d)
		// In the provided keymap, find all widgets which keybinding
		// has the key as prefix (like ya, yiW, for key y)

		// If we have a single widget, or none, we are done reading keys. Break

		// Now we must read a new key.
		// CHECK WHY AND HOW zsh-vi-mode matches the default widget to save if fully matching here.
		// If not matching, note that are entering operator pending mode here.
		key = ""

		// Start again with this new key, which is going to filter again the widgets list.
	}

	// We have either a widget or none of them.
	// First exit operator pending mode.

	// If the last key entered is not empty but we don't have a match, we return this key
	// to be used another way. Example: yb => b is not matched, but will actually trigger
	// everything back to the previous word to be yanked.

	return
}

// readNavigationKey tries to match a key against a navigation widget.
func (rl *Instance) readNavigationKey(key string, keymap keyMap) {
	// When no keys are provided, we return.

	// Catch some special key actions (fFtT) for forwarding to chars after looking them up.
	// This should give us also an action/widget to run

	// Else, match against movement keys.

	// Return if we have no widget.

	// Match any count in the action, or set it to one.

	// And run the widget we found 'count' times.
	// (At the first loop, check if cursor moved: if not, save time and break the loop)

	// Only reset the cursor to its position before the loop if the loop failed.
}

func (rl *Instance) viYank(b []byte, i int, r []rune) {
	// We always try to read further keys for a matching widget:
	// In some modes we will get a different one, while in others (like visual)
	// we will just fallback on this current widget (vi-yank), which will be executed
	// as is, since we won't get any remaining key.

	// If we got a remaining key with the widget, we
	// first check for special keys such as Escape.

	// If the widget we found is also returned with some remaining keys,
	// (such as Vi iterations, range keys, etc) we must keep reading them
	// with a range handler before coming back here.

	// All handlers have caught and ran, and we are now ready
	// to perform yanking itself, either on a visual range or not.
}

func (rl *Instance) viDeleteHandler(b []byte, i int, r []rune) {
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

func (rl *Instance) viChange(b []byte, i int, r []rune) {
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
