package readline

import (
	"regexp"
	"strings"
)

// lineWidgets maps widget names to their corresponding line widgets.
type lineWidgets map[string]lineWidget

// standardLineWidgets either need access to the input key,
// or need to return specific instructions and values.
func (rl *Instance) commonLineWidgets() lineWidgets {
	widgets := map[string]lineWidget{
		"accept-line":    rl.acceptLine,
		"self-insert":    rl.selfInsert,
		"digit-argument": rl.digitArgument,
	}

	return widgets
}

// standardWidgets don't need access to the input key.
func (rl *Instance) commonWidgets() baseWidgets {
	widgets := map[string]func(){
		"clear-screen":                   rl.clearScreen,
		"beginning-of-line":              rl.beginningOfLine,
		"end-of-line":                    rl.endOfLine,
		"kill-line":                      rl.killLine,
		"kill-whole-line":                rl.killWholeLine,
		"backward-kill-word":             rl.backwardKillWord,
		"kill-word":                      rl.killWord,
		"yank":                           rl.yank,
		"backward-delete-char":           rl.backwardDeleteChar,
		"delete-char":                    rl.deleteChar,
		"forward-char":                   rl.forwardChar,
		"backward-char":                  rl.backwardChar,
		"forward-word":                   rl.forwardWord,
		"backward-word":                  rl.backwardWord,
		"undo":                           rl.undo,
		"down-line-or-history":           rl.historyNext,
		"up-line-or-history":             rl.historyPrev,
		"down-history":                   rl.downHistory,
		"up-history":                     rl.upHistory,
		"infer-next-history":             rl.inferNextHistory,
		"overwrite-mode":                 rl.overwriteMode,
		"set-mark-command":               rl.setMarkCommand,
		"exhange-point-and-mark":         rl.exchangePointAndMark,
		"quote-region":                   rl.quoteRegion,
		"quote-line":                     rl.quoteLine,
		"neg-argument":                   rl.negArgument,
		"beginning-of-buffer-or-history": rl.beginningOfBufferOrHistory,
		"end-of-buffer-or-history":       rl.endOfBufferOrHistory,
		"capitalize-word":                rl.capitalizeWord,
		"down-case-word":                 rl.downCaseWord,
		"up-case-word":                   rl.upCaseWord,
		"transpose-words":                rl.transposeWords,
		"copy-region-as-kill":            rl.copyRegionAsKill,
		"copy-prev-word":                 rl.copyPrevWord,
		"kill-region":                    rl.killRegion,
		"redo":                           rl.redo,
		"switch-keyword":                 rl.switchKeyword,
	}

	return widgets
}

func (rl *Instance) selfInsert(r []rune) (read, ret bool, val string, err error) {
	rl.viUndoSkipAppend = true

	// Prepare the line
	// line, err := rl.mainHistory.GetLine(rl.mainHistory.Len() - 1)
	// if err != nil {
	// 	return
	// }
	// if !rl.mainHist {
	// 	line, err = rl.altHistory.GetLine(rl.altHistory.Len() - 1)
	// 	if err != nil {
	// 		return
	// 	}
	// }
	//
	// tokens, _, _ := tokeniseSplitSpaces([]rune(line), 0)
	// pos := int(r[1]) - 48 // convert ASCII to integer
	// if pos > len(tokens) {
	// 	return
	// }
	//
	// // The line is prepared and the actual runes to insert are as well.
	// r = []rune(tokens[pos-1])

	for {
		// I don't really understand why `0` is creaping in at the end of the
		// array but it only happens with unicode characters.
		if len(r) > 1 && r[len(r)-1] == 0 {
			r = r[:len(r)-1]
			continue
		}
		break
	}

	// We can ONLY have three fondamentally different cases:
	switch {
	// The line is empty
	case len(rl.line) == 0:
		rl.line = r

	// We are inserting somewhere in the middle
	case rl.pos < len(rl.line):
		forwardLine := append(r, rl.line[rl.pos:]...)
		rl.line = append(rl.line[:rl.pos], forwardLine...)

	// We are at the end of the input line
	case rl.pos == len(rl.line):
		rl.line = append(rl.line, r...)
	}

	rl.pos += len(r)

	// This should also update the rl.pos
	rl.updateHelpers()

	return
}

