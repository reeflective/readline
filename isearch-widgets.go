package readline

import (
	"bytes"
	"regexp"
	"unicode"

	"github.com/reiver/go-caret"
)

// The isearch keymap is empty by default: the widgets that can
// be used while in incremental search mode will be found in the
// main keymap, so that the same keybinds can be used.
//
// Completion widgets are added at bind time, so that completion
// can still be used while searching in them.
var isearchKeys = map[string]string{}

func (rl *Instance) isearchWidgets() lineWidgets {
	return map[string]widget{
		"incremental-search-forward":  rl.isearchForward,
		"incremental-search-backward": rl.isearchBackward,
		"isearch-insert":              rl.isearchInsert,
		"isearch-delete-char":         rl.isearchDeleteChar,
	}
}

// those widgets, generally found in the main keymap, are the only
// valid widgets to be used in the incremental search minibuffer.
var validIsearchWidgets = []string{
	"accept-and-infer-next-history",
	"accept-line",
	"accept-line-and-down-history",
	"accept-search",
	"backward-delete-char",
	"vi-backward-delete-char",
	"backward-kill-word",
	"backward-delete-word",
	"vi-backward-kill-word",
	"clear-screen",
	"history-incremental-search-forward",  // Not sure history- needed
	"history-incremental-search-backward", // same
	"space",
	"quoted-insert",
	"vi-quoted-insert",
	"vi-cmd-mode",
	"self-insert",
}

func (rl *Instance) isearchForward() {
	rl.skipUndoAppend()

	switch rl.local {
	case isearch:
	// case menuselect:
	default:
		// First initialize completions.
		if rl.completer != nil {
			rl.startMenuComplete(rl.completer)
		} else {
			rl.startMenuComplete(rl.generateCompletions)
		}

		// Then enter the isearch mode, which updates
		// the hint line, and initializes other things.
		rl.enterIsearchMode()
	}
}

func (rl *Instance) isearchBackward() {
}

func (rl *Instance) isearchInsert() {
}

func (rl *Instance) isearchDeleteChar() {
}

func (rl *Instance) isIsearchMode(mode keymapMode) bool {
	if mode != emacs && mode != viins && mode != vicmd {
		return false
	}

	if rl.local != isearch {
		return false
	}

	return true
}

func (rl *Instance) filterIsearchWidgets(mode keymapMode) (isearch widgets) {
	km := rl.config.Keymaps[mode]

	isearch = make(widgets)
	b := new(bytes.Buffer)
	decoder := caret.Decoder{Writer: b}

	for key, widget := range km {

		// Widget must be a valid isearch widget
		if !isValidIsearchWidget(widget) {
			continue
		}

		// Or bind to our temporary isearch keymap
		rl.bindWidget(key, widget, &isearch, decoder, b)
	}

	return
}

func isValidIsearchWidget(widget string) bool {
	for _, isw := range validIsearchWidgets {
		if isw == widget {
			return true
		}
	}

	return false
}

func (rl *Instance) enterIsearchMode() {
	rl.local = isearch
	rl.hintText = []rune(seqBold + seqFgCyan + "isearch: " + seqReset)
	rl.hintText = append(rl.hintText, rl.tfLine...)
}

// useIsearchLine replaces the input line with our current
// isearch buffer, the time for the widget to work on it.
func (rl *Instance) useIsearchLine() {
	rl.lineBuf = string(rl.line)
	rl.line = append([]rune{}, rl.tfLine...)

	cpos := rl.pos
	rl.pos = rl.tfPos
	rl.tfPos = cpos
}

// exitIsearchLine resets the input line to its original once
// the widget used in isearch mode has done its work.
func (rl *Instance) exitIsearchLine() {
	rl.tfLine = append([]rune{}, rl.line...)
	rl.line = []rune(rl.lineBuf)
	rl.lineBuf = ""

	cpos := rl.tfPos
	rl.tfPos = rl.pos
	rl.pos = cpos
}

func (rl *Instance) resetIsearch() {
	if rl.local != isearch {
		return
	}

	rl.tfLine = []rune{}
	rl.tfPos = 0
	rl.regexSearch = nil
}

// updateIsearch recompiles the isearch as a regex and
// filters matching candidates in the available completions.
func (rl *Instance) updateIsearch() {
	// First compile the search as regular expression
	var regexStr string
	if hasUpper(rl.tfLine) {
		regexStr = string(rl.tfLine)
	} else {
		regexStr = "(?i)" + string(rl.tfLine)
	}

	var err error
	rl.regexSearch, err = regexp.Compile(regexStr)
	if err != nil {
		rl.hintText = append(rl.hintText, []rune(seqFgRed+"Failed to compile search regexp")...)
	}

	rl.completer()

	// And filter out the completions.
	for _, g := range rl.tcGroups {
		g.updateTabFind(rl)
	}
}

func hasUpper(line []rune) bool {
	for _, r := range line {
		if unicode.IsUpper(r) {
			return true
		}
	}

	return false
}
