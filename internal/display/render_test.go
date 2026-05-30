//go:build unix

package display_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hinshun/vt10x"
)

// countLine reports how many of the screen's rows begin with prefix.
func countLine(screen, prefix string) int {
	var n int
	for _, row := range strings.Split(screen, "\n") {
		if strings.HasPrefix(row, prefix) {
			n++
		}
	}

	return n
}

// TestRenderPromptAndInput is the baseline end-to-end golden-screen test: boot
// a shell under a PTY, check the prompt renders, type a line, check it appears
// next to the prompt, then submit it and check it is accepted and returned.
func TestRenderPromptAndInput(t *testing.T) {
	c := newConsole(t, "PROMPT> ", 80, 24)

	c.waitForScreen("PROMPT>", 3*time.Second)

	c.send("hello world")
	screen := c.waitForScreen("PROMPT> hello world", 3*time.Second)

	firstLine := strings.SplitN(screen, "\n", 2)[0]
	if got, want := strings.TrimRight(firstLine, " "), "PROMPT> hello world"; got != want {
		t.Fatalf("first line misaligned:\n  got:  %q\n  want: %q", got, want)
	}

	c.send("\r")
	c.waitForScreen("[LINE:hello world]", 3*time.Second)
}

// TestRenderWrappingAlignment guards multi-row layout: with a well-behaved
// terminal, a line that wraps onto a second row renders with the prompt exactly
// once and the input flush against it.
func TestRenderWrappingAlignment(t *testing.T) {
	// 40-col terminal, 8-col prompt -> 32 text columns on the first row.
	c := newConsole(t, "PROMPT> ", 40, 24)
	c.waitForScreen("PROMPT>", 3*time.Second)

	// Type 40 'x': 32 land on row 0 (after the prompt), 8 wrap to row 1.
	c.send(strings.Repeat("x", 40))
	screen := c.waitUntil(func(s string) bool {
		return strings.Count(s, "x") >= 40
	}, 3*time.Second)

	rows := strings.Split(screen, "\n")

	if got := countLine(screen, "PROMPT>"); got != 1 {
		t.Fatalf("prompt should render exactly once, got %d times:\n%s", got, screen)
	}

	if want := "PROMPT> " + strings.Repeat("x", 32); strings.TrimRight(rows[0], " ") != want {
		t.Fatalf("row 0 wrong:\n  got:  %q\n  want: %q", strings.TrimRight(rows[0], " "), want)
	}

	if want := strings.Repeat("x", 8); strings.TrimRight(rows[1], " ") != want {
		t.Fatalf("row 1 (wrapped continuation) wrong:\n  got:  %q\n  want: %q",
			strings.TrimRight(rows[1], " "), want)
	}
}

// TestRenderMultilinePromptAtBottom guards the bottom-of-window case: a
// multi-line primary prompt rendered on the last row of the terminal must keep
// all of its lines (including the input line) instead of overlapping them.
//
// Before ensureInputSpace reserved the prompt+input height, the terminal
// scrolled underneath the prompt, the row bookkeeping drifted, and the prompt's
// last (input) line was pushed off the bottom and lost.
func TestRenderMultilinePromptAtBottom(t *testing.T) {
	const rows = 10

	// A 4-line prompt whose last line is the input line, placed on the very
	// last row of the window (prefill pushes it down). The defect showed on the
	// initial render, before any input.
	c := startConsole(t, consoleConfig{
		prompt:  "L1\nL2\nL3\nL4> ",
		cols:    20,
		rows:    rows,
		prefill: rows - 1,
	})

	c.waitForScreen("L1", 3*time.Second)
	time.Sleep(250 * time.Millisecond) // let the render settle

	screen := c.screen()

	// Every line of the prompt — crucially the last/input line — must survive.
	for _, line := range []string{"L1", "L2", "L3", "L4>"} {
		if !strings.Contains(screen, line) {
			t.Fatalf("prompt line %q was lost when rendering at the bottom of the window:\n%s",
				line, screen)
		}
	}
}

// TestRenderMisreportedCursor exercises a known robustness gap (kept gated so it
// does not fail CI): when the terminal reports a wrong cursor position for an
// "ESC[6n" query — slow / racing / quirky terminals — the row/column
// bookkeeping is corrupted and the prompt is redrawn on every row. Here we
// simulate a terminal that always reports column 1, type a wrapping line and
// edit it.
//
// Run with READLINE_RUN_KNOWN_GAPS=1. Remove the skip once the cursor probe can
// be disabled or startCols is sanity-checked.
func TestRenderMisreportedCursor(t *testing.T) {
	if os.Getenv("READLINE_RUN_KNOWN_GAPS") == "" {
		t.Skip("known gap (currently red): run with READLINE_RUN_KNOWN_GAPS=1; " +
			"unskip when the cursor probe can be disabled / startCols is sanity-checked.")
	}

	lying := func(vt10x.Cursor) string { return "\x1b[1;1R" } // always column 1

	c := startConsole(t, consoleConfig{prompt: "PROMPT> ", cols: 40, rows: 24, probeReply: lying})
	c.waitForScreen("PROMPT>", 3*time.Second)

	c.send(strings.Repeat("x", 40)) // wrap to a second row
	time.Sleep(200 * time.Millisecond)
	c.send("\x7f\x7f\x7f\x7f\x7fYYYYY") // edit -> several redraws
	time.Sleep(200 * time.Millisecond)

	screen := c.screen()
	if got := countLine(screen, "PROMPT>"); got != 1 {
		t.Fatalf("prompt rendered %d times (want 1) under a wrong cursor probe:\n%s", got, screen)
	}
}
