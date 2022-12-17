package readline

import (
	"os"
)

// Readline displays the readline prompt and reads for user input.
// It can return on several things:
// - When the user accepts the line (generally with Enter),
//   in which case the input line is returned to the caller.
// - If a particular keystroke mapping returns an error
//   (like Ctrl-C, Ctrl-D, etc), and an empty string.
func (rl *Instance) Readline() (string, error) {
	fd := int(os.Stdin.Fd())
	state, err := MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer Restore(fd, state)

	rl.initLine()
	rl.initHelpers()
	rl.initHistory()
	rl.initKeymap()

	// The prompt reevaluates itself when its corresponding
	// functions are bound. Some of its components (PS1/RPROMPT)
	// are normally only computed here (until the next Readline loop),
	// but other components (PS2/tips) are computed more than once.
	rl.Prompt.init(rl)

	// If the prompt is set as transient, print it once
	// our command line is returned to the caller.
	defer rl.Prompt.printTransient(rl)

	// Multisplit
	if len(rl.multilineSplit) > 0 {
		return rl.initMultiline()
	}

	// Finally, print any hints or completions
	// if the TabCompletion engines so desires
	rl.renderHelpers()

	// Start handling keystrokes.
	for {
		// Readline actualization/initialization. ------------------------------
		//
		// Since we always update helpers after being asked to read
		// for user input again, we do it before actually reading it.
		rl.updateHelpers()

		// The last key might have modified both the local keymap mode or
		// the global keymap (main), which is either emacs or viins/vicmd.
		//
		// Here we must ensure/actualize the reference to the main keymap:
		// - If we are now in viins, the main keymap is viins
		// - If in vicmd, the main keymap is vicmd.
		//
		// These are the only keymaps that actually can be bound to main:
		// If we are now in a viopp, or menu-select, or isearch, this main
		// keymap reference does NOT change, so that any of its keys that
		// are not hidden by the local keymap ones can still be used.
		rl.updateKeymaps()

		// Read user key stroke(s) ---------------------------------------------
		//
		// Read the input from stdin if any, and upon successfull
		// read, convert the input into runes for better scanning.
		b, i, readErr := rl.readInput()
		if readErr != nil {
			return "", err
		}

		// Only keep the portion that has been read.
		r := []rune(string(b))[:i]

		// We store the key in our key stack. which is used
		// when the key only matches some widgets as a prefix.
		// We use a copy for the matches below, as some actions
		// will reset this stack.
		rl.keys += string(r)
		keys := rl.keys

		// If the last input is a carriage return, process
		// according to configured multiline behavior.
		if isMultiline(r) || len(rl.multilineBuffer) > 0 {
			done, ret, val, err := rl.processMultiline(r, b, i)
			if ret {
				return val, err
			} else if done {
				continue
			}
		}

		// Interrupt keys (CtrlC/CtrlD, etc) are caught before any keymap:
		// These handlers adapt their behavior on their own, depending on
		// the current state of the shell, keymap, etc.
		if handler, yes := rl.isInterrupt(keys); yes && handler != nil {
			done, ret, val, err := handler(r)
			if ret {
				return val, err
			} else if done {
				continue
			}
		}

		//
		// Main dispatchers ----------------------------------------------------
		//

		// 1) First test the key against the local widget keymap, if any.
		// - In emacs mode, this local keymap is empty, except when performing
		//   completions or performing history/incremental search.
		// - In Vim, this can be either 'visual', 'viopp', 'completion' or
		//   'incremental' search.
		// - When completing/searching, can be 'menuselect' or 'isearch'
		widget, prefix := rl.matchKeymap(keys, rl.local)
		if widget != nil {
			_, ret, val, err := rl.run(widget, keys)
			if ret || err != nil {
				return val, err
				// } else if read {
				// 	continue
			}
			continue
		} else if prefix {
			continue
		}

		// Past the local keymap, our actions have a direct effect on the line
		// or on the cursor position, so we must first "reset" or accept any
		// completion state we're in, if any, such as a virtually inserted candidate.
		rl.updateCompletionState()

		// 2) If the key was not matched against any local widget, match it
		// against the global keymap, which can never be nil.
		// - In Emacs mode, this keymap is 'emacs'.
		// - In Vim mode, this can be 'viins' (Insert) or 'vicmd' (Normal).
		widget, prefix = rl.matchKeymap(keys, rl.main)
		if widget != nil {
			_, ret, val, err := rl.run(widget, keys)
			if ret || err != nil {
				return val, err
				// } else if read {
				// 	continue
			}
			continue
		} else if prefix {
			continue
		}

		// When the key has not matched any keybind pattern in any of the
		// active keymaps (perfectly or by prefix), we discard the key stack.
		rl.keys = ""
	}
}
