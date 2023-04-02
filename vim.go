package readline

import (
	"unicode"

	"github.com/reeflective/readline/internal/keymap"
	"github.com/reeflective/readline/internal/strutil"
	"github.com/xo/inputrc"
)

// lineWidgets maps widget names to their implementation.
type lineWidgets map[string]func()

// viWidgets returns all Vim commands.
// Under each comment are gathered all commands related to the comment's
// subject. When there are two subgroups separated by an empty line, the
// second one comprises commands that are not legacy readline commands.
//
// Modes
// Moving
// Changing text
// Killing and Yanking
// Selecting text
// Miscellaneous.
func (rl *Shell) viWidgets() lineWidgets {
	return map[string]func(){
		// Modes enter/exit
		"vi-append-mode":      rl.viAddNext,     // vi-add-next
		"vi-append-eol":       rl.viAddEol,      // vi-add-eol
		"vi-insertion-mode":   rl.viInsertMode,  // vi-insert-mode
		"vi-insert-beg":       rl.viInsertBol,   // vi-insert-bol
		"vi-movement-mode":    rl.viCommandMode, // vi-cmd-mode
		"vi-visual-mode":      rl.viVisualMode,
		"vi-visual-line-mode": rl.viVisualLineMode,
		"vi-editing-mode":     rl.viInsertMode,

		// Movement
		"vi-backward-char":    rl.viBackwardChar,        // TODO not multiline anymore
		"vi-forward-char":     rl.viForwardChar,         // TODO not multiline anymore
		"vi-prev-word":        rl.viBackwardWord,        // vi-backward-word
		"vi-next-word":        rl.viForwardWord,         // vi-forward-word
		"vi-backward-word":    rl.viBackwardWord,        // vi-backward-word
		"vi-forward-word":     rl.viForwardWord,         // vi-forward-word
		"vi-backward-bigword": rl.viBackwardBlankWord,   // vi-backward-blank-word
		"vi-forward-bigword":  rl.viForwardBlankWord,    // vi-forward-blank-word
		"vi-end-word":         rl.viForwardWordEnd,      // vi-forward-word-end
		"vi-end-bigword":      rl.viForwardBlankWordEnd, // vi-forward-blank-word-end
		"vi-match":            rl.viMatchBracket,        // vi-match-bracket
		"vi-column":           rl.viGotoColumn,          // vi-goto-column
		"vi-end-of-line":      rl.viEndOfLine,
		"vi-back-to-indent":   rl.viBackToIndent,
		"vi-first-print":      rl.viFirstPrint,
		"vi-goto-mark":        rl.viGotoMark,

		"vi-backward-end-word":    rl.viBackwardWordEnd,      // vi-backward-word-end
		"vi-backward-end-bigword": rl.viBackwardBlankWordEnd, // vi-backward-blank-word-end

		// Changing text
		"vi-change-to":   rl.viChangeTo,   // vi-change
		"vi-delete-to":   rl.viDeleteTo,   // vi-delete
		"vi-delete":      rl.viDeleteChar, // vi-delete-chars
		"vi-change-char": rl.viChangeChar, // vi-replace-chars
		"vi-replace":     rl.viReplace,    // vi-overstrike and vi-overstrike-delete
		"vi-change-case": rl.viChangeCase, // vi-swap-case
		"vi-down-case":   rl.viDownCase,
		"vi-up-case":     rl.viUpCase,
		"vi-subst":       rl.viSubstitute, // vi-substitute

		"vi-change-eol":      rl.viChangeEol,
		"vi-add-surround":    rl.viAddSurround,
		"vi-change-surround": rl.viChangeSurround,
		"vi-open-line-above": rl.viOpenLineAbove,
		"vi-open-line-below": rl.viOpenLineBelow,

		// Kill and Yanking
		"vi-kill-eol":         rl.viKillEol,
		"vi-unix-word-rubout": rl.backwardKillWord, // backward-kill-word
		"vi-rubout":           rl.viRubout,
		"vi-yank-to":          rl.viYankTo, // vi-yank
		// "vi-yank-pop":                       rl.viYank,
		// "vi-yank-arg":                       rl.viYank,

		"vi-kill-line":       rl.viKillLine,
		"vi-put":             rl.viPut,
		"vi-put-after":       rl.viPutAfter,
		"vi-put-before":      rl.viPutBefore,
		"vi-set-buffer":      rl.viSetBuffer,
		"vi-yank-whole-line": rl.viYankWholeLine,

		// Selecting text
		"select-a-blank-word":  rl.viSelectABlankWord,
		"select-a-shell-word":  rl.viSelectAShellWord,
		"select-a-word":        rl.viSelectAWord,
		"select-in-blank-word": rl.viSelectInBlankWord,
		"select-in-shell-word": rl.viSelectInShellWord,
		"select-in-word":       rl.viSelectInWord,
		"vi-select-surround":   rl.viSelectSurround,

		// Miscellaneous
		"vi-eof-maybe": rl.viEofMaybe,
		// "vi-search"
		// "vi-search-again"
		"vi-arg-digit":                rl.viArgDigit, // vi-digit-or-beginning-of-line
		"vi-char-search":              rl.viCharSearch,
		"vi-set-mark":                 rl.viSetMark,
		"vi-edit-and-execute-command": rl.viEditAndExecuteCommand,
		"vi-undo":                     rl.undoLast,
		"vi-redo":                     rl.redo,

		"vi-edit-command-line":   rl.viEditCommandLine,
		"vi-find-next-char":      rl.viFindNextChar,
		"vi-find-next-char-skip": rl.viFindNextCharSkip,
		"vi-find-prev-char":      rl.viFindPrevChar,
		"vi-find-prev-char-skip": rl.viFindPrevCharSkip,
	}
}

