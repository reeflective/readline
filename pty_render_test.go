//go:build unix

package readline

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

// TestPTYWrappingAlignment is a regression guard for the rendering engine: with
// a well-behaved terminal, a line that wraps onto a second row must render with
// the prompt exactly once and the input flush against it. This passes today and
// locks in correct multi-row layout.
func TestPTYWrappingAlignment(t *testing.T) {
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

// TestRepro98MultilinePromptAtBottom reproduces issue #98 (and its downstream
// report reeflective/console#78) with a truthful terminal — the representative
// real-world trigger.
//
// When a MULTI-LINE primary prompt is rendered at the bottom of the terminal
// window, the engine does not scroll the screen up to reserve the prompt's own
// height before printing it. The terminal then scrolls underneath the prompt,
// the row bookkeeping (display.Engine.startRows / cursorRow) is left off by the
// number of un-reserved lines, and the prompt's last line (the input line) is
// pushed off the bottom and lost — the prompt "overlaps with the input line"
// exactly as console#78 describes. (Ctrl-L fixes it there because clear-screen
// reprints the prompt from a clean top, resetting startRows.)
//
// ensureInputSpace() (internal/display/refresh.go) reserves space for the input
// rows (e.lineRows) but never for the prompt rows (prompt.PrimaryUsed()). The
// fix should reserve primaryRows+lineRows before printing near the bottom.
//
// The assertion encodes the DESIRED behavior (every prompt line survives). It
// currently FAILS, so it is gated behind RUN_REPRO98 to keep CI green. Remove
// the skip once the fix lands; this then becomes a permanent regression test.
func TestRepro98MultilinePromptAtBottom(t *testing.T) {
	if os.Getenv("RUN_REPRO98") == "" {
		t.Skip("reproduces #98 / console#78 (currently red): run with RUN_REPRO98=1; " +
			"unskip when the bottom-of-window multi-line prompt fix lands.")
	}

	const rows = 10

	// A 4-line prompt whose last line is the input line, placed on the very
	// last row of the window (prefill pushes it down). The bug shows on the
	// initial render — no input needed.
	c := startConsole(t, consoleConfig{
		prompt:  "L1\nL2\nL3\nL4> ",
		cols:    20,
		rows:    rows,
		prefill: rows - 1,
	})

	// Wait for the top of the prompt to render, then let it settle.
	c.waitForScreen("L1", 3*time.Second)
	time.Sleep(250 * time.Millisecond)

	screen := c.screen()

	// Every line of the prompt — crucially the last/input line — must survive.
	for _, line := range []string{"L1", "L2", "L3", "L4>"} {
		if !strings.Contains(screen, line) {
			t.Fatalf("#98: prompt line %q was lost when rendering at the bottom of the window:\n%s",
				line, screen)
		}
	}
}

// TestRepro98ProbeMisalignment is a second facet of #98: the same row/column
// bookkeeping is corrupted when the terminal reports a wrong cursor position
// for an "ESC[6n" query (slow / racing / quirky terminals — the fragility
// tracked by #101). Here we simulate a terminal that always reports column 1,
// type a wrapping line and edit it; the prompt ends up redrawn on every row.
//
// Gated like the test above; remove the skip when probing can be disabled or
// startCols is sanity-checked (#101).
func TestRepro98ProbeMisalignment(t *testing.T) {
	if os.Getenv("RUN_REPRO98") == "" {
		t.Skip("reproduces #98 (currently red): run with RUN_REPRO98=1; " +
			"unskip when #101 (disable-probe / robust startCols) lands.")
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
		t.Fatalf("#98: prompt rendered %d times (want 1) under a wrong cursor probe:\n%s",
			got, screen)
	}
}
