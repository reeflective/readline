package core

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
)

type fields struct {
	stype    string
	active   bool
	bpos     int
	epos     int
	fg       string
	bg       string
	surround []Selection
	line     *Line
	cursor   *Cursor
}

func newLine(line string) (Line, Cursor) {
	l := Line([]rune(line))
	c := Cursor{line: &l}

	return l, c
}

func fieldsWith(l Line, c *Cursor) fields {
	return fields{
		line:   &l,
		cursor: c,
		bpos:   -1,
		epos:   -1,
		stype:  "visual",
	}
}

func newTestSelection(fields fields) *Selection {
	return &Selection{
		stype:     fields.stype,
		active:    fields.active,
		bpos:      fields.bpos,
		epos:      fields.epos,
		fg:        fields.fg,
		bg:        fields.bg,
		surrounds: fields.surround,
		line:      fields.line,
		cursor:    fields.cursor,
	}
}

func TestSelection_Mark(t *testing.T) {
	line, cur := newLine("multiple-ambiguous 10.203.23.45")
	cur.Set(19)

	type args struct {
		pos int
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantBpos int
		wantEpos int
	}{
		{
			name:     "Valid position (current cursor)",
			fields:   fieldsWith(line, &cur),
			args:     args{cur.Pos()},
			wantBpos: cur.Pos(),
			wantEpos: cur.Pos(), // No movement, so both
		},
		{
			name:     "Invalid position (out of range)",
			fields:   fieldsWith(line, &cur),
			args:     args{line.Len() + 1},
			wantBpos: -1,
			wantEpos: -1,
		},
		{
			name:     "Invalid position (negative)",
			fields:   fieldsWith(line, &cur),
			args:     args{line.Len() + 1},
			wantBpos: -1,
			wantEpos: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sel := newTestSelection(test.fields)

			sel.Mark(test.args.pos)

			bpos, epos := sel.Pos()
			require.Equalf(t, test.wantBpos, bpos, "Bpos: '%d', want '%d'", bpos, test.wantBpos)
			require.Equalf(t, test.wantEpos, epos, "Epos: '%d', want '%d'", epos, test.wantEpos)
		})
	}
}

func TestSelection_MarkMove(t *testing.T) {
	line, cur := newLine("multiple-ambiguous 10.203.23.45 127.0.0.1")
	cur.Set(19)

	type args struct {
		pos int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantBpos   int
		wantEpos   int
		moveCursor int
	}{
		{
			name:       "Cursor backward (valid move)",
			fields:     fieldsWith(line, &cur),
			args:       args{cur.Pos()},
			moveCursor: -2,
			wantBpos:   cur.Pos() - 2,
			wantEpos:   cur.Pos(),
		},
		{
			name:       "Cursor forward (valid move)",
			fields:     fieldsWith(line, &cur),
			args:       args{cur.Pos()},
			moveCursor: 12,
			wantBpos:   cur.Pos(),
			wantEpos:   29,
		},
		{
			name:       "Cursor to end of line",
			fields:     fieldsWith(line, &cur),
			args:       args{cur.Pos()},
			moveCursor: line.Len() - cur.Pos(),
			wantBpos:   cur.Pos(),
			wantEpos:   line.Len(),
		},
		{
			name:     "Cursor out-of-range move (checked)",
			fields:   fieldsWith(line, &cur),
			args:     args{cur.Pos()},
			wantBpos: cur.Pos(),
			wantEpos: line.Len(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sel := newTestSelection(test.fields)

			// Mark and move the cursor
			sel.Mark(test.args.pos)
			cur.Move(test.moveCursor)

			bpos, epos := sel.Pos()
			require.Equalf(t, test.wantBpos, bpos, "Bpos: '%d', want '%d'", bpos, test.wantBpos)
			require.Equalf(t, test.wantEpos, epos, "Epos: '%d', want '%d'", epos, test.wantEpos)
		})
	}
}