//
// Modes ----------------------------------------------------------------
//

func (rl *Shell) viInsertMode() {
	// Reset any visual selection and iterations.
	rl.selection.Reset()
	rl.iterations.Reset()

	// Change the keymap and mark the insertion point.
	rl.keymaps.SetLocal("")
	rl.keymaps.SetMain(keymap.ViIns)
	rl.cursor.SetMark()
}

func (rl *Shell) viCommandMode() {
	// Reset any visual selection and iterations.
	rl.undo.SkipSave()
	rl.iterations.Reset()
	rl.selection.Reset()

	// Only go back if not in insert mode
	if rl.keymaps.Main() == keymap.ViIns && !rl.cursor.AtBeginningOfLine() {
		rl.cursor.Dec()
	}

	// Update the cursor position, keymap and insertion point.
	rl.cursor.CheckCommand()
	rl.keymaps.SetLocal("")
	rl.keymaps.SetMain(keymap.ViCmd)
}

func (rl *Shell) viVisualMode() {
	rl.undo.SkipSave()
	rl.iterations.Reset()

	// Mark the selection as visual at the current cursor position.
	rl.selection.Mark(rl.cursor.Pos())
	rl.selection.Visual(false)
	rl.keymaps.SetLocal(keymap.Visual)
}

func (rl *Shell) viVisualLineMode() {
	rl.undo.SkipSave()
	rl.iterations.Reset()

	// Mark the selection as visual at the current
	// cursor position, in visual line mode.
	rl.selection.Mark(rl.cursor.Pos())
	rl.selection.Visual(true)
	rl.keymaps.SetLocal(keymap.Visual)

	rl.keymaps.PrintCursor(keymap.Visual)
}

func (rl *Shell) viInsertBol() {
	rl.iterations.Reset()
	rl.beginningOfLine()
	rl.viInsertMode()
}

func (rl *Shell) viAddNext() {
	if rl.line.Len() > 0 {
		rl.cursor.Inc()
	}

	rl.viInsertMode()
}

func (rl *Shell) viAddEol() {
	rl.iterations.Reset()

	if rl.keymaps.Local() == keymap.Visual {
		rl.cursor.Inc()
		rl.viInsertMode()
		return
	}

	rl.endOfLine()
	rl.viInsertMode()
}

//
// Movement -------------------------------------------------------------
//

// TODO: autosuggestion insert.
func (rl *Shell) viForwardChar() {
	// Only exception where we actually don't forward a character.
	// if rl.config.HistoryAutosuggest && len(rl.histSuggested) > 0 && rl.cursor.Pos() == rl.line.Len()-1 {
	// 	rl.historyAutosuggestInsert()
	// 	return
	// }

	rl.undo.SkipSave()

	// In vi-cmd-mode, we don't go further than the
	// last character in the line, hence rl.line-1
	if rl.keymaps.Main() != keymap.ViIns && rl.cursor.Pos() < rl.line.Len()-1 {
		if (*rl.line)[rl.cursor.Pos()+1] != '\n' {
			rl.cursor.Inc()
		}

		return
	}
}

func (rl *Shell) viBackwardChar() {
	rl.undo.SkipSave()

	if (*rl.line)[rl.cursor.Pos()-1] != '\n' {
		rl.cursor.Dec()
	}
}

func (rl *Shell) viBackwardWord() {
	rl.undo.SkipSave()

	vii := rl.iterations.Get()
	for i := 1; i <= vii; i++ {
		backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(backward)
	}
}

