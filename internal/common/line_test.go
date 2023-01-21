package common

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: Add many more values to insert: Control characters,
// and everything as weird as one can find in a terminal.
func TestLine_Insert(t *testing.T) {
	line := Line("multiple-ambiguous 10.203.23.45")

	type args struct {
		pos int
		r   []rune
	}
	tests := []struct {
		name string
		l    *Line
		args args
		want string
	}{
		{
			name: "Append to end of line",
			l:    &line,
			args: args{pos: len(line), r: []rune(" 127.0.0.1")},
			want: "multiple-ambiguous 10.203.23.45 127.0.0.1",
		},
		{
			name: "Insert at beginning of line",
			l:    &line,
			args: args{pos: 0, r: []rune("root ")},
			want: "root multiple-ambiguous 10.203.23.45 127.0.0.1",
		},
		{
			name: "Insert with an out of range position",
			l:    &line,
			args: args{pos: 100, r: []rune("dropped")},
			want: "root multiple-ambiguous 10.203.23.45 127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Insert(tt.args.pos, tt.args.r...)
		})
		require.Equalf(t, tt.want, string(line), "Line: '%s', wanted '%s'", string(line), tt.want)
	}
}

func TestLine_InsertBetween(t *testing.T) {
	line := Line("multiple-ambiguous 10.203.23.45")

	type args struct {
		bpos int
		epos int
		r    []rune
	}
	tests := []struct {
		name string
		l    *Line
		args args
		want string
	}{
		{
			name: "Insert at beginning of line",
			l:    &line,
			args: args{bpos: 0, r: []rune("root ")},
			want: "root multiple-ambiguous 10.203.23.45",
		},
		{
			name: "Insert with a non-ending range (thus at other position)",
			l:    &line,
			args: args{bpos: 24, epos: -1, r: []rune("trail ")},
			want: "root multiple-ambiguous trail 10.203.23.45",
		},
		{
			name: "Append to end of line (no epos)",
			l:    &line,
			args: args{bpos: 42, epos: -1, r: []rune(" 127.0.0.1")},
			want: "root multiple-ambiguous trail 10.203.23.45 127.0.0.1",
		},

		{
			name: "Insert with cut",
			l:    &line,
			args: args{bpos: 23, epos: 29, r: []rune(" 10.10.10.10")},
			want: "root multiple-ambiguous 10.10.10.10 10.203.23.45 127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.InsertBetween(tt.args.bpos, tt.args.epos, tt.args.r...)
		})
		require.Equalf(t, tt.want, string(line), "Line: '%s', wanted '%s'", string(line), tt.want)
	}
}

func TestLine_Cut(t *testing.T) {
	line := Line("basic -f \"commands.go,line.go\" -cp=/usr")

	type args struct {
		bpos int
		epos int
	}
	tests := []struct {
		name string
		l    *Line
		args args
		want string
	}{
		{
			name: "Cut in the middle",
			l:    &line,
			args: args{bpos: 21, epos: 29},
			want: "basic -f \"commands.go\" -cp=/usr",
		},
		{
			name: "Cut at beginning of line",
			l:    &line,
			args: args{bpos: 0, epos: 9},
			want: "\"commands.go\" -cp=/usr",
		},
		{
			name: "Cut with range out of bounds",
			l:    &line,
			args: args{bpos: 13, epos: len(line) + 1},
			want: "\"commands.go\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Cut(tt.args.bpos, tt.args.epos)
		})
		require.Equalf(t, tt.want, string(line), "Line: '%s', wanted '%s'", string(line), tt.want)
	}
}

func TestLine_CutRune(t *testing.T) {
	line := Line("basic -f \"commands.go,line.go\" -cp=/usr")

	type args struct {
		pos int
	}
	tests := []struct {
		name string
		l    *Line
		args args
		want string
	}{
		{
			name: "Cut rune in the middle",
			l:    &line,
			args: args{pos: 22},
			want: "basic -f \"commands.go,ine.go\" -cp=/usr",
		},
		{
			name: "Cut rune at end of line, append mode",
			l:    &line,
			args: args{pos: len(line) - 1},
			want: "basic -f \"commands.go,ine.go\" -cp=/us",
		},
		{
			name: "Cut rune at invalid position (not removed)",
			l:    &line,
			args: args{pos: len(line) + 1},
			want: "basic -f \"commands.go,ine.go\" -cp=/us",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.CutRune(tt.args.pos)
		})
		require.Equalf(t, tt.want, string(line), "Line: '%s', wanted '%s'", string(line), tt.want)
	}
}

