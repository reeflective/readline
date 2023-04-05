package readline

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"unicode"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/keymap"
	"github.com/reeflective/readline/internal/strutil"
	"github.com/reeflective/readline/internal/term"
	"github.com/xo/inputrc"
)

// standardWidgets returns all standard/emacs commands.
// Under each comment are gathered all commands related to the comment's
// subject. When there are two subgroups separated by an empty line, the
// second one comprises commands that are not legacy readline commands.
//
// Modes
// Moving
// Changing text
// Killing and Yanking
// Numeric arguments.
// Macros
// Miscellaneous.
func (rl *Shell) standardWidgets() lineWidgets {
	widgets := map[string]func(){
		// Modes
		"emacs-editing-mode": rl.emacsEditingMode,

		// Moving
		"forward-char":         rl.forwardChar,
		"backward-char":        rl.backwardChar,
		"forward-word":         rl.forwardWord,
		"backward-word":        rl.backwardWord,
		"shell-forward-word":   rl.forwardShellWord,
		"shell-backward-word":  rl.backwardShellWord,
		"beginning-of-line":    rl.beginningOfLine,
		"end-of-line":          rl.endOfLine,
		"previous-screen-line": rl.upLine,   // up-line
		"next-screen-line":     rl.downLine, // down-line
		"clear-screen":         rl.clearScreen,
		"clear-display":        rl.clearDisplay,
		"redraw-current-line":  rl.display.Refresh,

		// Changing text
		"end-of-file":                  rl.endOfFile,
		"delete-char":                  rl.deleteChar,
		"backward-delete-char":         rl.backwardDeleteChar,
		"forward-backward-delete-char": rl.forwardBackwardDeleteChar,
		"self-insert":                  rl.selfInsert,
		"quoted-insert":                rl.quotedInsert,
		"bracketed-paste-begin":        rl.bracketedPasteBegin, // TODO: Finish and find how to do it.
		"transpose-chars":              rl.transposeChars,
		"transpose-words":              rl.transposeWords, // TODO: test.
		"shell-transpose-words":        rl.shellTransposeWords,
		"down-case-word":               rl.downCaseWord,
		"up-case-word":                 rl.upCaseWord,
		"capitalize-word":              rl.capitalizeWord,
		"overwrite-mode":               rl.overwriteMode,
		"delete-horizontal-whitespace": rl.deleteHorizontalWhitespace,

		"delete-word":      rl.deleteWord,
		"quote-region":     rl.quoteRegion,
		"quote-line":       rl.quoteLine,
		"keyword-increase": rl.keywordIncrease,
		"keyword-decrease": rl.keywordDecrease,

		// Killing & yanking
		"kill-line":                rl.killLine,
		"backward-kill-line":       rl.backwardKillLine,
		"unix-line-discard":        rl.backwardKillLine,
		"kill-whole-line":          rl.killWholeLine,
		"kill-word":                rl.killWord,
		"backward-kill-word":       rl.backwardKillWord,
		"shell-kill-word":          rl.shellKillWord,
		"shell-backward-kill-word": rl.shellBackwardKillWord,
		"unix-word-rubout":         rl.backwardKillWord,
		"kill-region":              rl.killRegion,
		"copy-region-as-kill":      rl.copyRegionAsKill,
		"copy-backward-word":       rl.copyBackwardWord,
		"copy-forward-word":        rl.copyForwardWord,
		"yank":                     rl.yank,
		"yank-pop":                 rl.yankPop,

		"kill-buffer":          rl.killBuffer,
		"copy-prev-shell-word": rl.copyPrevShellWord,

		// Numeric arguments
		"digit-argument": rl.digitArgument,

		// Macros
		"start-kbd-macro":      rl.startKeyboardMacro,
		"end-kbd-macro":        rl.endKeyboardMacro,
		"call-last-kbd-macro":  rl.callLastKeyboardMacro,
		"print-last-kbd-macro": rl.printLastKeyboardMacro,

		// Miscellaneous
		"re-read-init-file":         rl.reReadInitFile,
		"abort":                     rl.abort,
		"prefix-meta":               rl.prefixMeta,
		"undo":                      rl.undoLast,
		"revert-line":               rl.revertLine,
		"set-mark":                  rl.setMark, // set-mark-command
		"exchange-point-and-mark":   rl.exchangePointAndMark,
		"character-search":          rl.characterSearch,
		"character-search-backward": rl.characterSearchBackward,
		"insert-comment":            rl.insertComment,
		"dump-functions":            rl.dumpFunctions,
		"dump-variables":            rl.dumpVariables,
		"dump-macros":               rl.dumpMacros,
		"magic-space":               rl.magicSpace,
		"edit-and-execute-command":  rl.editAndExecuteCommand,
		"edit-command-line":         rl.editCommandLine,

		"redo": rl.redo,
	}

	return widgets
}