func (rl *Shell) viForwardWord() {
	rl.undo.SkipSave()

	vii := rl.iterations.Get()
	for i := 1; i <= vii; i++ {
		forward := rl.line.Forward(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(forward)
	}
}

func (rl *Shell) viBackwardBlankWord() {
	rl.undo.SkipSave()

	vii := rl.iterations.Get()
	for i := 1; i <= vii; i++ {
		backward := rl.line.Backward(rl.line.TokenizeSpace, rl.cursor.Pos())
		rl.cursor.Move(backward)
	}
}

func (rl *Shell) viForwardBlankWord() {
	rl.undo.SkipSave()

	vii := rl.iterations.Get()
	for i := 1; i <= vii; i++ {
		forward := rl.line.Forward(rl.line.TokenizeSpace, rl.cursor.Pos())
		rl.cursor.Move(forward)
	}
}

func (rl *Shell) viBackwardWordEnd() {
	rl.undo.SkipSave()

	vii := rl.iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.cursor.Inc()

		rl.cursor.Move(rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos()))
		rl.cursor.Move(rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos()))

		// Then move forward, adjusting if we are on a punctuation.
		if strutil.IsPunctuation((*rl.line)[rl.cursor.Pos()]) {
			rl.cursor.Dec()
		}

		rl.cursor.Move(rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos()))
	}
}

func (rl *Shell) viForwardWordEnd() {
	rl.undo.SkipSave()
	vii := rl.iterations.Get()
	for i := 1; i <= vii; i++ {
		forward := rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(forward)
	}
}

func (rl *Shell) viBackwardBlankWordEnd() {
	rl.undo.SkipSave()

	vii := rl.iterations.Get()

	for i := 1; i <= vii; i++ {
		rl.cursor.Inc()

		rl.cursor.Move(rl.line.Backward(rl.line.TokenizeSpace, rl.cursor.Pos()))
		rl.cursor.Move(rl.line.Backward(rl.line.TokenizeSpace, rl.cursor.Pos()))

		rl.cursor.Move(rl.line.ForwardEnd(rl.line.TokenizeSpace, rl.cursor.Pos()))
	}
}

func (rl *Shell) viForwardBlankWordEnd() {
	rl.undo.SkipSave()
	vii := rl.iterations.Get()
	for i := 1; i <= vii; i++ {
		rl.cursor.Move(rl.line.ForwardEnd(rl.line.TokenizeSpace, rl.cursor.Pos()))
	}
}

func (rl *Shell) viMatchBracket() {
	rl.undo.SkipSave()

	nextPos := rl.cursor.Pos()
	found := false

	// If we are on a bracket/brace/parenthesis, we just find the matcher
	if !strutil.IsBracket((*rl.line)[rl.cursor.Pos()]) {
		for i := rl.cursor.Pos() + 1; i < rl.line.Len(); i++ {
			char := (*rl.line)[i]
			if char == '}' || char == ')' || char == ']' {
				nextPos = i - rl.cursor.Pos()
				found = true

				break
			}
		}

		if !found {
			return
		}

		rl.cursor.Move(nextPos)
	}

	var adjust int

	split, index, pos := rl.line.TokenizeBlock(rl.cursor.Pos())

	switch {
	case len(split) == 0:
		return
	case pos == 0:
		adjust = len(split[index])
	default:
		adjust = pos * -1
	}

	rl.cursor.Move(adjust)
}

// TODO: Finish.
func (rl *Shell) viGotoColumn() {
	// rl.undo.SkipSave()
	// iterations := rl.iterations
	// column := rl.iterations.Get()
	//
	// if iterations == "" {
	// 	column = 0
	// } else if column < 0 {
	// 	column = rl.line.Len() + column
	// }
	//
	//    // Get the width of the current line.
	//
	// rl.pos = column
}

func (rl *Shell) viEndOfLine() {
	rl.undo.SkipSave()
	rl.cursor.EndOfLine()
}

func (rl *Shell) viFirstPrint() {
	rl.cursor.BeginningOfLine()
	rl.cursor.ToFirstNonSpace(true)
}

func (rl *Shell) viBackToIndent() {
	rl.cursor.BeginningOfLine()
	rl.cursor.ToFirstNonSpace(true)
}

func (rl *Shell) viGotoMark() {
	switch {
	case rl.selection.Active():
		// We either an active selection, in which case
		// we go to the position (begin or end) that is
		// set and not equal to the cursor.
		bpos, epos := rl.selection.Pos()
		if bpos != rl.cursor.Pos() {
			rl.cursor.Set(bpos)
		} else {
			rl.cursor.Set(epos)
		}

	case rl.cursor.Mark() != -1:
		// Or we go to the cursor mark, which was set when
		// entering insert mode. This might have no effect.
		rl.cursor.Set(rl.cursor.Mark())
	}
}

//
// Changing Text --------------------------------------------------------
//

