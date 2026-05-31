//go:build unix

package display_test

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestNoPromptDriftWithHint reproduces a reported bug: when a hint is displayed
// below the input line, every refresh repositioned the prompt one row too high,
// so the whole UI crept upward by one row per keystroke. The anchor (top) line
// of a multi-line prompt must stay on the same screen row as the user types.
func TestNoPromptDriftWithHint(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:    "ANCHOR\nP> ",
		cols:      80,
		rows:      24,
		transient: "STATUS-HINT",
	})
	c.waitForScreen("ANCHOR")
	c.waitForScreen("STATUS-HINT")

	chars := []string{"a", "b", "c", "d"}
	rows := make([]int, 0, len(chars))

	for _, ch := range chars {
		c.send(ch)
		time.Sleep(150 * time.Millisecond)

		screen := c.screen()
		rows = append(rows, rowIndex(screen, "ANCHOR"))
	}

	for i, r := range rows {
		if r != rows[0] {
			t.Fatalf("prompt drifted: ANCHOR row per keystroke = %v (moved at step %d)\nlast screen:\n%s",
				rows, i, c.screen())
		}
	}
}

// TestNoPromptDriftWithHintAndAutocomplete is the compound case from the report:
// autocomplete menu + hint, multi-line prompt. The anchor line must not drift.
func TestNoPromptDriftWithHintAndAutocomplete(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:       "ANCHOR\nP> ",
		cols:         80,
		rows:         24,
		autocomplete: true,
		hintProvider: true,
	})
	c.waitForScreen("ANCHOR")

	chars := strings.Split("alp", "")
	rows := make([]int, 0, len(chars))

	for _, ch := range chars {
		c.send(ch)
		time.Sleep(200 * time.Millisecond)

		screen := c.screen()
		rows = append(rows, rowIndex(screen, "ANCHOR"))
	}

	for i, r := range rows {
		if r != rows[0] {
			t.Fatalf("prompt drifted with autocomplete: ANCHOR row per keystroke = %v (moved at step %d)\nlast screen:\n%s",
				rows, i, c.screen())
		}
	}
}

// TestNoPromptDriftAcrossAsyncWakes reproduces the core bug: each async-refresh
// wake (e.g. a transient hint pushed from another goroutine) repainted the UI
// one row too high, so repeated async updates crept the prompt upward — even
// with no keystroke. The anchor line must stay put across many wakes.
func TestNoPromptDriftAcrossAsyncWakes(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:      "ANCHOR\nP> ",
		cols:        80,
		rows:        24,
		asyncMS:     150,
		asyncRepeat: 5,
	})
	c.waitForScreen("ANCHOR")

	rows := make([]int, 0, 5)

	for i := range 5 {
		c.waitForScreen(fmt.Sprintf("ASYNCPING-%d", i))
		rows = append(rows, rowIndex(c.screen(), "ANCHOR"))
	}

	for i, r := range rows {
		if r != rows[0] {
			t.Fatalf("prompt drifted across async wakes: ANCHOR row per update = %v (moved at update %d)\nlast screen:\n%s",
				rows, i, c.screen())
		}
	}
}

// TestNoPromptDriftAsyncWakesWithRightPrompt is the closest match to the example
// console: a multi-line prompt with a right-side prompt (reaching the far-right
// column), updated by repeated async transient hints.
func TestNoPromptDriftAsyncWakesWithRightPrompt(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:      "ANCHOR\nP> ",
		cols:        80,
		rows:        24,
		rightPrompt: "12:00:00.000",
		asyncMS:     150,
		asyncRepeat: 5,
	})
	c.waitForScreen("ANCHOR")

	rows := make([]int, 0, 5)

	for i := range 5 {
		c.waitForScreen(fmt.Sprintf("ASYNCPING-%d", i))
		rows = append(rows, rowIndex(c.screen(), "ANCHOR"))
	}

	for i, r := range rows {
		if r != rows[0] {
			t.Fatalf("prompt drifted across async wakes (with right prompt): ANCHOR row = %v (moved at %d)\n%s",
				rows, i, c.screen())
		}
	}
}

// TestNoPromptDriftTwoHintLanes reproduces the root cause directly: with TWO
// active hint lanes (a provider hint plus an async transient), the hint-row
// count was over-counted, so each refresh crept the prompt up one row.
func TestNoPromptDriftTwoHintLanes(t *testing.T) {
	c := startConsole(t, consoleConfig{
		prompt:       "ANCHOR\nP> ",
		cols:         80,
		rows:         24,
		hintProvider: true, // provider hint = lane 1 (once the line is non-empty)
		asyncMS:      150,
		asyncRepeat:  5, // transient hint = lane 2
	})
	c.waitForScreen("ANCHOR")

	c.send("x") // make the provider hint non-empty
	c.waitForScreen("HINT:x")

	rows := make([]int, 0, 5)

	for i := range 5 {
		c.waitForScreen(fmt.Sprintf("ASYNCPING-%d", i))
		rows = append(rows, rowIndex(c.screen(), "ANCHOR"))
	}

	for i, r := range rows {
		if r != rows[0] {
			t.Fatalf("prompt drifted with two hint lanes: ANCHOR row = %v (moved at %d)\n%s",
				rows, i, c.screen())
		}
	}
}
