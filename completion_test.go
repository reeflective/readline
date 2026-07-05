package readline

import (
	"strings"
	"testing"
)

// TestCommandCompletionRecoversFromPanic ensures a panic in the
// application-provided completer is recovered and surfaced as a completion
// message instead of propagating and crashing the shell. readline calls the
// completer on nearly every keystroke, so a single bad completer or transient
// state must not take down the process.
func TestCommandCompletionRecoversFromPanic(t *testing.T) {
	rl := NewShell()
	rl.Completer = func([]rune, int) Completions {
		panic("boom")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("commandCompletion did not recover, panicked with: %v", r)
		}
	}()

	values := rl.commandCompletion()

	if values.Messages.IsEmpty() {
		t.Fatal("recovered completion produced no message, want a completion error message")
	}

	var found bool
	for _, msg := range values.Messages.Get() {
		if strings.Contains(msg, "completion error") && strings.Contains(msg, "boom") {
			found = true
		}
	}

	if !found {
		t.Fatalf("recovered messages = %v, want one containing the panic value", values.Messages.Get())
	}
}

// TestCommandCompletionNilCompleter verifies the no-completer case stays a clean
// no-op (empty values, no message).
func TestCommandCompletionNilCompleter(t *testing.T) {
	rl := NewShell()
	rl.Completer = nil

	values := rl.commandCompletion()

	if !values.Messages.IsEmpty() {
		t.Fatalf("nil completer produced messages %v, want none", values.Messages.Get())
	}
}
