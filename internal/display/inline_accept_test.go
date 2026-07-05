package display

import (
	"testing"

	"github.com/reeflective/readline/internal/core"
)

func newInlineEngine(t *testing.T, buf, suggestion string) (*Engine, *core.Line, *core.Cursor) {
	t.Helper()

	line := &core.Line{}
	line.Set([]rune(buf)...)
	cursor := core.NewCursor(line)
	cursor.Set(line.Len())

	engine := &Engine{}
	engine.SetInlineSuggestion(suggestion)

	return engine, line, cursor
}

func TestAcceptInlineFull(t *testing.T) {
	engine, line, cursor := newInlineEngine(t, "sta", "status --verbose")

	if !engine.AcceptInline(line, cursor) {
		t.Fatal("AcceptInline returned false, want the suggestion accepted")
	}
	if got := string(*line); got != "status --verbose" {
		t.Fatalf("line = %q, want %q", got, "status --verbose")
	}
	if engine.GetInlineSuggestion() != "" {
		t.Fatal("inline suggestion not cleared after full accept")
	}
}

func TestAcceptInlineWordThenRemainder(t *testing.T) {
	engine, line, cursor := newInlineEngine(t, "sta", "status --verbose")

	engine.AcceptInlineWord(line, cursor)
	if got := string(*line); got != "status" {
		t.Fatalf("line after word accept = %q, want %q", got, "status")
	}
	if engine.GetInlineSuggestion() != "status --verbose" {
		t.Fatal("suggestion cleared too early: it still extends the line")
	}

	engine.AcceptInlineWord(line, cursor)
	if got := string(*line); got != "status --verbose" {
		t.Fatalf("line after second word accept = %q, want %q", got, "status --verbose")
	}
	if engine.GetInlineSuggestion() != "" {
		t.Fatal("inline suggestion not cleared after remainder consumed")
	}
}

func TestAcceptInlineGated(t *testing.T) {
	// Cursor not at line end.
	engine, line, cursor := newInlineEngine(t, "sta", "status")
	cursor.Set(1)
	if engine.AcceptInline(line, cursor) {
		t.Fatal("AcceptInline accepted with cursor away from line end")
	}

	// Suggestion does not extend the current line.
	engine2, line2, cursor2 := newInlineEngine(t, "xyz", "status")
	if engine2.AcceptInline(line2, cursor2) {
		t.Fatal("AcceptInline accepted a non-extending suggestion")
	}
}
