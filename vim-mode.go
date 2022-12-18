package readline

import (
	"fmt"
)

//
// Mode switching ------------------------------------------------------------------ //
//

func (rl *Instance) enterVisualMode() {
	rl.local = visual
	rl.visualLine = false
	rl.mark = rl.pos
}

func (rl *Instance) enterVisualLineMode() {
	rl.local = visual
	rl.visualLine = true
	rl.mark = 0 // start at the beginning of the line.
}

func (rl *Instance) exitVisualMode() {
	for i, reg := range rl.regions {
		if reg.regionType == "visual" {
			if len(rl.regions) > i {
				rl.regions = append(rl.regions[:i], rl.regions[i+1:]...)
			}
		}
	}

	rl.visualLine = false
	rl.mark = -1

	if rl.local != visual {
		return
	}

	rl.local = ""
}

// enterVioppMode adds a widget to the list of widgets waiting for an operator/action,
// enters the vi operator pending mode and updates the cursor.
func (rl *Instance) enterVioppMode(widget string) {
	// When the widget is empty, we just want to update the cursor.
	if widget == "" {
		rl.isViopp = true
		return
	}

	rl.local = viopp

	act := action{
		widget:     widget,
		iterations: rl.getIterations(),
	}

	// Push the widget on the stack of widgets
	rl.pendingActions = append(rl.pendingActions, act)
}

func (rl *Instance) exitVioppMode() {
	if rl.local == viopp {
		rl.local = ""
	}
	rl.isViopp = false
}

// isVimEscape checks if the key matches the custom Vim mode escapes,
// and returns the corresponding callback if it matches.
func (rl *Instance) isVimEscape(key string) (cb EventCallback, yes bool) {
	if rl.main == emacs {
		return
	}

	// Make the callback even if not used
	cb = func(_ string, _ []rune, _ int) *EventReturn {
		// Reset any incremental search parameters
		// if we were currently editing its buffer.
		rl.resetIsearch()
		rl.resetHintText()

		// Most of the time this will be caught as if this
		// callback/widget was found on the local keymap,
		// thus avoiding the update of the completion system:
		// do it here instead.
		rl.updateCompletionState()

		event := &EventReturn{
			Widget:  "vi-cmd-mode",
			NewLine: rl.line,
			NewPos:  rl.pos,
		}

		return event
	}

	// Escape is builtin
	if len(key) == 1 && key[0] == charEscape {
		return cb, true
	}

	return
}

//
// Mode printing ------------------------------------------------------------------ //
//

const (
	vimInsertStr      = "[I]"
	vimReplaceOnceStr = "[V]"
	vimReplaceManyStr = "[R]"
	vimDeleteStr      = "[D]"
	vimKeysStr        = "[N]"
)

func (rl *Instance) refreshVimStatus() {
	rl.Prompt.compute(rl)
	rl.updateHelpers()
}

// viHintMessage - lmorg's way of showing Vim status is to overwrite the hint.
// Currently not used, as there is a possibility to show the current Vim mode in the prompt.
func (rl *Instance) viHintMessage() {
	defer func() {
		rl.clearHelpers()
		rl.renderHelpers()
	}()

	// The internal VI operator pending is used when
	// we don't bother changing the keymap just to read a key.
	if rl.isViopp {
		vioppMsg := fmt.Sprintf("viopp: %s", rl.keys)
		rl.hintText = []rune(vioppMsg)
		return
	}

	// The local keymap, most of the time, has priority
	switch rl.local {
	case viopp:
		vioppMsg := fmt.Sprintf("viopp: %s", rl.keys)
		rl.hintText = []rune(vioppMsg)
		return
	case visual:
		return
	}

	// But if not, we check for the global keymap
	switch rl.main {
	case viins:
		rl.hintText = []rune("-- INSERT --")
	case vicmd:
		rl.hintText = []rune("-- VIM KEYS -- (press `i` to return to normal editing mode)")
	default:
		rl.getHintText()
	}
}
