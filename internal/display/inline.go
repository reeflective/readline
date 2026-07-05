package display

import (
	"strings"
	"unicode"

	"github.com/reeflective/readline/internal/core"
)

// AcceptInline replaces the line buffer with the inline suggestion when one
// applies at the cursor, clears the suggestion, and reports whether it did so.
// Callers pass the line/cursor to accept onto, since the engine's own line may
// point at a minibuffer during completion or search.
func (e *Engine) AcceptInline(line *core.Line, cursor *core.Cursor) bool {
	if !e.inlineAcceptable(line, cursor) {
		return false
	}

	line.Set([]rune(e.inline)...)
	cursor.Set(line.Len())
	e.inline = ""

	return true
}

// AcceptInlineWord accepts the next word of the inline suggestion, leaving the
// rest as a suggestion. It follows forward-word semantics: any leading
// whitespace plus the following run of non-whitespace runes are inserted.
func (e *Engine) AcceptInlineWord(line *core.Line, cursor *core.Cursor) {
	if !e.inlineAcceptable(line, cursor) {
		return
	}

	sug := []rune(e.inline)
	from := line.Len()
	remainder := sug[from:]

	pos := 0
	for pos < len(remainder) && unicode.IsSpace(remainder[pos]) {
		pos++
	}

	for pos < len(remainder) && !unicode.IsSpace(remainder[pos]) {
		pos++
	}

	line.Set(sug[:from+pos]...)
	cursor.Set(line.Len())

	if from+pos >= len(sug) {
		e.inline = ""
	}
}

// inlineAcceptable reports whether the inline suggestion can be accepted onto
// the given line: the cursor must be at the end of the line and the suggestion
// must extend it (mirrors the display gate in inlineSuggestionApplies).
func (e *Engine) inlineAcceptable(line *core.Line, cursor *core.Cursor) bool {
	return e.inline != "" &&
		cursor.Pos() == line.Len() &&
		strings.HasPrefix(e.inline, string(*line)) &&
		len([]rune(e.inline)) > line.Len()
}
