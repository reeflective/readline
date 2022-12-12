package readline

import (
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

// run is in charge of executing the matched EventCallback, unwrapping its values and return behavior
// parameters (errors/lines/read), and optionally to execute pending widgets (vi operator pending mode),
func (rl *Instance) run(cb EventCallback, keys string) (read, ret bool, val string, err error) {
	if cb == nil {
		read = true
		return
	}

	// Run the callback, and by default, use its behavior for return values
	event := cb(keys, rl.line, rl.pos)
	ret = event.CloseReadline
	rl.line = append(event.NewLine, []rune{}...)
	rl.pos = event.NewPos

	// Update/reset helpers
	if event.ClearHelpers {
		rl.resetHelpers()
	}

	if len(event.HintText) > 0 {
		rl.hintText = event.HintText
		rl.updateHelpers()
	}

	// If the callback has a widget, run it. Any instruction to return, or an error
	// being raised has precedence over other callback read/return settings.
	if event.Widget != "" {
		ret, val, err = rl.runWidget(event.Widget, []rune(keys))
		if ret || err != nil {
			return
		}
	}

	// If we are asked to close the readline, we don't care about pending operations.
	if event.CloseReadline {
		rl.clearHelpers()
		ret = true
		val = string(rl.line)

		return
	}

	// If we don't have to dispatch the key to next keymaps
	// (in the same loop), we are done with this callback.
	// This is the default for all builtin widgets.
	if !event.ForwardKey {
		read = true

		return
	}

	return
}

// bindWidget wraps a widget into an EventCallback and binds it to the corresponding keymap.
// The event callback is basically empty as far as functionality is concerned: it just returns
// the name of a widget to be run, and specifies some additional behavior.
func (rl *Instance) bindWidget(key, widget string, km *widgets) {
	cb := func(_ string, line []rune, pos int) *EventReturn {
		event := &EventReturn{
			Widget:  widget,
			NewLine: line,
			NewPos:  pos,
		}

		return event
	}

	// Bind the wrapped widget.
	(*km)[key] = cb
}

// getWidget looks in the various widget lists for a target widget,
// and if it finds it, sometimes will wrap it into a function so that
// all widgets look the same to the shell instance.
// This is so because some widgets, like Vim ones, don't return anything.
//
// The order in which those widgets maps are tested should not matter as
// long as there are no duplicates across any two of them.
func (rl *Instance) getWidget(name string) lineWidget {
	// Error widgets

	// Standard widgets (all editing modes/styles)
	if widget, found := rl.commonWidgets()[name]; found && widget != nil {
		return func(_ []rune) (bool, bool, string, error) {
			widget()
			return false, false, "", nil
		}
	}

	// Standard line widgets, wrapped inside a compliant handler.
	if widget, found := rl.commonLineWidgets()[name]; found && widget != nil {
		return func(keys []rune) (bool, bool, string, error) {
			read, ret, val, err := widget(keys)
			return read, ret, val, err
		}
	}

	// Emacs

	// Vim standard widgets don't return anything, wrap them in a simple call.
	if widget, found := rl.viWidgets()[name]; found && widget != nil {
		return func(_ []rune) (bool, bool, string, error) {
			widget()
			return false, false, "", nil
		}
	}

	// Incremental search

	// Completion

	return nil
}

// runWidget wraps a few calls for finding a widget and executing it, returning some basic
// instructions pertaining to what to do next: either keep reading input, or return the line.
func (rl *Instance) runWidget(name string, keys []rune) (ret bool, val string, err error) {
	widget := rl.getWidget(name)
	if widget == nil {
		return
	}

	// We matched a single widget, so reset
	// the current key as stored by the shell.
	defer func() {
		rl.prefixMatchedWidget = nil
		rl.keys = ""
	}()

	// Execute the widget
	read, ret, val, err := widget(keys)
	if read || ret {
		return
	}

	// Any keymap caught before (if amy) has to expressly ask us
	// not to push "its effect" onto our undo stack. Thus if we're
	// here, we store the key in our Undo history (Vim mode).
	rl.undoAppendHistory()

	return
}

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

	// Any remaining pending widget
	// will wait for the following key.
	rl.keys = ""

	// Permutate viIterations and pending iterations,
	// so that further operator iterations are used
	// within the widgets themselves.
	times := pending.iterations

	keys := []rune(key)

	// Run the widget with all navigation keys
	for i := 0; i < times; i++ {
		widget(keys)
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

func findBindkeyWidget(key string, wid widgets) widgets {
	kmWidgets := make(widgets)

	for wkey := range wid {
		// for wkey := range bindings {
		if strings.HasPrefix(wkey, key) {
			kmWidgets[wkey] = wid[wkey]
		}
	}

	return kmWidgets
}

// getWidget returns the first widget in the keymap
func getWidget(km widgets) (key string, widget EventCallback) {
	for key, widget := range km {
		return key, widget
	}

	return
}

func getWidgetMatch(key string, km widgets) (widget EventCallback) {
	widget = km[key]
	return
}
