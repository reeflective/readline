package readline

type keyMap map[string]string

// vimDefaultKeymaps binds Vim widgets to their default keys.
// var vimDefaultKeymaps = keymaps{
// 	// Non-standard
// 	'[': "vi-jump-previous-brace",
// 	']': "vi-jump-next-brace",
// }

// viinsKeymaps are the default keymaps in Vim Insert mode
var viinsKeymaps = keyMap{
	// Standard
	string(charEscape): "vi-cmd-mode",
	string(charCtrlM):  "accept-line",
	// Emacs
	// Bunch of self-insert characters: How to use them, like Ctrl-C ?
	string(charCtrlL):          "clear-screen",
	string(charCtrlY):          "yank",
	string(charCtrlE):          "end-of-line",
	string(charCtrlA):          "beginning-of-line",
	string(charCtrlF):          "forward-char",
	string(charCtrlB):          "backward-char",
	string(charCtrlK):          "kill-line",
	string(charCtrlN):          "down-line-or-history",
	string(charCtrlP):          "up-line-or-history",
	string(charCtrlW):          "backward-kill-word",
	string(charBackspace):      "backward-delete-char",
	string(charBackspace2):     "backward-delete-char",
	string(charCtrlUnderscore): "undo",
	seqDelete:                  "delete-char",
	seqHome:                    "beginning-of-line",
	seqEnd:                     "end-of-line",
	seqPageUp:                  "history-search",
	seqPageDown:                "menu-select",
	seqArrowRight:              "vi-forward-char",
	seqArrowLeft:               "vi-backward-char",
	seqCtrlDelete:              "kill-word",
	seqCtrlArrowRight:          "forward-word",
	seqCtrlArrowLeft:           "backward-word",
	seqCtrlArrowRight:          "forward-word",
	seqCtrlArrowLeft:           "backward-word",

	// History
	seqArrowUp:   "up-line-or-search",
	seqArrowDown: "down-line-or-select",

	// Vim
	// string(charCtrlH): "vi-backward-delete-char",
	seqArrowRight: "vi-forward-char",
	seqArrowLeft:  "vi-backward-char",
	// TODO: Important; magic-space
}

// viinsKeymaps are the default keymaps in Vim Command mode
var vicmdKeymaps = keyMap{
	// Standard
	"i":               "vi-insert-mode",
	string(charCtrlM): "accept-line",

	// Emacs
	string(charCtrlL): "clear-screen",
	string(charCtrlN): "down-history",
	string(charCtrlP): "up-history",
	seqDelete:         "delete-char",
	seqPageDown:       "down-line-or-history",
	seqPageUp:         "up-line-or-history",
	seqHome:           "beginning-of-line",
	seqEnd:            "end-of-line",
	seqArrowUp:        "history-search",
	seqArrowDown:      "menu-select",
	seqCtrlDelete:     "kill-word",
	seqCtrlArrowRight: "forward-word",
	seqCtrlArrowLeft:  "backward-word",

	// History

	// Vim
	string(charCtrlA): "switch-keyword", // SPECIAL HANDLER
	// string(charCtrlH):      "vi-backward-char",
	string(charCtrlR):      "redo",
	string(charBackspace):  "backward-delete-char",
	string(charBackspace2): "backward-delete-char",
	seqArrowRight:          "vi-forward-char",
	seqArrowLeft:           "vi-backward-char",
	" ":                    "vi-forward-char",
	"$":                    "vi-end-of-line",
	"%":                    "vi-match-bracket",
	"\"":                   "vi-set-buffer",
	"0":                    "vi-digit-or-beginning-of-line",
	"a":                    "vi-add-next", // SPECIAL HANDLER
	"A":                    "vi-add-eol",
	"b":                    "vi-backward-word",
	"B":                    "vi-backward-blank-word",
	"c":                    "vi-change-surround", // SPECIAL HANDLER  vi-change-surround-text-object
	"C":                    "vi-change-eol",
	"d":                    "vi-delete", // SPECIAL HANDLER
	"D":                    "vi-kill-eol",
	"e":                    "vi-forward-word-end",
	"E":                    "vi-forward-blank-word-end",
	"f":                    "vi-find-next-char",
	"t":                    "vi-find-next-char-skip",
	"I":                    "vi-insert-bol",
	"h":                    "vi-backward-char",
	"l":                    "vi-forward-char",
	"j":                    "down-line-or-history",
	"k":                    "up-line-or-history",
	"p":                    "vi-put-after",
	"P":                    "vi-put-before",
	"r":                    "vi-replace-chars",
	"R":                    "vi-replace",
	"F":                    "vi-find-prev-char",
	"T":                    "vi-find-prev-char-skip",
	"u":                    "undo",
	"v":                    "visual-mode", // Detects if v or V
	"V":                    "visual-mode", // Detects if v or V
	"w":                    "vi-forward-word",
	"W":                    "vi-forward-blank-word",
	"x":                    "vi-delete-char",
	"X":                    "vi-backward-delete-char",
	"y":                    "vi-yank",
	"Y":                    "vi-yank-whole-line",
	"|":                    "vi-goto-column",
	"~":                    "vi-swap-case",
}

// viinsKeymaps are the default keymaps in Vim Operating Pending mode
var vioppKeymaps = keyMap{
	string(charEscape): "vi-cmd-mode",
	"a":                "vi-select-surround", // SPECIAL HANDLER, probably different from 'a' in visual mode, less choices.
	"i":                "vi-select-surround", // SAME THING.
	"j":                "down-line",          // Not sure since no multiline
	"k":                "up-line",            // Not sure since no multiline
}

// viinsKeymaps are the default keymaps in Vim Visual mode
var visualKeymaps = keyMap{
	string(charEscape): "vi-cmd-mode",

	"S": "vi-change-surround", // SPECIAL HANDLER vi-change-surround (no text object)
	"a": "vi-select-surround", // SPECIAL HANDLER
	"c": "vi-change",          // SPECIAL HANDLER ?
	"d": "vi-delete",
	"i": "vi-select-surround", // SPECIAL HANDLER
	"j": "down-line",          // Not sure since no multiline
	"k": "up-line",            // Not sure since no multiline
	"u": "vi-down-case",
	"v": "vi-edit-command-line",
	"x": "vi-delete",
	"y": "vi-yank",      // SPECIAL HANDLER
	"~": "vi-swap-case", // Need to be a separate widget from ~ in cmd mode ?
}

var vicmdSpecialKeymaps = keyMap{
	`^([1-9]{1})$`: "digit-argument",
}
