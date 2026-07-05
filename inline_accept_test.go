package readline

import "testing"

// setLine replaces the shell's line buffer and puts the cursor at its end.
func setLine(rl *Shell, s string) {
	rl.line.Set([]rune(s)...)
	rl.cursor.Set(rl.line.Len())
}

func TestInlineSuggestAcceptFull(t *testing.T) {
	rl := NewShell()
	setLine(rl, "sta")
	rl.SetInlineSuggestion("status --verbose")

	rl.inlineSuggestAccept()

	if got := string(*rl.line); got != "status --verbose" {
		t.Fatalf("line after accept = %q, want %q", got, "status --verbose")
	}
	if got := rl.GetInlineSuggestion(); got != "" {
		t.Fatalf("inline suggestion after accept = %q, want cleared", got)
	}
}

func TestInlineSuggestAcceptWord(t *testing.T) {
	rl := NewShell()
	setLine(rl, "sta")
	rl.SetInlineSuggestion("status --verbose")

	rl.inlineSuggestAcceptWord()

	if got := string(*rl.line); got != "status" {
		t.Fatalf("line after word accept = %q, want %q", got, "status")
	}
	// Suggestion still extends the line, so it must not be cleared.
	if got := rl.GetInlineSuggestion(); got != "status --verbose" {
		t.Fatalf("inline suggestion after word accept = %q, want unchanged", got)
	}

	// A second word-accept consumes the remainder and clears the suggestion.
	rl.inlineSuggestAcceptWord()
	if got := string(*rl.line); got != "status --verbose" {
		t.Fatalf("line after second word accept = %q, want %q", got, "status --verbose")
	}
	if got := rl.GetInlineSuggestion(); got != "" {
		t.Fatalf("inline suggestion after full consumption = %q, want cleared", got)
	}
}

func TestInlineSuggestAcceptNotApplicable(t *testing.T) {
	rl := NewShell()
	setLine(rl, "xyz") // suggestion does not extend the current line
	rl.SetInlineSuggestion("status")

	rl.inlineSuggestAccept()

	if got := string(*rl.line); got != "xyz" {
		t.Fatalf("line = %q, want unchanged %q", got, "xyz")
	}

	// Cursor not at end of line: must not accept.
	setLine(rl, "sta")
	rl.cursor.Set(1)
	rl.SetInlineSuggestion("status")

	rl.inlineSuggestAccept()

	if got := string(*rl.line); got != "sta" {
		t.Fatalf("line = %q, want unchanged %q (cursor not at end)", got, "sta")
	}
}