// acceptLine returns the line to the readline caller for being executed/evaluated.
func (rl *Instance) acceptLine(_ []rune) (read, ret bool, val string, err error) {
	if rl.modeTabCompletion {
		cur := rl.getCurrentGroup()

		// Check that there is a group indeed, as we might have no completions.
		if cur == nil {
			rl.clearHelpers()
			rl.resetTabCompletion()
			rl.renderHelpers()
			read = true

			return
		}

		// IF we have a prefix and completions printed, but no candidate
		// (in which case the completion is ""), we immediately return.
		completion := cur.getCurrentCell(rl)
		prefix := len(rl.tcPrefix)
		if prefix > len(completion.Value) {
			rl.carriageReturn()

			val = string(rl.line)
			ret = true

			return
		}

		// Else, we insert the completion candidate in the real input line.
		// By default we add a space, unless completion group asks otherwise.
		rl.compAddSpace = true
		rl.resetVirtualComp(false)

		// If we were in history completion, immediately execute the line.
		if rl.modeAutoFind && rl.searchMode == HistoryFind {
			rl.carriageReturn()

			val = string(rl.line)
			ret = true

			return
		}

		// Reset completions and update input line
		rl.clearHelpers()
		rl.resetTabCompletion()
		rl.renderHelpers()

		read = true
		return
	}

	rl.carriageReturn()

	val = string(rl.line)
	ret = true

	return
}

func (rl *Instance) clearScreen() {
	rl.viUndoSkipAppend = true

	print(seqClearScreen)
	print(seqCursorTopLeft)

	// Print the prompt, all or part of it.
	print(rl.Prompt.getPrimary())
	print(seqClearScreenBelow)

	rl.resetHintText()
	rl.getHintText()
	rl.renderHelpers()
}

func (rl *Instance) beginningOfLine() {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}

	rl.viUndoSkipAppend = true
	rl.pos = 0
}

func (rl *Instance) endOfLine() {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}

	if len(rl.line) > 0 {
		rl.pos = len(rl.line)
	}

	rl.viUndoSkipAppend = true
}

func (rl *Instance) killLine() {
	rl.saveBufToRegister(rl.line[rl.pos-1:])
	rl.line = rl.line[:rl.pos]
	rl.resetHelpers()
	rl.updateHelpers()
	rl.addIteration("")
}

func (rl *Instance) killWholeLine() {
	if len(rl.line) == 0 {
		return
	}

	rl.saveBufToRegister(rl.line)
	rl.clearLine()
}

func (rl *Instance) backwardKillWord() {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}

	rl.saveToRegister(rl.viJumpB(tokeniseLine))
	rl.viDeleteByAdjust(rl.viJumpB(tokeniseLine))
	rl.updateHelpers()
}

func (rl *Instance) killWord() {
	rl.saveToRegisterTokenize(tokeniseLine, rl.viJumpE, 1)
	rl.viDeleteByAdjust(rl.viJumpE(tokeniseLine) + 1)
	// WARN: HERE THE +1 SHOULD BE CHECKED, because panic when at the end of line.
}

func (rl *Instance) yank() {
	if rl.modeTabCompletion {
		rl.resetVirtualComp(false)
	}
	// paste after the cursor position
	rl.viUndoSkipAppend = true
	buffer := rl.pasteFromRegister()
	rl.insert(buffer)
	rl.updateHelpers()
}

