//go:build unix

package display_test

import (
	"strings"
	"testing"
)

// rowIndex returns the index of the first screen row containing substr, or -1.
func rowIndex(screen, substr string) int {
	for i, row := range strings.Split(screen, "\n") {
		if strings.Contains(row, substr) {
			return i
		}
	}

	return -1
}

// TestHintProviderTracksLine checks that a registered passive hint provider is
// re-evaluated as the input changes: the provided lane echoes the current line.
func TestHintProviderTracksLine(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:       "PROMPT> ",
		cols:         80,
		rows:         24,
		hintProvider: true,
	})
	c.waitForScreen("PROMPT>")

	c.send("abc")
	c.waitForScreen("HINT:abc")

	// As the line grows, the provided hint tracks it.
	c.send("d")
	c.waitForScreen("HINT:abcd")
}

// TestHintProviderAboveTransient verifies lane precedence: the passive provider
// hint renders strictly above the transient (async status) hint.
func TestHintProviderAboveTransient(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:       "PROMPT> ",
		cols:         80,
		rows:         24,
		hintProvider: true,
		transient:    "ASYNCMSG",
	})
	c.waitForScreen("PROMPT>")

	c.send("abc")
	screen := c.waitUntil(func(s string) bool {
		return strings.Contains(s, "HINT:abc") && strings.Contains(s, "ASYNCMSG")
	})

	provided := rowIndex(screen, "HINT:abc")
	transient := rowIndex(screen, "ASYNCMSG")

	if provided < 0 || transient < 0 {
		t.Fatalf("both hints must be present (provided=%d transient=%d):\n%s",
			provided, transient, screen)
	}

	if provided >= transient {
		t.Fatalf("provider hint (row %d) must render above transient hint (row %d):\n%s",
			provided, transient, screen)
	}
}

// TestTransientSurvivesIsearchAndRendersAbove checks the two guarantees that
// make the transient lane useful for async status: entering incremental search
// (which owns the completion/text lane) must NOT clear the transient hint, and
// the transient hint must render above that completion-lane hint.
func TestTransientSurvivesIsearchAndRendersAbove(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:    "PROMPT> ",
		cols:      80,
		rows:      24,
		transient: "ASYNCMSG",
	})
	c.waitForScreen("ASYNCMSG")

	// Ctrl-R enters incremental search, which sets the completion/text lane.
	c.send("\x12")
	screen := c.waitForScreen("inc-search")

	// The transient hint must have survived the isearch taking over the
	// completion lane.
	transient := rowIndex(screen, "ASYNCMSG")
	isearch := rowIndex(screen, "inc-search")

	if transient < 0 {
		t.Fatalf("transient hint was lost when entering incremental search:\n%s", screen)
	}

	if transient >= isearch {
		t.Fatalf("transient hint (row %d) must render above the isearch hint (row %d):\n%s",
			transient, isearch, screen)
	}
}

// TestAsyncTransientWhileIdle proves the async-refresh wake: a transient hint
// pushed from another goroutine while the shell is idle (blocked waiting for
// input) must appear WITHOUT any keystroke being sent.
func TestAsyncTransientWhileIdle(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:  "PROMPT> ",
		cols:    80,
		rows:    24,
		asyncMS: 200,
	})
	c.waitForScreen("PROMPT>")

	// Deliberately send NO input: the only thing that can make this appear is
	// the wake repainting the idle loop.
	c.waitForScreen("ASYNCPING")
}

// TestAsyncRefreshKeepsLayout guards against the async repaint corrupting the
// render bookkeeping ("prints a mess / jumps around"): after an idle async
// repaint, the prompt must still render exactly once and accept aligned input.
func TestAsyncRefreshKeepsLayout(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:  "PROMPT> ",
		cols:    80,
		rows:    24,
		asyncMS: 200,
	})
	c.waitForScreen("ASYNCPING")

	// Now type after the async repaint and confirm the line is intact.
	c.send("abc")
	screen := c.waitForScreen("PROMPT> abc")

	if got := countLine(screen, "PROMPT>"); got != 1 {
		t.Fatalf("prompt should render exactly once after async repaint, got %d:\n%s", got, screen)
	}

	first := strings.TrimRight(strings.SplitN(screen, "\n", 2)[0], " ")
	if first != "PROMPT> abc" {
		t.Fatalf("input misaligned after async repaint:\n  got:  %q\n  want: %q", first, "PROMPT> abc")
	}
}