func TestSelection_MarkRange(t *testing.T) {
	line, cur := newLine("multiple-ambiguous 10.203.23.45")
	pos := cur.Pos()

	type args struct {
		bpos       int
		epos       int
		moveCursor int
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantBpos int
		wantEpos int
	}{
		{
			name:     "Valid range (cursor to end of line)",
			fields:   fieldsWith(line, &cur),
			args:     args{cur.Pos(), line.Len(), 0},
			wantBpos: cur.Pos(),
			wantEpos: line.Len(),
		},
		{
			name:     "Invalid range (both positive out-of-line values)",
			fields:   fieldsWith(line, &cur),
			args:     args{line.Len() + 1, line.Len() + 10, 0},
			wantBpos: -1,
			wantEpos: -1,
		},
		{
			name:     "Range with negative epos (uses cursor pos instead)",
			fields:   fieldsWith(line, &cur),
			args:     args{cur.Pos(), -1, 5},
			wantBpos: cur.Pos(),
			wantEpos: cur.Pos() + 5,
		},
		{
			name:     "Range with negative bpos (uses cursor pos instead)",
			fields:   fieldsWith(line, &cur),
			args:     args{-1, cur.Pos(), 5},
			wantBpos: cur.Pos(),
			wantEpos: cur.Pos() + 5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cur.Set(pos)

			sel := newTestSelection(test.fields)

			sel.MarkRange(test.args.bpos, test.args.epos)
			cur.Move(test.args.moveCursor)

			bpos, epos := sel.Pos()
			require.Equalf(t, test.wantBpos, bpos, "Bpos: '%d', want '%d'", bpos, test.wantBpos)
			require.Equalf(t, test.wantEpos, epos, "Epos: '%d', want '%d'", epos, test.wantEpos)
		})
	}
}

func TestSelection_MarkSurround(t *testing.T) {
	line, cur := newLine("multiple-ambiguous '10.203.23.45 127.0.0.1' ::1")
	type args struct {
		bpos int
		epos int
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantSelections int
		wantBpos       int
		wantEpos       int
		wantBposS2     int
		wantEposS2     int
	}{
		{
			name:           "Valid surround (single quotes)",
			fields:         fieldsWith(line, &cur),
			args:           args{19, 42},
			wantSelections: 2,
			wantBpos:       19,
			wantEpos:       20,
			wantBposS2:     42,
			wantEposS2:     43,
		},
		{
			name:           "Invalid (epos out of range)",
			fields:         fieldsWith(line, &cur),
			args:           args{19, line.Len() + 1},
			wantSelections: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sel := newTestSelection(test.fields)

			sel.MarkSurround(test.args.bpos, test.args.epos)

			require.Lenf(t, sel.Surrounds(), test.wantSelections,
				"Surround areas: '%d', want '%d'", sel.Surrounds(), test.wantSelections)

			if len(sel.Surrounds()) == 0 {
				return
			}

			// Surround 1
			bpos, epos := sel.Surrounds()[0].Pos()
			require.Equalf(t, test.wantBpos, bpos, "Bpos: '%d', want '%d'", bpos, test.wantBpos)
			require.Equalf(t, test.wantEpos, epos, "Epos: '%d', want '%d'", epos, test.wantEpos)

			// Surround 2
			bpos, epos = sel.Surrounds()[1].Pos()
			require.Equalf(t, test.wantBposS2, bpos, "BposS2: '%d', want '%d'", bpos, test.wantBposS2)
			require.Equalf(t, test.wantEposS2, epos, "EposS2: '%d', want '%d'", epos, test.wantEposS2)
		})
	}
}

