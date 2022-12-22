package readline

var menuselectKeys = map[string]string{
	"^I":   "menu-complete",
	"^[[Z": "reverse-menu-complete",
	"^@":   "accept-and-menu-complete",
	"^F":   "incremental-search-forward",
	"^[[A": "reverse-menu-complete",
	"^[[B": "menu-complete",
	"^[[C": "menu-complete",
	"^[[D": "reverse-menu-complete",
}
