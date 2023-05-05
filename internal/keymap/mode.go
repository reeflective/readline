package keymap

// Mode is a root keymap mode for the shell.
// To each of these keymap modes is bound a keymap.
type Mode string

// These are the root keymaps used in the readline shell.
// Their functioning is similar to how ZSH organizes keymaps.
const (
	// Editor.
	Emacs         Mode = "emacs"
	EmacsMeta     Mode = "emacs-meta"
	EmacsCtrlX    Mode = "emacs-ctlx"
	EmacsStandard Mode = "emacs-standard"

	ViInsert  Mode = "vi-insert"
	Vi        Mode = "vi"
	ViCommand Mode = "vi-command"
	ViMove    Mode = "vi-move"
	Visual    Mode = "vi-visual"
	ViOpp     Mode = "vi-opp"

	// Completion and search.
	Isearch    Mode = "isearch"
	MenuSelect Mode = "menu-select"
)