//
// Modes ----------------------------------------------------------------
//

func (rl *Shell) emacsEditingMode() {
	rl.keymaps.SetMain(keymap.Emacs)
}

//
// Movement -------------------------------------------------------------
//

// TODO: multiline support.
func (rl *Shell) forwardChar() {
	rl.undo.SkipSave()

	if rl.cursor.Pos() < rl.line.Len() {
		rl.cursor.Inc()
	}
}

func (rl *Shell) backwardChar() {
	rl.undo.SkipSave()

	if rl.cursor.Pos() > 0 {
		rl.cursor.Dec()
	}
}

func (rl *Shell) forwardWord() {
	rl.undo.SkipSave()

	vii := rl.iterations.Get()
	for i := 1; i <= vii; i++ {
		forward := rl.line.Forward(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(forward)
	}
}

func (rl *Shell) backwardWord() {
	rl.undo.SkipSave()

	vii := rl.iterations.Get()
	for i := 1; i <= vii; i++ {
		backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(backward)
	}
}

func (rl *Shell) forwardShellWord() {
	// Try to find enclosing quotes from here
	sBpos, sEpos, _, _ := rl.line.FindSurround('\'', rl.cursor.Pos())
	dBpos, dEpos, _, _ := rl.line.FindSurround('"', rl.cursor.Pos())
	mark, cpos := strutil.AdjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos)

	// And only move the cursor if we found them.
	if mark != -1 && cpos != -1 {
		rl.cursor.Set(cpos)
	}

	// Then move forward to the next word
	forward := rl.line.Forward(rl.line.TokenizeSpace, rl.cursor.Pos())
	rl.cursor.Move(forward)
}

func (rl *Shell) backwardShellWord() {
	// First go the beginning of the blank word
	startPos := rl.cursor.Pos()
	backward := rl.line.Backward(rl.line.TokenizeSpace, startPos)
	rl.cursor.Move(backward)

	// Now try to find enclosing quotes from here.
	sBpos, sEpos, _, _ := rl.line.FindSurround('\'', rl.cursor.Pos())
	dBpos, dEpos, _, _ := rl.line.FindSurround('"', rl.cursor.Pos())
	mark, cpos := strutil.AdjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos)

	// Either exit because we didn't match any, or select.
	if mark == -1 && cpos == -1 {
		return
	}

	rl.cursor.Set(mark)
}

func (rl *Shell) beginningOfLine() {
	rl.undo.SkipSave()
	rl.cursor.BeginningOfLine()
}

func (rl *Shell) endOfLine() {
	rl.undo.SkipSave()
	// If in Vim command mode, cursor
	// will be brought back once later.
	rl.cursor.EndOfLineAppend()
}

func (rl *Shell) upLine() {
	lines := rl.iterations.Get()
	rl.cursor.LineMove(lines * -1)
}

func (rl *Shell) downLine() {
	lines := rl.iterations.Get()
	rl.cursor.LineMove(lines)
}

func (rl *Shell) clearScreen() {
	rl.undo.SkipSave()

	print(term.CursorTopLeft)
	print(term.ClearScreen)

	rl.prompt.PrimaryPrint()
}

func (rl *Shell) clearDisplay() {
	rl.undo.SkipSave()

	print(term.CursorTopLeft)
	print(term.ClearDisplay)

	rl.prompt.PrimaryPrint()
}

//
// Changing Text -----------------------------------------------------------------
//

func (rl *Shell) endOfFile() {
	switch rl.line.Len() {
	case 0:
		rl.display.AcceptLine()
		rl.histories.Accept(false, false, io.EOF)
	default:
		rl.deleteChar()
	}
}

