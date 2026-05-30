//go:build unix

package display_test

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
//
// It lives as an external test package (display_test) so it can drive the full
// shell via the root readline package without an import cycle.

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/hinshun/vt10x"

	"github.com/reeflective/readline"
)

const (
	childEnvVar        = "READLINE_PTY_CHILD"
	promptEnvVar       = "READLINE_PTY_PROMPT"
	prefillEnvVar      = "READLINE_PTY_PREFILL"
	noProbeEnvVar      = "READLINE_PTY_NOPROBE"
	hintProviderEnvVar = "READLINE_PTY_HINTPROVIDER"
	transientEnvVar    = "READLINE_PTY_TRANSIENT"
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
	// Optionally push the prompt down the screen first, so tests can place it
	// at (or near) the bottom of the terminal window. The value is the number
	// of blank lines to print before starting the shell.
	if n, err := strconv.Atoi(os.Getenv(prefillEnvVar)); err == nil && n > 0 {
		fmt.Fprint(os.Stdout, strings.Repeat("\r\n", n))
	}

	rl := readline.NewShell()

	if os.Getenv(noProbeEnvVar) == "1" {
		rl.Config.Set("cursor-position-probe", false)
	}

	// Register a passive hint provider that echoes the current line, so tests
	// can observe the provided lane tracking input.
	if os.Getenv(hintProviderEnvVar) == "1" {
		rl.Hint.SetProvider(func(line []rune, _ int) []rune {
			if len(line) == 0 {
				return nil
			}

			return []rune("HINT:" + string(line))
		})
	}

	// Seed a transient (async status) hint before reading, so tests can assert
	// it renders and survives completion/isearch activity.
	if msg := os.Getenv(transientEnvVar); msg != "" {
		rl.Hint.SetTransient(msg)
	}

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

	mu         sync.Mutex // guards term (writer + reads), probeCount and DSR replies
	done       chan struct{}
	probeCount int // number of "ESC[6n" cursor-position queries observed

	// probeReply, if non-nil, computes the DSR reply bytes for an "ESC[6n"
	// cursor-position query, letting tests simulate a misbehaving terminal.
	// If nil, the emulator's true cursor position is reported (1-based).
	probeReply func(cur vt10x.Cursor) string
}

// consoleConfig configures a PTY-backed test console.
type consoleConfig struct {
	prompt     string
	cols, rows int
	// prefill is the number of blank lines printed before the prompt, used to
	// push the prompt down the window (e.g. to the bottom row).
	prefill int
	// noProbe disables the shell's cursor-position probing in the child.
	noProbe bool
	// hintProvider registers a passive hint provider in the child that echoes
	// the current line as "HINT:<line>".
	hintProvider bool
	// transient, if non-empty, seeds a transient (async status) hint in the
	// child before the read loop starts.
	transient string
	// probeReply, if non-nil, computes the DSR reply for an "ESC[6n" query,
	// letting tests simulate a terminal that reports a wrong cursor position.
	probeReply func(vt10x.Cursor) string
}

// newConsole spawns the child shell under a PTY of the given size, with the
// given primary prompt, and starts mirroring its output into the emulator.
func newConsole(t *testing.T, prompt string, cols, rows int) *console {
	return startConsole(t, consoleConfig{prompt: prompt, cols: cols, rows: rows})
}

// startConsole spawns the child shell under a PTY using the full config and
// starts mirroring its output into the emulator.
func startConsole(t *testing.T, cfg consoleConfig) *console {
	t.Helper()

	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(),
		childEnvVar+"=1",
		promptEnvVar+"="+cfg.prompt,
		fmt.Sprintf("%s=%d", prefillEnvVar, cfg.prefill),
		"INPUTRC=/dev/null", // don't pick up a developer's ~/.inputrc
		"TERM=xterm-256color",
	)

	if cfg.noProbe {
		cmd.Env = append(cmd.Env, noProbeEnvVar+"=1")
	}

	if cfg.hintProvider {
		cmd.Env = append(cmd.Env, hintProviderEnvVar+"=1")
	}

	if cfg.transient != "" {
		cmd.Env = append(cmd.Env, transientEnvVar+"="+cfg.transient)
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(cfg.rows), Cols: uint16(cfg.cols)})
	if err != nil {
		t.Fatalf("start child under pty: %v", err)
	}

	c := &console{
		t:          t,
		cmd:        cmd,
		ptmx:       ptmx,
		term:       vt10x.New(vt10x.WithSize(cfg.cols, cfg.rows)),
		done:       make(chan struct{}),
		probeReply: cfg.probeReply,
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
				c.mu.Lock()
				c.probeCount++
				c.mu.Unlock()

				c.replyCursorPos()
			}
		}

		if err != nil {
			return
		}
	}
}

// replyCursorPos writes a DSR cursor-position report (1-based row;col), or the
// custom reply from probeReply when a test wants to simulate a bad terminal.
func (c *console) replyCursorPos() {
	c.mu.Lock()
	cur := c.term.Cursor()
	fn := c.probeReply
	c.mu.Unlock()

	reply := fmt.Sprintf("\x1b[%d;%dR", cur.Y+1, cur.X+1)
	if fn != nil {
		reply = fn(cur)
	}

	if reply != "" {
		_, _ = c.ptmx.WriteString(reply)
	}
}

// send writes raw bytes to the child as if typed by the user.
func (c *console) send(s string) {
	c.t.Helper()

	if _, err := c.ptmx.WriteString(s); err != nil {
		c.t.Fatalf("send %q: %v", s, err)
	}
}

// probeQueries returns how many "ESC[6n" cursor-position queries the child has
// sent so far.
func (c *console) probeQueries() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.probeCount
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

// waitUntil polls the rendered screen until cond returns true, or fails the
// test on timeout. Returns the last screen contents.
func (c *console) waitUntil(cond func(screen string) bool, timeout time.Duration) string {
	c.t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		s := c.screen()
		if cond(s) {
			return s
		}

		if time.Now().After(deadline) {
			c.t.Fatalf("timed out waiting for screen condition; got:\n%s", s)
			return s
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// close tears the child and PTY down. Registered via t.Cleanup, so it must
// never hang even when a test fails mid-flight: we close the PTY (which EOFs
// the child's reads), then force-kill as a backstop before reaping.
func (c *console) close() {
	_ = c.ptmx.Close()

	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}

	_ = c.cmd.Wait()
	<-c.done
}
