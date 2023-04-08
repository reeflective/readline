package keymap

import "github.com/reeflective/readline/inputrc"

// menuselectKeys are the default keymaps in menuselect mode.
var menuselectKeys = map[string]inputrc.Bind{
	"^I":      {Action: "menu-complete"},
	"^[[Z":    {Action: "reverse-menu-complete"},
	"^@":      {Action: "accept-and-menu-complete"},
	"^F":      {Action: "menu-incremental-search"},
	"^[[A":    {Action: "reverse-menu-complete"},
	"^[[B":    {Action: "menu-complete"},
	"^[[C":    {Action: "menu-complete"},
	"^[[D":    {Action: "reverse-menu-complete"},
	"^[[1;5A": {Action: "menu-complete-prev-tag"},
	"^[[1;5B": {Action: "menu-complete-next-tag"},
}