// func (rl *Instance) errorCtrlC() error {
// 	rl.keys = ""
//
// 	// When we have a completion inserted, just cancel the completions.
// 	if len(rl.comp) > 0 {
// 		rl.resetVirtualComp(true)
// 		rl.resetCompletion()
// 		rl.resetIsearch()
// 		rl.resetHintText()
// 		rl.completer = nil
//
// 		return nil
// 	}
//
// 	// Or return the current command line
// 	rl.clearHelpers()
// 	moveCursorDown(rl.fullY - rl.posY)
// 	print("\r\n")
//
// 	return ErrCtrlC

func (rl *Shell) deleteChar() {
	// Extract from bash documentation of readline:
	// Delete the character at point.  If this function is bound
	// to the same character as the tty EOF character, as C-d
	//
	// TODO: We should match the same behavior here.

	rl.undo.Save(*rl.line, *rl.cursor)

	vii := rl.iterations.Get()

	// Delete the chars in the line anyway
	for i := 1; i <= vii; i++ {
		rl.line.CutRune(rl.cursor.Pos())
	}
}

func (rl *Shell) backwardDeleteChar() {
	if rl.keymaps.Main() == keymap.ViIns {
		rl.undo.SkipSave()
	} else {
		rl.undo.Save(*rl.line, *rl.cursor)
	}

	rl.completer.Update()

	if rl.cursor.Pos() == 0 {
		return
	}

	vii := rl.iterations.Get()

	switch vii {
	case 1:
		var toDelete rune
		var isSurround, matcher bool

		if rl.line.Len() > rl.cursor.Pos() {
			toDelete = (*rl.line)[rl.cursor.Pos()-1]
			isSurround = strutil.IsBracket(toDelete) || toDelete == '\'' || toDelete == '"'
			matcher = strutil.IsSurround(toDelete, (*rl.line)[rl.cursor.Pos()])
		}

		rl.cursor.Dec()
		rl.line.CutRune(rl.cursor.Pos())

		if isSurround && matcher {
			rl.cursor.Inc()
			rl.line.CutRune(rl.cursor.Pos())
		}

	default:
		for i := 1; i <= vii; i++ {
			rl.cursor.Dec()
			rl.line.CutRune(rl.cursor.Pos())
		}
	}
}

func (rl *Shell) forwardBackwardDeleteChar() {
	switch rl.cursor.Pos() {
	case rl.line.Len():
		rl.backwardDeleteChar()
	default:
		rl.deleteChar()
	}
}

func (rl *Shell) selfInsert() {
	rl.undo.SkipSave()

	// Handle suffix-autoremoval for inserted completions.
	rl.completer.TrimSuffix()
	rl.completer.Update()

	key, empty := rl.keys.Peek()
	if empty {
		return
	}

	// Insert the unescaped version of the key, and update cursor position.
	unescaped := inputrc.Unescape(string(key))
	rl.line.Insert(rl.cursor.Pos(), []rune(unescaped)...)
	rl.cursor.Move(len(unescaped))
}

func (rl *Shell) quotedInsert() {
	rl.undo.SkipSave()
	rl.completer.TrimSuffix()

	keys, _ := rl.keys.ReadArgument()

	quoted := []rune{}

	for _, key := range keys {
		switch {
		case inputrc.IsControl(key):
			quoted = append(quoted, '^')
			quoted = append(quoted, inputrc.Decontrol(key))
		default:
			quoted = append(quoted, key)
		}
	}

	rl.line.Insert(rl.cursor.Pos(), quoted...)
	rl.cursor.Move(len(quoted))
}

func (rl *Shell) bracketedPasteBegin() {
	println("Keys:")
	keys, _ := rl.keys.PeekAll()
	println(string(keys))
}

func (rl *Shell) transposeChars() {
	if rl.cursor.Pos() < 2 || rl.line.Len() < 2 {
		rl.undo.SkipSave()
		return
	}

	switch {
	case rl.cursor.Pos() == rl.line.Len():
		last := (*rl.line)[rl.cursor.Pos()-1]
		blast := (*rl.line)[rl.cursor.Pos()-2]
		(*rl.line)[rl.cursor.Pos()-2] = last
		(*rl.line)[rl.cursor.Pos()-1] = blast
	default:
		last := (*rl.line)[rl.cursor.Pos()]
		blast := (*rl.line)[rl.cursor.Pos()-1]
		(*rl.line)[rl.cursor.Pos()-1] = last
		(*rl.line)[rl.cursor.Pos()] = blast
	}
}

