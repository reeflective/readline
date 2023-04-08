package macro

import (
	"fmt"
	"strings"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/ui"
	"github.com/reeflective/readline/inputrc"
)

// Engine manages all things related to keyboard macros:
// recording, dumping and feeding (running) them to the shell.
type Engine struct {
	recording bool
	current   []rune   // The current macro being recorded.
	macros    []string // All previously recorded macros.

	keys   *core.Keys // The engine feeds macros directly in the key stack.
	hint   *ui.Hint   // The engine notifies when macro recording starts/stops.
	status string     // The hint status displaying the currently recorded macro.
}

// NewEngine is a required constructor to setup a working macro engine.
func NewEngine(keys *core.Keys, hint *ui.Hint) *Engine {
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
	e.status = color.Dim + "Recording macro: " + color.Bold
	e.hint.Set(e.status)
}

// StopRecord stops using key input as part of a macro.
// A notification is given through the hint section.
func (e *Engine) StopRecord() {
	e.recording = false

	// Remove the hint.
	if strings.HasPrefix(e.hint.Text(), color.Dim+"Recording macro:") {
		e.hint.Reset()
	}

	if len(e.current) == 0 {
		return
	}

	macro := inputrc.EscapeMacro(string(e.current))
	e.macros = append(e.macros, macro)
	e.current = make([]rune, 0)
}

// RecordKeys is being passed every key read by the shell, and will save
// those entered while the engine is in record mode. All others are ignored.
func (e *Engine) RecordKeys(bind inputrc.Bind) {
	if !e.recording || bind.Action == "end-kbd-macro" {
		return
	}

	keys, empty := e.keys.PeekAll()
	if empty || len(keys) == 0 {
		return
	}

	e.current = append(e.current, keys...)
	e.hint.Set(e.status + inputrc.EscapeMacro(string(e.current)))
}

// RunLastMacro feeds keys the last recorded macro to
// the shell's key stack, so that the macro is replayed.
func (e *Engine) RunLastMacro() {
	if len(e.macros) == 0 {
		return
	}

	macro := inputrc.Unescape(e.macros[len(e.macros)-1])
	e.keys.Feed(false, true, []rune(macro)...)
}

// RunMacro runs a given macro, injecting its
// key sequence back into the shell key stack.
func (e *Engine) RunMacro() {}

// PrintLastMacro dumps the last recorded macro to the screen.
func (e *Engine) PrintLastMacro() {
	if len(e.macros) == 0 {
		return
	}

	// Print the macro and the prompt.
	// The shell takes care of clearing itself
	// before printing, and refreshing after.
	fmt.Printf("\n%s\n", e.macros[len(e.macros)-1])
}

// PrintAllMacros dumps all macros to the screen.
func (e *Engine) PrintAllMacros() {}
