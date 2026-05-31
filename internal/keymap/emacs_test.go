package keymap

import (
	"testing"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/core"
)

// TestEmacsBackspaceDeletesChar guards the \C-h binding: on many terminals
// Backspace sends ^H, so \C-h must resolve to backward-delete-char (the GNU
// default), not backward-kill-word, which would delete a whole word on every
// Backspace. The emacsKeys overlay must therefore NOT override the default.
func TestEmacsBackspaceDeletesChar(t *testing.T) {
	keys := &core.Keys{}
	iters := &core.Iterations{}

	_, cfg := NewEngine(keys, iters)

	got := cfg.Binds[string(Emacs)][inputrc.Unescape(`\C-h`)].Action

	if got == "backward-kill-word" {
		t.Fatal(`\C-h is bound to backward-kill-word: Backspace would delete a whole word on terminals that send ^H`)
	}

	if got != "backward-delete-char" {
		t.Fatalf(`\C-h resolved to %q, want "backward-delete-char"`, got)
	}
}
