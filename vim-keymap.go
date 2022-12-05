package readline

// viinsKeymaps are the default keymaps in Vim Insert mode
var viinsKeymaps = keyMap{
	// Standard
	string(charEscape): "vi-cmd-mode",
	string(charCtrlM):  "accept-line",
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
	seqPageUp:                  "history-search", // TODO
	seqPageDown:                "menu-select",    // TODO
	seqArrowRight:              "vi-forward-char",
	seqArrowLeft:               "vi-backward-char",
	seqCtrlDelete:              "kill-word", // TODO
	seqCtrlArrowRight:          "forward-word",
	seqCtrlArrowLeft:           "backward-word",
	seqCtrlArrowRight:          "forward-word",
	seqCtrlArrowLeft:           "backward-word",

	// History
	seqArrowUp:   "up-line-or-search",   // TODO
	seqArrowDown: "down-line-or-select", // TODO

	// Vim
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
	string(charCtrlA):      "switch-keyword", // SPECIAL HANDLER TODO
	string(charCtrlR):      "redo",           // TODO
	string(charBackspace):  "backward-delete-char",
	string(charBackspace2): "backward-delete-char",
	seqArrowRight:          "vi-forward-char",
	seqArrowLeft:           "vi-backward-char",
	" ":                    "vi-forward-char",
	"$":                    "vi-end-of-line",
	"%":                    "vi-match-bracket",
	"\"":                   "vi-set-buffer",
	"0":                    "vi-digit-or-beginning-of-line",
	"a":                    "vi-add-next",
	"A":                    "vi-add-eol",
	"b":                    "vi-backward-word",
	"B":                    "vi-backward-blank-word",
	"c":                    "vi-change-surround", // SPECIAL HANDLER  vi-change-surround-text-object
	"C":                    "vi-change-eol",
	"d":                    "vi-delete",
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
	"v":                    "visual-mode",
	"V":                    "visual-line-mode",
	"w":                    "vi-forward-word",
	"W":                    "vi-forward-blank-word",
	"x":                    "vi-delete-char",
	"X":                    "vi-backward-delete-char",
	"y":                    "vi-yank",
	"Y":                    "vi-yank-whole-line",
	"|":                    "vi-goto-column", // TODO
	"~":                    "vi-swap-case",   // TODO
}

// viinsKeymaps are the default keymaps in Vim Operating Pending mode
var vioppKeymaps = keyMap{
	string(charEscape): "vi-cmd-mode",
	"aW":               "select-a-blank-word",
	"aa":               "select-a-shell-word",
	"aw":               "select-a-word",
	"iW":               "select-in-blank-word",
	"ia":               "select-in-shell-word",
	"iw":               "select-in-word",
	"j":                "down-line", // Not sure since-test no multiline
	"k":                "up-line",   // Not sure since no multiline
}

// viinsKeymaps are the default keymaps in Vim Visual mode
var visualKeymaps = keyMap{
	string(charEscape): "vi-cmd-mode",
	"aW":               "select-a-blank-word",
	"aa":               "select-a-shell-word",
	"aw":               "select-a-word",
	"iW":               "select-in-blank-word",
	"ia":               "select-in-shell-word",
	"iw":               "select-in-word",
	"S":                "vi-change-surround", // SPECIAL HANDLER vi-change-surround (no text object)
	"a":                "vi-select-surround", // SPECIAL HANDLER
	"c":                "vi-change",          // SPECIAL HANDLER ?
	"d":                "vi-delete",
	"i":                "vi-select-surround", // SPECIAL HANDLER
	"j":                "down-line",          // Not sure since no multiline
	"k":                "up-line",            // Not sure since no multiline
	"u":                "vi-down-case",
	"v":                "vi-edit-command-line",
	"x":                "vi-delete",
	"y":                "vi-yank",
	"~":                "vi-swap-case", // Need to be a separate widget from ~ in cmd mode ?
}

var vicmdSpecialKeymaps = keyMap{
	`^([1-9]{1})$`: "digit-argument",
}
