package macro

import (
	"github.com/reeflective/readline/internal/common"
	"github.com/reeflective/readline/internal/ui"
	"github.com/xo/inputrc"
)

// Engine manages all things related to keyboard macros:
// recording, dumping and feeding (running) them to the shell.
type Engine struct {
	recording bool
	current   []rune   // The current macro being recorded.
	macros    []string // All previously recorded macros.

	keys *common.Keys // The engine feeds macros directly in the key stack.
	hint *ui.Hint     // The engine notifies when macro recording starts/stops.
}

// NewEngine is a required constructor to setup a working macro engine.
func NewEngine(keys *common.Keys, hint *ui.Hint) *Engine {
	return &Engine{
		current: make([]rune, 0),
		keys:    keys,
		hint:    hint,
	}
}

// StartRecord uses all key input to record a macro.
// A notification is given through the hint section.
func (e *Engine) StartRecord() {
	e.recording = true
}

// StopRecord stops using key input as part of a macro.
// A notification is given through the hint section.
func (e *Engine) StopRecord() {
	e.recording = false

	if len(e.current) == 0 {
		return
	}

	macro := inputrc.EscapeMacro(string(e.current))
	e.macros = append(e.macros, macro)
	e.current = make([]rune, 0)
}

// RecordKey is being passed every key read by the shell, and will save
// those entered while the engine is in record mode. All others are ignored.
func (e *Engine) RecordKey(key rune) {
	if !e.recording {
		return
	}

	e.current = append(e.current, key)
}

// RunLastMacro feeds keys the last recorded macro to
// the shell's key stack, so that the macro is replayed.
func (e *Engine) RunLastMacro() {
	if len(e.macros) == 0 {
		return
	}

	macro := inputrc.Unescape(e.macros[len(e.macros)-1])

	// Since this method is called within a command,
	// and that the key having triggered that command
	// have not been flushed yet, we flush them first.
	// The key management routine will not flush them
	e.keys.Feed(false, []rune(macro)...)
}

// RunMacro runs a given macro, injecting its
// key sequence back into the shell key stack.
func (e *Engine) RunMacro() {}

// PrintLastMacro dumps the last recorded macro to the screen.
func (e *Engine) PrintLastMacro() {
	if len(e.macros) == 0 {
		return
	}

	// Go below the current line, clear helpers,
	// print the macro and the prompt, then let
	// the shell do its job refreshing.

	print(e.macros[len(e.macros)-1])
}

// PrintAllMacros dumps all macros to the screen.
func (e *Engine) PrintAllMacros() {}