func TestSelection_Active(t *testing.T) {
	line, cur := newLine("multiple-ambiguous 10.203.23.45")
	type args struct {
		bpos int
		epos int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Valid position (current cursor)",
			fields: fieldsWith(line, &cur),
			args:   args{cur.Pos(), cur.Pos() + 1},
			want:   true,
		},
		{
			name:   "Invalid range (both positive out-of-line values)",
			fields: fieldsWith(line, &cur),
			args:   args{line.Len() + 1, line.Len() + 10},
			want:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sel := newTestSelection(test.fields)

			sel.MarkRange(test.args.bpos, test.args.epos)

			if got := sel.Active(); got != test.want {
				t.Errorf("Selection.Active() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestSelection_Visual(t *testing.T) {
	single, cSingle := newLine("multiple-ambiguous 10.203.23.45 127.0.0.1")
	multi, cMulti := newLine("basic -f \"commands.go \nanother testing\" --alternate \"another\nquote\" -c")

	type args struct {
		bpos       int
		epos       int
		cursorMove int
		visualLine bool
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantBpos int
		wantEpos int
	}{
		// Visual
		{
			name:     "Cursor position (single character)",
			fields:   fieldsWith(single, &cSingle),
			args:     args{bpos: cSingle.Pos(), epos: -1},
			wantBpos: 0,
			wantEpos: 1,
		},
		{
			name:     "Cursor to end of line",
			fields:   fieldsWith(single, &cSingle),
			args:     args{bpos: cSingle.Pos(), epos: -1, cursorMove: single.Len() - cSingle.Pos()},
			wantBpos: 0,
			wantEpos: single.Len(),
		},
		// Visual line
		{
			name:     "Visual line (single line)",
			fields:   fieldsWith(single, &cSingle),
			args:     args{bpos: 20, epos: -1, visualLine: true},
			wantBpos: 0,
			wantEpos: single.Len(),
		},
		{
			name:     "Visual line (multiline)",
			fields:   fieldsWith(multi, &cMulti),
			args:     args{bpos: 24, epos: -1, visualLine: true, cursorMove: 24},
			wantBpos: 23,
			wantEpos: 61,
		},
		{
			name:     "Visual line cursor movement (multiline)",
			fields:   fieldsWith(multi, &cMulti),
			args:     args{bpos: 24, epos: -1, visualLine: true, cursorMove: 0},
			wantBpos: 0,
			wantEpos: 61,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.fields.cursor.Set(0)

			sel := newTestSelection(test.fields)

			// First create the mark, then only after set it to visual/line
			sel.MarkRange(test.args.bpos, test.args.epos)
			sel.Visual(test.args.visualLine)

			test.fields.cursor.Move(test.args.cursorMove)

			gotBpos, gotEpos := sel.Pos()
			if gotBpos != test.wantBpos {
				t.Errorf("Selection.Pos() gotBpos = %v, want %v", gotBpos, test.wantBpos)
			}
			if gotEpos != test.wantEpos {
				t.Errorf("Selection.Pos() gotEpos = %v, want %v", gotEpos, test.wantEpos)
			}
		})
	}
}

// TODO: selection.Pos() is actually used in many/all other tests in this file,
// so this test should be used to check additional cases and setups.
// func TestSelection_Pos(t *testing.T) {
// 	l, c := newLine("multiple-ambiguous 10.203.23.45 127.0.0.1")
// 	tests := []struct {
// 		name     string
// 		fields   fields
// 		wantBpos int
// 		wantEpos int
// 	}{
// 		// Visual
// 		{
// 			name:   "Valid position (current cursor)",
// 			fields: fieldsWith(l, &c),
// 		},
// 		// Visual line
// 	}
// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			s := &Selection{
// 				stype:     test.fields.stype,
// 				active:    test.fields.active,
// 				bpos:      test.fields.bpos,
// 				epos:      test.fields.epos,
// 				fg:        test.fields.fg,
// 				bg:        test.fields.bg,
// 				surrounds: test.fields.surround,
// 				line:      test.fields.line,
// 				cursor:    test.fields.cursor,
// 			}
// 			gotBpos, gotEpos := s.Pos()
// 			if gotBpos != test.wantBpos {
// 				t.Errorf("Selection.Pos() gotBpos = %v, want %v", gotBpos, test.wantBpos)
// 			}
// 			if gotEpos != test.wantEpos {
// 				t.Errorf("Selection.Pos() gotEpos = %v, want %v", gotEpos, test.wantEpos)
// 			}
// 		})
// 	}
// }

func TestSelection_Cursor(t *testing.T) {
	single, cSingle := newLine("multiple-ambiguous 10.203.23.45 127.0.0.1")
	multi, cMulti := newLine("basic -f \"commands.go \nanother testing\" --alternate \"another\nquote\" -c")

	type args struct {
		bpos       int
		cursorMove int
		visualLine bool
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantCpos int
	}{
		{
			name:     "Cursor forward word",
			fields:   fieldsWith(single, &cSingle),
			args:     args{bpos: 0, cursorMove: 6},
			wantCpos: 0, // Forward selection does not move the cursor.
		},
		{
			name:     "Cursor backward word",
			fields:   fieldsWith(single, &cSingle),
			args:     args{bpos: 6, cursorMove: -6},
			wantCpos: 0, // Backward selection, if deleted, would move the cursor back.
		},
		{
			name:     "Cursor on last line (visual line selection)",
			fields:   fieldsWith(multi, &cMulti),
			args:     args{bpos: multi.Len() - 1, visualLine: true},
			wantCpos: 31, // Position of the cursor on the previous line if we deleted our selected line.
		},
		{
			name:   "Current line longer than next (visual line selection)",
			fields: fieldsWith(multi, &cMulti),
			args:   args{bpos: multi.Len() - 11, visualLine: true}, // end of before-last line.
			// Same here: vertical position of cursor greater than last line length, becomes last line length.
			wantCpos: 31,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sel := newTestSelection(test.fields)

			// Place the cursor where we want to start selecting.
			test.fields.cursor.Set(test.args.bpos)
			sel.Mark(test.fields.cursor.Pos())
			if test.args.visualLine {
				sel.Visual(test.args.visualLine)
			}

			// Move the cursor when needed.
			test.fields.cursor.Move(test.args.cursorMove)

			cpos := sel.Cursor()
			if cpos != test.wantCpos {
				t.Errorf("Selection.Cursor() cpos = %v, want %v", cpos, test.wantCpos)
			}
		})
	}
}

func TestSelection_Pop(t *testing.T) {
	single, cSingle := newLine("multiple-ambiguous 10.203.23.45 127.0.0.1")
	multi, cMulti := newLine("basic -f \"commands.go \nanother testing\" --alternate \"another\nquote\" -c")

	type args struct {
		bpos       int
		moveCursor int
		visual     bool
		visualLine bool
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantBuf  string
		wantBpos int
		wantEpos int
		wantCpos int
	}{
		// Single line
		{
			name:     "Invalid range (not visual, no move, epos=bpos)",
			fields:   fieldsWith(single, &cSingle),
			args:     args{bpos: cSingle.Pos()},
			wantBpos: cSingle.Pos(),
			wantEpos: cSingle.Pos(),
			wantCpos: cSingle.Pos(),
		},
		{
			name:     "Valid cursor (visual, no move)",
			fields:   fieldsWith(single, &cSingle),
			args:     args{bpos: cSingle.Pos(), visual: true},
			wantBpos: cSingle.Pos(),
			wantEpos: cSingle.Pos() + 1,
			wantCpos: cSingle.Pos(),
			wantBuf:  "m",
		},
		{
			name:     "Valid range (cursor to end of line)",
			fields:   fieldsWith(single, &cSingle),
			args:     args{bpos: cSingle.Pos(), moveCursor: single.Len() - cSingle.Pos()},
			wantBpos: cSingle.Pos(),
			wantEpos: single.Len(),
			wantBuf:  string(single),
		},
		// Multiline
		{
			name:     "Cursor on last line (visual line selection)",
			fields:   fieldsWith(multi, &cMulti),
			args:     args{bpos: multi.Len() - 1, visualLine: true},
			wantCpos: 31, // Position of the cursor on the previous line if we deleted our selected line.
			wantBpos: 61,
			wantEpos: multi.Len(),
			wantBuf:  "quote\" -c",
		},
		{
			name:   "Current line longer than next (visual line selection)",
			fields: fieldsWith(multi, &cMulti),
			args:   args{bpos: multi.Len() - 11, visualLine: true}, // end of before-last line.
			// Same here: vertical position of cursor greater than last line length, becomes last line length.
			wantCpos: 31,
			wantBpos: 23,
			wantEpos: 61,
			wantBuf:  "another testing\" --alternate \"another\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sel := newTestSelection(test.fields)

			// Place the cursor where we want to start selecting.
			test.fields.cursor.Set(test.args.bpos)
			sel.Mark(test.fields.cursor.Pos())
			if test.args.visualLine || test.args.visual {
				sel.Visual(test.args.visualLine)
			}

			// Move the cursor when needed.
			test.fields.cursor.Move(test.args.moveCursor)

			gotBuf, gotBpos, gotEpos, gotCpos := sel.Pop()
			if gotBuf != test.wantBuf {
				t.Errorf("Selection.Pop() gotBuf = %v, want %v", gotBuf, test.wantBuf)
			}
			if gotBpos != test.wantBpos {
				t.Errorf("Selection.Pop() gotBpos = %v, want %v", gotBpos, test.wantBpos)
			}
			if gotEpos != test.wantEpos {
				t.Errorf("Selection.Pop() gotEpos = %v, want %v", gotEpos, test.wantEpos)
			}
			if gotCpos != test.wantCpos {
				t.Errorf("Selection.Pop() gotCpos = %v, want %v", gotCpos, test.wantCpos)
			}

			// Selection should be reset.
			testSelectionReset(t, sel)
		})
	}
}

