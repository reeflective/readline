package readline

import (
	"testing"

	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/history"
)

// feedPaste builds a minimal shell, feeds a bracketed-paste payload terminated
// by the end sequence, runs the paste handler, and returns the resulting line.
func feedPaste(t *testing.T, payload string) string {
	t.Helper()

	return feedPasteWithTransformer(t, payload, nil)
}

func feedPasteWithTransformer(t *testing.T, payload string, transformer func(string) string) string {
	t.Helper()

	line := new(core.Line)
	rl := &Shell{
		Keys:   new(core.Keys),
		line:   line,
		cursor: core.NewCursor(line),
	}
	rl.PasteTransformer = transformer

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

func TestBracketedPasteTransformer(t *testing.T) {
	got := feedPasteWithTransformer(t, "a\r\nb", func(text string) string {
		if text != "a\nb" {
			t.Fatalf("transformer saw %q, want normalized text", text)
		}

		return "rewritten"
	})

	if got != "rewritten" {
		t.Fatalf("transformed paste = %q, want rewritten", got)
	}
}

func TestBracketedPasteTransformerCanDropText(t *testing.T) {
	got := feedPasteWithTransformer(t, "secret", func(string) string {
		return ""
	})

	if got != "" {
		t.Fatalf("dropped paste = %q, want empty line", got)
	}
}

// drainKeys pops everything remaining in the key buffer and returns it.
func drainKeys(rl *Shell) string {
	var b []byte

	for {
		key, empty := core.PopKey(rl.Keys)
		if empty {
			return string(b)
		}

		b = append(b, key)
	}
}

// TestSkipCsiSequence drives skip-csi-sequence over the bytes that remain after
// the dispatcher has matched the leading "\e[": it must swallow the parameter
// bytes and the single final byte, while leaving any following keystrokes
// untouched.
func TestSkipCsiSequence(t *testing.T) {
	cases := []struct {
		name     string
		feed     string // bytes left in the buffer after "\e[" was matched
		wantLeft string // what must remain unconsumed afterwards
	}{
		{"function key params and final", "15~", ""},
		{"single final byte", "Z", ""},
		{"modified arrow with params", "1;5D", ""},
		{"stops at final, keeps following keys", "15~abc", "abc"},
		{"empty buffer is a no-op", "", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rl := &Shell{Keys: new(core.Keys), History: &history.Sources{}}

			if c.feed != "" {
				rl.Keys.Feed(false, []rune(c.feed)...)
			}

			rl.skipCsiSequence()

			if got := drainKeys(rl); got != c.wantLeft {
				t.Fatalf("skip of %q left %q, want %q", c.feed, got, c.wantLeft)
			}
		})
	}
}