func (rl *Shell) viChangeTo() {
	switch {
	case rl.keymaps.IsCaller():
		// In vi operator pending mode, it's that we've been called
		// twice in a row (eg. `cc`), so copy the entire current line.
		rl.undo.SaveWithPos(*rl.line, rl.cursor.Pos())
		rl.undo.SkipSave()

		rl.selection.Mark(rl.cursor.Pos())
		rl.selection.Visual(true)
		rl.selection.Cut()
		rl.viInsertMode()

	case rl.selection.Active():
		// In visual mode, we have just have a selection to delete.
		rl.undo.Save(*rl.line, *rl.cursor)
		rl.undo.SkipSave()

		cpos := rl.selection.Cursor()
		rl.selection.Cut()
		rl.cursor.Set(cpos)

		rl.viInsertMode()

	default:
		// Since we must emulate the default readline behavior,
		// we vary our behavior depending on the caller key.
		key, _ := rl.keys.Peek()

		switch key {
		case 'c':
			rl.keymaps.Pending()
			rl.selection.Mark(rl.cursor.Pos())
			// isSurround := action == "vi-select-surround"
			// if isSurround {
			// 	rl.keys = key
			// }
		case 'C':
			rl.viChangeEol()
		}
	}
}

func (rl *Shell) viDeleteTo() {
	switch {
	case rl.keymaps.IsCaller():
		// In vi operator pending mode, it's that we've been called
		// twice in a row (eg. `dd`), so delete the entire current line.
		rl.undo.SaveWithPos(*rl.line, rl.cursor.Pos())
		rl.undo.SkipSave()

		rl.selection.Mark(rl.cursor.Pos())
		rl.selection.Visual(true)
		rl.cursor.Set(rl.selection.Cursor())

		cut := rl.selection.Cut()
		rl.buffers.Write([]rune(cut)...)

	case rl.selection.Active():
		// In visual mode, or with a non-empty selection, just cut it.
		rl.undo.SaveWithPos(*rl.line, rl.cursor.Pos())
		rl.undo.SkipSave()

		cpos := rl.selection.Cursor()
		cut := rl.selection.Cut()
		rl.buffers.Write([]rune(cut)...)
		rl.cursor.Set(cpos)

		rl.viCommandMode()

	default:
		// Since we must emulate the default readline behavior,
		// we vary our behavior depending on the caller key.
		key, _ := rl.keys.Peek()

		switch key {
		case 'd':
			rl.keymaps.Pending()
			rl.selection.Mark(rl.cursor.Pos())
		case 'D':
			rl.viKillEol()
		}
	}
}

func (rl *Shell) viDeleteChar() {
	rl.undo.Save(*rl.line, *rl.cursor)

	cutBuf := make([]rune, 0)

	vii := rl.iterations.Get()

	for i := 1; i <= vii; i++ {
		cutBuf = append(cutBuf, (*rl.line)[rl.cursor.Pos()])
		rl.line.CutRune(rl.cursor.Pos())
	}

	rl.buffers.Write(cutBuf...)
}

func (rl *Shell) viChangeChar() {
	// We read a character to use first.
	key, isAbort := rl.keys.ReadArgument()
	if isAbort || len(key) == 0 {
		rl.undo.SkipSave()
		return
	}

	switch {
	case rl.selection.Active() && rl.selection.IsVisual():
		// In visual mode, we replace all chars of the selection
		rl.selection.ReplaceWith(func(r rune) rune {
			return key[0]
		})
	default:
		// Or simply the character under the cursor.
		rl.line.CutRune(rl.cursor.Pos())
		rl.line.Insert(rl.cursor.Pos(), []rune(key)...)
	}
}

func (rl *Shell) viReplace() {
	// We store the current line as an undo item first, but will not
	// store any intermediate changes (in the loop below) as undo items.
	rl.undo.Save(*rl.line, *rl.cursor)

	// All replaced characters are stored, to be used with backspace
	cache := make([]rune, 0)

	// Don't use the delete cache past the end of the line
	lineStart := rl.line.Len()

	// The replace mode is quite special in that it does escape back
	// to the main readline loop: it keeps reading characters and inserts
	// them as long as the escape key is not pressed.
	for {
		// We read a character to use first.
		keys, isAbort := rl.keys.ReadArgument()
		if isAbort {
			break
		}

		key := keys[0]

		// If the key is a backspace, we go back one character
		if string(key) == inputrc.Unescape(string(`\C-?`)) {
			if rl.cursor.Pos() > lineStart {
				rl.backwardDeleteChar()
			} else if rl.cursor.Pos() > 0 {
				rl.cursor.Dec()
			}

			// And recover the last replaced character
			if len(cache) > 0 && rl.cursor.Pos() < lineStart {
				key = cache[len(cache)-1]
				cache = cache[:len(cache)-1]
				(*rl.line)[rl.cursor.Pos()] = key
			}
		} else {
			// If the cursor is at the end of the line,
			// we insert the character instead of replacing.
			if rl.line.Len() == rl.cursor.Pos() {
				rl.line.Insert(rl.cursor.Pos(), key)
			} else {
				cache = append(cache, (*rl.line)[rl.cursor.Pos()])
				(*rl.line)[rl.cursor.Pos()] = key
			}

			rl.cursor.Inc()
		}

		// Update the line
		rl.display.Refresh()
	}

	// When exiting the replace mode, move the cursor back
	rl.cursor.Dec()
}

