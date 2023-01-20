package common

// Selection contains all regions of an input line that are currently selected/marked
// with either a begin and/or end position. The main selection is the visual one, used
// with the default cursor mark and position, and contains a list of additional surround
// selections used to change/select multiple parts of the line at once.
type Selection struct {
	stype    string
	active   bool
	bpos     int
	epos     int
	fg       string
	bg       string
	surround []Selection

	line   *Line
	cursor *Cursor
}

// Mark starts a pending selection at the specified position in the line.
// If the position is out of the line bounds, no selection is started.
// If this function is called on a surround selection, nothing happens.
func (s *Selection) Mark(pos int) {}

// MarkRange starts a selection as a range in the input line. If either of
// the begin or end positions are out of bounds, no selection is started.
func (s *Selection) MarkRange(bpos, epos int) {}

// MarkSurround starts a selection like MarkRange, but stores the created
// selection as a surround one. Any number of surround selections can be created.
func (s *Selection) MarkSurround(bpos, epos int) {}

// Active return true if the selection is active.
// When created, all selections are marked active,
// so that visual modes in Vim can work properly.
func (s *Selection) Active() bool { return false }

// Pos returns the begin and end positions of the selection.
// If any of these is not set, it is set to the cursor position.
// This is generally the case with "pending" visual selections.
func (s *Selection) Pos() (bpos, epos int) { return 0, 0 }

// Cursor returns what should be the cursor position if the active
// selection is to be deleted.
// Note that this applies quite well to yank operations as well.
func (s *Selection) Cursor() {}

// Pop returns the contents of the current selection as a string, as well as its
// begin and end position in the line, and the cursor position as given by the
// Cursor() method. Then, the selection is reset (deleted).
func (s *Selection) Pop() (buf string, bpos, epos, cpos int) { return "", 0, 0, 0 }

// InsertAt insert the contents of the selection into the line, between the
// begin and end position, effectively deleting everything in between those.
//
// If either or these positions is equal to -1, the selection content
// is inserted at the other position. If both are -1, nothing is done.
// This is equivalent to selection.Pop(), and line.InsertAt() combined.
func (s *Selection) InsertAt(bpos, epos int) {}

// Surround surrounds the selection with a begin and end character,
// effectively inserting those characters into the current input line.
// After insertion, the selection is reset.
func (s *Selection) Surround(bchar, echar rune) {}

// ReplaceWith replaces all characters of the line within the current
// selection range by applying to each rune the provided replacer function.
func (s *Selection) ReplaceWith(replacer func(r rune) rune) {}

// Cut deletes the current selection from the line, updates the cursor position
// and returns the deleted content, which can then be passed to the shell registers.
// After that, the selection is reset.
func (s *Selection) Cut() {}

// Yank gets the current selection and returns it.
// The selection is not reset after this.
func (s *Selection) Yank() (buf string) { return "" }

// Reset makes the current selection inactive, resetting all of its values.
func (s *Selection) Reset() {}
