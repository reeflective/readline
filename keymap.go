package readline

import (
	"regexp"
	"strings"
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

// runWidget wraps a few calls for finding a widget and executing it, returning some basic
// instructions pertaining to what to do next: either keep reading input, or return the line.
func (rl *Instance) runWidget(name string, b []byte, i int, r []rune) (ret bool, val string, err error) {
	widget := rl.getWidget(name)
	if widget == nil {
		return
	}

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

func findBindkeyWidget(key rune, keymap keyMap) keyMap {
	widgets := make(keyMap)

	for wkey, widget := range keymap {
		if strings.HasPrefix(wkey, string(key)) {
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

func getWidgetMatch(key rune, keymap keyMap) (widget string) {
	for wkey, widget := range keymap {
		if wkey == string(key) {
			return widget
		}
	}
	return
}