func (rl *Shell) transposeWords() {
	rl.undo.Save(*rl.line, *rl.cursor)
	startPos := rl.cursor.Pos()

	// Save the current word and move the cursor to its beginning
	rl.cursor.ToFirstNonSpace(true)
	rl.selection.Mark(rl.cursor.Pos())
	forward := rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(forward + 1)
	epos := rl.cursor.Pos()

	backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(backward)
	bpos := rl.cursor.Pos()

	rl.selection.MarkRange(bpos, epos)
	toTranspose, tbpos, tepos, _ := rl.selection.Pop()
	rl.cursor.Set(tbpos)

	// Then move back some number of words
	vii := rl.iterations.Get()
	for i := 1; i <= vii; i++ {
		backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(backward)
	}

	// Save the word to transpose with
	rl.selection.Mark(rl.cursor.Pos())
	forward = rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(forward + 1)

	transposeWith, wbpos, wepos, _ := rl.selection.Pop()

	// We might be on the first word of the line,
	// in which case we don't do anything.
	if wepos > tbpos {
		rl.cursor.Set(startPos)
		return
	}

	// Assemble the newline
	begin := string((*rl.line)[:wbpos])
	newLine := append([]rune(begin), []rune(toTranspose)...)
	newLine = append(newLine, (*rl.line)[wepos:tbpos]...)
	newLine = append(newLine, []rune(transposeWith)...)
	newLine = append(newLine, (*rl.line)[tepos:]...)
	rl.line.Set(newLine...)

	// And replace the cursor
	if vii < 0 {
		rl.cursor.Set(epos)
	} else {
		backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
		rl.cursor.Move(backward)

		for i := 0; i <= vii; i++ {
			forward := rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
			rl.cursor.Move(forward + 1)
		}
	}
}

// TODO: finish when visual and iw/ia operators are done.
func (rl *Shell) shellTransposeWords() {
	rl.undo.Save(*rl.line, *rl.cursor)

	startPos := rl.cursor.Pos()

	// Save the current word
	rl.viSelectAShellWord()
	toTranspose, tbpos, tepos, _ := rl.selection.Pop()

	// First move back the number of words
	rl.cursor.Set(tbpos)
	rl.backwardShellWord()

	// Save the word to transpose with
	rl.viSelectAShellWord()
	transposeWith, wbpos, wepos, _ := rl.selection.Pop()

	// We might be on the first word of the line,
	// in which case we don't do anything.
	if wepos > tbpos {
		rl.cursor.Set(startPos)
		return
	}

	// Assemble the newline
	begin := string((*rl.line)[:wbpos])
	newLine := append([]rune(begin), []rune(toTranspose)...)
	newLine = append(newLine, (*rl.line)[wepos:tbpos]...)
	newLine = append(newLine, []rune(transposeWith)...)
	newLine = append(newLine, (*rl.line)[tepos:]...)
	rl.line.Set(newLine...)

	// And replace cursor
	rl.cursor.Set(startPos)
}

func (rl *Shell) downCaseWord() {
	rl.undo.Save(*rl.line, *rl.cursor)

	startPos := rl.cursor.Pos()

	// Save the current word
	rl.cursor.Inc()
	backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(backward)

	rl.selection.Mark(rl.cursor.Pos())
	forward := rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(forward)

	rl.selection.ReplaceWith(unicode.ToLower)
	rl.cursor.Set(startPos)
}

func (rl *Shell) upCaseWord() {
	rl.undo.Save(*rl.line, *rl.cursor)

	startPos := rl.cursor.Pos()

	// Save the current word
	rl.cursor.Inc()
	backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(backward)

	rl.selection.Mark(rl.cursor.Pos())
	forward := rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(forward)

	rl.selection.ReplaceWith(unicode.ToUpper)
	rl.cursor.Set(startPos)
}

func (rl *Shell) capitalizeWord() {
	rl.undo.Save(*rl.line, *rl.cursor)

	startPos := rl.cursor.Pos()

	rl.cursor.Inc()
	backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(backward)

	letter := (*rl.line)[rl.cursor.Pos()]
	upper := unicode.ToUpper(letter)
	(*rl.line)[rl.cursor.Pos()] = upper
	rl.cursor.Set(startPos)
}

