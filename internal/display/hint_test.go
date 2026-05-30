//go:build unix

package display_test

import (
	"strings"
	"testing"
	"time"
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
	c.waitForScreen("PROMPT>", 3*time.Second)

	c.send("abc")
	c.waitForScreen("HINT:abc", 3*time.Second)

	// As the line grows, the provided hint tracks it.
	c.send("d")
	c.waitForScreen("HINT:abcd", 3*time.Second)
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
	c.waitForScreen("PROMPT>", 3*time.Second)

	c.send("abc")
	screen := c.waitUntil(func(s string) bool {
		return strings.Contains(s, "HINT:abc") && strings.Contains(s, "ASYNCMSG")
	}, 3*time.Second)

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
	c.waitForScreen("ASYNCMSG", 3*time.Second)

	// Ctrl-R enters incremental search, which sets the completion/text lane.
	c.send("\x12")
	screen := c.waitForScreen("inc-search", 3*time.Second)

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
