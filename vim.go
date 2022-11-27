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