func TestLine_Len(t *testing.T) {
	line := Line("basic -f \"commands.go,line.go\" -cp=/usr")

	tests := []struct {
		name string
		l    *Line
		want int
	}{
		{
			name: "Length of non-empty line",
			l:    &line,
			want: len(line),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.Len(); got != tt.want {
				t.Errorf("Line.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_SelectWord(t *testing.T) {
	line := Line("basic -c true -p on")

	type args struct {
		pos int
	}
	tests := []struct {
		name     string
		l        *Line
		args     args
		wantBpos int
		wantEpos int
	}{
		{
			name:     "Select word from start",
			l:        &line,
			args:     args{0},
			wantBpos: 0,
			wantEpos: 4,
		},
		{
			name:     "Select word in middle of word",
			l:        &line,
			args:     args{2},
			wantBpos: 0,
			wantEpos: 4,
		},
		{
			name:     "Select shell word flag",
			l:        &line,
			args:     args{10},
			wantBpos: 9,
			wantEpos: 12,
		},
		{
			name:     "Select numeric expression",
			l:        &line,
			args:     args{len(line) - 1},
			wantBpos: len(line) - 2,
			wantEpos: len(line) - 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotBpos, gotEpos := test.l.SelectWord(test.args.pos)
			if gotBpos != test.wantBpos {
				t.Errorf("Line.SelectWord() gotBpos = %v, want %v", gotBpos, test.wantBpos)
			}
			if gotEpos != test.wantEpos {
				t.Errorf("Line.SelectWord() gotEpos = %v, want %v", gotEpos, test.wantEpos)
			}
		})
	}
}

// TODO: Add special characters.
func TestLine_Find(t *testing.T) {
	line := Line("basic -f \"commands.go,line.go\" -cp=/usr --option [value1 value2]")
	pos := 0 // We reuse the same updated position for each next test case.

	type args struct {
		r       rune
		pos     int
		forward bool
	}
	tests := []struct {
		name     string
		l        *Line
		args     args
		wantRpos int
	}{
		// Forward
		{
			name:     "Find first quote from beginning of line",
			l:        &line,
			args:     args{r: '"', pos: pos, forward: true},
			wantRpos: 9,
		},
		{
			name:     "Find first opening bracket from start",
			l:        &line,
			args:     args{r: '[', pos: pos, forward: true},
			wantRpos: 49,
		},
		{
			name:     "Search for non existent rune in the line",
			l:        &line,
			args:     args{r: '%', pos: pos, forward: true},
			wantRpos: -1,
		},
		// Backward
		{
			name:     "Find first quote from end of line",
			l:        &line,
			args:     args{r: '"', pos: len(line), forward: false},
			wantRpos: 29,
		},
		{
			name:     "Find first opening bracket from end of line",
			l:        &line,
			args:     args{r: '[', pos: len(line), forward: false},
			wantRpos: 49,
		},
		{
			name:     "Search for non existent rune in the line",
			l:        &line,
			args:     args{r: '%', pos: len(line), forward: false},
			wantRpos: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRpos := tt.l.Find(tt.args.r, tt.args.pos, tt.args.forward); gotRpos != tt.wantRpos {
				t.Errorf("Line.Next() = %v, want %v", gotRpos, tt.wantRpos)
			}
			pos = tt.wantRpos
		})
	}
}

func TestLine_Forward(t *testing.T) {
	line := Line("basic -f \"commands.go,line.go\" -cp=/usr --option [value1 value2]")

	type args struct {
		split Tokenizer
		pos   int
	}
	tests := []struct {
		name       string
		l          *Line
		args       args
		wantAdjust int
	}{
		{
			name:       "Forward word",
			l:          &line,
			args:       args{split: line.Tokenize, pos: 0},
			wantAdjust: 6,
		},
		{
			name:       "Forward blank word",
			l:          &line,
			args:       args{split: line.TokenizeSpace, pos: 10},
			wantAdjust: 21,
		},
		{
			name:       "Forward bracket",
			l:          &line,
			args:       args{split: line.TokenizeBlock, pos: 49},
			wantAdjust: 64,
		},

		{
			name:       "Forward bracket (no match)",
			l:          &line,
			args:       args{split: line.TokenizeBlock, pos: 48},
			wantAdjust: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if gotAdjust := test.l.Forward(test.args.split, test.args.pos); gotAdjust != test.wantAdjust {
				t.Errorf("Line.Forward() = %v, want %v", gotAdjust, test.wantAdjust)
			}
		})
	}
}

func TestLine_ForwardEnd(t *testing.T) {
	line := Line("basic -f \"commands.go,line.go\" -cp=/usr --option [value1 value2]")

	type args struct {
		split Tokenizer
		pos   int
	}
	tests := []struct {
		name       string
		l          *Line
		args       args
		wantAdjust int
	}{
		{
			name:       "Forward word end",
			l:          &line,
			args:       args{split: line.Tokenize, pos: 0},
			wantAdjust: 4,
		},
		{
			name:       "Forward blank word end",
			l:          &line,
			args:       args{split: line.TokenizeSpace, pos: 10},
			wantAdjust: 19,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAdjust := tt.l.ForwardEnd(tt.args.split, tt.args.pos); gotAdjust != tt.wantAdjust {
				t.Errorf("Line.ForwardEnd() = %v, want %v", gotAdjust, tt.wantAdjust)
			}
		})
	}
}

