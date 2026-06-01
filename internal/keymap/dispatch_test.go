package keymap

import (
	"testing"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/strutil"
)

// builtinBindMaps returns every keymap exactly as the dispatcher sees it after
// a default engine is built (GNU defaults + this library's builtin overlays).
func builtinBindMaps(t *testing.T) map[string]map[string]inputrc.Bind {
	t.Helper()

	_, cfg := NewEngine(&core.Keys{}, &core.Iterations{})

	return cfg.Binds
}

// collisions returns, for a keymap, the ConvertMeta key bytes that more than one
// sequence maps to with DIFFERING binds — inputs whose exact match would depend
// on iteration order if matchBind did not impose a deterministic order.
func collisions(binds map[string]inputrc.Bind) map[string][]inputrc.Bind {
	bySeq := make(map[string][]inputrc.Bind)

	for sequence, bind := range binds {
		seq := strutil.ConvertMeta([]rune(sequence))
		bySeq[seq] = append(bySeq[seq], bind)
	}

	out := make(map[string][]inputrc.Bind)

	for seq, list := range bySeq {
		differ := false

		for _, b := range list[1:] {
			if b != list[0] {
				differ = true

				break
			}
		}

		if differ {
			out[seq] = list
		}
	}

	return out
}

// TestMatchBindSortIsLoadBearing documents why matchBind must impose a
// deterministic order: the default keymaps DO contain sequences that collapse
// (via ConvertMeta) to the same key bytes with different actions, e.g. ESC-prefixed
// meta bindings overlapping a self-insert. With those collisions present, dropping
// the ordering (as a naive optimization would) makes the exact `match` depend on
// Go's randomized map iteration — i.e. nondeterministic keybind resolution. If
// this ever reports zero collisions, the determinism requirement can be revisited.
func TestMatchBindSortIsLoadBearing(t *testing.T) {
	binds := builtinBindMaps(t)[string(Emacs)]

	cols := collisions(binds)
	if len(cols) == 0 {
		t.Fatal("expected ConvertMeta collisions in the emacs keymap; if truly none, matchBind no longer needs to order its matches")
	}

	t.Logf("emacs keymap has %d colliding key sequences whose exact match is disambiguated only by matchBind's ordering", len(cols))
}

// TestMatchBindResolvesDeterministically guards the property the ordering buys:
// for a colliding input, matchBind must return the SAME exact match on every
// call despite map iteration randomness. Runs many times so a regression to
// unordered iteration would be caught.
func TestMatchBindResolvesDeterministically(t *testing.T) {
	eng := &Engine{}
	binds := builtinBindMaps(t)[string(Emacs)]

	cols := collisions(binds)
	if len(cols) == 0 {
		t.Skip("no collisions to probe")
	}

	// Pick one colliding input deterministically (smallest key bytes).
	var probe string

	for seq := range cols {
		if probe == "" || seq < probe {
			probe = seq
		}
	}

	want, _ := eng.matchBind([]byte(probe), binds)

	for range 64 {
		got, _ := eng.matchBind([]byte(probe), binds)
		if got != want {
			t.Fatalf("matchBind(%q) is nondeterministic: got %+v, first saw %+v", probe, got, want)
		}
	}
}
