package readline

import (
	"io"
)

// loadInterruptHandlers maps all interrupt handlers to the shell.
func (rl *Instance) loadInterruptHandlers() {
	rl.interruptHandlers = map[string]lineWidget{
		string(charCtrlC): rl.errorCtrlC,
		string(charEOF):   rl.errorEOF,
	}
}

// isInterrupt returns true if the input key is an interrupt key.
func (rl *Instance) isInterrupt(keys string) (lineWidget, bool) {
	handler, found := rl.interruptHandlers[keys]

	return handler, found
}

// errorCtrlC is one of the special interrupt handlers, which behavior depends
// on our current shell mode: this is because this handler is not directly registered
// on one of our keymaps, and every input key is checked against this before keymaps.
func (rl *Instance) errorCtrlC(_ []rune) (read, ret bool, val string, err error) {
	err = CtrlC
	val = string(rl.line)
	rl.keys = ""

	if rl.modeTabCompletion {
		rl.resetVirtualComp(true)
		rl.resetHelpers()
		rl.renderHelpers()

		read = true
		return
	}
	rl.clearHelpers()

	ret = true
	return
}

// errorEOF is also a special interrupt handler, and has the
// same effect regardless of the current mode the shell is in.
func (rl *Instance) errorEOF(_ []rune) (read, ret bool, val string, err error) {
	err = io.EOF
	val = string(rl.line)
	rl.keys = ""

	rl.clearHelpers()

	ret = true
	return
}