func (rl *Shell) overwriteMode() {
	// We store the current line as an undo item first, but will not
	// store any intermediate changes (in the loop below) as undo items.
	rl.undo.Save(*rl.line, *rl.cursor)

	// The replace mode is quite special in that it does escape back
	// to the main readline loop: it keeps reading characters and inserts
	// them as long as the escape key is not pressed.
	for {
		// Read a new key
		keys, esc := rl.keys.ReadArgument()
		if esc {
			return
		}

		key := keys[0]

		// If the key is a backspace, we go back one character
		if key == inputrc.Backspace {
			rl.backwardDeleteChar()
		} else {
			// If the cursor is at the end of the line,
			// we insert the character instead of replacing.
			if rl.cursor.Pos() == rl.line.Len() {
				rl.line.Insert(rl.cursor.Pos(), key)
			} else {
				(*rl.line)[rl.cursor.Pos()] = key
			}

			rl.cursor.Inc()
		}

		rl.display.Refresh()
	}
}

func (rl *Shell) deleteHorizontalWhitespace() {
	rl.undo.Save(*rl.line, *rl.cursor)

	startPos := rl.cursor.Pos()

	rl.cursor.ToFirstNonSpace(false)

	if rl.cursor.Pos() != startPos {
		rl.cursor.Inc()
	}
	bpos := rl.cursor.Pos()

	rl.cursor.ToFirstNonSpace(true)

	if rl.cursor.Pos() != startPos {
		rl.cursor.Dec()
	}
	epos := rl.cursor.Pos()

	rl.line.Cut(bpos, epos)
	rl.cursor.Set(bpos)
}

