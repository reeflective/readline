package readline

import (
	"errors"
	"io"
)

// ErrCtrlC is returned when ctrl+c is pressed.
var ErrCtrlC = errors.New("Ctrl+C")

// loadInterruptHandlers maps all interrupt handlers to the shell.
func (rl *Instance) loadInterruptHandlers() {
	rl.interruptHandlers = map[rune]func() error{
		charCtrlC: rl.errorCtrlC,
		charEOF:   rl.errorEOF,
	}
}

// isInterrupt returns true if the input key is an interrupt key.
func (rl *Instance) isInterrupt(keys string) (func() error, bool) {
	if len(keys) > 1 {
		return nil, false
	}

	key := rune(keys[0])
	handler, found := rl.interruptHandlers[key]

	return handler, found
}

// errorCtrlC is one of the special interrupt handlers, which behavior depends
// on our current shell mode: this is because this handler is not directly registered
// on one of our keymaps, and every input key is checked against this before keymaps.
func (rl *Instance) errorCtrlC() error {
	rl.keys = ""

	// When we have a completion inserted, just cancel the completions.
	if len(rl.comp) > 0 {
		rl.resetVirtualComp(true)
		rl.resetCompletion()
		rl.resetIsearch()
		rl.resetHintText()
		rl.completer = nil

		return nil
	}

	// Or return the current command line
	rl.clearHelpers()
	print("\r\n")

	return ErrCtrlC
}

// errorEOF is also a special interrupt handler, and has the
// same effect regardless of the current mode the shell is in.
func (rl *Instance) errorEOF() error {
	rl.keys = ""

	rl.clearHelpers()
	print("\r\n")

	return io.EOF
}