func TestSelection_InsertAt(t *testing.T) {
	line, cur := newLine("multiple-ambiguous 10.203.23.45 127.0.0.1")
	type args struct {
		bpos int
		epos int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantBuf string
	}{
		{
			name:    "Valid range insertion",
			args:    args{bpos: line.Len() - 10, epos: line.Len()}, // The line won't actually change.
			wantBuf: string(line),
		},
		{
			name:    "Insert at end of line",
			args:    args{bpos: line.Len(), epos: -1},
			wantBuf: string(line) + " 127.0.0.1",
		},
		{
			name:    "Insert at begin position (epos == -1)",
			args:    args{bpos: 18, epos: -1},
			wantBuf: "multiple-ambiguous 127.0.0.1 10.203.23.45 127.0.0.1",
		},
		{
			name:    "Insert at begin position (bpos == epos)",
			args:    args{bpos: 18, epos: 18},
			wantBuf: "multiple-ambiguous 127.0.0.1 10.203.23.45 127.0.0.1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			line, cur = newLine("multiple-ambiguous 10.203.23.45 127.0.0.1")
			test.fields.line, test.fields.cursor = &line, &cur

			sel := &Selection{
				stype:     test.fields.stype,
				active:    test.fields.active,
				bpos:      test.fields.bpos,
				epos:      test.fields.epos,
				fg:        test.fields.fg,
				bg:        test.fields.bg,
				surrounds: test.fields.surround,
				line:      test.fields.line,
				cursor:    test.fields.cursor,
			}
			// Select the last IP.
			sel.MarkRange(test.fields.line.Len()-10, test.fields.line.Len())

			// Insert according to test spec.
			sel.InsertAt(test.args.bpos, test.args.epos)

			// Check line contents and selection reset.
			gotBuf := string(*test.fields.line)
			if gotBuf != test.wantBuf {
				t.Errorf("Selection.InsertAt() gotBuf = %v, want %v", gotBuf, test.wantBuf)
			}

			testSelectionReset(t, sel)
		})
	}
}

