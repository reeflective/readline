package keymap

import "github.com/reeflective/readline/inputrc"

// menuselectKeys are the default keymaps in menuselect mode.
var menuselectKeys = map[string]inputrc.Bind{
	unescape(`\C-i`):    {Action: "menu-complete"},
	unescape(`\e[Z`):    {Action: "menu-complete-backward"},
	unescape(`\C-@`):    {Action: "accept-and-menu-complete"},
	unescape(`\C-F`):    {Action: "menu-incremental-search"},
	unescape(`\e[A`):    {Action: "menu-complete-backward"},
	unescape(`\e[B`):    {Action: "menu-complete"},
	unescape(`\e[C`):    {Action: "menu-complete"},
	unescape(`\e[D`):    {Action: "menu-complete-backward"},
	unescape(`\e[1;5A`): {Action: "menu-complete-prev-tag"},
	unescape(`\e[1;5B`): {Action: "menu-complete-next-tag"},
}

// The isearch keymap is empty by default: the widgets that can
// be used while in incremental search mode will be found in the
// main keymap, so that the same keybinds can be used.
var isearchKeys = map[string]string{}

// those widgets, generally found in the main keymap, are the only
// valid widgets to be used in the incremental search minibuffer.
var IsearchCommands = []string{
	"abort",
	"accept-and-infer-next-history",
	"accept-line",
	"accept-and-hold",
	"operate-and-get-next",
	"backward-delete-char",
	"backward-kill-word",
	"backward-kill-line",
	"unix-line-discard",
	"unix-word-rubout",
	"vi-unix-word-rubout",
	"clear-screen",
	"clear-display",
	"history-incremental-search-forward",
	"history-incremental-search-backward",
	"magic-space", //
	"vi-movement-mode",
	"yank",
	"self-insert",
}

// IsearchCommands returns all commands that are available in incremental-search mode.
// These commands are a restricted set of edit/movement/history functions.
func (m *Engine) IsearchCommands(mode Mode) map[string]inputrc.Bind {
	isearch := make(map[string]inputrc.Bind)

	for seq, command := range m.config.Binds[string(mode)] {
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
