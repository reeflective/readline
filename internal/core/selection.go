package core

import (
	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/strutil"
	"github.com/reeflective/readline/inputrc"
)

// Selection contains all regions of an input line that are currently selected/marked
// with either a begin and/or end position. The main selection is the visual one, used
// with the default cursor mark and position, and contains a list of additional surround
// selections used to change/select multiple parts of the line at once.
type Selection struct {
	stype      string
	active     bool
	visual     bool
	visualLine bool
	bpos       int
	epos       int
	fg         string
	bg         string
	surrounds  []Selection

	line   *Line
	cursor *Cursor
}

// NewSelection is a required constructor to use for initializing
// a selection, as some numeric values must be negative by default.
func NewSelection(line *Line, cursor *Cursor) *Selection {
	return &Selection{
		bpos:   -1,
		epos:   -1,
		line:   line,
		cursor: cursor,
	}
}

// Mark starts a pending selection at the specified position in the line.
// If the position is out of the line bounds, no selection is started.
// If this function is called on a surround selection, nothing happens.
func (s *Selection) Mark(pos int) {
	if pos < 0 || pos > s.line.Len() {
		return
	}

	s.MarkRange(pos, -1)
}

// MarkRange starts a selection as a range in the input line. If either of
// begin/end pos are negative, it is replaced with the current cursor position.
// Any out of range positive value is replaced by the length of the line.
func (s *Selection) MarkRange(bpos, epos int) {
	bpos, epos, valid := s.checkRange(bpos, epos)
	if !valid {
		return
	}

	s.stype = "visual"
	s.active = true
	s.bpos = bpos
	s.epos = epos
	s.bg = color.BgBlue
}

// MarkSurroundA creates two distinct selections each containing one rune.
// The first area starts at bpos, and the second one at epos. If either bpos
// is negative or epos is > line.Len()-1, no selection is created.
func (s *Selection) MarkSurround(bpos, epos int) {
	if bpos < 0 || epos > s.line.Len() {
		return
	}

	for _, pos := range []int{bpos, epos} {
		s.surrounds = append(s.surrounds, Selection{
			stype:  "surround",
			active: true,
			visual: true,
			bpos:   pos,
			epos:   pos,
			bg:     color.BgRed,
			line:   s.line,
			cursor: s.cursor,
		})
	}
}

// Active return true if the selection is active.
// When created, all selections are marked active,
// so that visual modes in Vim can work properly.
func (s *Selection) Active() bool {
	return s.active
}

// Visual sets the selection as a visual one (highlighted),
// which is core.y seen in Vim.
// If line is true, the visual is extended to entire lines.
func (s *Selection) Visual(line bool) {
	s.visual = true
	s.visualLine = line
}

// IsVisual indicates whether the selection should be highlighted.
func (s *Selection) IsVisual() bool {
	return s.visual
}

// Pos returns the begin and end positions of the selection.
// If any of these is not set, it is set to the cursor position.
// This is generally the case with "pending" visual selections.
func (s *Selection) Pos() (bpos, epos int) {
	if s.line.Len() == 0 || !s.active {
		return -1, -1
	}

	bpos, epos, valid := s.checkRange(s.bpos, s.epos)
	if !valid {
		return
	}

	// Use currently set values, or update if one is pending.
	bpos, epos = s.bpos, s.epos

	if epos == -1 {
		bpos, epos = s.selectToCursor(bpos)
	}

	if s.visual {
		epos++
	}

	// Always check that neither the initial values
	// nor the ones that we might have updated are wrong.
	bpos, epos, valid = s.checkRange(bpos, epos)
	if !valid {
		return -1, -1
	}

	return bpos, epos
}

