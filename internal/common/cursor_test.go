package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// cursorLine is a simple command line to test basic things on the cursor.
	cursorLine = Line("git command -c BranchName --another-opt value")

	// cursorMultiline is used for tests requiring multiline input (horizontal positions, etc).
	cursorMultiline = Line("git command -c \n second line of input before an empty line \n\n and then a last one")
)

func TestCursor_Set(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	type args struct {
		pos int
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		expected int
	}{
		{
			name:     "Valid position",
			args:     args{10},
			fields:   fields{line: &cursorLine},
			expected: 10,
		},
		{
			name:     "Bigger than line length",
			args:     args{100},
			fields:   fields{line: &cursorLine},
			expected: len(cursorLine),
		},
		{
			name:     "Negative",
			args:     args{-1},
			fields:   fields{line: &cursorLine, pos: 5},
			expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			c.Set(test.args.pos)
			require.Equalf(t, test.expected, c.pos, "Cursor position: %d, should be %d", c.pos, test.expected)
		})
	}
}

func TestCursor_Pos(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "Valid cursor position",
			fields: fields{line: &cursorLine, pos: 10},
			want:   10,
		},
		{
			name:   "Out-of-range cursor position",
			fields: fields{line: &cursorLine, pos: 100},
			want:   len(cursorLine),
		},
		{
			name:   "Negative cursor position",
			fields: fields{line: &cursorLine, pos: -1},
			want:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			if got := c.Pos(); got != test.want {
				t.Errorf("Cursor.Pos() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCursor_Inc(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "Start cursor position",
			fields: fields{line: &cursorLine, pos: 0},
			want:   1,
		},
		{
			name:   "Before end of line",
			fields: fields{line: &cursorLine, pos: len(cursorLine) - 1},
			want:   len(cursorLine),
		},
		{
			name:   "End of line",
			fields: fields{line: &cursorLine, pos: len(cursorLine)},
			want:   len(cursorLine),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			c.Inc()
			if got := c.Pos(); got != test.want {
				t.Errorf("Cursor.Pos() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCursor_Dec(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "Middle of line",
			fields: fields{line: &cursorLine, pos: 10},
			want:   9,
		},
		{
			name:   "Before beginning of line",
			fields: fields{line: &cursorLine, pos: 1},
			want:   0,
		},
		{
			name:   "Beginning of line",
			fields: fields{line: &cursorLine, pos: 0},
			want:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			c.Dec()
			if got := c.Pos(); got != test.want {
				t.Errorf("Cursor.Pos() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCursor_Move(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	type args struct {
		offset int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name:   "Valid positive offset",
			args:   args{10},
			fields: fields{line: &cursorLine, pos: 5},
			want:   15,
		},
		{
			name:   "Valid negative offset",
			args:   args{-10},
			fields: fields{line: &cursorLine, pos: 15},
			want:   5,
		},
		{
			name:   "Out-of-bound positive offset",
			args:   args{10},
			fields: fields{line: &cursorLine, pos: len(cursorLine) - 5},
			want:   len(cursorLine),
		},
		{
			name:   "Out-of-bound negative offset",
			args:   args{-10},
			fields: fields{line: &cursorLine, pos: 5},
			want:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			c.Move(test.args.offset)
			if got := c.Pos(); got != test.want {
				t.Errorf("Cursor.Pos() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCursor_BeginningOfLine(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "Beginning of line",
			fields: fields{line: &cursorLine, pos: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cursor{
				pos:  tt.fields.pos,
				mark: tt.fields.mark,
				line: tt.fields.line,
			}
			c.BeginningOfLine()
			require.Equalf(t, 0, c.pos, "Cursor: %d, should be 0", c.pos)
		})
	}
}

func TestCursor_EndOfLine(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "End of line",
			fields: fields{line: &cursorLine, pos: 5},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			c.EndOfLine()
			require.Equalf(t, len(*test.fields.line)-1, c.pos, "Cursor: %d, should be %d", c.pos, len(*test.fields.line)-1)
		})
	}
}

func TestCursor_EndOfLineAppend(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "End of line (append)",
			fields: fields{line: &cursorLine, pos: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cursor{
				pos:  tt.fields.pos,
				mark: tt.fields.mark,
				line: tt.fields.line,
			}
			c.EndOfLineAppend()
			require.Equalf(t, len(*tt.fields.line), c.pos, "Cursor: %d, should be %d", c.pos, len(*tt.fields.line))
		})
	}
}

func TestCursor_SetMark(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name     string
		fields   fields
		expected int
	}{
		{
			name:     "Set Mark",
			fields:   fields{line: &cursorLine, pos: 10},
			expected: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			c.SetMark()
			require.Equalf(t, test.expected, c.pos, "Mark: %d, should be %d", c.mark, test.expected)
			require.Equalf(t, c.pos, c.mark, "Cpos: %d should be equal to mark: %d", c.pos, c.mark)
		})
	}
}

