package completion

// The isearch keymap is empty by default: the widgets that can
// be used while in incremental search mode will be found in the
// main keymap, so that the same keybinds can be used.
var isearchKeys = map[string]string{}

// those widgets, generally found in the main keymap, are the only
// valid widgets to be used in the incremental search minibuffer.
var validIsearchWidgets = []string{
	"accept-and-infer-next-history",
	"accept-line",
	"accept-line-and-down-history",
	"accept-search",
	"backward-delete-char",
	"vi-backward-delete-char",
	"backward-kill-word",
	"backward-delete-word",
	"vi-backward-kill-word",
	"clear-screen",
	"history-incremental-search-forward",  // Not sure history- needed
	"history-incremental-search-backward", // same
	"space",
	"quoted-insert",
	"vi-quoted-insert",
	"vi-cmd-mode",
	"self-insert",
}

// func (rl *Shell) enterIsearchMode() {
// 	// rl.local = isearch
// 	// rl.hint = []rune(seqBold + seqFgCyan + "isearch: " + seqReset)
// 	// rl.hint = append(rl.hint, rl.tfLine...)
// }
//
// // useIsearchLine replaces the input line with our current
// // isearch buffer, the time for the widget to work on it.
// func (rl *Shell) useIsearchLine() {
// 	// rl.lineBuf = string(rl.line)
// 	// rl.line = append([]rune{}, rl.tfLine...)
// 	//
// 	// cpos := rl.pos
// 	// rl.pos = rl.tfPos
// 	// rl.tfPos = cpos
// }
//
// // exitIsearchLine resets the input line to its original once
// // the widget used in isearch mode has done its work.
// func (rl *Shell) exitIsearchLine() {
// 	// rl.tfLine = append([]rune{}, rl.line...)
// 	// rl.line = []rune(rl.lineBuf)
// 	// rl.lineBuf = ""
// 	//
// 	// cpos := rl.tfPos
// 	// rl.tfPos = rl.pos
// 	// rl.pos = cpos
// }
//
// // updateIsearch recompiles the isearch as a regex and
// // filters matching candidates in the available completions.
// func (rl *Shell) updateIsearch() {
// 	// First compile the search as regular expression
// 	// var regexStr string
// 	// if hasUpper(rl.tfLine) {
// 	// 	regexStr = string(rl.tfLine)
// 	// } else {
// 	// 	regexStr = "(?i)" + string(rl.tfLine)
// 	// }
// 	//
// 	// var err error
// 	// rl.isearch, err = regexp.Compile(regexStr)
// 	// if err != nil {
// 	// 	rl.hint = append(rl.hint, []rune(seqFgRed+"Failed to compile search regexp")...)
// 	// }
// 	//
// 	// if rl.completer != nil {
// 	// 	rl.completer()
// 	// }
// 	//
// 	// // And filter out the completions.
// 	// for _, g := range rl.tcGroups {
// 	// 	g.updateIsearch(rl)
// 	// }
// 	//
// 	// // In history isearch, insert the first matching candidate.
// 	// // This candidate will be cleared/replaced as soon as another
// 	// // key/change is operated on the isearch buffer.
// 	// if len(rl.histHint) > 0 && len(rl.tcGroups) > 0 && len(rl.tcGroups[0].values) > 0 {
// 	// 	rl.resetVirtualComp(true)
// 	// 	cur := rl.currentGroup()
// 	// 	cur.tcPosY = 0
// 	// 	cur.tcPosX = 0
// 	// 	rl.updateVirtualComp()
// 	// 	cur.tcPosY = -1
// 	// 	cur.tcPosX = -1
// 	// }
// }
//
// func (rl *Shell) resetIsearch() {
// 	// if rl.local != isearch {
// 	// 	return
// 	// }
//
// 	// rl.local = ""
// 	// rl.tfLine = []rune{}
// 	// rl.tfPos = 0
// 	// rl.isearch = nil
// }
//
// func (rl *Shell) isIsearchMode(mode keymap.Mode) bool {
// 	// if mode != emacsC && mode != viinsC && mode != vicmdC {
// 	// 	return false
// 	// }
//
// 	// if rl.local != isearch {
// 	// 	return false
// 	// }
//
// 	return true
// }
//
// // func (rl *Shell) filterIsearchWidgets(mode keymapMode) (isearch widgets) {
// // 	km := rl.config.Keymaps[mode]
// //
// // 	isearch = make(widgets)
// // 	b := new(bytes.Buffer)
// // 	decoder := caret.Decoder{Writer: b}
// //
// // 	for key, widget := range km {
// //
// // 		// Widget must be a valid isearch widget
// // 		if !isValidIsearchWidget(widget) {
// // 			continue
// // 		}
// //
// // 		// Or bind to our temporary isearch keymap
// // 		// rl.bindWidget(key, widget, &isearch, decoder, b)
// // 	}
// //
// // 	return
// // }
//
// func isValidIsearchWidget(widget string) bool {
// 	for _, isw := range validIsearchWidgets {
// 		if isw == widget {
// 			return true
// 		}
// 	}
//
// 	return false
// }
//
// func hasUpper(line []rune) bool {
// 	for _, r := range line {
// 		if unicode.IsUpper(r) {
// 			return true
// 		}
// 	}
//
// 	return false
// }
