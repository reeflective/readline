//go:build unix

package readline

// This file provides a PTY-backed, virtual-terminal test harness for the shell.
//
// Why a subprocess under a real PTY (instead of swapping os.Stdin/os.Stdout
// in-process)? The library writes raw escape sequences directly to os.Stdout
// and the internal/term package captures the stdout/stderr file handles in its
// init(), before any test could redirect them. Running the shell in a child
// process that is given the PTY as its real std{in,out,err} is therefore the
// faithful way to exercise the true byte stream a terminal would see.
//
// The harness feeds the child's output into a vt10x virtual terminal so tests
// can assert on the *rendered* screen, and it auto-responds to cursor-position
// (DSR "ESC[6n") queries so that GetCursorPos() does not block forever
// consuming our keystrokes (see internal/core/keys_unix.go).

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/hinshun/vt10x"
)

const (
	childEnvVar  = "READLINE_PTY_CHILD"
	promptEnvVar = "READLINE_PTY_PROMPT"
)

// TestMain lets this test binary double as the process-under-test: when the
// harness re-execs us with READLINE_PTY_CHILD=1, we run a minimal readline app
// instead of the test suite.
func TestMain(m *testing.M) {
	if os.Getenv(childEnvVar) == "1" {
		runPTYChild()
		return
	}

	os.Exit(m.Run())
}

// runPTYChild runs one Readline() round with a deterministic prompt, then
// prints the accepted line (or error) wrapped in markers the harness matches on.
func runPTYChild() {
	rl := NewShell()

	prompt := os.Getenv(promptEnvVar)
	if prompt == "" {
		prompt = "> "
	}

	rl.Prompt.Primary(func() string { return prompt })

	line, err := rl.Readline()
	if err != nil {
		fmt.Fprintf(os.Stdout, "\r\n[ERR:%s]\r\n", err)
		os.Exit(0)
	}

	fmt.Fprintf(os.Stdout, "\r\n[LINE:%s]\r\n", line)
	os.Exit(0)
}

// console drives a child shell over a PTY and mirrors its output into a vt10x
// virtual terminal for screen assertions.
type console struct {
	t    *testing.T
	cmd  *exec.Cmd
	ptmx *os.File
	term vt10x.Terminal

	mu   sync.Mutex // guards term (writer + reads) and serialises DSR replies
	done chan struct{}
}

// newConsole spawns the child shell under a PTY of the given size, with the
// given primary prompt, and starts mirroring its output into the emulator.
func newConsole(t *testing.T, prompt string, cols, rows int) *console {
	t.Helper()

	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(),
		childEnvVar+"=1",
		promptEnvVar+"="+prompt,
		"INPUTRC=/dev/null", // don't pick up a developer's ~/.inputrc
		"TERM=xterm-256color",
	)

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
	if err != nil {
		t.Fatalf("start child under pty: %v", err)
	}

	c := &console{
		t:    t,
		cmd:  cmd,
		ptmx: ptmx,
		term: vt10x.New(vt10x.WithSize(cols, rows)),
		done: make(chan struct{}),
	}

	go c.readLoop()

	t.Cleanup(c.close)

	return c
}

// readLoop copies child output into the emulator and answers DSR queries.
func (c *console) readLoop() {
	defer close(c.done)

	buf := make([]byte, 4096)
	for {
		n, err := c.ptmx.Read(buf)
		if n > 0 {
			chunk := buf[:n]

			c.mu.Lock()
			_, _ = c.term.Write(chunk)
			c.mu.Unlock()

			// Respond to "ESC[6n" cursor-position reports. The prompt bytes
			// preceding the query are already applied above, so the emulator's
			// cursor reflects the true start-of-line column.
			//
			// NOTE: this assumes the query is not split across read boundaries,
			// which holds in practice because the shell emits prompt+query in
			// one write. A buffering scanner can be added if that changes.
			if bytes.Contains(chunk, []byte("\x1b[6n")) {
				c.replyCursorPos()
			}
		}

		if err != nil {
			return
		}
	}
}

// replyCursorPos writes a DSR cursor-position report (1-based row;col).
func (c *console) replyCursorPos() {
	c.mu.Lock()
	cur := c.term.Cursor()
	c.mu.Unlock()

	fmt.Fprintf(c.ptmx, "\x1b[%d;%dR", cur.Y+1, cur.X+1)
}

// send writes raw bytes to the child as if typed by the user.
func (c *console) send(s string) {
	c.t.Helper()

	if _, err := c.ptmx.WriteString(s); err != nil {
		c.t.Fatalf("send %q: %v", s, err)
	}
}

// screen returns the current rendered contents of the virtual terminal.
func (c *console) screen() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.term.String()
}

// waitForScreen polls until the rendered screen contains substr, or fails the
// test on timeout. It returns the (last) screen contents either way.
func (c *console) waitForScreen(substr string, timeout time.Duration) string {
	c.t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		s := c.screen()
		if strings.Contains(s, substr) {
			return s
		}

		if time.Now().After(deadline) {
			c.t.Fatalf("timed out waiting for %q on screen; got:\n%s", substr, s)
			return s
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// close tears the child and PTY down. Registered via t.Cleanup.
func (c *console) close() {
	_ = c.ptmx.Close()
	_, _ = c.cmd.Process.Wait()
	<-c.done
}
