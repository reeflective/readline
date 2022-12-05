package readline

import (
	"regexp"
)

// keymapMode is a root keymap mode for the shell.
// To each of these keymap modes is bound a keymap.
type keymapMode string

type keyMap map[string]string

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
type keyHandler func(rl *Instance, b []byte, i int, r []rune) (bool, bool, string, error)

// errorHandlers maps some special keys directly to their handlers, since
// those handlers do not have a corresponding name, eg. they are anonymous.
var errorHandlers = map[byte]keyHandler{
	// charCtrlC: errorCtrlC,
	// charEOF:   errorEOF,
}

// var baseHandlers = readlineHandlers{
// // HERE DO CTRLC AND EOF
// case '\r':
// 	fallthrough
//
// // Completion
// "menu-complete":         menuSelect,          // Tab
// "search-complete":       searchComplete,      // CtrlF
// "exit-complete":         exitComplete,        // CtrlG
// "history-menu-complete": historyMenuComplete, // CtrlR
// }

// setBaseKeymap is ran once at the beginning of an instance start.
// It is in charge of setting the configured/default input mode,
// which will have an effect on which and how subsequent keymaps
// will be interpreted.
func (rl *Instance) setBaseKeymap() {
	// Bind all default keymaps
	rl.keymaps = map[keymapMode]keyMap{
		emacs:  emacsKeymaps,
		viins:  viinsKeymaps,
		vicmd:  vicmdKeymaps,
		visual: visualKeymaps,
		viopp:  vioppKeymaps,
	}

	// TODO: Change this hardcoding
	// Link the configured/current keymap to keymap 'main'
	rl.main = viins

	// TODO here if emacs main, bind special regexp keymap.
}

func (rl *Instance) initKeymap() {
	// TODO: Maybe we should keep the same current Vi mode instead of defaulting to one.
	// In Vim mode, we always start in Input mode. The prompt needs this.
	if rl.main == vicmd {
		rl.main = viins
	}

	// There is no local keymap set by default
	rl.local = ""
	rl.localKeymap = make(keyMap)

	if rl.main == viins {
		print(cursorBlinkingBeam)
	}
}

// updateKeymaps is in charge of ensuring the correct referencing of the main/global
// keymap, so that correct key dispatching can occur with local keymaps as well.
func (rl *Instance) updateKeymaps() {
	// Ensure the main keymap is valid, defaulting to emacs if not.
	if rl.main != emacs && rl.main != viins && rl.main != vicmd {
		rl.main = emacs
	}

	// Bind the corresponding keymaps for the main one.
	globalKeymap, found := rl.keymaps[rl.main]
	if !found {
		rl.mainKeymap = emacsKeymaps
	} else {
		rl.mainKeymap = globalKeymap
	}

	// Bind the corresponding keymaps for the local one.
	localKeymap, found := rl.keymaps[rl.local]
	if !found {
		rl.localKeymap = make(keyMap)
	} else {
		rl.localKeymap = localKeymap
	}

	// Finally, set the special regexp keymap according to the main one.
	switch rl.main {
	case emacs:
		rl.specialKeymap = emacsSpecialKeymaps
	case vicmd, viins:
		rl.specialKeymap = vicmdSpecialKeymaps
	}
}

// matchKeymap checks if the provided key matches a precise widget, or if only a prefix
// is matched. When only a prefix is matched, the shell keeps reading for another key.
func (rl *Instance) matchKeymap(key string, keymap keyMap) (widget string, prefix bool) {
	filtered := findBindkeyWidget(key, keymap)

	// We either have no match, so we reset the keys.
	if len(filtered) == 0 {
		rl.keys = ""
		return "", false
	}

	// The escape key is a special key that bypass the entire process.
	if len(key) == 1 && key[0] == charEscape {
		widget = keymap[key]
		rl.keys = ""
		return
	}

	// Or several matches, in which case we must read another key.
	if len(filtered) > 1 {
		return "", true
	}

	// Or only one, but we might only have prefix,
	// in which case the widget is still empty.
	if widget = getWidgetMatch(key, filtered); widget == "" {
		return "", true
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