func TestLine_Backward(t *testing.T) {
	line := Line("basic -f \"commands.go,line.go\" -cp=/usr --option [value1 value2]")

	type args struct {
		split Tokenizer
		pos   int
	}
	tests := []struct {
		name       string
		l          *Line
		args       args
		wantAdjust int
	}{
		{
			name:       "Backward word",
			l:          &line,
			args:       args{split: line.Tokenize, pos: 6},
			wantAdjust: -6,
		},
		{
			name:       "Backward blank word",
			l:          &line,
			args:       args{split: line.TokenizeSpace, pos: 22},
			wantAdjust: -13,
		},
		{
			name:       "Backward bracket",
			l:          &line,
			args:       args{split: line.TokenizeBlock, pos: line.Len() - 1},
			wantAdjust: -14,
		},

		{
			name:       "Backward bracket (no match)",
			l:          &line,
			args:       args{split: line.TokenizeBlock, pos: line.Len() - 2},
			wantAdjust: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAdjust := tt.l.Backward(tt.args.split, tt.args.pos); gotAdjust != tt.wantAdjust {
				t.Errorf("Line.Backward() = %v, want %v", gotAdjust, tt.wantAdjust)
			}
		})
	}
}

func TestLine_Tokenize(t *testing.T) {
	line := Line("basic -f \"commands.go \nanother testing\" --alternate \"another\nquote\" -c")

	type args struct {
		pos int
	}
	tests := []struct {
		name  string
		l     *Line
		args  args
		want  []string
		want1 int
		want2 int
	}{
		{
			name: "Tokenize line",
			args: args{pos: line.Len() - 1},
			l:    &line,
			want: []string{
				"basic ", "-", "f ", "\"", "commands", ".", "go \n",
				"another ", "testing", "\" ", "--", "alternate ", "\"", "another\n",
				"quote", "\" ", "-", "c",
			},
			want1: 17,
			want2: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, got1, got2 := test.l.Tokenize(test.args.pos)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Line.Tokenize() got = %v, want %v", got, test.want)
			}
			if got1 != test.want1 {
				t.Errorf("Line.Tokenize() got1 = %v, want %v", got1, test.want1)
			}
			if got2 != test.want2 {
				t.Errorf("Line.Tokenize() got2 = %v, want %v", got2, test.want2)
			}
		})
	}
}

