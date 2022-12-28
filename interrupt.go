package readline

import (
	"errors"
	"io"
)

// CtrlC is returned when ctrl+c is pressed.
var ErrCtrlC = errors.New("Ctrl+C")

// loadInterruptHandlers maps all interrupt handlers to the shell.
func (rl *Instance) loadInterruptHandlers() {
	rl.interruptHandlers = map[string]func() error{
		string(charCtrlC): rl.errorCtrlC,
		string(charEOF):   rl.errorEOF,
	}
}

// isInterrupt returns true if the input key is an interrupt key.
func (rl *Instance) isInterrupt(keys string) (func() error, bool) {
	handler, found := rl.interruptHandlers[keys]

	return handler, found
}

// errorCtrlC is one of the special interrupt handlers, which behavior depends
// on our current shell mode: this is because this handler is not directly registered
// on one of our keymaps, and every input key is checked against this before keymaps.
func (rl *Instance) errorCtrlC() (err error) {
	rl.keys = ""

	// When we have a completion inserted, just cancel the completions.
	if len(rl.comp) > 0 {
		rl.resetVirtualComp(true)
		rl.resetCompletion()
		rl.resetIsearch()
		rl.resetHintText()
		rl.completer = nil

		return
	}

	// Or return the current command line
	err = ErrCtrlC

	rl.clearHelpers()
	print("\r\n")

	return
}

// errorEOF is also a special interrupt handler, and has the
// same effect regardless of the current mode the shell is in.
func (rl *Instance) errorEOF() (err error) {
	err = io.EOF
	rl.keys = ""

	rl.clearHelpers()
	print("\r\n")

	return
}