// Cursor returns what should be the cursor position if the active
// selection is to be deleted, but also works for yank operations.
func (s *Selection) Cursor() int {
	bpos, epos := s.Pos()
	if bpos == -1 && epos == -1 {
		return s.cursor.Pos()
	}

	cpos := bpos

	if !s.visual || !s.visualLine {
		return cpos
	}

	var indent int
	pos := s.cursor.Pos()

	// Get the indent of the cursor line.
	for cpos = pos - 1; cpos >= 0; cpos-- {
		if (*s.line)[cpos] == '\n' {
			break
		}
	}

	indent = pos - cpos - 1

	// If the selection includes the last line,
	// the cursor will move up the above line.
	var hpos, rpos int

	if epos < s.line.Len() {
		hpos = epos + 1
		rpos = bpos
	} else {
		for hpos = bpos - 2; hpos >= 0; hpos-- {
			if (*s.line)[hpos] == '\n' {
				break
			}
		}
		if hpos < -1 {
			hpos = -1
		}
		hpos++
		rpos = hpos
	}

	// Now calculate the cursor position, the indent
	// must be less than the line characters.
	for cpos = hpos; cpos < s.line.Len(); cpos++ {
		if (*s.line)[cpos] == '\n' {
			break
		}

		if hpos+indent <= cpos {
			break
		}
	}

	// That cursor position might be bigger than the line itself:
	// it should be controlled when the line is redisplayed.
	cpos = rpos + cpos - hpos

	return cpos
}

// Len returns the length of the current selection. If any
// of begin/end pos is not set, the cursor position is used.
func (s *Selection) Len() int {
	if s.line.Len() == 0 || (s.bpos == s.epos) {
		return 0
	}

	bpos, epos := s.Pos()
	buf := (*s.line)[bpos:epos]

	return len(buf)
}

// Text returns the current selection as a string, but does not reset it.
func (s *Selection) Text() string {
	if s.line.Len() == 0 {
		return ""
	}

	bpos, epos := s.Pos()
	if bpos == -1 || epos == -1 {
		return ""
	}

	return string((*s.line)[bpos:epos])
}

// Pop returns the contents of the current selection as a string,
// as well as its begin and end position in the line, and the cursor
// position as given by the Cursor() method. Then, the selection is reset.
func (s *Selection) Pop() (buf string, bpos, epos, cpos int) {
	if s.line.Len() == 0 {
		return "", -1, -1, 0
	}

	defer s.Reset()

	bpos, epos = s.Pos()
	if bpos == -1 || epos == -1 {
		return "", -1, -1, 0
	}

	// End position is increased by one so
	// that we capture the entire selection.
	epos++

	cpos = s.Cursor()
	buf = string((*s.line)[bpos:epos])

	return buf, bpos, epos, cpos
}

// InsertAt insert the contents of the selection into the line, between the
// begin and end position, effectively deleting everything in between those.
//
// If either or these positions is equal to -1, the selection content
// is inserted at the other position. If both are negative, nothing is done.
// This is equivalent to selection.Pop(), and line.InsertAt() combined.
//
// After insertion, the selection is reset.
func (s *Selection) InsertAt(bpos, epos int) {
	bpos, epos, valid := s.checkRange(bpos, epos)
	if !valid {
		return
	}

	defer s.Reset()

	// Get and reset the selection.
	buf := s.Text()
	if len(buf) == 0 {
		return
	}

	switch {
	case bpos == -1:
		s.line.Insert(epos, []rune(buf)...)
	case epos == -1, bpos == epos:
		s.line.Insert(bpos, []rune(buf)...)
	default:
		s.line.InsertBetween(bpos, epos, []rune(buf)...)
	}
}

// Surround surrounds the selection with a begin and end character,
// effectively inserting those characters into the current input line.
// If epos is greater than the line length, the line length is used.
// After insertion, the selection is reset.
func (s *Selection) Surround(bchar, echar rune) {
	if s.line.Len() == 0 || s.Len() == 0 {
		return
	}

	defer s.Reset()

	var buf []rune
	buf = append(buf, bchar)
	buf = append(buf, []rune(s.Text())...)
	buf = append(buf, echar)

	// The begin and end positions of the selection
	// is where we insert our new buffer.
	bpos, epos := s.Pos()
	if bpos == -1 || epos == -1 {
		return
	}

	s.line.InsertBetween(bpos, epos, buf...)
}

