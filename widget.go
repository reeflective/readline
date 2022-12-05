package readline

import (
	// "regexp"
	"strings"
)

// action is represents the action of a widget, the number of times
// this widget needs to be run, and an optional operator argument.
// Most of the time we don't need this operator.
//
// Those actions are mostly used by widgets which make the shell enter
// the Vim operator pending mode, and thus require another key to be read.
type action struct {
	widget     string
	iterations int
	key        string
	operator   string
}

// getWidget looks in the various widget lists for a target widget,
// and if it finds it, sometimes will wrap it into a function so that
// all widgets look the same to the shell instance.
// This is so because some widgets, like Vim ones, don't return anything.
//
// The order in which those widgets maps are tested should not matter as
// long as there are no duplicates across any two of them.
func (rl *Instance) getWidget(name string) keyHandler {
	// Error widgets

	// Standard widgets (all editing modes/styles)
	if widget, found := standardWidgets[name]; found && widget != nil {
		return widget
	}

	// Standard line widgets, wrapped inside a compliant handler.
	if widget, found := standardLineWidgets[name]; found && widget != nil {
		return func(rl *Instance, _ []byte, _ int, _ []rune) (bool, bool, string, error) {
			read, ret, err := widget(rl)
			return read, ret, "", err
		}
	}

	// Emacs

	// Vim standard widgets don't return anything, wrap them in a simple call.
	if widget, found := standardViWidgets[name]; found && widget != nil {
		return func(rl *Instance, _ []byte, _ int, _ []rune) (bool, bool, string, error) {
			widget(rl)
			return false, false, "", nil
		}
	}

	// Non-standard Vim widgets require some input.
	if widget, found := viinsWidgets[name]; found && widget != nil {
		return widget
	}

	// Incremental search

	// Completion

	return nil
}

// run wraps a few calls for finding a widget and executing it, returning some basic
// instructions pertaining to what to do next: either keep reading input, or return the line.
func (rl *Instance) run(name string, b []byte, i int, r []rune) (ret bool, val string, err error) {
	widget := rl.getWidget(name)
	if widget == nil {
		return
	}

	// We matched a single widget, so reset
	// the current key as stored by the shell.
	rl.keys = ""

	// Execute the widget
	read, ret, val, err := widget(rl, b, i, r)
	if read || ret {
		return
	}

	// Any keymap caught before (if amy) has to expressly ask us
	// not to push "its effect" onto our undo stack. Thus if we're
	// here, we store the key in our Undo history (Vim mode).
	rl.undoAppendHistory()

	return
}

// matchPendingWidget processes a key against pending operation, first considering
// the key as an operator to this action. It executes anything that should be done
// here, updating any mode if needed, and notifies the caller if it must keep trying
// to match the key against the other (main) keymaps.
// func (rl *Instance) matchPendingWidget(key string) (read, ret bool, val string, err error) {
// If the key is a digit, we add it to the viopp-specific iterations
// if isDigit, _ := regexp.MatchString(`^([1-9]{1})$`, key); isDigit {
// 	rl.viIteration += key
//
// 	read = true
// 	return
// }

// // Since we can stack pending actions (like in 'y2Ft', where both y and F need
// // pending operators), if the current key matches any of the main keymap bindings,
// // we don't use it now as an argument, and must match it against the main keymap.
// if widget, found := rl.mainKeymap[key]; found && widget != "" {
// 	return
// }
//
// // Else, the key is taken as an argument to the last pending widget.
// rl.runPendingWidget(key)

// return
// }

// runPendingWidget finds the last widget pushed onto the
// pending stack and runs it against the provided input key.
func (rl *Instance) runPendingWidget(key string) {
	// Exit the pending operator mode if no more
	// widgets waiting for an argument operator.
	defer func() {
		if len(rl.pendingActions) == 0 {
			rl.exitVioppMode()
			rl.updateCursor()
		}
	}()

	pending := rl.getPendingWidget()

	if pending.widget == "" {
		return
	}

	widget := rl.getWidget(pending.widget)
	if widget == nil {
		return
	}

	// We matched a single widget, so reset
	// the current key as stored by the shell.
	rl.keys = ""

	// Permutate viIterations and pending iterations,
	// so that further operator iterations are used
	// within the widgets themselves.
	times := pending.iterations

	keys := []rune(key)

	// Run the widget with all navigation keys
	for i := 0; i < times; i++ {
		widget(rl, []byte{}, len(keys), keys)
	}
}

// getPendingWidget returns the last widget pushed onto the pending stack.
func (rl *Instance) getPendingWidget() (act action) {
	if len(rl.pendingActions) > 0 {
		act = rl.pendingActions[len(rl.pendingActions)-1]
		rl.pendingActions = rl.pendingActions[:len(rl.pendingActions)-1]
	}

	return
}

func findBindkeyWidget(key string, keymap keyMap) keyMap {
	widgets := make(keyMap)

	for wkey, widget := range keymap {
		if strings.HasPrefix(wkey, key) {
			widgets[wkey] = widget
		}
	}

	return widgets
}

// getWidget returns the first widget in the keymap
func getWidget(keymap keyMap) (key, widget string) {
	for key, widget := range keymap {
		return key, widget
	}

	return
}

func getWidgetMatch(key string, keymap keyMap) (widget string) {
	for wkey, widget := range keymap {
		if wkey == key {
			return widget
		}
	}
	return
}