func (rl *Shell) viChangeCase() {
	switch {
	case rl.selection.Active() && rl.selection.IsVisual():
		rl.selection.ReplaceWith(func(char rune) rune {
			if unicode.IsLower(char) {
				return unicode.ToUpper(char)
			}

			return unicode.ToLower(char)
		})

	default:
		char := (*rl.line)[rl.cursor.Pos()]
		if unicode.IsLower(char) {
			char = unicode.ToUpper(char)
		} else {
			char = unicode.ToLower(char)
		}

		(*rl.line)[rl.cursor.Pos()] = char
	}
}

func (rl *Shell) viDownCase() {
	rl.undo.SkipSave()

	switch {
	case rl.keymaps.IsCaller():
		// In vi operator pending mode, it's that we've been called
		// twice in a row (eg. `uu`), so modify the entire current line.
		rl.undo.Save(*rl.line, *rl.cursor)
		rl.undo.SkipSave()

		rl.selection.Mark(rl.cursor.Pos())
		rl.selection.Visual(true)
		rl.selection.ReplaceWith(unicode.ToLower)
		rl.viCommandMode()

	case rl.selection.Active():
		rl.selection.ReplaceWith(unicode.ToLower)
		rl.viCommandMode()

	default:
		// Else if we are actually starting a yank action.
		rl.undo.SkipSave()
		rl.keymaps.Pending()
		rl.selection.Mark(rl.cursor.Pos())
	}
}

func (rl *Shell) viUpCase() {
	rl.undo.SkipSave()

	switch {
	case rl.keymaps.IsCaller():
		// In vi operator pending mode, it's that we've been called
		// twice in a row (eg. `uu`), so modify the entire current line.
		rl.undo.Save(*rl.line, *rl.cursor)
		rl.undo.SkipSave()

		rl.selection.Mark(rl.cursor.Pos())
		rl.selection.Visual(true)
		rl.selection.ReplaceWith(unicode.ToUpper)
		rl.viCommandMode()

	case rl.selection.Active():
		rl.selection.ReplaceWith(unicode.ToUpper)
		rl.viCommandMode()

	default:
		// Else if we are actually starting a yank action.
		rl.undo.SkipSave()
		rl.keymaps.Pending()
		rl.selection.Mark(rl.cursor.Pos())
	}
}

func (rl *Shell) viSubstitute() {
	switch {
	case rl.selection.Active():
		// Delete the selection and enter insert mode.
		cpos := rl.selection.Cursor()
		rl.selection.Cut()
		rl.cursor.Set(cpos)
		rl.viInsertMode()

	default:
		// Delete next characters and enter insert mode.
		vii := rl.iterations.Get()
		for i := 1; i <= vii; i++ {
			rl.line.CutRune(rl.cursor.Pos())
		}

		rl.viInsertMode()
	}
}

func (rl *Shell) viChangeEol() {
	rl.undo.Save(*rl.line, *rl.cursor)
	rl.undo.SkipSave()

	pos := rl.cursor.Pos()
	rl.selection.Mark(pos)
	rl.cursor.EndOfLineAppend()
	rl.selection.Cut()
	rl.cursor.Set(pos)

	rl.iterations.Reset()
	rl.display.ResetHelpers()
	rl.viInsertMode()
}

func (rl *Shell) viAddSurround() {
	// Get the surround character to change.
	key, isAbort := rl.keys.ReadArgument()
	if isAbort {
		rl.undo.SkipSave()
		return
	}

	bchar, echar := strutil.MatchSurround(rune(key[0]))

	rl.undo.Save(*rl.line, *rl.cursor)

	// Surround the selection
	rl.selection.Surround(bchar, echar)
}

func (rl *Shell) viChangeSurround() {
	rl.undo.SkipSave()

	// Read a key as a rune to search for
	key, isAbort := rl.keys.ReadArgument()
	if isAbort {
		return
	}

	char := rune(key[0])

	// Find the corresponding enclosing chars
	bpos, epos, _, _ := rl.line.FindSurround(char, rl.cursor.Pos())
	if bpos == -1 && epos == -1 {
		return
	}

	// Add those two positions to highlighting and update.
	rl.selection.MarkSurround(bpos, epos)
	rl.display.Refresh()

	defer func() { rl.selection.Reset() }()

	// Now read another key
	key, isAbort = rl.keys.ReadArgument()
	if isAbort {
		return
	}

	rl.undo.Save(*rl.line, *rl.cursor)

	rchar := rune(key[0])

	// There might be a matching equivalent.
	bchar, echar := strutil.MatchSurround(rchar)

	(*rl.line)[bpos] = bchar
	(*rl.line)[epos] = echar
}