func (rl *Shell) deleteWord() {
	rl.undo.Save(*rl.line, *rl.cursor)

	rl.selection.Mark(rl.cursor.Pos())
	forward := rl.line.ForwardEnd(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(forward)

	rl.selection.Cut()
}

func (rl *Shell) quoteRegion() {
	rl.undo.Save(*rl.line, *rl.cursor)

	rl.selection.Surround('\'', '\'')
	rl.cursor.Inc()
}

func (rl *Shell) quoteLine() {
	if rl.line.Len() == 0 {
		return
	}

	rl.line.Insert(0, '\'')

	for pos, char := range *rl.line {
		if char == '\n' {
			break
		}

		if char == '\'' {
			(*rl.line)[pos] = '"'
		}
	}

	rl.line.Insert(rl.line.Len(), '\'')
}

func (rl *Shell) keywordIncrease() {
	rl.undo.Save(*rl.line, *rl.cursor)
	rl.keywordSwitch(true)
}

func (rl *Shell) keywordDecrease() {
	rl.undo.Save(*rl.line, *rl.cursor)
	rl.keywordSwitch(false)
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
// 2 -a  => -a.
func (rl *Shell) keywordSwitch(increase bool) {
	cpos := strutil.AdjustNumberOperatorPos(rl.cursor.Pos(), *rl.line)

	// Select in word and get the selection positions
	bpos, epos := rl.line.SelectWord(cpos)
	epos++

	// Move the cursor backward if needed/possible
	if bpos != 0 && ((*rl.line)[bpos-1] == '+' || (*rl.line)[bpos-1] == '-') {
		bpos--
	}

	// Get the selection string
	selection := string((*rl.line)[bpos:epos])

	// For each of the keyword handlers, run it, which returns
	// false/none if didn't operate, then continue to next handler.
	for _, switcher := range strutil.KeywordSwitchers() {
		vii := rl.iterations.Get()

		changed, word, obpos, oepos := switcher(selection, increase, vii)
		if !changed {
			continue
		}

		// We are only interested in the end position after all runs
		epos = bpos + oepos
		bpos += obpos

		if cpos < bpos || cpos >= epos {
			continue
		}

		// Update the line and the cursor, and return
		// since we have a handler that has been ran.
		begin := string((*rl.line)[:bpos])
		end := string((*rl.line)[epos:])

		newLine := append([]rune(begin), []rune(word)...)
		newLine = append(newLine, []rune(end)...)
		rl.line.Set(newLine...)
		rl.cursor.Set(bpos + len(word) - 1)

		return
	}
}

//
// Killing & Yanking -------------------------------------------------------------
//

func (rl *Shell) killLine() {
	rl.iterations.Reset()
	rl.undo.Save(*rl.line, *rl.cursor)

	cut := []rune(*rl.line)[rl.cursor.Pos():]
	rl.buffers.Write(cut...)

	rl.line.Cut(rl.cursor.Pos(), rl.line.Len())
}

func (rl *Shell) backwardKillLine() {
	rl.iterations.Reset()
	rl.undo.Save(*rl.line, *rl.cursor)

	cut := []rune(*rl.line)[:rl.cursor.Pos()]
	rl.buffers.Write(cut...)

	rl.line.Cut(0, rl.cursor.Pos())
}

func (rl *Shell) killWholeLine() {
	rl.undo.Save(*rl.line, *rl.cursor)

	if rl.line.Len() == 0 {
		return
	}

	rl.buffers.Write(*rl.line...)
	rl.line.Cut(0, rl.line.Len())
}

func (rl *Shell) killBuffer() {
	rl.undo.Save(*rl.line, *rl.cursor)

	if rl.line.Len() == 0 {
		return
	}

	rl.buffers.Write(*rl.line...)
	rl.line.Cut(0, rl.line.Len())
}

func (rl *Shell) killWord() {
	rl.undo.Save(*rl.line, *rl.cursor)

	bpos := rl.cursor.Pos()

	rl.cursor.ToFirstNonSpace(true)
	forward := rl.line.Forward(rl.line.TokenizeSpace, rl.cursor.Pos())
	rl.cursor.Move(forward - 1)
	epos := rl.cursor.Pos()

	rl.selection.MarkRange(bpos, epos)
	rl.buffers.Write([]rune(rl.selection.Cut())...)
	rl.cursor.Set(bpos)
}

func (rl *Shell) backwardKillWord() {
	rl.undo.Save(*rl.line, *rl.cursor)
	rl.undo.SkipSave()

	rl.selection.Mark(rl.cursor.Pos())
	adjust := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(adjust)

	rl.buffers.Write([]rune(rl.selection.Cut())...)
}

func (rl *Shell) shellKillWord() {
	startPos := rl.cursor.Pos()

	// select the shell word, and if the cursor position
	// has changed, we delete the part after the initial one.
	rl.cursor.ToFirstNonSpace(true)
	rl.viSelectAShellWord()

	_, epos := rl.selection.Pos()

	rl.buffers.Write([]rune((*rl.line)[startPos:epos])...)
	rl.line.Cut(startPos, epos)
	rl.cursor.Set(startPos)

	rl.selection.Reset()
}

// TODO: Fix not catching the word when only one in line cursor at end of line.
func (rl *Shell) shellBackwardKillWord() {
	startPos := rl.cursor.Pos()

	rl.cursor.ToFirstNonSpace(false)
	backward := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(backward)

	rl.viSelectAShellWord()
	bpos, _ := rl.selection.Pos()

	rl.buffers.Write([]rune((*rl.line)[bpos:startPos])...)
	rl.line.Cut(bpos, startPos)
	rl.cursor.Set(bpos)

	rl.selection.Reset()
}

func (rl *Shell) killRegion() {
	rl.undo.Save(*rl.line, *rl.cursor)

	if !rl.selection.Active() {
		return
	}

	rl.buffers.Write([]rune(rl.selection.Cut())...)
}

func (rl *Shell) copyRegionAsKill() {
	rl.undo.SkipSave()

	if !rl.selection.Active() {
		return
	}

	rl.buffers.Write([]rune(rl.selection.Text())...)
	rl.selection.Reset()
}

func (rl *Shell) copyBackwardWord() {
	rl.undo.Save(*rl.line, *rl.cursor)

	rl.selection.Mark(rl.cursor.Pos())
	adjust := rl.line.Backward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(adjust)

	rl.buffers.Write([]rune(rl.selection.Text())...)
	rl.selection.Reset()
}

func (rl *Shell) copyForwardWord() {
	rl.undo.Save(*rl.line, *rl.cursor)

	rl.selection.Mark(rl.cursor.Pos())
	adjust := rl.line.Forward(rl.line.Tokenize, rl.cursor.Pos())
	rl.cursor.Move(adjust + 1)

	rl.buffers.Write([]rune(rl.selection.Text())...)
	rl.selection.Reset()
}

func (rl *Shell) yank() {
	buf := rl.buffers.Get(rune(0))
	rl.line.Insert(rl.cursor.Pos(), buf...)
	rl.cursor.Move(len(buf))
}

func (rl *Shell) yankPop() {
	buf := rl.buffers.Pop()
	rl.line.Insert(rl.cursor.Pos(), buf...)
	rl.cursor.Move(len(buf))
}

func (rl *Shell) copyPrevShellWord() {
	rl.undo.Save(*rl.line, *rl.cursor)

	posInit := rl.cursor.Pos()

	// First go back to the beginning of the current word,
	// then go back again to the beginning of the previous.
	rl.backwardShellWord()
	rl.backwardShellWord()

	// Select the current shell word
	rl.viSelectAShellWord()

	word := rl.selection.Text()

	// Replace the cursor before reassembling the line.
	rl.cursor.Set(posInit)
	rl.selection.InsertAt(rl.cursor.Pos(), -1)
	rl.cursor.Move(len(word))
}

//
// Numeric Arguments -----------------------------------------------------------
//

// digitArgument is used both in Emacs and Vim modes,
// but strips the Alt modifier used in Emacs mode.
func (rl *Shell) digitArgument() {
	rl.undo.SkipSave()

	// If we were called in the middle of a pending
	// operation, we should not yet trigger the caller.
	// This boolean is recomputed at the next key read:
	// This just postpones running the caller a little.
	// rl.isViopp = false

	keys, empty := rl.keys.PeekAll()
	if empty {
		return
	}

	rl.iterations.Add(string(keys))
}

//
// Macros ----------------------------------------------------------------------
//

func (rl *Shell) startKeyboardMacro() {
	rl.macros.StartRecord()
}

func (rl *Shell) endKeyboardMacro() {
	rl.macros.StopRecord()
}

func (rl *Shell) callLastKeyboardMacro() {
	rl.macros.RunLastMacro()
}

func (rl *Shell) printLastKeyboardMacro() {
	rl.display.ClearHelpers()

	rl.macros.PrintLastMacro()

	rl.prompt.PrimaryPrint()
	rl.display.Refresh()
}

//
// Miscellaneous ---------------------------------------------------------------
//

func (rl *Shell) reReadInitFile() {
	config := filepath.Join(os.Getenv("HOME"), ".inputrc")

	err := inputrc.ParseFile(config, rl.opts)
	if err != nil {
		rl.hint.Set(color.FgRed + "Inputrc reload error: " + err.Error())
	} else {
		rl.hint.Set(color.FgGreen + "Inputrc reloaded: " + config)
	}
}

func (rl *Shell) abort() {}

func (rl *Shell) prefixMeta() {}

func (rl *Shell) undoLast() {
	rl.undo.Undo(rl.line, rl.cursor)
}

func (rl *Shell) revertLine() {}

func (rl *Shell) setMark() {
}

func (rl *Shell) exchangePointAndMark() {
}

func (rl *Shell) characterSearch()         {}
func (rl *Shell) characterSearchBackward() {}
func (rl *Shell) insertComment()           {}

func (rl *Shell) dumpFunctions() {
	rl.display.ClearHelpers()
	fmt.Println()

	defer func() {
		rl.prompt.PrimaryPrint()
		rl.display.Refresh()
	}()

	inputrcFormat := rl.iterations.Get() != 1
	rl.keymaps.PrintBinds(inputrcFormat)
}

func (rl *Shell) dumpVariables() {
	rl.display.ClearHelpers()
	fmt.Println()

	defer func() {
		rl.prompt.PrimaryPrint()
		rl.display.Refresh()
	}()

	// Get all variables and their values, alphabetically sorted.
	var variables []string

	for variable := range rl.opts.Vars {
		variables = append(variables, variable)
	}

	sort.Strings(variables)

	// Either print in inputrc format, or wordly one.
	if rl.iterations.Get() != 1 {
		for _, variable := range variables {
			value := rl.opts.Vars[variable]
			fmt.Printf("set %s %v\n", variable, value)
		}
	} else {
		for _, variable := range variables {
			value := rl.opts.Vars[variable]
			fmt.Printf("%s is set to `%v'\n", variable, value)
		}
	}
}

func (rl *Shell) dumpMacros() {
	rl.display.ClearHelpers()
	fmt.Println()

	defer func() {
		rl.prompt.PrimaryPrint()
		rl.display.Refresh()
	}()

	// We print the macros bound to the current keymap only.
	binds := rl.opts.Binds[string(rl.keymaps.Main())]
	if len(binds) == 0 {
		return
	}

	var macroBinds []string

	for keys, bind := range binds {
		if bind.Macro {
			macroBinds = append(macroBinds, inputrc.Escape(keys))
		}
	}

	sort.Strings(macroBinds)

	if rl.iterations.Get() != 1 {
		for _, key := range macroBinds {
			action := inputrc.Escape(binds[inputrc.Unescape(key)].Action)
			fmt.Printf("\"%s\": \"%s\"\n", key, action)
		}
	} else {
		for _, key := range macroBinds {
			action := inputrc.Escape(binds[inputrc.Unescape(key)].Action)
			fmt.Printf("%s outputs %s\n", key, action)
		}
	}
}

func (rl *Shell) magicSpace()            {}
func (rl *Shell) editAndExecuteCommand() {}
func (rl *Shell) editCommandLine()       {}

func (rl *Shell) redo() {
	rl.undo.Redo(rl.line, rl.cursor)
}

// func (rl *Shell) setMarkCommand() {
// 	rl.undo.SkipSave()
//
// 	vii := rl.iterations.Get()
// 	switch {
// 	case vii < 0:
// 		rl.resetSelection()
// 		rl.visualLine = false
// 	default:
// 		rl.markSelection(rl.pos)
// 	}
// }.
//
// func (rl *Shell) copyPrevShellWord() {
// 	rl.undo.Save(*rl.line, *rl.cursor)
//
// 	posInit := rl.pos
//
// 	// First go back a single blank word
// 	rl.moveCursorByAdjust(rl.viJumpB(tokeniseSplitSpaces))
//
// 	// Now try to find enclosing quotes from here.
// 	sBpos, sEpos, _, _ := rl.searchSurround('\'')
// 	dBpos, dEpos, _, _ := rl.searchSurround('"')
//
// 	mark, cpos := adjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos)
// 	if mark == -1 && cpos == -1 {
// 		rl.markSelection(rl.pos)
// 		rl.moveCursorByAdjust(rl.viJumpE(tokeniseSplitSpaces))
// 	} else {
// 		rl.markSelection(mark)
// 		rl.pos = cpos
// 	}
//
// 	word, _, _, _ := rl.popSelection()
//
// 	// Replace the cursor before reassembling the line.
// 	rl.pos = posInit
//
// 	rl.insertBlock(rl.pos, rl.pos, word, "")
// 	rl.pos += len(word)
// }
//
// func (rl *Shell) exchangePointAndMark() {
// 	rl.undo.SkipSave()
// 	vii := rl.iterations.Get()
//
// 	visual := rl.visualSelection()
// 	if visual == nil {
// 		return
// 	}
//
// 	switch {
// 	case vii < 0:
// 		rl.pos, visual.bpos = visual.bpos, rl.pos
// 	case vii > 0:
// 		rl.pos, visual.bpos = visual.bpos, rl.pos
// 		visual.active = true
// 	case vii == 0:
// 		visual.active = true
// 	}
// }
//
// func (rl *Shell) editCommandLine() {
// 	rl.clearHelpers()
//
// 	buffer := rl.line
//
// 	edited, err := editor.EditBuffer(buffer, "", "")
// 	// edited, err := rl.StartEditorWithBuffer(buffer, "")
// 	if err != nil || (len(edited) == 0 && len(buffer) != 0) {
// 		rl.undo.SkipSave()
// 		errStr := strings.ReplaceAll(err.Error(), "\n", "")
// 		changeHint := fmt.Sprintf(seqFgRed+"Editor error: %s", errStr)
// 		rl.hint = append([]rune{}, []rune(changeHint)...)
// 		return
// 	}
//
// 	// Update our line
// 	rl.line = edited
//
// 	// We're done with visual mode when we were in.
// 	if (rl.main == vicmdC || rl.main == viinsC) && rl.local == visualC {
// 		rl.exitVisualMode()
// 	}
// }
