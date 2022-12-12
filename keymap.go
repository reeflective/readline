package readline

import (
	"bytes"
	"regexp"

	"github.com/reiver/go-caret"
)

// keymapMode is a root keymap mode for the shell.
// To each of these keymap modes is bound a keymap.
type keymapMode string

// keymap maps a key (either in caret or hex notation)
// to the name of the widget to run when key is pressed.
type keymap map[string]string

// widgets maps keys (either in caret or hex notation) to an EventCallback,
// which wraps the corresponding widget for this key. Those widgets maps are
// built at start/config reload time.
type widgets map[string]EventCallback

// These are the root keymaps used in the readline shell.
// Their functioning is similar to how ZSH organizes keymaps.
const (
	// Editor
	emacs  keymapMode = "emacs"
	viins  keymapMode = "viins"
	vicmd  keymapMode = "vicmd"
	viopp  keymapMode = "viopp"
	visual keymapMode = "visual"
	// Completion and search
	isearch    keymapMode = "isearch"
	menuselect keymapMode = "menuselect"
)

// These handlers are mostly (if not only) used in the main readline loop (entrypoint)
// and are thus the first dispatcher used when receiving a key sequence.
// Thus, they are the only handlers that can tell the shell either to keep
// reading input, or to return the entire line to the readline caller.
//
// These handlers return the following values:
// @read =>     read the next character at the input line
// @return =>   Return the line read before starting a new readline loop
// @val    =>   The string returned to the readline caller, generally the line input, or nothing.
// @error =>    Any error caught, generally those returned on signals like CtrlC
type lineWidget func(r []rune) (bool, bool, string, error)

// loadKeymapWidgets is ran once at the beginning of an instance start.
// It is in charge of setting the configured/default input mode, which will
// have an effect on which and how subsequent keymaps will be interpreted.
func (rl *Instance) loadKeymapWidgets() {
	rl.widgets = make(map[keymapMode]widgets)

	// Since the key might be in caret notation, we decode the key
	// first, so that when we can match the key as detected by the
	// shell (in ASCII notation).
	b := new(bytes.Buffer)
	decoder := caret.Decoder{Writer: b}

	// And for each keymap, initialize the widget
	// map and load the widgets into it.
	for mode, km := range rl.config.Keymaps {
		keymapWidgets := make(widgets)
		for key, widget := range km {

			// First decode the key, if in caret notation.
			if _, err := decoder.Write([]byte(key)); err == nil {
				key = b.String()
				b.Reset()
			}

			// And use the potentially decoded key to map the widget.
			rl.bindWidget(key, widget, &keymapWidgets)
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

func (rl *Instance) initKeymap() {
	switch rl.config.InputMode {
	case Emacs:
		rl.main = emacs
	case Vim:
		rl.main = viins
	}

	if rl.main == vicmd {
		rl.viInsertMode()
	}

	rl.local = ""

	rl.updateCursor()
}

// updateKeymaps is in charge of ensuring the correct referencing of the main/global
// keymap, so that correct key dispatching can occur with local keymaps as well.
func (rl *Instance) updateKeymaps() {
	// Ensure the main keymap is valid.
	if rl.main != emacs && rl.main != viins && rl.main != vicmd {
		rl.main = emacs
	}

	// And set the special regexp keymap according to the main one.
	switch rl.main {
	case emacs:
		rl.specialKeymap = emacsSpecialKeymaps
	case vicmd, viins:
		rl.specialKeymap = vicmdSpecialKeymaps
	}

	// When matching a widget, we need to know if the shell was in operator
	// pending mode before trying to match the key against our keymaps.
	rl.viopp = rl.local == viopp
}

// matchKeymap checks if the provided key matches a precise widget, or if only a prefix
// is matched. When only a prefix is matched, the shell keeps reading for another key.
func (rl *Instance) matchKeymap(key string, kmode keymapMode) (cb EventCallback, prefix bool) {
	if kmode == "" {
		return nil, false
	}

	matchWidgets := rl.widgets[kmode]
	filtered := findBindkeyWidget(key, matchWidgets)

	// When we have absolutely no matching widget for the keys,
	// we either return, or if we have a perfectly matching one
	// waiting for an input, we execute it.
	if len(filtered) == 0 {
		if rl.prefixMatchedWidget != nil {
			cb = rl.prefixMatchedWidget
			rl.keys = key
			rl.prefixMatchedWidget = nil
		} else {
			rl.keys = ""
		}
		return
	}

	// The escape key is a special key that bypass the entire process.
	// TODO: HERE IS WHERE WE SHOULD CHECK FOR SPECIAL ESCAPES.
	if len(key) == 1 && key[0] == charEscape && rl.main != emacs {
		cb = matchWidgets[key]
		rl.keys = ""
		return
	}

	// Or several matches, in which case we must read another key.
	if len(filtered) > 1 {
		rl.prefixMatchedWidget = matchWidgets[key]
		return nil, true
	}

	// Or only one, but we might only have prefix,
	// in which case the widget is still empty.
	if cb = getWidgetMatch(key, filtered); cb == nil {
		return nil, true
	}

	return
}

// matchRegexKeymap sequentially tests for a matching regexp in the special keymap
func (rl *Instance) matchRegexKeymap(key string) (widget string) {
	for regex, widget := range rl.specialKeymap {

		matcher, err := regexp.Compile(regex)
		if err != nil {
			continue
		}

		if match := matcher.MatchString(key); match {
			return widget
		}
	}

	return
}
