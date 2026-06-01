package readline

import (
	"testing"

	"github.com/reeflective/readline/internal/core"
)

// feedPaste builds a minimal shell, feeds a bracketed-paste payload terminated
// by the end sequence, runs the paste handler, and returns the resulting line.
func feedPaste(t *testing.T, payload string) string {
	t.Helper()

	line := new(core.Line)
	rl := &Shell{
		Keys:   new(core.Keys),
		line:   line,
		cursor: core.NewCursor(line),
	}

	// The handler consumes keys until it sees the paste-end sequence, so the
	// terminator must be fed too — otherwise it would block waiting for input.
	rl.Keys.Feed(false, []rune(payload+"\x1b[201~")...)
	rl.bracketedPasteBegin()

	return string(*line)
}

// TestBracketedPasteNormalizesCarriageReturns guards that pasted line breaks,
// which terminals deliver as \r or \r\n, are stored as \n. Raw carriage
// returns corrupt the line buffer and its multiline rendering.
func TestBracketedPasteNormalizesCarriageReturns(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		want    string
	}{
		{"crlf", "a\r\nb", "a\nb"},
		{"bare cr", "a\rb", "a\nb"},
		{"mixed", "one\r\ntwo\rthree", "one\ntwo\nthree"},
		{"no newline", "plain", "plain"},
		{"already lf", "a\nb", "a\nb"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := feedPaste(t, c.payload); got != c.want {
				t.Fatalf("paste %q => %q, want %q", c.payload, got, c.want)
			}
		})
	}
}