func (rl *Shell) viOpenLineAbove() {
	rl.undo.Save(*rl.line, *rl.cursor)
	rl.beginningOfLine()
	rl.line.Insert(rl.cursor.Pos(), '\n')
	rl.viInsertMode()
}

func (rl *Shell) viOpenLineBelow() {
	rl.undo.Save(*rl.line, *rl.cursor)
	rl.endOfLine()
	rl.line.Insert(rl.cursor.Pos(), '\n')
	rl.cursor.Inc()
	rl.viInsertMode()
}

//
// Killing & Yanking ----------------------------------------------------
//

func (rl *Shell) viKillEol() {
	rl.undo.Save(*rl.line, *rl.cursor)
	rl.undo.SkipSave()

	pos := rl.cursor.Pos()
	rl.selection.Mark(rl.cursor.Pos())
	rl.cursor.EndOfLineAppend()

	cut := rl.selection.Cut()
	rl.buffers.Write([]rune(cut)...)
	rl.cursor.Set(pos)

	if !rl.cursor.AtBeginningOfLine() {
		rl.cursor.Dec()
	}

	rl.iterations.Reset()
	rl.display.ResetHelpers()
}

func (rl *Shell) viRubout() {
	if rl.keymaps.Main() != keymap.ViIns {
		rl.undo.Save(*rl.line, *rl.cursor)
	}

	vii := rl.iterations.Get()

	cut := make([]rune, 0)

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {
		if rl.cursor.Pos() == 0 {
			break
		}

		rl.cursor.Dec()
		cut = append(cut, (*rl.line)[rl.cursor.Pos()])
		rl.line.CutRune(rl.cursor.Pos())
	}

	rl.buffers.Write(cut...)
}

func (rl *Shell) viYankTo() {
	rl.undo.SkipSave()

	switch {
	case rl.keymaps.IsCaller():
		// In vi operator pending mode, it's that we've been called
		// twice in a row (eg. `yy`), so copy the entire current line.
		rl.selection.Mark(rl.cursor.Pos())
		rl.selection.Visual(true)

		text, _, _, cpos := rl.selection.Pop()
		rl.buffers.Write([]rune(text)...)
		rl.cursor.Set(cpos)

	case rl.selection.Active():
		// In visual mode, or with a non-empty selection, just yank.
		text, _, _, cpos := rl.selection.Pop()

		rl.buffers.Write([]rune(text)...)
		rl.cursor.Set(cpos)

		rl.viCommandMode()

	default:
		// Since we must emulate the default readline behavior,
		// we vary our behavior depending on the caller key.
		key, _ := rl.keys.Peek()

		switch key {
		case 'y':
			rl.keymaps.Pending()
			rl.selection.Mark(rl.cursor.Pos())
		case 'Y':
			rl.viYankWholeLine()
		}
	}
}

func (rl *Shell) viYankWholeLine() {
	rl.undo.SkipSave()

	// calculate line selection.
	rl.selection.Mark(rl.cursor.Pos())
	rl.selection.Visual(true)

	bpos, epos := rl.selection.Pos()

	// If selection has a new line, remove it.
	if (*rl.line)[epos-1] == '\n' {
		epos--
	}

	// Pass the buffer to register.
	buffer := (*rl.line)[bpos:epos]
	rl.buffers.Write(buffer...)

	// Done with any selection.
	rl.selection.Reset()
}

func (rl *Shell) viKillLine() {
	if rl.cursor.Pos() <= rl.cursor.Mark() || rl.cursor.Pos() == 0 {
		return
	}

	rl.undo.Save(*rl.line, *rl.cursor)
	rl.undo.SkipSave()

	rl.selection.MarkRange(rl.cursor.Mark(), rl.line.Len())
	rl.cursor.Dec()
	cut := rl.selection.Cut()
	rl.buffers.Write([]rune(cut)...)
}

func (rl *Shell) viPut() {
	key, _ := rl.keys.Peek()

	switch key {
	case 'P':
		rl.viPutBefore()
	case 'p':
		fallthrough
	default:
		rl.viPutAfter()
	}
}

func (rl *Shell) viPutAfter() {
	rl.undo.Save(*rl.line, *rl.cursor)

	buffer := rl.buffers.Active()

	// Add newlines when pasting an entire line.
	if buffer[len(buffer)-1] == '\n' {
		if !rl.cursor.OnEmptyLine() {
			rl.cursor.EndOfLineAppend()
		}

		if rl.cursor.Pos() == rl.line.Len() {
			buffer = append([]rune{'\n'}, buffer[:len(buffer)-2]...)
		}
	}

	rl.cursor.Inc()
	pos := rl.cursor.Pos()

	vii := rl.iterations.Get()
	for i := 1; i <= vii; i++ {
		rl.line.Insert(pos, buffer...)
	}

	rl.cursor.Set(pos)
}