func (rl *Instance) backwardDeleteChar() {
	vii := rl.getViIterations()

	// We might be on an active register, but not yanking...
	rl.saveToRegister(vii)

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {
		rl.deleteX()
	}

	if rl.main == viins || rl.main == emacs {
		rl.viUndoSkipAppend = true
	}
}

func (rl *Instance) deleteChar() {
	vii := rl.getViIterations()

	// We might be on an active register, but not yanking...
	rl.saveToRegister(vii)

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {
		rl.deletex()
	}
	if rl.pos == len(rl.line) && len(rl.line) > 0 {
		rl.pos--
	}
}

func (rl *Instance) forwardChar() {
	if rl.pos < len(rl.line) {
		rl.pos++
	}
	rl.viUndoSkipAppend = true
}

func (rl *Instance) backwardChar() {
	if rl.pos > 0 {
		rl.pos--
	}
	rl.viUndoSkipAppend = true
}

func (rl *Instance) forwardWord() {
	// If we were not yanking
	rl.viUndoSkipAppend = true

	// If the input line is empty, we don't do anything
	if rl.pos == 0 && len(rl.line) == 0 {
		return
	}

	// Get iterations and move
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpW(tokeniseLine))
	}
}

func (rl *Instance) backwardWord() {
	rl.viUndoSkipAppend = true

	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	}
}

func (rl *Instance) downHistory() {
	rl.mainHist = true
	rl.walkHistory(-1)
}

func (rl *Instance) upHistory() {
	rl.mainHist = true
	rl.walkHistory(1)
}

// digitArgument is used both in Emacs and Vim modes,
// but strips the Alt modifier used in Emacs mode.
func (rl *Instance) digitArgument(r []rune) (read, ret bool, val string, err error) {
	if len(r) > 1 {
		// The first rune is the alt modifier.
		rl.addIteration(string(r[1:]))
	} else {
		rl.addIteration(string(r))
	}

	rl.viUndoSkipAppend = true

	return
}

func (rl *Instance) historyNext() {
	rl.viUndoSkipAppend = true
	rl.mainHist = true
	rl.walkHistory(-1)
}

func (rl *Instance) historyPrev() {
	rl.viUndoSkipAppend = true
	rl.mainHist = true
	rl.walkHistory(1)
}

func (rl *Instance) killBuffer() {
	if len(rl.line) == 0 {
		return
	}
	rl.saveBufToRegister(rl.line)
	rl.clearLine()
}

func (rl *Instance) inferNextHistory() {
	rl.viUndoSkipAppend = true
	matchIndex := 0
	histSuggested := make([]rune, 0)
	rl.mainHist = true

	// Work with correct history source (depends on CtrlR/CtrlE)
	var history History
	if !rl.mainHist {
		history = rl.altHistory
	} else {
		history = rl.mainHistory
	}

	// Nothing happens if the history is nil or empty.
	if history == nil || history.Len() == 0 {
		return
	}

	for i := 1; i <= history.Len(); i++ {
		histline, err := history.GetLine(history.Len() - i)
		if err != nil {
			return
		}

		// If too short
		if len(histline) <= len(rl.line) {
			continue
		}

		// Or if not fully matching
		match := false
		for i, char := range rl.line {
			if byte(char) == histline[i] {
				match = true
			} else {
				match = false
				break
			}
		}

		// If the line fully matches, we have our suggestion
		if match {
			matchIndex = history.Len() - i
			histSuggested = append(histSuggested, []rune(histline)...)
			break
		}
	}

	// If we have no match we return, or check for the next line.
	if (len(histSuggested) == 0 && matchIndex <= 0) || history.Len() <= matchIndex+1 {
		return
	}

	// Get the next history line
	nextLine, err := history.GetLine(matchIndex + 1)
	if err != nil {
		return
	}

	rl.line = []rune(nextLine)
	rl.pos = len(nextLine)
	// TODO: How to adjust conditionally on keymap ? Many widgets need this.
	// if rl.pos > 0 {
	// 	rl.pos--
	// }
}