func TestCursor_Mark(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name     string
		fields   fields
		expected int
	}{
		{
			name:     "Get Mark",
			fields:   fields{line: &cursorLine, mark: 10},
			expected: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			c.Mark()
			require.Equalf(t, c.Mark(), test.expected, "Mark: %d, should be %d", c.Mark(), test.expected)
		})
	}
}

func TestCursor_Line(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "Single line input",
			fields: fields{line: &cursorLine, pos: 10},
			want:   0,
		},
		{
			name:   "Multiline input (second line)",
			fields: fields{line: &cursorMultiline, pos: 20},
			want:   1, // Second line.
		},
		{
			name:   "Multiline input (last line, eol)",
			fields: fields{line: &cursorMultiline, pos: len(cursorMultiline) - 1},
			want:   len(strings.Split(string(cursorMultiline), "\n")) - 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			if got := c.Line(); got != test.want {
				t.Errorf("Cursor.Line() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCursor_LineMove(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	type args struct {
		offset int
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantLine int
		wantPos  int
	}{
		{
			name:     "Single line down (on non-multiline)",
			fields:   fields{line: &cursorLine, pos: 0},
			args:     args{1},
			wantLine: 0,
			wantPos:  0,
		},
		{
			name:     "Single line down",
			fields:   fields{line: &cursorMultiline, pos: 0},
			args:     args{1},
			wantLine: 1,
			wantPos:  15,
		},
		{
			name:     "Single line up (lands on empty line)",
			fields:   fields{line: &cursorMultiline, pos: len(cursorMultiline) - 1}, // end of last line
			args:     args{-1},
			wantLine: len(strings.Split(string(cursorMultiline), "\n")) - 2,
			wantPos:  60,
		},
		{
			name:     "Out of range line up",
			fields:   fields{line: &cursorMultiline, pos: 61}, // beginning of last line
			args:     args{-5},
			wantLine: 0,
			wantPos:  0,
		},
		{
			name:     "Out of range line down",
			fields:   fields{line: &cursorMultiline, pos: 15}, // end of first line
			args:     args{5},
			wantLine: 3,
			wantPos:  len(cursorMultiline),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			c.LineMove(test.args.offset)
			require.Equalf(t, test.wantPos, c.Pos(), "Cursor: %d, want %d", c.Pos(), test.wantPos)
		})
	}
}

func TestCursor_OnEmptyLine(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "On empty line",
			fields: fields{line: &cursorMultiline, pos: 60},
			want:   true,
		},
		{
			name:   "On non-empty line",
			fields: fields{line: &cursorMultiline, pos: 61},
			want:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			if got := c.OnEmptyLine(); got != test.want {
				t.Errorf("Cursor.OnEmptyLine() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCursor_CheckAppend(t *testing.T) {
	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "Check with valid position",
			fields: fields{line: &cursorLine, pos: 10},
			want:   10,
		},
		{
			name:   "Check with out-of-range position",
			fields: fields{line: &cursorMultiline, pos: len(cursorMultiline) + 10},
			want:   len(cursorMultiline),
		},
		{
			name:   "Check with negative position",
			fields: fields{line: &cursorMultiline, pos: -1},
			want:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			c.CheckAppend()
			if got := c.Pos(); got != test.want {
				t.Errorf("Cursor.Pos() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCursor_Coordinates(t *testing.T) {
	indent := 2 // Assumes the prompt strings uses two columns

	// Reassign the function for getting the terminal width to a fixed value
	getTermWidth = func() int { return 80 }

	type fields struct {
		pos  int
		mark int
		line *Line
	}
	tests := []struct {
		name   string
		fields fields
		wantX  int
		wantY  int
	}{
		{
			name:   "Cursor at end of buffer",
			fields: fields{line: &cursorMultiline, pos: len(cursorMultiline) - 1},
			wantX:  indent + 19,
			wantY:  len(strings.Split(string(cursorMultiline), "\n")) - 1,
		},
		{
			name:   "Cursor at beginning of buffer",
			fields: fields{line: &cursorMultiline, pos: 0},
			wantX:  indent,
			wantY:  0,
		},
		{
			name:   "Cursor on empty line",
			fields: fields{line: &cursorMultiline, pos: 60},
			wantX:  indent,
			wantY:  len(strings.Split(string(cursorMultiline), "\n")) - 2,
		},
		{
			name:   "Cursor at end of line",
			fields: fields{line: &cursorMultiline, pos: 58},
			wantX:  indent + 42,
			wantY:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Cursor{
				pos:  test.fields.pos,
				mark: test.fields.mark,
				line: test.fields.line,
			}
			gotX, gotY := c.Coordinates(indent)
			if gotX != test.wantX {
				t.Errorf("Cursor.Coordinates() gotX = %v, want %v", gotX, test.wantX)
			}
			if gotY != test.wantY {
				t.Errorf("Cursor.Coordinates() gotY = %v, want %v", gotY, test.wantY)
			}
		})
	}
}
