package readline

var menuselectKeys = map[string]string{
	"^I":      "menu-complete",
	"^[[Z":    "reverse-menu-complete",
	"^@":      "accept-and-menu-complete",
	"^F":      "menu-incremental-search",
	"^[[A":    "reverse-menu-complete",
	"^[[B":    "menu-complete",
	"^[[C":    "menu-complete",
	"^[[D":    "reverse-menu-complete",
	"^[[1;5A": "menu-complete-prev-tag",
	"^[[1;5B": "menu-complete-next-tag",
	"^M":      "accept-completion-or-line",
}
