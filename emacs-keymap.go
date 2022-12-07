package readline

// emacsKeys are the default keymaps in Emacs mode
var emacsKeys = keymap{
	string(charCtrlM): "accept-line",
	string(charCtrlA): "beginning-of-line",
	string(charCtrlB): "backward-char",
	// TODO: Here EOF (CtrlD) deletes a char.
	string(charCtrlE): "end-of-line",
	string(charCtrlF): "forward-char",
	string(charCtrlG): "send-break", // DON'T KNOW WHAT THAT IS
	string(charCtrlK): "kill-line",  // Similar to forward-kill-line
	string(charCtrlL): "clear-screen",
	string(charCtrlN): "down-line-or-history",
	string(charCtrlP): "up-line-or-history",
	string(charCtrlO): "accept-line-and-down-history",
	// TODO: CtrlR CtrlS incremental-search-history.
	string(charCtrlT): "transpose-chars",
	string(charCtrlU): "kill-whole-line",
	string(charCtrlV): "quoted-insert", // WHAT IS THIS ?
	string(charCtrlY): "yank",

	// Special combinations, how to match them ?
	"^X^B": "vi-match-bracket",
	"^X^E": "vi-edit-command-line",
	"^X^K": "kill-buffer", // IS THAT KILL REGISTER ?
	"^X^N": "infer-next-history",
	"^X^O": "overwrite-mode", // Probably can reuse Vim replace mode here.
	"^X^U": "undo",
	"^X^V": "vi-cmd-mode",
	"^Xu":  "undo",

	// TODO: ^Xr ^Xs incremental-search-history. same as CtrlR above.

	"^[^D":   "list-choices",
	"^[^G":   "send-break",
	"^[^H":   "backward-kill-word",
	"^[^I":   "self-insert-unmeta",
	"^[^J":   "self-insert-unmeta",
	"^[^L":   "clear-screen",
	"^[^M":   "self-insert-unmeta",
	"^[^[OA": "history-search",
	"^[^[OB": "menu-select",
	"^[^[[A": "history-search",
	"^[^[[B": "menu-select",
	"^[^_":   "copy-prev-word",

	// "^[ ":  "expand-history",
	// "^[!":  "expand-history",
	// "^[\"": "quote-region",
	// "^[$":  "spell-word",
	// "^['":  "quote-line",
	// "^[-":  "neg-argument",
	// "^[.":  "insert-last-word",
	// "^[<":  "beginning-of-buffer-or-history",
	// "^[>":  "end-of-buffer-or-history",

	"^[A": "accept-and-hold",
	"^[B": "backward-word",
	"^[C": "capitalize-word",
	"^[D": "kill-word",
	"^[F": "forward-word",
	"^[G": "get-line",
	"^[H": "run-help",
	"^[L": "down-case-word",
	"^[N": "history-search-forward",
	"^[P": "history-search-backward",
	"^[Q": "push-line",
	"^[S": "spell-word",
	"^[T": "transpose-words",
	"^[U": "up-case-word",
	"^[W": "copy-region-as-kill",

	"^[[1;3A": "history-search",
	"^[[1;3B": "menu-select",
	"^[[1;5C": "forward-word",
	"^[[1;5D": "backward-word",
	"^[[200~": "bracketed-paste",
	"^[[3;5~": "kill-word",
	"^[[3~":   "delete-char",
	"^[[5~":   "history-search",
	"^[[6~":   "menu-select",
	"^[[A":    "up-line-or-search",
	"^[[B":    "down-line-or-select",
	"^[[C":    "forward-char",
	"^[[D":    "backward-char",
	"^[[Z":    "menu-select",
	"^[_":     "insert-last-word",
	"^[a":     "accept-and-hold",
	"^[b":     "backward-word",
	"^[c":     "capitalize-word",
	"^[d":     "kill-word",
	"^[f":     "forward-word",
	"^[g":     "get-line",
	"^[h":     "run-help",
	"^[l":     "ls^J",
	"^[m":     "copy-prev-shell-word",
	"^[n":     "history-search-forward",
	"^[p":     "history-search-backward",
	"^[q":     "push-line",
	"^[s":     "spell-word",
	"^[t":     "transpose-words",
	"^[u":     "up-case-word",
	"^[w":     "kill-region",
	"^[x":     "execute-named-cmd",
	"^[y":     "yank-pop",
	"^[z":     "execute-last-named-cmd",
	"^[|":     "vi-goto-column",
	"^[^?":    "backward-kill-word",
	"^_":      "undo",
	// " ": "magic-space",
	// "!"-"~": "self-insert",
	"^?": "backward-delete-char",
	// "\M-^@"-"\M-^?": "self-insert",
}

var emacsSpecialKeymaps = keymap{
	`^\^\[([0-9]{1})$`: "digit-argument",
}
