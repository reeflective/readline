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

	rl.Prompt.PrimaryPrint()
	defer rl.Display.PrintTransientPrompt()

	rl.init()

	for {
		// Since we always update helpers after being asked to read
		// for user input again, we do it before actually reading it.
		rl.Display.Refresh()

		// Block and wait for user input.
		rl.Keys.Read()

		// 1 - Local keymap (completion/isearch/viopp)
		bind, command, prefixed := rl.Keymap.MatchLocal()
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
		rl.Completions.Update()

		// 2 - Main keymap (vicmd/viins/emacs-*)
		bind, command, prefixed = rl.Keymap.MatchMain()
		if prefixed {
			continue
		}

		accepted, line, err = rl.run(bind, command)
		if accepted {
			return line, err
		}

		rl.Keys.FlushUsed()

		// Reaching this point means the last key/sequence has not
		// been dispatched down to a command: therefore this key is
		// undefined for the current local/main keymaps.
		rl.handleUndefined(bind, command)
	}
}

// init gathers all steps to perform at the beginning of readline loop.
func (rl *Shell) init() {
	// Reset core editor components.
	rl.selection.Reset()
	rl.Buffers.Reset()
	rl.History.Reset()
	rl.Keys.Flush()
	rl.cursor.ResetMark()
	rl.cursor.Set(0)
	rl.line.Set([]rune{}...)
	rl.History.Save()
	rl.Iterations.Reset()

	// Some accept-* command must fetch a specific
	// line outright, or keep the accepted one.
	rl.History.Init()

	// Reset/initialize user interface components.
	rl.Hint.Reset()
	rl.Completions.ResetForce()
	rl.Display.Init(rl.SyntaxHighlighter)
}

// run wraps the execution of a target command/sequence with various pre/post actions
// and setup steps (buffers setup, cursor checks, iterations, key flushing, etc...)
func (rl *Shell) run(bind inputrc.Bind, command func()) (bool, string, error) {
	// Whether or not the command is resolved, let the macro
	// engine record the keys if currently recording a macro.
	rl.Macros.RecordKeys(bind)

	// If the resolved bind is a macro itself, reinject its
	// bound sequence back to the key stack.
	if bind.Macro {
		macro := inputrc.Unescape(bind.Action)
		rl.Keys.Feed(false, true, []rune(macro)...)
	}

	// And don't do anything else if we don't have a command.
	if command == nil {
		return false, "", nil
	}

	// The completion system might have control of the
	// input line and be using it with a virtual insertion,
	// so it knows which line and cursor we should work on.
	rl.line, rl.cursor, rl.selection = rl.Completions.GetBuffer()

	// The line and cursor are ready, we can run the command
	// along with any pending ones, and reset iterations.
	rl.execute(command)

	// When iterations are active, show them in hint section.
	hint := rl.Iterations.ResetPostCommand()

	if hint != "" {
		rl.Hint.Persist(hint)
	} else {
		rl.Hint.ResetPersist()
	}

	// If the command just run was using the incremental search
	// buffer (acting on it), update the list of matches.
	rl.Completions.UpdateIsearch()

	// Work is done: ask the completion system to
	// return the correct input line and cursor.
	rl.line, rl.cursor, rl.selection = rl.Completions.GetBuffer()

	// History: save the last action to the line history,
	// and return with the call to the history system that
	// checks if the line has been accepted (entered), in
	// which case this will automatically write the history
	// sources and set up errors/returned line values.
	rl.History.SaveWithCommand(bind)

	return rl.History.LineAccepted()
}

func (rl *Shell) execute(command func()) {
	command()

	// Only run pending-operator commands when the command we
	// just executed has not had any influence on iterations.
	if !rl.Iterations.IsPending() {
		rl.Keymap.RunPending()
	}

	// Flush the keys used to match against the command.
	rl.Keys.FlushUsed()

	// Update/check cursor positions after run.
	switch rl.Keymap.Main() {
	case keymap.ViCmd, keymap.ViMove, keymap.Vi:
		rl.cursor.CheckCommand()
	default:
		rl.cursor.CheckAppend()
	}
}

// handleUndefined is in charge of all actions to take when the
// last key/sequence was not dispatched down to a readline command.
func (rl *Shell) handleUndefined(bind inputrc.Bind, cmd func()) {
	if bind.Action != "" || cmd != nil {
		return
	}

	// Undefined keys incremental-search mode cancels it.
	if rl.Keymap.Local() == keymap.Isearch {
		rl.Hint.Reset()
		rl.Completions.Reset()
	}
}
