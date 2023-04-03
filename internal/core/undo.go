package core

// LineHistory contains all the history modifications
// for the current line, and manages all undo/redo actions.
type LineHistory struct {
	pos     int
	skip    bool
	undoing bool
	items   []undoItem
	lineBuf string
	linePos int
}

type undoItem struct {
	line string
	pos  int
}

// Save saves the current line and cursor position as an undo state item.
// If this was called while the shell in the middle of its undo history
// (eg. the caller has undone one or more times), all undone steps are dropped.
func (lh *LineHistory) Save(line Line, cursor Cursor) {
	defer lh.Reset()

	if lh.skip {
		return
	}

	// When the line is identical to the previous undo, we just update
	// the cursor position if it's a different one.
	if len(lh.items) > 0 && lh.items[len(lh.items)-1].line == string(line) {
		lh.items[len(lh.items)-1].pos = cursor.Pos()
		return
	}

	// When we add an item to the undo history, the history
	// is cut from the current undo hist position onwards.
	if lh.pos > len(lh.items) {
		lh.pos = len(lh.items)
	}
	lh.items = lh.items[:len(lh.items)-lh.pos]

	// Make a copy of the cursor and ensure its position.
	cur := NewCursor(&line)
	cur.Set(cursor.Pos())
	cur.CheckCommand()

	// And save the item.
	lh.items = append(lh.items, undoItem{
		line: string(line),
		pos:  cur.Pos(),
	})
}

// SkipSave will not save the current line when the target command is done.
func (lh *LineHistory) SkipSave() {
	lh.skip = true
}

// Undo restores the line and cursor position to their last saved state.
func (lh *LineHistory) Undo(line *Line, cursor *Cursor) {
	lh.skip = true
	lh.undoing = true

	if len(lh.items) == 0 {
		return
	}

	// Keep the current line buffer
	if lh.pos == 0 {
		lh.lineBuf = string(*line)
		lh.linePos = cursor.Pos()
	}

	var undo undoItem

	// When undoing, we loop through preceding undo items
	// as long as they are identical to the current line.
	for {
		lh.pos++

		// Exit if we reached the end.
		if lh.pos > len(lh.items) {
			lh.pos = len(lh.items)
			return
		}

		// Break as soon as we find a non-matching line.
		undo = lh.items[len(lh.items)-lh.pos]
		if undo.line != string(*line) {
			break
		}
	}

	// Use the undo we found
	line.Set([]rune(undo.line)...)
	cursor.Set(undo.pos)
}

// Redo cancels an undo action if any has been made, or if
// at the begin of the undo history, restores the original
// line's contents as their were before starting undoing.
func (lh *LineHistory) Redo(line *Line, cursor *Cursor) {
	lh.skip = true
	lh.undoing = true

	if len(lh.items) == 0 {
		return
	}

	lh.pos--

	if lh.pos < 1 {
		lh.pos = 0
		line.Set([]rune(lh.lineBuf)...)
		cursor.Set(lh.linePos)

		return
	}

	undo := lh.items[len(lh.items)-lh.pos]
	line.Set([]rune(undo.line)...)
	cursor.Set(undo.pos)
}

// Pos returns the current position in the undo history, which is
// equal to its length minus the number of previous undo calls.
func (lh *LineHistory) Pos() int {
	return lh.pos
}

// Reset will reset the current position in the list
// of undo items, but will not delete any of them.
func (lh *LineHistory) Reset() {
	lh.skip = false

	if !lh.undoing {
		lh.pos = 0
	}

	lh.undoing = false
}