// TODO: Find a way to catch on other keymaps ? How and when to exit the mode if not with escape ?
func (rl *Instance) overwriteMode() {
	// We store the current line as an undo item first, but will not
	// store any intermediate changes (in the loop below) as undo items.
	rl.undoAppendHistory()
	rl.viUndoSkipAppend = true

	// The replace mode is quite special in that it does escape back
	// to the main readline loop: it keeps reading characters and inserts
	// them as long as the escape key is not pressed.
	for {
		// Read a new key
		keys, esc := rl.readArgumentKey()
		if esc {
			break
		}
		key := rune(keys[0])

		// If the key is a backspace, we go back one character
		if key == charBackspace || key == charBackspace2 {
			rl.backwardDeleteChar()
		} else {
			// If the cursor is at the end of the line,
			// we insert the character instead of replacing.
			if len(rl.line)-1 < rl.pos {
				rl.line = append(rl.line, key)
			} else {
				rl.line[rl.pos] = key
			}

			rl.pos++
		}

		// Update the line
		rl.updateHelpers()
	}
}

func (rl *Instance) setMarkCommand() {
	rl.viUndoSkipAppend = true
	vii := rl.getViIterations()
	switch {
	case vii < 0:
		rl.mark = -1
		rl.activeRegion = false
		rl.visualLine = false
	default:
		rl.mark = rl.pos
		rl.activeRegion = true
	}
}

func (rl *Instance) quoteRegion() {
	bpos, epos, cpos := rl.getSelection()
	selection := string(rl.line[bpos:epos])
	begin := string(rl.line[:bpos])
	end := string(rl.line[epos:])
	quoted := "'" + selection + "'"

	newLine := append([]rune(begin), []rune(quoted)...)
	newLine = append(newLine, []rune(end)...)
	rl.line = newLine
	rl.pos = cpos + 1

	if rl.activeRegion {
		rl.activeRegion = false
		rl.mark = -1
	}
}

func (rl *Instance) quoteLine() {
	newLine := make([]rune, 0)
	newLine = append(newLine, '\'')

	for _, r := range rl.line {
		if r == '\n' {
			break
		}
		if r == '\'' {
			newLine = append(newLine, []rune("\\'")...)
		} else {
			newLine = append(newLine, r)
		}
	}

	newLine = append(newLine, '\'')

	rl.line = newLine
}

func (rl *Instance) negArgument() {
	rl.negativeArg = true
}

func (rl *Instance) beginningOfBufferOrHistory() {
	rl.viUndoSkipAppend = true

	if rl.pos == 0 {
		var history History
		rl.mainHist = true
		if !rl.mainHist {
			history = rl.altHistory
		} else {
			history = rl.mainHistory
		}

		if history == nil {
			return
		}

		new, err := history.GetLine(0)
		if err != nil {
			rl.resetHelpers()
			print(rl.Prompt.primary)
			return
		}

		rl.clearLine()
		rl.line = []rune(new)
		rl.pos = len(rl.line)
		if rl.pos > 0 {
			rl.pos--
		}

		return
	}

	rl.beginningOfLine()
}

func (rl *Instance) endOfBufferOrHistory() {
	rl.viUndoSkipAppend = true
	if rl.pos == len(rl.line) {
		var history History
		rl.mainHist = true
		if !rl.mainHist {
			history = rl.altHistory
		} else {
			history = rl.mainHistory
		}

		if history == nil {
			return
		}

		new, err := history.GetLine(history.Len() - 1)
		if err != nil {
			rl.resetHelpers()
			print(rl.Prompt.primary)
			return
		}

		rl.clearLine()
		rl.line = []rune(new)
		rl.pos = len(rl.line)
		if rl.pos > 0 {
			rl.pos--
		}
		return
	}

	rl.endOfLine()
}

