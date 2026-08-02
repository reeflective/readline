//go:build unix

package core

import (
	"os"
	"testing"

	"github.com/creack/pty"

	"github.com/reeflective/readline/internal/term"
)

func TestGetCursorPosPreservesCoalescedInput(t *testing.T) {
	ptmx, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("pty: %v", err)
	}
	defer ptmx.Close()
	defer tty.Close()

	// Raw mode, as Readline always sets before the probe runs — in the
	// slave's default canonical mode a read would block waiting for \n.
	if _, err := term.MakeRaw(int(tty.Fd())); err != nil {
		t.Fatalf("raw mode: %v", err)
	}

	origStdin := os.Stdin
	os.Stdin = tty
	defer func() { os.Stdin = origStdin }()

	// The typed line and the cursor answer arrive as one chunk: written
	// before the probe reads, they sit together in the tty input queue.
	if _, err := ptmx.Write([]byte("echo hello\r\x1b[5;10R")); err != nil {
		t.Fatalf("write: %v", err)
	}

	keys := &Keys{}

	x, y := keys.GetCursorPos()
	if x != 10 || y != 5 {
		t.Errorf("cursor position = (%d, %d), want (10, 5)", x, y)
	}
	if got := string(keys.buf); got != "echo hello\r" {
		t.Errorf("pending input = %q, want %q (typed bytes must survive the probe)", got, "echo hello\r")
	}
}
