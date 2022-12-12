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

		// Ensure the completion system is in a sane
		// state before processing an input key.
		rl.ensureCompState()

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
		//
		// TODO: REWRITE THIS AND COMMENTS BELOW
		// The dispatching process grossly works in two steps:
		// 1)  It tries to match against the local keymap (visual/viopp/etc),
		// 2)  Just after, if some local keymap is active but has not matched
		//     the key, the key is checked/processed against some secondary stuff.
		//
		// 3)  If none of the above have notified us to read a key again, we check
		//     against the main keymap (vicmd/viins/emacs).
		// 4)  If not matched, we again process the key against some secondary stuff
		//     (like in 2.), but here for inserting the key if in insert mode.
		//
		// 5)  If the key was not inserted on the line, check against special patterns.

		// 1) First test the key against the local widget keymap, if any.
		// - In emacs mode, this local keymap is empty, except when performing
		// completions or performing history/incremental search.
		// - In Vim, this can be either 'visual', 'viopp', 'completion' or
		//   'incremental' search.
		handler, prefix := rl.matchKeymap(keys, rl.local)
		if handler != nil {
			read, ret, val, err := rl.run(handler, keys)
			if ret || err != nil {
				return val, err
			}

			// Only continue to next key if not asked to forward the key.
			if read {
				continue
			}
		} else if prefix {
			continue
		}

		// 3) If the key was not matched against any local widget,
		// check the global widget, which can never be nil.
		// - In Emacs mode, this widget is 'emacs'.
		// - In Vim mode, this can be 'viins' (Insert) or 'vicmd' (Normal).
		handler, prefix = rl.matchKeymap(keys, rl.main)
		if handler != nil {
			read, ret, val, err := rl.run(handler, keys)
			if ret || err != nil {
				return val, err
			}

			// If a widget of the main keymap was executed while the shell
			// was in operator pending mode (only Vim), then the caller widget
			// is waiting to be executed again.
			if rl.viopp {
				rl.runPendingWidget(keys)
			}

			// Only continue to next key if not asked to forward the key.
			if read {
				continue
			}
		} else if prefix {
			continue
		}

		// 4) When no widgets are matched neither locally nor globally,
		// and if we are in an insert mode (either 'emacs' or 'viins'),
		// we run the self-insert widget to input the key in the line.
		if rl.main == emacs || rl.main == viins {
			rl.viUndoSkipAppend = true
			ret, val, err := rl.runWidget("self-insert", r)
			if ret {
				return val, err
			}

			continue
		}

		// 5) Else, we are not in an insert mode either.
		// We try to match the key against the special keymap, which
		// is done using regular expressions. This allows to use digit
		// arguments, or other special patterns and ranges.
		if widget := rl.matchRegexKeymap(keys); widget != "" {
			ret, val, err := rl.runWidget(widget, r)
			if ret {
				return val, err
			}

			continue
		}

		// When the input key has neither:
		// - matched at least one widget in either keymap, be it only by prefix.
		// - not been inserted into the line.
		// - not matched the special regexp-based keymap.
		// We reset the current key stack.
		rl.keys = ""
	}
}
