package common

// Tokenizer is a method used by a (line) type to split itself according to
// different rules (split between spaces, punctuation, brackets, quotes, etc.).
type Tokenizer func(cursorPos int) (split []string, index int, newPos int)

// Line is an input line buffer.
// Contains methods to search and modify its contents,
// split itself with tokenizers, and displaying itself.
type Line []rune

// Insert inserts one or more runes at the given position.
// If the position is either negative or greater than the
// length of the line, nothing is inserted.
func (l *Line) Insert(pos int, r ...rune) {}

// InsertAt inserts one or more runes into the line, between the specified
// begin and end position, effectively deleting everything in between those.
// If either or these positions is equal to -1, the selection content
// is inserted at the other position. If both are -1, nothing is done.
func (l *Line) InsertBetween(bpos, epos int, r ...rune) {}

// Cut deletes a slice of runes between a beginning and end position on the line.
// If the begin/end pos is negative/greater than the line, all runes located on
// valid indexes in the given range are removed.
func (l *Line) Cut(bpos, epos int) {}

// CutRune deletes a rune at the given position in the line.
// If the position is out of bounds, nothing is deleted.
func (l *Line) CutRune(pos int) {}

// Len returns the length of the line.
func (l *Line) Len() int { return 0 }

// SelectWord returns the full non-blank word around the specified position.
func (l *Line) SelectWord(pos int) (bpos, epos int) { return 0, 0 }

// Next returns the offset to a rune, searching forward, relative to the pos passed as parameter.
// If the target rune is not found, the adjust will be computed relatively to the end of the line.
func (l *Line) Next(r rune, pos int) (adjust int) { return }

// Prev returns the position to a rune, searching backward, relative to the pos passed as parameter.
// If the target rune is not found, the adjust will be computed relatively to the beginning of the line.
func (l *Line) Prev(r rune, pos int) (adjust int) { return }

// Forward returns the offset to the beginning of the next
// (forward) token determined by the tokenizer function.
func (l *Line) Forward(split Tokenizer) (adjust int) { return }

// ForwardEnd returns the offset to the end of the next
// (forward) token determined by the tokenizer function.
func (l *Line) ForwardEnd(split Tokenizer) (adjust int) { return }

// Backward returns the offset to the beginning position of the previous
// (backward) token determined by the tokenizer function.
func (l *Line) Backward(split Tokenizer) (adjust int) { return }

// Tokenize splits the line on each word, that is, split on every punctuation or space.
func (l *Line) Tokenize(pos int) ([]string, int, int) { return nil, 0, 0 }

// Tokenize splits the line on each WORD (blank word), that is, split on every space.
func (l *Line) TokenizeSpace(pos int) ([]string, int, int) { return nil, 0, 0 }

// TokenizeBlock splits the line into arguments delimited either by
// brackets, braces and parenthesis, and/or single and double quotes.
func (l *Line) TokenizeBlock(pos int) ([]string, int, int) { return nil, 0, 0 }

// Display prints the line to stdout, starting at the current terminal
// cursor position, assuming it is at the end of the shell prompt string.
// Params:
// @indent -    Used to align all lines (except the first) together on a single column.
// @suggested - An optional string to append to the line, for things like command autosuggestion.
func (l *Line) Display(indent int, suggested string) {}

// Used returns the number of real terminal lines on which the input line spans, considering
// any contained newlines, any overflowing line, and the indent passed as parameter. The values
// also take into account an eventual suggestion added to the line before printing.
// Params:
// @indent -    Used to align all lines (except the first) together on a single column.
// @suggested - An optional string to append to the line, for things like command autosuggestion.
// Returns:
// @x - The number of columns, starting from the terminal left, to the end of the last line.
// @y - The number of actual lines on which the line spans.
func (l *Line) Used(indent int, suggested string) (x, y int) { return 0, 0 }
