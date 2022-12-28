package readline

import (
	"bytes"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/reiver/go-caret"
)

// These handlers are mostly (if not only) used in the main readline loop (entrypoint)
// and are thus the first dispatcher used when receiving a key sequence.
// Thus, they are the only handlers that can tell the shell either to keep
// reading input, or to return the entire line to the readline caller.
type widget func()

// widgets maps keys (either in caret or hex notation) to an EventCallback,
// which wraps the corresponding widget for this key.
// Those widgets maps are built at start/config reload time.
//
// We compile each keybind as a regular expression, to allow
// for ranges of characters to share the same widget.
type widgets map[*regexp.Regexp]EventCallback

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

// bindWidgets goes through all "key-sequence":"widget" pair in all keymaps,
// decoding the key from caret notation (if used) and compiling it as a regular
// expression, and binds it to the internal widget list for the given mode.
func (rl *Instance) bindWidgets() {
	rl.widgets = make(map[keymapMode]widgets)

	// Since the key might be in caret notation, we decode the key
	// first, so that when we can match the key as detected by the
	// shell (in ASCII notation).
	b := new(bytes.Buffer)
	decoder := caret.Decoder{Writer: b}

	for mode, km := range rl.config.Keymaps {
		keymapWidgets := make(widgets)

		for key, widget := range km {
			rl.bindWidget(key, widget, &keymapWidgets, decoder, b)
		}

		rl.widgets[mode] = keymapWidgets
	}

	switch rl.config.InputMode {
	case Emacs:
		rl.main = emacs
	case Vim:
		rl.main = viins
	}
}

// run is in charge of executing the matched EventCallback, unwrapping its values and return behavior
// parameters (errors/lines/read), and optionally to execute pending widgets (vi operator pending mode),.
func (rl *Instance) run(cb EventCallback, keys string, mode keymapMode) {
	if cb == nil {
		return
	}

	// Use the minibuffer if currently working in isearch mode.
	if rl.isIsearchMode(mode) {
		rl.useIsearchLine()
		defer rl.updateIsearch() // Order matters: defer is executed in inverse order.
		defer rl.exitIsearchLine()
	}

	// Run the callback, and by default, use its behavior for return values
	event := cb(keys, rl.line, rl.pos)
	rl.accepted = event.CloseReadline
	rl.line = append(event.NewLine, []rune{}...)
	rl.pos = event.NewPos

	// Update/reset helpers
	if event.ClearHelpers {
		rl.resetHelpers()
	}

	if len(event.HintText) > 0 {
		rl.hint = event.HintText
	}

	if len(event.ToolTip) > 0 {
		rl.Prompt.tooltip = event.ToolTip
	}

	// If the callback has a widget, run it. Any instruction to return, or an error
	// being raised has precedence over other callback read/return settings.
	if event.Widget != "" {
		rl.runWidget(event.Widget)
		if rl.accepted || rl.err != nil {
			return
		}
	}

	// If we are asked to close the readline, we don't care about pending operations.
	if event.CloseReadline {
		rl.clearHelpers()
		rl.accepted = true

		return
	}

	// If we don't have to dispatch the key to next keymaps
	// (in the same loop), we are done with this callback.
	// This is the default for all builtin widgets.
	// TODO: What to do here
	if !event.ForwardKey {
	}

	// Finally, we might have any pending widget to run.
	if rl.isViopp {
		rl.runPendingWidget()
	}
}

// bindWidget wraps a widget into an EventCallback and binds it to the corresponding keymap.
func (rl *Instance) bindWidget(key, widget string, keymap *widgets, decoder caret.Decoder, buf *bytes.Buffer) {
	// When the key is a regular expression range, we add some metacharacters
	// to force the regex to match the entire string that we will give later.
	if isRegexCapturingGroup(key) {
		key = "^" + key + "$"
	} else {
		// Or decode the key in case its in caret notation.
		if _, err := decoder.Write([]byte(key)); err == nil {
			key = buf.String()
			buf.Reset()
		}

		// Quote all metacharacters before compiling to regex
		key = regexp.QuoteMeta(key)
	}

	reg, err := regexp.Compile(key)
	if err != nil || reg == nil {
		return
	}

	callback := func(_ string, line []rune, pos int) *EventReturn {
		event := &EventReturn{
			Widget:  widget,
			NewLine: line,
			NewPos:  pos,
		}

		return event
	}

	// Bind the wrapped widget.
	(*keymap)[reg] = callback
}

