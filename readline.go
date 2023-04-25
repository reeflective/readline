package readline

import (
	"os"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/keymap"
	"github.com/reeflective/readline/internal/term"
)

// Readline displays the readline prompt and reads for user input.
// It can return from the call because of different several things:
//
//   - When the user accepts the line (generally with Enter).
//   - If a particular keystroke mapping returns an error
//     (like Ctrl-C, Ctrl-D, etc).
//
// In all cases, the current input line is returned along with any error,
// and it is up to the caller to decide what to do with the line result.
func (rl *Shell) Readline() (string, error) {
	descriptor := int(os.Stdin.Fd())

	state, err := term.MakeRaw(descriptor)
	if err != nil {
		return "", err
	}
	defer term.Restore(descriptor, state)

	rl.init()

	rl.prompt.PrimaryPrint()
	defer rl.display.PrintTransientPrompt()

	for {
		// Since we always update helpers after being asked to read
		// for user input again, we do it before actually reading it.
		rl.display.Refresh()

		// Block and wait for user input.
		rl.keys.Read()

		// 1 - Local keymap (completion/isearch/viopp)
		bind, command, prefixed := rl.keymaps.MatchLocal()
		if prefixed {
			continue
		}

		accepted, line, err := rl.run(bind, command)
		if accepted {
			return line, err
		}

		if command != nil {
			continue
		}

		// Past the local keymap, our actions have a direct effect
		// on the line or on the cursor position, so we must first
		// "reset" or accept any completion state we're in, if any,
		// such as a virtually inserted candidate.
		rl.completer.Update()

		// 2 - Main keymap (vicmd/viins/emacs-*)
		bind, command, prefixed = rl.keymaps.MatchMain()
		if prefixed {
			continue
		}

		accepted, line, err = rl.run(bind, command)
		if accepted {
			return line, err
		}

		rl.keys.FlushUsed()
	}
}

func (rl *Shell) run(bind inputrc.Bind, command func()) (bool, string, error) {
	// Whether or not the command is resolved, let the macro
	// engine record the keys if currently recording a macro.
	rl.macros.RecordKeys(bind)

	// If the resolved bind is a macro itself, reinject its
	// bound sequence back to the key stack.
	if bind.Macro {
		macro := inputrc.Unescape(bind.Action)
		rl.keys.Feed(false, true, []rune(macro)...)
	}

	// And don't do anything else if we don't have a command.
	if command == nil {
		return false, "", nil
	}

	// The completion system might have control of the
	// input line and be using it with a virtual insertion,
	// so it knows which line and cursor we should work on.
	rl.line, rl.cursor, rl.selection = rl.completer.GetBuffer()

	// Run the command
	command()

	// Only run pending-operator commands when the command we
	// just executed has not had any influence on iterations.
	if !rl.iterations.IsPending() {
		rl.keymaps.RunPending()
	}

	rl.keys.FlushUsed() // Drop some or all keys (used ones)
	rl.checkCursor()    // Ensure cursor position is correct.

	// Drop iterations if required, or show them in the hint.
	hint := rl.iterations.ResetPostCommand()

	if hint != "" {
		rl.hint.Persist(hint)
	} else {
		rl.hint.ResetPersist()
	}

	// If the command just run was using the incremental search
	// buffer (acting on it), update the list of matches.
	rl.completer.UpdateIsearch()

	// Work is done: ask the completion system to
	// return the correct input line and cursor.
	rl.line, rl.cursor, rl.selection = rl.completer.GetBuffer()

	// History: save the last action to the line history,
	// and return with the call to the history system that
	// checks if the line has been accepted (entered), in
	// which case this will automatically write the history
	// sources and set up errors/returned line values.
	rl.undo.SaveWithCommand(bind)

	return rl.histories.LineAccepted()
}

func (rl *Shell) checkCursor() {
	switch rl.keymaps.Main() {
	case keymap.ViCmd, keymap.ViMove, keymap.Vi:
		rl.cursor.CheckCommand()
	default:
		rl.cursor.CheckAppend()
	}
}
