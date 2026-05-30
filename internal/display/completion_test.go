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
	c.waitForScreen("P>", 3*time.Second)

	// Open the completion menu (possible-completions): it displays the initial
	// candidates, and "charlie" is not among them yet.
	c.send("\x1b?")
	screen := c.waitForScreen("alpha", 3*time.Second)

	if strings.Contains(screen, "charlie") {
		t.Fatalf("charlie should not be present before the async refresh:\n%s", screen)
	}

	// Send NO further input: the async producer grows the result set and calls
	// RefreshCompletions, which must regenerate the open menu in place so the
	// new candidate appears on its own.
	c.waitForScreen("charlie", 3*time.Second)
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
	c.waitForScreen("P>", 3*time.Second)

	// Do not open a menu. Wait past the async RefreshCompletions call; with no
	// active menu it must do nothing, so no candidate ever appears.
	time.Sleep(900 * time.Millisecond)

	screen := c.screen()
	if strings.Contains(screen, "alpha") || strings.Contains(screen, "charlie") {
		t.Fatalf("RefreshCompletions must no-op when no menu is active, but a menu appeared:\n%s", screen)
	}
}
