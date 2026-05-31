//go:build unix

package display_test

import (
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

	c.waitForScreen("PROMPT>")

	c.send("hello world")
	screen := c.waitForScreen("PROMPT> hello world")

	firstLine := strings.SplitN(screen, "\n", 2)[0]
	if got, want := strings.TrimRight(firstLine, " "), "PROMPT> hello world"; got != want {
		t.Fatalf("first line misaligned:\n  got:  %q\n  want: %q", got, want)
	}

	c.send("\r")
	c.waitForScreen("[LINE:hello world]")
}

// TestRenderWrappingAlignment guards multi-row layout: with a well-behaved
// terminal, a line that wraps onto a second row renders with the prompt exactly
// once and the input flush against it.
func TestRenderWrappingAlignment(t *testing.T) {
	// 40-col terminal, 8-col prompt -> 32 text columns on the first row.
	c := newConsole(t, "PROMPT> ", 40, 24)
	c.waitForScreen("PROMPT>")

	// Type 40 'x': 32 land on row 0 (after the prompt), 8 wrap to row 1.
	c.send(strings.Repeat("x", 40))
	screen := c.waitUntil(func(s string) bool {
		return strings.Count(s, "x") >= 40
	})

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

	c.waitForScreen("L1")
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

// TestRenderMultilinePromptWrapAtBottom guards the harder compound case: a
// multi-line prompt at the bottom of the window whose input also wraps onto
// extra rows. The prompt's upper lines (printed once, not otherwise refreshed)
// must survive the scrolling instead of being lost.
func TestRenderMultilinePromptWrapAtBottom(t *testing.T) {
	const rows = 10

	c := startConsole(t, consoleConfig{
		prompt:  "TOP\nP> ",
		cols:    20,
		rows:    rows,
		prefill: rows - 1,
	})

	c.waitForScreen("P>")

	// Type enough to wrap the input onto a second/third visual row.
	c.send(strings.Repeat("y", 50))
	screen := c.waitUntil(func(s string) bool {
		return strings.Count(s, "y") >= 50
	})

	for _, line := range []string{"TOP", "P> "} {
		if !strings.Contains(screen, line) {
			t.Fatalf("prompt line %q was lost when a multi-line prompt wrapped at the bottom:\n%s",
				line, screen)
		}
	}
}

// TestCursorProbeEnabledByDefault confirms probing is on by default: a normal
// render issues at least one "ESC[6n" cursor-position query.
func TestCursorProbeEnabledByDefault(t *testing.T) {
	c := newConsole(t, "PROMPT> ", 80, 24)
	c.waitForScreen("PROMPT>")

	// The prompt string is printed before the first probe, so poll for the
	// query rather than checking immediately after the prompt appears.
	deadline := time.Now().Add(3 * time.Second)
	for c.probeQueries() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if c.probeQueries() == 0 {
		t.Fatal("expected at least one ESC[6n cursor-position query with probing enabled")
	}
}

// TestRenderWithCursorProbeDisabled verifies the cursor-position-probe option:
// when turned off, the shell sends no "ESC[6n" query and renders correctly from
// the printed prompt width alone — even against a terminal that would have lied
// about the cursor position. This is the supported fallback for PTY harnesses
// and constrained terminals (#101).
func TestRenderWithCursorProbeDisabled(t *testing.T) {
	// This responder would report a bogus column; with probing disabled it must
	// never be consulted.
	lying := func(vt10x.Cursor) string { return "\x1b[1;1R" }

	c := startConsole(t, consoleConfig{
		prompt:     "PROMPT> ",
		cols:       40,
		rows:       24,
		noProbe:    true,
		probeReply: lying,
	})
	c.waitForScreen("PROMPT>")

	c.send(strings.Repeat("x", 40)) // wrap to a second row
	screen := c.waitUntil(func(s string) bool {
		return strings.Count(s, "x") >= 40
	})

	if got := c.probeQueries(); got != 0 {
		t.Fatalf("expected no ESC[6n queries when probing is disabled, got %d", got)
	}

	if got := countLine(screen, "PROMPT>"); got != 1 {
		t.Fatalf("prompt should render exactly once with probing disabled, got %d:\n%s", got, screen)
	}

	row0 := strings.TrimRight(strings.SplitN(screen, "\n", 2)[0], " ")
	if want := "PROMPT> " + strings.Repeat("x", 32); row0 != want {
		t.Fatalf("row 0 misaligned with probing disabled:\n  got:  %q\n  want: %q", row0, want)
	}
}