func (rl *Shell) viPutBefore() {
	rl.undo.Save(*rl.line, *rl.cursor)

	buffer := rl.buffers.Active()

	if buffer[len(buffer)-1] == '\n' {
		rl.cursor.BeginningOfLine()

		if rl.cursor.OnEmptyLine() {
			buffer = append(buffer, '\n')
			rl.cursor.Dec()
		}
	}

	pos := rl.cursor.Pos()

	vii := rl.iterations.Get()
	for i := 1; i <= vii; i++ {
		rl.line.Insert(pos, buffer...)
	}

	rl.cursor.Set(pos)
}

func (rl *Shell) viSetBuffer() {
	rl.undo.SkipSave()

	// Always reset the active register.
	rl.buffers.Reset()

	// Then read a key to select the register
	key, isAbort := rl.keys.ReadArgument()
	if isAbort {
		return
	}

	rl.buffers.SetActive(key[0])
}

//
// Selecting Text -------------------------------------------------------
//

func (rl *Shell) viSelectABlankWord() {
	rl.undo.SkipSave()

	// Go the beginning of the word and start mark
	rl.cursor.ToFirstNonSpace(true)
	backward := rl.line.Backward(rl.line.TokenizeSpace, rl.cursor.Pos())
	rl.cursor.Move(backward)
	bpos := rl.cursor.Pos()

	// Then go to the end of the blank word
	forward := rl.line.Forward(rl.line.TokenizeSpace, rl.cursor.Pos())
	rl.cursor.Move(forward - 1)
	epos := rl.cursor.Pos()

	// Adjust if we are at the end of the line.
	if rl.cursor.Pos() == rl.line.Len()-1 {
		rl.cursor.Set(bpos)

		backward := rl.line.Backward(rl.line.TokenizeSpace, rl.cursor.Pos())
		rl.cursor.Move(backward)

		forward := rl.line.Forward(rl.line.TokenizeSpace, rl.cursor.Pos())
		rl.cursor.Move(forward + 1)

		bpos = rl.cursor.Pos()
		rl.cursor.Set(epos)
	}

	rl.selection.MarkRange(bpos, epos)
}

func (rl *Shell) viSelectAShellWord() {
	rl.undo.SkipSave()

	rl.cursor.ToFirstNonSpace(true)

	sBpos, sEpos, _, _ := rl.line.FindSurround('\'', rl.cursor.Pos())
	dBpos, dEpos, _, _ := rl.line.FindSurround('"', rl.cursor.Pos())

	mark, cpos := strutil.AdjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos)

	// If none matched, use blankword
	if mark == -1 && cpos == -1 {
		rl.viSelectABlankWord()

		return
	}

	// Adjust for spaces preceding
	rl.cursor.Set(mark)
	rl.cursor.ToFirstNonSpace(false)
	bpos := rl.cursor.Pos() + 1

	// Else set the region inside those quotes
	rl.cursor.Set(cpos)
	rl.selection.MarkRange(bpos, rl.cursor.Pos())
}

func (rl *Shell) viSelectAWord() {
	rl.undo.SkipSave()

	rl.cursor.ToFirstNonSpace(true)

	// Go to the beginning of the current word, and select all spaces before it.
	rl.cursor.Move(rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos()))

	spaceBefore := rl.cursor.Pos() > 0 && (*rl.line)[rl.cursor.Pos()-1] == inputrc.Space
	if spaceBefore {
		rl.cursor.Dec()
		rl.cursor.ToFirstNonSpace(false)
		if rl.cursor.Pos() != 0 {
			rl.cursor.Inc()
		}
	}

	bpos := rl.cursor.Pos()

	// Then go the end of the word, and only select further spaces
	// if the word selected is not preceded by spaces as well.
	rl.cursor.Move(rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos()) + 1)

	if !spaceBefore {
		rl.cursor.Inc()
		rl.cursor.ToFirstNonSpace(true)
		rl.cursor.Dec()
	}

	epos := rl.cursor.Pos()

	// Select the range and return: the caller will decide what
	// to do with the cursor position and the selection itself.
	rl.selection.MarkRange(bpos, epos)
}

