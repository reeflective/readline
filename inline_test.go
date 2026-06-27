package readline

import "testing"

func TestInlineSuggestionAPI(t *testing.T) {
	rl := NewShell()

	rl.SetInlineSuggestion("status --verbose")
	if got := rl.GetInlineSuggestion(); got != "status --verbose" {
		t.Fatalf("inline suggestion = %q, want status --verbose", got)
	}

	rl.ClearInlineSuggestion()
	if got := rl.GetInlineSuggestion(); got != "" {
		t.Fatalf("inline suggestion after clear = %q, want empty", got)
	}
}
