package history

import (
	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/core"
)

// lineHistory contains all state changes for a given input line,
// whether it is the current input line or one of the history ones.
type lineHistory struct {
	pos     int
	lineBuf string
	linePos int
	items   []undoItem
}

type undoItem struct {
	line string
	pos  int
}

// Save saves the current line and cursor position as an undo state item.
// If this was called while the shell in the middle of its undo history
// (eg. the caller has undone one or more times), all undone steps are dropped.
func (h *Sources) Save() {
	defer h.Reset()

	if h.skip {
		return
	}

	// Get the undo states for the current line.
	lh := h.getLineHistory()
	if lh == nil {
		return
	}

	// When the line is identical to the previous undo, we just update
	// the cursor position if it's a different one.
	if len(lh.items) > 0 && lh.items[len(lh.items)-1].line == string(*h.line) {
		lh.items[len(lh.items)-1].pos = h.cursor.Pos()
		return
	}

	// When we add an item to the undo history, the history
	// is cut from the current undo hist position onwards.
	if lh.pos > len(lh.items) {
		lh.pos = len(lh.items)
	}

	lh.items = lh.items[:len(lh.items)-lh.pos]

	// Make a copy of the cursor and ensure its position.
	cur := core.NewCursor(h.line)
	cur.Set(h.cursor.Pos())
	cur.CheckCommand()

	// And save the item.
	lh.items = append(lh.items, undoItem{
		line: string(*h.line),
		pos:  cur.Pos(),
	})
}

// SkipSave will not save the current line when the target command is done.
func (h *Sources) SkipSave() {
	h.skip = true
}

// SaveWithCommand is only meant to be called in the main readline loop of the shell,
// and not from within commands themselves: it does the same job as Save(), but also
// keeps the command that has just been executed.
func (h *Sources) SaveWithCommand(bind inputrc.Bind) {
	h.last = bind
	h.Save()
}

// Undo restores the line and cursor position to their last saved state.
func (h *Sources) Undo() {
	h.skip = true
	h.undoing = true

	// Get the undo states for the current line.
	lh := h.getLineHistory()
	if lh == nil || len(lh.items) == 0 {
		return
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
		if undo.line != string(*h.line) {
			break
		}
	}

	// Use the undo we found
	h.line.Set([]rune(undo.line)...)
	h.cursor.Set(undo.pos)
}

// Revert goes back to the initial state of the line, which is what it was
// like when the shell started reading user input. Note that this state might
// be a line that was inferred, accept-and-held from the previous readline run.
func (h *Sources) Revert() {
	lh := h.getLineHistory()
	if lh == nil || len(lh.items) == 0 {
		return
	}

	// Reuse the first saved state.
	undo := lh.items[0]

	h.line.Set([]rune(undo.line)...)
	h.cursor.Set(undo.pos)

	// And reset everything
	lh.items = make([]undoItem, 0)

	h.Reset()
}

// Redo cancels an undo action if any has been made, or if
// at the begin of the undo history, restores the original
// line's contents as their were before starting undoing.
func (h *Sources) Redo() {
	h.skip = true
	h.undoing = true

	lh := h.getLineHistory()
	if lh == nil || len(lh.items) == 0 {
		return
	}

	lh.pos--

	if lh.pos < 1 {
		lh.pos = 0
		h.line.Set([]rune(lh.lineBuf)...)
		h.cursor.Set(lh.linePos)

		return
	}

	undo := lh.items[len(lh.items)-lh.pos]
	h.line.Set([]rune(undo.line)...)
	h.cursor.Set(undo.pos)
}

// Last returns the last command ran by the shell.
func (h *Sources) Last() inputrc.Bind {
	return h.last
}

// Pos returns the current position in the undo history, which is
// equal to its length minus the number of previous undo calls.
func (h *Sources) Pos() int {
	lh := h.getLineHistory()
	if lh == nil {
		return 0
	}

	return lh.pos
}

// Reset will reset the current position in the list
// of undo items, but will not delete any of them.
func (h *Sources) Reset() {
	h.skip = false

	lh := h.getLineHistory()
	if lh == nil {
		return
	}

	if !h.undoing {
		lh.pos = 0
	}

	h.undoing = false
}

// Always returns a non-nil map, whether or not a history source is found.
func (h *Sources) getHistoryLineChanges() map[int]*lineHistory {
	history := h.Current()
	if history == nil {
		return map[int]*lineHistory{}
	}

	// Get the state changes of all history lines
	// for the current history source.
	source := h.names[h.sourcePos]

	hist := h.lines[source]
	if hist == nil {
		h.lines[source] = make(map[int]*lineHistory)
		hist = h.lines[source]
	}

	return hist
}

func (h *Sources) getLineHistory() *lineHistory {
	hist := h.getHistoryLineChanges()
	if hist == nil {
		return &lineHistory{}
	}

	if hist[h.hpos] == nil {
		hist[h.hpos] = &lineHistory{}
	}

	// Return the state changes of the current line.
	return hist[h.hpos]
}

func (h *Sources) restoreLineBuffer() {
	hist := h.getHistoryLineChanges()
	if hist == nil {
		return
	}

	// Get the undo states for the line buffer
	// (the last one, not any of the history ones)
	lh := hist[0]
	if lh == nil || len(lh.items) == 0 {
		return
	}

	undo := lh.items[len(lh.items)-1]

	// Restore the line to the last known state.
	h.line.Set([]rune(undo.line)...)
	h.cursor.Set(undo.pos)
}
