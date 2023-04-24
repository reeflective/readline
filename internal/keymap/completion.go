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
