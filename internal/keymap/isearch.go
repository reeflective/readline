package keymap

import "github.com/reeflective/readline/inputrc"

// The isearch keymap is empty by default: the widgets that can
// be used while in incremental search mode will be found in the
// main keymap, so that the same keybinds can be used.
var isearchKeys = map[string]string{}

// those widgets, generally found in the main keymap, are the only
// valid widgets to be used in the incremental search minibuffer.
var IsearchCommands = []string{
	"accept-and-infer-next-history",
	"accept-line",
	"operate-and-get-next",
	"accept-search",
	"delete-char",
	"backward-delete-char",
	"backward-kill-word",
	"vi-unix-word-rubout",
	"clear-screen",
	"history-incremental-search-forward",  // Not sure history- needed
	"history-incremental-search-backward", // same
	"magic-space",
	"quoted-insert",
	"vi-movement-mode",
	"self-insert",
}

// IsearchCommands returns all commands that are available in incremental-search mode.
// These commands are a restricted set of edit/movement/history functions.
func (m *Modes) IsearchCommands(mode Mode) map[string]inputrc.Bind {
	isearch := make(map[string]inputrc.Bind)

	for seq, command := range m.opts.Binds[string(mode)] {
		// Widget must be a valid isearch widget
		if !isValidIsearchWidget(command.Action) {
			continue
		}

		// Or bind to our temporary isearch keymap
		isearch[seq] = command
	}

	return isearch
}

func isValidIsearchWidget(widget string) bool {
	for _, isw := range IsearchCommands {
		if isw == widget {
			return true
		}
	}

	return false
}
