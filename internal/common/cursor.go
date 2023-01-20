package common

// Cursor is the cursor position in the current line buffer.
// Contains methods to set, move, describe and check itself.
type Cursor struct {
	pos  int
	mark int
	line *Line
}

// Set sets the position of the cursor to an absolute value.
// If either negative or greater than the length of the line,
// the cursor will be set to either 0, or the length of the line.
func (c *Cursor) Set(pos int) {}

// Pos returns the current cursor position.
// This function cannot return an invalid cursor position: it cannot be negative, nor it
// can be greater than the length of the line (note it still can be out of line by -1).
func (c *Cursor) Pos() int { return 0 }

// Inc increments the cursor position by 1, if its not at the end of the line.
func (c *Cursor) Inc() {}

// Dec decrements the cursor position by 1, if its not at the beginning of the line.
func (c *Cursor) Dec() {}

// Move moves the cursor position by a relative value. If the end result is negative,
// the cursor is set to 0. If longer than the line, the cursor is set to length of line.
func (c *Cursor) Move(offset int) {}

// BeginningOfLine moves the cursor to the beginning of the current line,
// (marked by a newline) or if no newline found, to the beginning of the buffer.
func (c *Cursor) BeginningOfLine() {}

// EndOfLine moves the cursor to the end of the current line, (marked by
// a newline) or if no newline found, to the position of the last character.
func (c *Cursor) EndOfLine() {}

// EndOfLineAppend moves the cursor to the very end of the line,
// that is, equal to len(Line), as in when appending in insert mode.
func (c *Cursor) EndOfLineAppend() {
}

// SetMark sets the current cursor position as the mark.
func (c *Cursor) SetMark() {}

// Mark returns the current mark value of the cursor, or -1 if not set.
func (c *Cursor) Mark() int { return 0 }

// Line returns the index of the current line on which the cursor is.
// A line is defined as a sequence of runes between one or two newline
// characters, between end and/or beginning of buffer, or a mix of both.
func (c *Cursor) Line() int { return 0 }

// LineMove moves the cursor by n lines either up (if the value is negative),
// or down (if positive). If greater than the length of possible lines above/below,
// the cursor will be set to either the first, or the last line of the buffer.
func (c *Cursor) LineMove(lines int) {}

// OnEmptyLine returns true if the rune under the current cursor position is a newline
// and that the preceding rune in the line is also a newline, or returns false.
func (c *Cursor) OnEmptyLine() bool { return false }

// Check verifies that the current cursor position is neither negative,
// nor greater than the length of the input line. If either is true, the
// cursor will set its value as either 0, or the length of the line.
func (c *Cursor) Check() {}

// Used returns the number of real terminal lines above the cursor position (y value),
// and the number of columns since the beginning of the current line (x value).
func (c *Cursor) Used() (x, y int) { return }
