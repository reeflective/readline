package readline

import (
	"bytes"
	"regexp"
	"unicode"

	"github.com/reiver/go-caret"
)

func (rl *Instance) enterIsearchMode() {
	rl.local = isearch
	rl.hint = []rune(seqBold + seqFgCyan + "isearch: " + seqReset)
	rl.hint = append(rl.hint, rl.tfLine...)
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
	rl.isearch, err = regexp.Compile(regexStr)
	if err != nil {
		rl.hint = append(rl.hint, []rune(seqFgRed+"Failed to compile search regexp")...)
	}

	rl.completer()

	// And filter out the completions.
	for _, g := range rl.tcGroups {
		g.updateIsearch(rl)
	}
}

func (rl *Instance) isearchHint() {
	rl.hint = []rune(seqBold + seqFgCyan + "isearch: " + seqReset + seqBgDarkGray)
	rl.hint = append(rl.hint, rl.tfLine...)

	if rl.isearch == nil && len(rl.tfLine) > 0 {
		rl.hint = append(rl.hint, []rune(seqFgRed+" ! failed to compile search regexp")...)
	} else if rl.noCompletions() {
		rl.hint = append(rl.hint, []rune(seqFgRed+" ! no matches")...)
	}

	rl.hint = append(rl.hint, []rune(seqReset)...)
}

func (rl *Instance) resetIsearch() {
	if rl.local != isearch {
		return
	}

	rl.tfLine = []rune{}
	rl.tfPos = 0
	rl.isearch = nil
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

func hasUpper(line []rune) bool {
	for _, r := range line {
		if unicode.IsUpper(r) {
			return true
		}
	}

	return false
}
