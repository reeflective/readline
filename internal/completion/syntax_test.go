package completion

import (
	"testing"

	"github.com/reeflective/readline/internal/core"
)

func TestAutopairInsertOrJump(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		cursor     int
		key        rune
		wantLine   string
		wantSkip   bool
		wantCursor int // Relative to original if not specified? No, absolute.
	}{
		{
			name:       "Empty line, insert quote",
			line:       "",
			cursor:     0,
			key:        '"',
			wantLine:   "\"",     // Function inserts closer
			wantSkip:   false,    // selfInsert will insert opener
			wantCursor: 0,        // Cursor stays same (caller handles insert)
		},
		{
			name:       "Inside quote, type closing quote",
			line:       "\"foo",
			cursor:     4,
			key:        '"',
			wantLine:   "\"foo",  // Should NOT insert pair
			wantSkip:   false,    // selfInsert will insert '"' -> "foo"
			wantCursor: 4,
		},
		{
			name:       "Balanced quotes, type new quote",
			line:       "\"foo\"",
			cursor:     5,
			key:        '"',
			wantLine:   "\"foo\"\"", // Inserts closer
			wantSkip:   false,
			wantCursor: 5,
		},
        {
            name:       "Escaped quote inside double, type quote",
            line:       "\"foo \\\"",
            cursor:     7,
            key:        '"',
            wantLine:   "\"foo \\\"", // Should detect unclosed and NOT insert pair
            wantSkip:   false,
            wantCursor: 7,
        },
        {
            name:       "Jump over closing quote",
            line:       "\"foo\"",
            cursor:     4, // before last "
            key:        '"',
            wantLine:   "\"foo\"",
            wantSkip:   true,
            wantCursor: 5, // Inc
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := core.Line([]rune(tt.line))
			cur := core.NewCursor(&line)
			cur.Set(tt.cursor)

			skip := AutopairInsertOrJump(tt.key, &line, cur)

			if skip != tt.wantSkip {
				t.Errorf("AutopairInsertOrJump() skip = %v, want %v", skip, tt.wantSkip)
			}

			if string(line) != tt.wantLine {
				t.Errorf("AutopairInsertOrJump() line = %q, want %q", string(line), tt.wantLine)
			}
            
            if cur.Pos() != tt.wantCursor {
                t.Errorf("AutopairInsertOrJump() cursor = %v, want %v", cur.Pos(), tt.wantCursor)
            }
		})
	}
}
