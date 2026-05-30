//go:build unix

package readline

import (
	"strings"
	"testing"
	"time"
)

// TestPTYPromptAndInput is the first end-to-end golden-screen test: it boots a
// shell under a PTY, checks the prompt renders, types a line, checks it appears
// next to the prompt, then submits it and checks it is accepted and returned.
func TestPTYPromptAndInput(t *testing.T) {
	c := newConsole(t, "PROMPT> ", 80, 24)

	// 1. The primary prompt should render.
	c.waitForScreen("PROMPT>", 3*time.Second)

	// 2. Typed input should appear immediately after the prompt.
	c.send("hello world")
	screen := c.waitForScreen("PROMPT> hello world", 3*time.Second)

	// The input line must be rendered on the first row, flush against the
	// prompt with no misalignment (a regression guard for issue #98).
	firstLine := strings.SplitN(screen, "\n", 2)[0]
	if got, want := strings.TrimRight(firstLine, " "), "PROMPT> hello world"; got != want {
		t.Fatalf("first line misaligned:\n  got:  %q\n  want: %q", got, want)
	}

	// 3. Submitting the line should accept it and return it to the caller.
	c.send("\r")
	c.waitForScreen("[LINE:hello world]", 3*time.Second)
}
