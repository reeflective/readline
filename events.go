package readline

import (
	"github.com/reeflective/readline/internal/keymap"
)

// EventCallback is a function that is called with a given key input (as typed by the user in the shell),
// to which is passed the current line and the current cursor position on this line.
// It returns an EventReturn, which specifies a target state (new line/cursor position, exit state, etc).
type EventCallback func(key string, line []rune, cursor int) EventReturn

// EventReturn is a structure returned by the callback event function.
// This is used by readline to determine what state the API should
// return to after the readline event.
type EventReturn struct {
	// Widget is the name of a widget to execute when the handler is called.
	// This can be used when the user wants an existing widget (eg. 'kill-line',
	// 'backward-word', etc) to modify the input line, rather than doing it himself.
	// An empty string is not a valid widget.
	Widget string

	// ForwardKey indicates if the key should be dispatched down to the shell widgets.
	// Example:
	// A handler modifying the line is bound by a user to the key "^X^Y", in Emacs mode.
	// The handler also returns a new line as 'NewLine', and a new cursor pos 'NewPos'.
	//
	// - If true: The shell first replaces its line and cursor pos with the ones given
	//   by the event return here. Then, it goes through its local/main keymaps to find
	//   a widget mapped to '^X^Y', which is also subsequently executed. The latter
	//   might again update the line and cursor position.
	// - If false: The shell replaces its line and cursor pos with the ones given by
	//   the even return, but will not try to find another widget mapped to '^X^Y'.
	ForwardKey bool

	// AcceptLine indicates if the shell should return from its read loop,
	// eg. close itself. If true, it will return the input line given in NewLine.
	AcceptLine bool

	// ClearHelpers indicates if completion and hint helpers should be cleared out of display.
	ClearHelpers bool

	// ToolTip is similar to HintText, except that it is displayed as a prompt
	// tooltip (similar to a right-side prompt) rather than below the input line.
	ToolTip string

	// HintText is a usage string printed below the input line when the callback is executed.
	HintText []rune

	NewLine []rune // NewLine is the new input line to use after the callback is executed.
	NewPos  int    // NewPos is the new cursor position to use.
}

// AddEventTest registers a bindkey handler for the given keyPress.
// It accepts an optional list of keymap modes for which to register the handler (eg. Vim visual/cmd/insert,
// emacs, completion, history, etc). If no list is passed, the event callback is mapped to all main keymaps
// of the shell, which is either emacs (in Emacs input mode), or viins/vicmd (in Vim input mode).
func (rl *Shell) AddEvent(key string, callback EventCallback, keymaps ...keymap.Mode) {
	if len(keymaps) == 0 {
		// keymaps = append(keymaps, emacsC)
		// keymaps = append(keymaps, viinsC)
	}

	// Prepare the caret decoder to be used.
	// buf := new(bytes.Buffer)
	// decoder := caret.Decoder{Writer: buf}
	//
	// // Add the callback to all keymaps
	// for _, mode := range keymaps {
	// 	if widgets, found := rl.widgets[mode]; found {
	// 		rl.bindWidget(key, "", &widgets, decoder, buf)
	// 	}
	// }
}

// DelEventTest deregisters an existing bindkey handler.
// It accepts an optional list of keymaps for which to deregister the handler.
// If this list is empty (or not passed), the bindkey handler is deregistered
// of all keymaps in which it is present. If the list is not empty, the bindkey
// handler is only deregistered from those keymaps, if it is found in them.
func (rl *Shell) DelEvent(key string, keymaps ...keymap.Mode) {
	// if len(keymaps) == 0 {
	// 	// for mode := range rl.config.Keymaps {
	// 	// 	keymaps = append(keymaps, mode)
	// 	// }
	// }
	//
	// // Decode the key
	// buf := new(bytes.Buffer)
	// decoder := caret.Decoder{Writer: buf}
	//
	// // Only decode the keys if the keybind is not a regexp expression
	// if !strings.HasPrefix(key, "[") || !strings.HasSuffix(key, "]") {
	// 	if _, err := decoder.Write([]byte(key)); err == nil {
	// 		key = buf.String()
	// 		buf.Reset()
	// 	}
	// }
	//
	// reg, err := regexp.Compile(key)
	// if err != nil || reg == nil {
	// 	return
	// }

	// Remove the callback from all keymaps
	// for _, mode := range keymaps {
	// 	if widgets, found := rl.widgets[mode]; found {
	// 		delete(widgets, reg)
	// 	}
	// }
}

func (rl *Shell) useEventHelpers(event EventReturn) {
	if event.ClearHelpers {
		rl.display.ResetHelpers()
	}

	// if len(event.HintText) > 0 {
	// 	rl.hint = event.HintText
	// }
	//
	// if len(event.ToolTip) > 0 {
	// 	rl.Prompt.tooltip = event.ToolTip
	// }
}
