package common

// LineHistory contains all the history modifications
// for the current line, and manages all undo/redo actions.
type LineHistory struct {
	pos     int
	undoing bool
	items   []undoItem
}

type undoItem struct {
	line string
	pos  int
}

// Save saves the current line as an undo state item.
func (lh *LineHistory) Save() {
}

// SkipSave will not save the current line when the target command is done.
func (lh *LineHistory) SkipSave() {
}

func (lh *LineHistory) Undo() {
}

func (lh *LineHistory) Redo() {
}

// Reset will reset the current position in the list
// of undo items, but will not delete any of them.
func (lh *LineHistory) Reset() {
}