// getWidget looks in the various widget lists for a target widget,
// and if it finds it, sometimes will wrap it into a function so that
// all widgets look the same to the shell instance.
func (rl *Instance) getWidget(name string) widget {
	// Standard widgets (all editing modes/styles)
	if wg, found := rl.standardWidgets()[name]; found && wg != nil {
		return wg
	}

	// Vim standard widgets don't return anything, wrap them in a simple call.
	if wg, found := rl.viWidgets()[name]; found && wg != nil {
		return wg
	}

	// History control widgets
	if wg, found := rl.historyWidgets()[name]; found && wg != nil {
		return wg
	}

	// Completion & incremental search
	if wg, found := rl.completionWidgets()[name]; found && wg != nil {
		return wg
	}

	return nil
}

// matchWidgets returns all widgets matching the current key either perfectly, as a prefix,
// or as one of the possible values matched by a regular expression.
func (rl *Instance) matchWidgets(key string, wids widgets) (cb EventCallback, all widgets) {
	all = make(widgets)

	// Sort binds
	binds := make([]*regexp.Regexp, 0)
	for regex := range wids {
		binds = append(binds, regex)
	}
	sort.Slice(binds, func(i, j int) bool {
		return strconv.QuoteToASCII(binds[i].String()) < strconv.QuoteToASCII(binds[j].String())
	})

	// Test against each regular expression.
	for _, r := range binds {
		reg := *r
		widget := wids[r]

		match := reg.MatchString(key)

		// No match is only valid if the keys are a valid prefix to the keybind.
		if !match {
			if strings.HasPrefix(reg.String(), key) && reg.String() != key && key != "" {
				all[&reg] = widget
			}
			continue
		}

		// When we only match as a prefix, and that the key is not a lone caret.
		if strings.HasPrefix(reg.String(), key) && len(key) < len(reg.String()) && key != "^" {
			all[&reg] = widget

			continue
		}

		// If the match is perfect, then we have
		// a default callback to use/store.
		cb = widget
	}

	return
}

// runWidget wraps a few calls for finding a widget and executing it, returning some basic
// instructions pertaining to what to do next: either keep reading input, or return the line.
func (rl *Instance) runWidget(name string) {
	targetWidget := rl.getWidget(name)
	if targetWidget == nil {
		return
	}

	// We matched a single widget, so reset
	// the current key as stored by the shell.
	defer func() {
		rl.widgetPrefixMatched = nil
		rl.keys = ""
	}()

	// Execute the widget
	targetWidget()
	if rl.accepted {
		return
	}

	// Any keymap caught before (if any) has to expressly ask us
	// not to push "its effect" onto our undo stack. Thus if we're
	// here, we store the key in our Undo history (Vim mode).
	rl.undoHistoryAppend()
}

// runPendingWidget finds the last widget pushed onto the
// pending stack and runs it against the provided input key.
func (rl *Instance) runPendingWidget() {
	defer rl.donePending()

	pending := rl.getPendingWidget()

	if pending.widget == "" {
		return
	}

	pendingWidget := rl.getWidget(pending.widget)
	if pendingWidget == nil {
		return
	}

	for i := 0; i < pending.iterations; i++ {
		pendingWidget()
		if rl.accepted {
			return
		}
	}

	// Any remaining pending widget will wait for the following key.
	rl.keys = ""

	// The pending widget might have its own effect on the line.
	rl.undoHistoryAppend()
}

// getPendingWidget returns the last widget pushed onto the pending stack.
func (rl *Instance) getPendingWidget() (act action) {
	if len(rl.pendingActions) > 0 {
		act = rl.pendingActions[len(rl.pendingActions)-1]
		rl.pendingActions = rl.pendingActions[:len(rl.pendingActions)-1]
	}

	return
}

func (rl *Instance) donePending() {
	if len(rl.pendingActions) == 0 {
		rl.exitVioppMode()
		rl.updateCursor()
	}
}

// regular expressions as keybinds are only allowed when expressed within a (global) capturing group.
func isRegexCapturingGroup(key string) bool {
	if (strings.HasPrefix(key, "[") && strings.HasSuffix(key, "]")) ||
		(strings.HasPrefix(key, "(") && strings.HasSuffix(key, ")")) {
		return true
	}

	return false
}

// not used in the current scheme of things.
func (rl *Instance) selfInsertWidget() EventCallback {
	return func(_ string, line []rune, pos int) *EventReturn {
		event := &EventReturn{
			Widget:  "self-insert",
			NewLine: line,
			NewPos:  pos,
		}

		return event
	}
}
