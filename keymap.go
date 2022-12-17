package readline

// keymapMode is a root keymap mode for the shell.
// To each of these keymap modes is bound a keymap.
type keymapMode string

// keymap maps a key (either in caret or hex notation)
// to the name of the widget to run when key is pressed.
type keymap map[string]string

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

// initKeymap ensures that all keymaps are set
// at the beginning of a readline run loop.
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

	// When matching a widget, we need to know if the shell was in operator
	// pending mode before trying to match the key against our keymaps.
	rl.viopp = rl.local == viopp
}

// matchKeymap checks if the provided key matches a precise widget, or if only a prefix
// is matched. When only a prefix is matched, the shell keeps reading for another key.
func (rl *Instance) matchKeymap(key string, mode keymapMode) (cb EventCallback, prefix bool) {
	if mode == "" {
		return nil, false
	}

	// The escape key is a special key that bypasses the entire process.
	// This never returns true (and a callback) when shell is in Emacs mode.
	if escape, yes := rl.isVimEscape(key); yes {
		cb = escape
		rl.keys = ""
		return
	}

	// Get all widgets matched by the key, either exactly or by prefix.
	matchWidgets := rl.widgets[mode]
	cb, prefixed := rl.matchWidgets(key, matchWidgets)

	// When we have absolutely no matching widget for the keys,
	// we either return, or if we have a perfectly matching one
	// waiting for an input, we execute it.
	if cb == nil && len(prefixed) == 0 {
		cb = rl.widgetPrefixMatched
		rl.widgetPrefixMatched = nil

		return
	}

	// Or several matches, in which case we must read another key.
	// If any widget perfectly matches the key, save it, so that
	// the next key, if not matching any of those prefix-matched
	// widgets, is passed as argument to this one.
	if cb != nil && len(prefixed) > 0 {
		rl.widgetPrefixMatched = cb
		return nil, true
	}

	// If we have a non-empty list of prefix-matched widgets,
	// we must keep reading a key as well.
	if len(prefixed) > 0 {
		return nil, true
	}

	// We either have a single widget callback, or nothing.
	return
}
