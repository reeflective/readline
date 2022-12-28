package readline

import (
	"os"
)

// Readline displays the readline prompt and reads for user input.
// It can return from the call because of different several things:
//
// - When the user accepts the line (generally with Enter),
//   in which case the input line is returned to the caller.
// - If a particular keystroke mapping returns an error
//   (like Ctrl-C, Ctrl-D, etc).
//
// In all cases, the current input line is returned along with any
// potential error, and it is up to the caller to decide what to do
// with the line result.
func (rl *Instance) Readline() (string, error) {
	fd := int(os.Stdin.Fd())
	state, err := MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer Restore(fd, state)

	rl.initLine()        // Clear the line in most cases
	rl.initHelpers()     // Prepare hints/completions
	rl.initHistory()     // Reset undo/history indexes in most cases.
	rl.initHistoryLine() // Retrieve a line from history when asked.
	rl.initKeymap()      // Verify key mappings and widget binds

	// The prompt reevaluates itself when its corresponding
	// functions are bound. Some of its components (PS1/RPROMPT)
	// are normally only computed here (until the next Readline loop),
	// but other components (PS2/tips) are computed more than once.
	// Also print the primary prompt (or most of it if multiline).
	rl.Prompt.init(rl)

	// If the prompt is set as transient, we will print it
	// once our command line is returned to the caller.
	defer rl.Prompt.printTransient(rl)

	// Multisplit
	if len(rl.multilineSplit) > 0 {
		return rl.initMultiline()
	}

	// Finally, print any hints or completions if needed.
	rl.renderHelpers()

	// Start handling keystrokes.
	for {
		// Readline actualization/initialization. ------------------------------
		//
		// Since we always update helpers after being asked to read
		// for user input again, we do it before actually reading it.
		rl.redisplay()

		// The last key might have modified both the local keymap mode or
		// the global keymap (main), which is either emacs or viins/vicmd.
		//
		// These are the only keymaps that actually can be bound to main:
		// If we are now in a viopp, or menu-select, or isearch, this main
		// keymap reference does NOT change, so that any of its keys that
		// are not hidden by the local keymap ones can still be used.
		rl.updateKeymaps()

		// Read user key stroke(s) ---------------------------------------------
		//
		// Read the input from stdin if any, and upon successful
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
		// will reset/consume this stack.
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
			err := handler()
			if err != nil {
				return string(rl.line), err
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
			rl.run(widget, keys, rl.local)
			if rl.accepted || rl.err != nil {
				return string(rl.line), rl.err
			}
			continue
		} else if prefix {
			continue
		}

		// Past the local keymap, our actions have a direct effect on the line
		// or on the cursor position, so we must first "reset" or accept any
		// completion state we're in, if any, such as a virtually inserted candidate.
		rl.updateCompletion()

		// 2) If the key was not matched against any local widget, match it
		// against the global keymap, which can never be nil.
		// - In Emacs mode, this keymap is 'emacs'.
		// - In Vim mode, this can be 'viins' (Insert) or 'vicmd' (Normal).
		widget, prefix = rl.matchKeymap(keys, rl.main)
		if widget != nil {
			rl.run(widget, keys, rl.main)
			if rl.accepted || rl.err != nil {
				return string(rl.line), rl.err
			}
			continue
		} else if prefix {
			continue
		}

		// If the key was not matched against any keymap (exact or by prefix)
		// we discard the input stack, and exit some local keymaps (isearch)
		rl.resetUndefinedKey()
	}
}