func TestLine_TokenizeSpace(t *testing.T) {
	line := Line("basic -f \"commands.go \nanother testing\" --alternate \"another\nquote\" -c")

	type args struct {
		pos int
	}
	tests := []struct {
		name  string
		l     *Line
		args  args
		want  []string
		want1 int
		want2 int
	}{
		{
			name: "Tokenize spaces",
			args: args{pos: line.Len() - 1},
			l:    &line,
			want: []string{
				"basic ", "-f ", "\"commands.go \nanother ", "testing\" ", "--alternate ", "\"another\nquote\" ", "-c",
			},
			want1: 6,
			want2: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, got1, got2 := test.l.TokenizeSpace(test.args.pos)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Line.TokenizeSpace() got = %v, want %v", got, test.want)
			}
			if got1 != test.want1 {
				t.Errorf("Line.TokenizeSpace() got1 = %v, want %v", got1, test.want1)
			}
			if got2 != test.want2 {
				t.Errorf("Line.TokenizeSpace() got2 = %v, want %v", got2, test.want2)
			}
		})
	}
}

func TestLine_TokenizeBlock(t *testing.T) {
	line := Line("basic -f \"commands.go \nanother testing\" --alternate \"another\nquote\" -v { expression here } -a [value1 value2]")

	type args struct {
		pos int
	}
	tests := []struct {
		name  string
		l     *Line
		args  args
		want  []string
		want1 int
		want2 int
	}{
		{
			name: "Tokenize blocks",
			args: args{pos: line.Len()}, // Note that we are in append-eol mode
			l:    &line,
			want: []string{
				"basic -f \"commands.go \nanother testing\" --alternate \"another\nquote\" -v { expression here } -a", "[value1 value2",
			},
			want1: 1,
			want2: 14,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, got1, got2 := test.l.TokenizeBlock(test.args.pos)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Line.TokenizeBlock() got = %v, want %v", got, test.want)
			}
			if got1 != test.want1 {
				t.Errorf("Line.TokenizeBlock() got1 = %v, want %v", got1, test.want1)
			}
			if got2 != test.want2 {
				t.Errorf("Line.TokenizeBlock() got2 = %v, want %v", got2, test.want2)
			}
		})
	}
}

func TestLine_Display(t *testing.T) {
	type args struct {
		indent    int
		suggested string
	}
	tests := []struct {
		name string
		l    *Line
		args args
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Display(tt.args.indent, tt.args.suggested)
		})
	}
}

func TestLine_Coordinates(t *testing.T) {
	indent := 10
	line := Line("basic -f \"commands.go,line.go\" -cp=/usr --option [value1 value2]")
	lineAutosuggest := " --option \" value entered earlier\" -m user@host.com"
	multiline := Line("basic -f \"commands.go \nanother testing\" --alternate \"another\nquote\" -v { expression here } -a [value1 value2]")
	autosuggest := " --option \" value entered earlier\"\n -m user@host.com"

	// Reassign the function for getting the terminal width to a fixed value
	getTermWidth = func() int { return 80 }

	type args struct {
		indent    int
		suggested string
	}
	tests := []struct {
		name  string
		l     *Line
		args  args
		wantX int
		wantY int
	}{
		{
			name:  "Single line buffer, (no autosuggestion)",
			l:     &line,
			args:  args{indent: indent},
			wantY: 0,
			wantX: indent + 64,
		},
		{
			name:  "Single line buffer, (with autosuggestion)",
			l:     &line,
			args:  args{indent: indent, suggested: lineAutosuggest},
			wantY: 1,
			wantX: 45,
		},
		{
			name:  "Multiline buffer, (no autosuggestion)",
			l:     &multiline,
			args:  args{indent: indent},
			wantY: 2,
			wantX: indent + 49,
		},
		{
			name:  "Multiline buffer (with autosuggestion)",
			l:     &multiline,
			args:  args{indent: indent, suggested: autosuggest},
			wantY: 4,
			wantX: indent + 18,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotX, gotY := test.l.Coordinates(test.args.indent, test.args.suggested)
			if gotX != test.wantX {
				t.Errorf("Line.Used() gotX = %v, want %v", gotX, test.wantX)
			}
			if gotY != test.wantY {
				t.Errorf("Line.Used() gotY = %v, want %v", gotY, test.wantY)
			}
		})
	}
}