// SelectAWord selects a word around the current cursor position,
// selecting leading or trailing spaces depending on where the cursor
// is: if on a blank space, in a word, or at the end of the line.
func (s *Selection) SelectAWord() {
	var bpos, epos int

	bpos = s.cursor.Pos()
	cpos := bpos

	spaceBefore, spaceUnder := s.spacesAroundWord(bpos)

	bpos, epos = s.line.SelectWord(cpos)
	s.cursor.Set(epos)
	cpos = s.cursor.Pos()

	spaceAfter := cpos < s.line.Len()-1 && (*s.line)[cpos+1] == inputrc.Space

	// And only select spaces after it if the word selected is not preceded
	// by spaces as well, or if we started the selection within this word.
	bpos, _ = s.adjustWordSelection(spaceBefore, spaceUnder, spaceAfter, bpos)

	if !s.Active() || bpos < cpos {
		s.Mark(bpos)
	}
}

// SelectABlankWord selects a bigword around the current cursor position,
// selecting leading or trailing spaces depending on where the cursor is:
// if on a blank space, in a word, or at the end of the line.
func (s *Selection) SelectABlankWord() (bpos, epos int) {
	bpos = s.cursor.Pos()
	spaceBefore, spaceUnder := s.spacesAroundWord(bpos)

	// If we are out of a word or in the middle of one, find its beginning.
	if !spaceUnder && !spaceBefore {
		s.cursor.Inc()
		s.cursor.Move(s.line.Backward(s.line.TokenizeSpace, s.cursor.Pos()))
		bpos = s.cursor.Pos()
	} else {
		s.cursor.ToFirstNonSpace(true)
	}

	// Then go to the end of the blank word
	s.cursor.Move(s.line.ForwardEnd(s.line.TokenizeSpace, s.cursor.Pos()))
	spaceAfter := s.cursor.Pos() < s.line.Len()-1 && (*s.line)[s.cursor.Pos()+1] == inputrc.Space

	// And only select spaces after it if the word selected is not preceded
	// by spaces as well, or if we started the selection within this word.
	bpos, _ = s.adjustWordSelection(spaceBefore, spaceUnder, spaceAfter, bpos)

	if !s.Active() || bpos < s.cursor.Pos() {
		s.Mark(bpos)
	}

	return bpos, s.cursor.Pos()
}

// SelectAShellWord selects a shell word around the cursor position,
// selecting leading or trailing spaces depending on where the cursor
// is: if on a blank space, in a word, or at the end of the line.
func (s *Selection) SelectAShellWord() (bpos, epos int) {
	s.cursor.CheckCommand()
	s.cursor.ToFirstNonSpace(true)

	sBpos, sEpos := s.line.SurroundQuotes(true, s.cursor.Pos())
	dBpos, dEpos := s.line.SurroundQuotes(false, s.cursor.Pos())

	mark, cpos := strutil.AdjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos)

	// If none of the quotes matched, use blank word
	if mark == -1 && cpos == -1 {
		mark, cpos = s.line.SelectBlankWord(s.cursor.Pos())
	}

	s.cursor.Set(mark)

	// The quotes might be followed by non-blank characters,
	// in which case we must keep expanding our selection.
	for {
		spaceBefore := mark > 0 && (*s.line)[mark-1] == inputrc.Space
		if spaceBefore {
			s.cursor.Dec()
			s.cursor.ToFirstNonSpace(false)
			s.cursor.Inc()
			break
		} else if mark == 0 {
			break
		}

		s.cursor.Move(s.line.Backward(s.line.TokenizeSpace, s.cursor.Pos()))
		mark = s.cursor.Pos()
	}

	bpos = s.cursor.Pos()
	s.cursor.Set(cpos)

	// Adjust if no spaces after.
	for {
		spaceAfter := cpos < s.line.Len()-1 && (*s.line)[cpos+1] == inputrc.Space
		if spaceAfter || cpos == s.line.Len()-1 {
			break
		}

		s.cursor.Move(s.line.ForwardEnd(s.line.TokenizeSpace, cpos))
		cpos = s.cursor.Pos()
	}

	// Else set the region inside those quotes
	if !s.Active() || bpos < s.cursor.Pos() {
		s.Mark(bpos)
	}

	return bpos, cpos
}