func TestSelection_Surround(t *testing.T) {
	line, cur := newLine("multiple-ambiguous 10.203.23.45 127.0.0.1")
	type args struct {
		bchar rune
		echar rune
		bpos  int
		epos  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantBuf string
	}{
		{
			name:    "Valid range",
			fields:  fieldsWith(line, &cur),
			args:    args{bchar: '"', echar: '"', bpos: 19, epos: 19 + 12},
			wantBuf: "multiple-ambiguous \"10.203.23.45\" 127.0.0.1",
		},
		{
			name:    "Valid range (epos at end of line)",
			fields:  fieldsWith(line, &cur),
			args:    args{bchar: '\'', echar: '\'', bpos: 32, epos: line.Len()},
			wantBuf: "multiple-ambiguous 10.203.23.45 '127.0.0.1'",
		},
		{
			name:    "Invalid range (epos out of range)",
			fields:  fieldsWith(line, &cur),
			args:    args{bchar: '\'', echar: '\'', bpos: 32, epos: line.Len() + 1},
			wantBuf: "multiple-ambiguous 10.203.23.45 '127.0.0.1'",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			line, cur = newLine("multiple-ambiguous 10.203.23.45 127.0.0.1")
			test.fields.line, test.fields.cursor = &line, &cur

			sel := newTestSelection(test.fields)

			// Mark and surround the selection.
			sel.MarkRange(test.args.bpos, test.args.epos)
			sel.Surround(test.args.bchar, test.args.echar)

			// Check line contents and selection reset.
			gotBuf := string(*test.fields.line)
			if gotBuf != test.wantBuf {
				t.Errorf("Selection.Surround() gotBuf = %v, want %v", gotBuf, test.wantBuf)
			}
			testSelectionReset(t, sel)
		})
	}
}

