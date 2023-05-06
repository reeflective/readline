package macro

import (
	"fmt"
	"sort"
	"strings"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/ui"
)

// validMacroKeys - All valid macro IDs (keys) for read/write Vim registers.
var validMacroKeys = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789\""

// Engine manages all things related to keyboard macros:
// recording, dumping and feeding (running) them to the shell.
type Engine struct {
	recording  bool
	current    []rune          // Key sequence of the current macro being recorded.
	currentKey rune            // The identifier of the macro being recorded.
	macros     map[rune]string // All previously recorded macros.

	keys   *core.Keys // The engine feeds macros directly in the key stack.
	hint   *ui.Hint   // The engine notifies when macro recording starts/stops.
	status string     // The hint status displaying the currently recorded macro.
}

// NewEngine is a required constructor to setup a working macro engine.
func NewEngine(keys *core.Keys, hint *ui.Hint) *Engine {
	return &Engine{
		current: make([]rune, 0),
		macros:  make(map[rune]string),
		keys:    keys,
		hint:    hint,
	}
}

// StartRecord uses all key input to record a macro.
// If the key parameter is an alphanumeric character, the macro recorded will be
// stored and used through this letter argument, just like macros work in Vim.
// If the key is neither valid nor the null value, the engine does not start.
// A notification is given through the hint section.
func (e *Engine) StartRecord(key rune) {
	switch {
	case isValidMacroID(key), key == 0:
		e.currentKey = key
	default:
		return
	}

	e.recording = true
	e.status = color.Dim + "Recording macro: " + color.Bold
	e.hint.Persist(e.status)
}

// StopRecord stops using key input as part of a macro.
// A notification is given through the hint section.
func (e *Engine) StopRecord(keys []rune) {
	e.recording = false

	// Remove the hint.
	e.hint.ResetPersist()

	if len(e.current) == 0 {
		return
	}

	seq := strings.TrimSuffix(string(e.current), string(keys))
	macro := inputrc.EscapeMacro(seq)

	e.macros[e.currentKey] = macro
	e.macros[rune(0)] = macro

	e.current = make([]rune, 0)
}

// RecordKeys is being passed every key read by the shell, and will save
// those entered while the engine is in record mode. All others are ignored.
func (e *Engine) RecordKeys() {
	if !e.recording {
		return
	}

	// TODO: Should we only record the used keys ?
	keys, empty := e.keys.PeekAll()
	if empty || len(keys) == 0 {
		return
	}

	e.current = append(e.current, keys...)
	e.hint.Persist(e.status + inputrc.EscapeMacro(string(e.current)) + color.Reset)
}

// Recording returns true if the macro engine is recording the keys for a macro.
func (e *Engine) Recording() bool {
	return e.recording
}

// RunLastMacro feeds keys the last recorded macro to
// the shell's key stack, so that the macro is replayed.
func (e *Engine) RunLastMacro() {
	if len(e.macros) == 0 {
		return
	}

	macro := inputrc.Unescape(e.macros[rune(0)])

	if len(macro) == 0 {
		return
	}

	e.keys.Feed(false, []rune(macro)...)
}

// RunMacro runs a given macro, injecting its key sequence back into the shell key stack.
// The key argument should either be one of the valid alphanumeric macro identifiers, or
// a nil rune (in which case the last recorded macro is ran).
func (e *Engine) RunMacro(key rune) {
	if !isValidMacroID(key) && key != 0 {
		return
	}

	macro := e.macros[key]
	if len(macro) == 0 {
		return
	}

	e.keys.Feed(false, []rune(macro)...)
}

// PrintLastMacro dumps the last recorded macro to the screen.
func (e *Engine) PrintLastMacro() {
	if len(e.macros) == 0 {
		return
	}

	// Print the macro and the prompt.
	// The shell takes care of clearing itself
	// before printing, and refreshing after.
	fmt.Printf("\n%s\n", e.macros[e.currentKey])
}

// PrintAllMacros dumps all macros to the screen.
func (e *Engine) PrintAllMacros() {
	var macroIDs []rune

	for key := range e.macros {
		macroIDs = append(macroIDs, key)
	}

	sort.Slice(macroIDs, func(i, j int) bool {
		return macroIDs[i] < macroIDs[j]
	})

	for _, macro := range macroIDs {
		sequence := e.macros[macro]
		if sequence == "" {
			continue
		}

		if macro == 0 {
			macro = '"'
		}

		fmt.Printf("\"%s\": %s\n", string(macro), sequence)
	}
}

func isValidMacroID(key rune) bool {
	for _, char := range validMacroKeys {
		if char == key {
			return true
		}
	}

	return false
}