// ReplaceWith replaces all characters of the line within the current
// selection range by applying to each rune the provided replacer function.
// After replacement, the selection is reset.
func (s *Selection) ReplaceWith(replacer func(r rune) rune) {
	if s.line.Len() == 0 || s.Len() == 0 {
		return
	}

	defer s.Reset()

	bpos, epos := s.Pos()
	if bpos == -1 || epos == -1 {
		return
	}

	for pos := bpos; pos < epos; pos++ {
		char := (*s.line)[pos]
		char = replacer(char)
		(*s.line)[pos] = char
	}
}

// Cut deletes the current selection from the line, updates the cursor position
// and returns the deleted content, which can then be passed to the shell registers.
// After deletion, the selection is reset.
func (s *Selection) Cut() (buf string) {
	if s.line.Len() == 0 {
		return
	}

	defer s.Reset()

	bpos, epos := s.Pos()
	if bpos == -1 || epos == -1 {
		return
	}

	buf = s.Text()

	s.line.Cut(bpos, epos)

	return
}

// Surrounds returns all surround-selected regions contained by the selection.
func (s *Selection) Surrounds() []Selection {
	return s.surrounds
}

// Highlights returns the highlighting sequences for the selection.
func (s *Selection) Highlights() (fg, bg string) {
	return s.fg, s.bg
}

// Reset makes the current selection inactive, resetting all of its values.
func (s *Selection) Reset() {
	s.stype = ""
	s.active = false
	s.visual = false
	s.visualLine = false
	s.bpos = -1
	s.epos = -1
	s.fg = ""
	s.bg = ""
	s.surrounds = make([]Selection, 0)
}

func (s *Selection) checkRange(bpos, epos int) (int, int, bool) {
	// Return on some on unfixable cases.
	switch {
	case s.line.Len() == 0:
		return -1, -1, false
	case bpos < 0 && epos < 0:
		return -1, -1, false
	case bpos > s.line.Len() && epos > s.line.Len():
		return -1, -1, false
	}

	// Adjust positive out-of-range values
	if bpos > s.line.Len() {
		bpos = s.line.Len()
	}

	if epos > s.line.Len() {
		epos = s.line.Len()
	}

	// Adjust negative values when pending.
	if bpos < 0 {
		bpos, epos = epos, -1
	} else if epos < 0 {
		epos = -1
	}

	// And reorder if inversed.
	if bpos > epos && epos != -1 {
		bpos, epos = epos, bpos
	}

	return bpos, epos, true
}

func (s *Selection) selectToCursor(bpos int) (int, int) {
	var epos int

	// The cursor might be now before its original mark,
	// in which case we invert before doing any move.
	if s.cursor.Pos() < bpos {
		bpos, epos = s.cursor.Pos(), bpos
	} else {
		epos = s.cursor.Pos()
	}

	if s.visualLine {
		// To beginning of line
		for bpos--; bpos >= 0; bpos-- {
			if (*s.line)[bpos] == '\n' {
				bpos++
				break
			}
		}

		if bpos == -1 {
			bpos = 0
		}

		// To end of line
		for ; epos < s.line.Len(); epos++ {
			if epos == -1 {
				epos = 0
			}

			if (*s.line)[epos] == '\n' {
				break
			}
		}
	}

	// Check again in case the visual line inverted both.
	if bpos > epos {
		bpos, epos = epos, bpos
	}

	return bpos, epos
}

func (s *Selection) spacesAroundWord(cpos int) (before, under bool) {
	under = (*s.line)[cpos] == inputrc.Space
	before = cpos > 0 && (*s.line)[cpos-1] == inputrc.Space

	return
}

// adjustWordSelection adjust the beginning and end of a word (blank or not) selection, depending
// on whether it's surrounded by spaces, and if selection started from a whitespace or within word.
func (s *Selection) adjustWordSelection(before, under, after bool, bpos int) (int, int) {
	var epos int

	if after && !under {
		s.cursor.Inc()
		s.cursor.ToFirstNonSpace(true)
		s.cursor.Dec()
	} else if !after {
		epos = s.cursor.Pos()
		s.cursor.Set(bpos - 1)
		s.cursor.ToFirstNonSpace(false)
		s.cursor.Inc()
		bpos = s.cursor.Pos()
		s.cursor.Set(epos)
	}

	epos = s.cursor.Pos()

	return bpos, epos
}