func (rl *Instance) capitalizeWord() {
	posInit := rl.pos
	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	letter := rl.line[rl.pos]
	upper := strings.ToUpper(string(letter))
	rl.line[rl.pos] = rune(upper[0])
	rl.pos = posInit
}

func (rl *Instance) downCaseWord() {
	posInit := rl.pos
	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))

	rl.mark = rl.pos
	rl.activeRegion = true
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
	bpos, epos, _ := rl.getSelection()
	word := string(rl.line[bpos:epos])
	word = strings.ToLower(word)

	begin := string(rl.line[:bpos])
	end := string(rl.line[epos:])

	newLine := append([]rune(begin), []rune(word)...)
	newLine = append(newLine, []rune(end)...)

	rl.line = newLine
	rl.pos = posInit
	rl.mark = -1
	rl.activeRegion = false
}

func (rl *Instance) upCaseWord() {
	posInit := rl.pos
	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))

	rl.mark = rl.pos
	rl.activeRegion = true
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
	bpos, epos, _ := rl.getSelection()
	word := string(rl.line[bpos:epos])
	word = strings.ToUpper(word)

	begin := string(rl.line[:bpos])
	end := string(rl.line[epos:])

	newLine := append([]rune(begin), []rune(word)...)
	newLine = append(newLine, []rune(end)...)

	rl.line = newLine
	rl.pos = posInit
	rl.mark = -1
	rl.activeRegion = false
}

func (rl *Instance) transposeWords() {
	posInit := rl.pos

	// Save the current word
	rl.pos++
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))

	rl.mark = rl.pos
	rl.activeRegion = true
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
	tbpos, tepos, _ := rl.getSelection()
	toTranspose := string(rl.line[tbpos:tepos])
	rl.mark = -1
	rl.activeRegion = false

	// First move the number of words
	vii := rl.getViIterations()
	for i := 0; i <= vii; i++ {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
	}

	// Save the word to transpose with
	rl.mark = rl.pos
	rl.activeRegion = true
	rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
	wbpos, wepos, _ := rl.getSelection()
	transposeWith := string(rl.line[wbpos:wepos])
	rl.mark = -1
	rl.activeRegion = false

	// Assemble the newline
	begin := string(rl.line[:wbpos])
	newLine := append([]rune(begin), []rune(toTranspose)...)
	newLine = append(newLine, rl.line[wepos:tbpos]...)
	newLine = append(newLine, []rune(transposeWith)...)
	newLine = append(newLine, rl.line[tepos:]...)
	rl.line = newLine

	// And replace cursor
	if vii < 0 {
		rl.pos = posInit
	} else {
		rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))
		for i := 0; i <= vii; i++ {
			rl.moveCursorByAdjust(rl.viJumpE(tokeniseLine))
		}
	}
}

func (rl *Instance) copyRegionAsKill() {
	rl.viUndoSkipAppend = true
	rl.yankSelection()
	rl.resetSelection()
}

func (rl *Instance) copyPrevWord() {
	posInit := rl.pos
	rl.mark = rl.pos
	rl.activeRegion = true
	rl.moveCursorByAdjust(rl.viJumpB(tokeniseLine))

	bpos, epos, _ := rl.getSelection()
	word := string(rl.line[bpos:epos])
	rl.mark = -1
	rl.activeRegion = false

	begin := string(rl.line[:epos])
	end := string(rl.line[epos:])
	newLine := append([]rune(begin), []rune(word)...)
	newLine = append(newLine, []rune(end)...)
	rl.line = newLine

	rl.pos = posInit + len(word)
}

// 	"^[m":     "copy-prev-shell-word",
func (rl *Instance) copyPrevShellWord() {
	// First find the previous quote, and automatically
	// detect where is the corresponding one.
	rl.findAndMoveCursor("'", 1, false, false)
	rl.mark = rl.pos
	rl.activeRegion = true
}

func (rl *Instance) killRegion() {
	rl.deleteSelection()
	rl.resetSelection()
}

