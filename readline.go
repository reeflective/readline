package readline

import (
	"os"
	"regexp"
)

var rxMultiline = regexp.MustCompile(`[\r\n]+`)

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
		// Keymaps actualization/initialization. ------------------------------
		//
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

		// Read user key stroke(s) --------------------------------------------
		//
		// Read the input from stdin if any, and upon successfull
		// read, convert the input into runes for better scanning.
		b, i, readErr := rl.readInput()
		if readErr != nil {
			return "", err
		}

		r := []rune(string(b))
		key := string(r[:i]) // This allows to read all special sequences.

		// If the last input is a carriage return, process
		// according to configured multiline behavior.
		if isMultiline(r[:i]) || len(rl.multilineBuffer) > 0 {
			done, ret, val, err := rl.processMultiline(r, b, i)
			if ret {
				return val, err
			} else if done {
				continue
			}
		}

		// Main dispatchers ---------------------------------------------------
		//
		// Test the key against the local widget keymap, if any.
		// - In emacs mode, this local keymap is empty, except when performing
		// completions or performing history/incremental search.
		// - In Vim, this can be either 'visual', 'viopp', 'completion' or
		//   'incremental' search.
		if widget, found := rl.localKeymap[key]; found && widget != "" {
			ret, val, err := rl.runWidget(widget, b, i, r)
			if ret {
				return val, err
			}

			rl.updateHelpers()
			continue
		}

		// If the key was not matched against any local widget,
		// check the global widget, which can never be nil.
		// - In Emacs mode, this widget is 'emacs'.
		// - In Vim mode, this can be 'viins' (Insert) or 'vicmd' (Normal).
		if widget, found := rl.mainKeymap[key]; found && widget != "" {
			ret, val, err := rl.runWidget(widget, b, i, r)
			if ret {
				return val, err
			}

			rl.updateHelpers()
			continue
		}

		// When no widgets are matched neither locally nor globally,
		// and if we are in an insert mode (either 'emacs' or 'viins'),
		// we run the self-insert widget to input the key in the line.
		if rl.main == emacs || rl.main == viins {
			ret, val, err := rl.runWidget("self-insert", b, i, r)
			if ret {
				return val, err
			}

			rl.updateHelpers()
			continue
		}

		// Else, we are not in an insert mode either.
		// We try to match the key against the special keymap, which
		// is done using regular expressions. This allows to use digit
		// arguments, or other special patterns and ranges.
		if widget := rl.matchRegexKeymap(key); widget != "" {
			ret, val, err := rl.runWidget(widget, b, i, r)
			if ret {
				return val, err
			}

			rl.updateHelpers()
			continue
		}
	}
}
