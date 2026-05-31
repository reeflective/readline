//go:build unix

package display_test

import (
	"strings"
	"testing"
	"time"
)

// TestAsyncRefreshCompletions proves #99 on the async-refresh rail: completions
// produced asynchronously (a background goroutine grows the result set and
// calls Shell.RefreshCompletions) rebuild an already-open menu in place, with
// no keystroke from the user.
func TestAsyncRefreshCompletions(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:    "P> ",
		cols:      80,
		rows:      24,
		asyncComp: true,
		asyncMS:   400,
	})
	c.waitForScreen("P>")

	// Open the completion menu (possible-completions): it displays the initial
	// candidates, and "alpaca" is not among them yet.
	c.send("\x1b?")
	screen := c.waitForScreen("alpha")

	if strings.Contains(screen, "alpaca") {
		t.Fatalf("alpaca should not be present before the async refresh:\n%s", screen)
	}

	// Send NO further input: the async producer grows the result set and calls
	// RefreshCompletions, which must regenerate the open menu in place so the
	// new candidate appears on its own.
	c.waitForScreen("alpaca")
}

// TestAsyncRefreshCompletionsNoMenuIsNoop verifies RefreshCompletions is a clean
// no-op when no menu is active: it must not spontaneously open one.
func TestAsyncRefreshCompletionsNoMenuIsNoop(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:    "P> ",
		cols:      80,
		rows:      24,
		asyncComp: true,
		asyncMS:   300,
	})
	c.waitForScreen("P>")

	// Do not open a menu. Wait past the async RefreshCompletions call; with no
	// active menu it must do nothing, so no candidate ever appears.
	time.Sleep(900 * time.Millisecond)

	screen := c.screen()
	if strings.Contains(screen, "alpha") || strings.Contains(screen, "alpaca") {
		t.Fatalf("RefreshCompletions must no-op when no menu is active, but a menu appeared:\n%s", screen)
	}
}

// TestAsyncRefreshKeepsSelection verifies #99 option B for an explicit menu:
// when a candidate is selected, an async RefreshCompletions both (a) updates the
// menu in place with newly produced candidates and (b) keeps the selection on
// the same candidate (matched by content), rather than freezing the menu or
// dropping the selection. The new candidate sorts first, so the grid re-sorts —
// proving the restore is content-based, not positional.
func TestAsyncRefreshKeepsSelection(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:    "P> ",
		cols:      80,
		rows:      24,
		asyncComp: true,
		asyncMS:   1500, // fire after we have selected a candidate
	})
	c.waitForScreen("P>")

	// Show the menu (no selection yet) and wait for the initial candidates.
	c.send("\x1b?")
	c.waitForScreen("alpha")

	// Select a candidate by navigating down to "alpine" (virtual insertion shows
	// it in the input line).
	c.send("\x1b[B") // select alpha
	c.send("\x1b[B") // select alpine
	screen := c.waitForScreen("P> alpine")

	if strings.Contains(screen, "alpaca") {
		t.Fatalf("alpaca should not be present before the async refresh:\n%s", screen)
	}

	// The async producer adds "alpaca" (which sorts first) and calls
	// RefreshCompletions while "alpine" is selected: the menu must update AND
	// the selection must persist.
	c.waitForScreen("alpaca")

	final := c.screen()
	first := strings.TrimRight(strings.SplitN(final, "\n", 2)[0], " ")

	if first != "P> alpine" {
		t.Fatalf("selection not preserved across async refresh: first line = %q want %q\n%s",
			first, "P> alpine", final)
	}
}