func TestSelection_ReplaceWith(t *testing.T) {
	line, cur := newLine("multiple-ambiguous lower UPPER")
	type args struct {
		bpos     int
		epos     int
		replacer func(r rune) rune
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantBuf string
	}{
		{
			name:    "Replace to upper",
			fields:  fieldsWith(line, &cur),
			args:    args{bpos: 19, epos: 24, replacer: unicode.ToUpper},
			wantBuf: "multiple-ambiguous LOWER UPPER",
		},
		{
			name:    "Replace to lower (with epos out-of-range)",
			fields:  fieldsWith(line, &cur),
			args:    args{bpos: 25, epos: line.Len() + 1, replacer: unicode.ToLower},
			wantBuf: "multiple-ambiguous lower upper",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			line, cur = newLine("multiple-ambiguous lower UPPER")
			test.fields.line, test.fields.cursor = &line, &cur

			sel := newTestSelection(test.fields)

			// Mark and replace the selection.
			sel.MarkRange(test.args.bpos, test.args.epos)
			sel.ReplaceWith(test.args.replacer)

			// Check line contents and selection reset.
			gotBuf := string(*test.fields.line)
			if gotBuf != test.wantBuf {
				t.Errorf("Selection.ReplaceWith() gotBuf = %v, want %v", gotBuf, test.wantBuf)
			}
			testSelectionReset(t, sel)
		})
	}
}

func TestSelection_Reset(t *testing.T) {
	line, cur := newLine("multiple-ambiguous test")
	type args struct {
		bpos int
		epos int
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantActive     bool
		wantVisual     bool
		wantVisualLine bool
		wantBpos       int
		wantEpos       int
		wantSurrounds  int
		wantFg         string
		wantBg         string
	}{
		{
			name:           "Select and reset",
			fields:         fieldsWith(line, &cur),
			args:           args{bpos: 0, epos: line.Len()},
			wantActive:     false,
			wantBpos:       -1,
			wantEpos:       -1,
			wantVisual:     false,
			wantVisualLine: false,
			wantFg:         "",
			wantBg:         "",
			wantSurrounds:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sel := newTestSelection(test.fields)

			// Mark the selection and reset it.
			sel.MarkRange(test.args.bpos, test.args.epos)
			sel.Reset()

			if sel.active != test.wantActive {
				t.Errorf("Selection.Reset() sel.active = %v, want %v", sel.active, test.wantActive)
			}
			if sel.bpos != test.wantBpos {
				t.Errorf("Selection.Reset() sel.bpos = %v, want %v", sel.bpos, test.wantBpos)
			}
			if sel.epos != test.wantEpos {
				t.Errorf("Selection.Reset() sel.epos = %v, want %v", sel.epos, test.wantEpos)
			}
			if sel.visual != test.wantVisual {
				t.Errorf("Selection.Reset() sel.visual = %v, want %v", sel.visual, test.wantVisual)
			}
			if sel.visualLine != test.wantVisualLine {
				t.Errorf("Selection.Reset() sel.visualLine = %v, want %v", sel.visualLine, test.wantVisualLine)
			}
			if sel.fg != test.wantFg {
				t.Errorf("Selection.Reset() sel.fg = %v, want %v", sel.fg, test.wantFg)
			}
			if sel.bg != test.wantBg {
				t.Errorf("Selection.Reset() sel.bg = %v, want %v", sel.bg, test.wantBg)
			}
			if len(sel.surrounds) != test.wantSurrounds {
				t.Errorf("Selection.Reset() len(sel.surrounds) = %v, want %v", len(sel.surrounds), test.wantSurrounds)
			}
		})
	}
}

func testSelectionReset(t *testing.T, sel *Selection) {
	t.Helper()

	if sel.Text() != "" {
		t.Errorf("Selection.Reset() gotBuf = %v, want %v", sel.Text(), "")
	}

	if sel.bpos != -1 {
		t.Errorf("Selection.Reset() gotBpos = %v, want %v", sel.bpos, -1)
	}

	if sel.epos != -1 {
		t.Errorf("Selection.Reset() epos = %v, want %v", sel.epos, -1)
	}

	if sel.active {
		t.Error("Selection.Reset() is still active, should not be")
	}
}