// Cursor position cases:
//
// 1. Cursor on symbol:
// 2+2   => +
// 2-2   => -
// 2 + 2 => +
// 2 +2  => +2
// 2 -2  => -2
// 2 -a  => -a
//
// 2. Cursor on number or alpha:
// 2+2   => +2
// 2-2   => -2
// 2 + 2 => 2
// 2 +2  => +2
// 2 -2  => -2
// 2 -a  => -a
func (rl *Instance) switchKeyword() {
	cpos := rl.pos
	increase := rl.keys == string(charCtrlA)

	if match, _ := regexp.MatchString(`[+-][0-9]`, rl.lineSlice(2)); match {
		// If cursor is on the `+` or `-`, we need to check if it is a
		// number with a sign or an operator, only the number needs to
		// forward the cursor.
		digit, _ := regexp.Compile(`[^0-9]`)
		if cpos == 0 || digit.MatchString(string(rl.line[cpos-1])) {
			cpos++
		}
	} else if match, _ := regexp.MatchString(`[+-][a-zA-Z]`, rl.lineSlice(2)); match {
		// If cursor is on the `+` or `-`, we need to check if it is a
		// short option, only the short option needs to forward the cursor.
		if cpos == 0 || rl.line[rl.pos-1] == ' ' {
			cpos++
		}
	}

	// Select in word and get the selection positions
	bpos, epos := rl.selectInWord(cpos)
	epos++

	// Move the cursor backward if needed/possible
	if bpos != 0 && (rl.line[bpos-1] == '+' || rl.line[bpos-1] == '-') {
		bpos--
	}

	// Get the selection string
	selection := string(rl.line[bpos:epos])

	// For each of the keyword handlers, run it, which returns
	// false/none if didn't operate, then continue to next handler.
	for _, switcher := range rl.keywordSwitchers() {

		changed, word, obpos, oepos := switcher(selection, increase)
		if !changed {
			continue
		}

		// We are only interested in the end position after all runs
		epos = bpos + oepos
		bpos = bpos + obpos
		if cpos < bpos || cpos >= epos {
			continue
		}

		// Update the line and the cursor, and return since we have a handler that has been ran.
		begin := string(rl.line[:bpos])
		end := string(rl.line[epos:])
		newLine := append([]rune(begin), []rune(word)...)
		newLine = append(newLine, []rune(end)...)
		rl.line = newLine
		rl.pos = bpos + len(word) - 1

		return
	}
}

func (rl *Instance) exchangePointAndMark() {
}

func (rl *Instance) transposeChars() {
}

// 	"^[N": "history-search-forward",
func (rl *Instance) historySearchForward() {
}

// 	"^[P": "history-search-backward",
func (rl *Instance) historySearchSackward() {
}

// "^[ ":  "expand-history",
// "^[!":  "expand-history",
func (rl *Instance) expandHistory() {
}

func (rl *Instance) acceptAndHold() {
}

func (rl *Instance) acceptAndInferNextHistory() {
}

func (rl *Instance) acceptAndDownHistory() {
}

// 	"^[y":     "yank-pop",
// func (rl *Instance) yankPop() {
// }

// "^[$":  "spell-word",
// func (rl *Instance) spellWord() {
// }

// "^[.":  "insert-last-word",
// func (rl *Instance) insertLastWord() {
// }
//

// 	"^[A": "accept-and-hold",
// func (rl *Instance) acceptAndHold() {
// }

// func (rl *Instance) getLine() {
// }

// 	"^[Q": "push-line",
// func (rl *Instance) pushLine() {
// }

// func (rl *Instance) runHelp() {
// }

// 	"^[x":     "execute-named-cmd",
// func (rl *Instance) executeNamedCmd() {
// }

// 	"^[z":     "execute-last-named-cmd",
// func (rl *Instance) executeLastNamedCmd() {
// }
