package readline

// The isearch keymap is empty by default: the widgets that can
// be used while in incremental search mode will be found in the
// main keymap, so that the same keybinds can be used.
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
			rl.startMenuComplete(rl.historyCompletion)
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