func (rl *Shell) viSelectInBlankWord() {
	rl.undo.SkipSave()

	// Go the beginning of the word and start mark
	rl.cursor.ToFirstNonSpace(true)
	backward := rl.line.Backward(rl.line.TokenizeSpace, rl.cursor.Pos())
	rl.cursor.Move(backward)
	bpos := rl.cursor.Pos()

	// Then go to the end of the blank word
	forward := rl.line.ForwardEnd(rl.line.TokenizeSpace, rl.cursor.Pos())
	rl.cursor.Move(forward)
	epos := rl.cursor.Pos()

	// Select the range and return: the caller will decide what
	// to do with the cursor position and the selection itself.
	rl.selection.MarkRange(bpos, epos)
}

func (rl *Shell) viSelectInShellWord() {
	rl.undo.SkipSave()

	rl.cursor.ToFirstNonSpace(true)

	sBpos, sEpos, _, _ := rl.line.FindSurround('\'', rl.cursor.Pos())
	dBpos, dEpos, _, _ := rl.line.FindSurround('"', rl.cursor.Pos())

	mark, cpos := strutil.AdjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos)

	// If none matched, use blankword
	if mark == -1 && cpos == -1 {
		rl.viSelectInBlankWord()

		return
	}

	// Select the range and return: the caller will decide what
	// to do with the cursor position and the selection itself.
	rl.selection.MarkRange(mark+1, cpos-1)
}

func (rl *Shell) viSelectInWord() {
	rl.undo.SkipSave()

	rl.cursor.ToFirstNonSpace(true)

	// Go to the beginning of the current word.
	rl.cursor.Move(rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos()))
	bpos := rl.cursor.Pos()

	// Then go to the end of the blank word
	rl.cursor.Move(rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos()) + 1)
	epos := rl.cursor.Pos()

	// Select the range and return: the caller will decide what
	// to do with the cursor position and the selection itself.
	rl.selection.MarkRange(bpos, epos)
}

func (rl *Shell) viSelectSurround() {
	rl.undo.SkipSave()

	var inside bool

	// The surround can be either inside or around a surrounding
	// character, so we look at the input keys: the first one is
	// the only that triggered this command, so check the second.
	// Use the first key to know if inside/around is used.
	key, _ := rl.keys.Pop()
	switch key {
	case 'i':
		inside = true
	}

	// Then use the next key as the surrounding character.
	char, empty := rl.keys.Pop()
	if empty {
		return
	}

	bpos, epos, _, _ := rl.line.FindSurround(char, rl.cursor.Pos())
	if bpos == -1 && epos == -1 {
		return
	}

	if inside {
		bpos++
		epos--
	}

	// Select the range and return: the caller will decide what
	// to do with the cursor position and the selection itself.
	rl.selection.MarkRange(bpos, epos)
}

//
// Miscellaneous ---------------------------------------------------------------
//

func (rl *Shell) viEofMaybe() {
	rl.endOfFile()
}
func (rl *Shell) viSearch()      {}
func (rl *Shell) viSearchAgain() {}

func (rl *Shell) viArgDigit() {
	rl.undo.SkipSave()

	// If the last command was a digit argument,
	// then our Vi iterations' length is not 0
	if len(*rl.iterations) > 0 {
		rl.iterations.Add("0")
		return
	}

	// Else we go the beginning of line.
	rl.beginningOfLine()
}

func (rl *Shell) viCharSearch() {
	var forward, skip bool

	// In order to keep readline compatibility,
	// we check the key triggering the command
	// so set the specific behavior.
	key, _ := rl.keys.Peek()

	switch key {
	case 'F':
		forward = false
		skip = false
	case 't':
		forward = true
		skip = true
	case 'T':
		forward = false
		skip = true
	case 'f':
		fallthrough
	default:
		forward = true
		skip = false
	}

	rl.viFindChar(forward, skip)
}

func (rl *Shell) viSetMark() {
	rl.undo.SkipSave()
	rl.selection.Mark(rl.cursor.Pos())
}

func (rl *Shell) viEditAndExecuteCommand() {}

func (rl *Shell) viEditCommandLine() {}

func (rl *Shell) viFindNextChar() {
	rl.viFindChar(true, false)
}

func (rl *Shell) viFindNextCharSkip() {
	rl.viFindChar(true, true)
}

func (rl *Shell) viFindPrevChar() {
	rl.viFindChar(false, false)
}

func (rl *Shell) viFindPrevCharSkip() {
	rl.viFindChar(false, true)
}

func (rl *Shell) viFindChar(forward, skip bool) {
	rl.undo.SkipSave()

	// Read the argument key to use as a pattern to search
	key, esc := rl.keys.ReadArgument()
	if esc {
		return
	}

	char := key[0]
	times := rl.iterations.Get()

	for i := 1; i <= times; i++ {
		pos := rl.line.Find(char, rl.cursor.Pos(), forward)

		if pos == rl.cursor.Pos() || pos == -1 {
			break
		}

		if forward && skip {
			pos--
		} else if !forward && skip {
			pos++
		}

		rl.cursor.Set(pos)
	}
}
