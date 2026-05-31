package term

import (
	"bytes"
	"testing"
)

// countWriter records how many times Write was called (i.e. how many syscalls a
// real terminal would see) and accumulates the bytes for content assertions.
type countWriter struct {
	writes int
	buf    bytes.Buffer
}

func (c *countWriter) Write(p []byte) (int, error) {
	c.writes++

	return c.buf.Write(p)
}

// TestWriteStringDirect: with no active buffer, each non-empty WriteString is a
// direct write, and empty strings are dropped.
func TestWriteStringDirect(t *testing.T) {
	cw := &countWriter{}
	renderOut = cw

	defer func() { renderOut = nil }()

	WriteString("hello")
	WriteString("") // dropped, no write
	WriteString(" world")

	if got := cw.buf.String(); got != "hello world" {
		t.Fatalf("content = %q, want %q", got, "hello world")
	}

	if cw.writes != 2 {
		t.Fatalf("direct writes = %d, want 2", cw.writes)
	}
}

// TestBufferCoalescesToSingleWrite: a whole frame's worth of writes is held
// until EndBuffer, then flushed as one write — the core flicker/syscall win.
func TestBufferCoalescesToSingleWrite(t *testing.T) {
	cw := &countWriter{}
	renderOut = cw

	defer func() { renderOut = nil }()

	BeginBuffer()
	WriteString("\x1b[?25l")
	WriteString("prompt> ")
	WriteString("line")

	if cw.writes != 0 {
		t.Fatalf("expected nothing on screen before EndBuffer, got %d writes", cw.writes)
	}

	EndBuffer()

	if cw.writes != 1 {
		t.Fatalf("expected a single coalesced write, got %d", cw.writes)
	}

	if got := cw.buf.String(); got != "\x1b[?25lprompt> line" {
		t.Fatalf("content = %q", got)
	}
}

// TestBufferNestingFlushesOnce: nested Begin/End pairs only flush at the
// outermost End, so render entrypoints can wrap independently.
func TestBufferNestingFlushesOnce(t *testing.T) {
	cw := &countWriter{}
	renderOut = cw

	defer func() { renderOut = nil }()

	BeginBuffer()
	WriteString("a")
	BeginBuffer() // nested
	WriteString("b")
	EndBuffer() // inner: must NOT flush

	if cw.writes != 0 {
		t.Fatalf("inner EndBuffer flushed early: %d writes", cw.writes)
	}

	WriteString("c")
	EndBuffer() // outer: flush

	if cw.writes != 1 {
		t.Fatalf("expected one flush, got %d", cw.writes)
	}

	if got := cw.buf.String(); got != "abc" {
		t.Fatalf("content = %q, want abc", got)
	}
}

// TestFlushBufferMidFrame: FlushBuffer pushes what's buffered so far (e.g.
// before an ESC[6n query) without ending the frame.
func TestFlushBufferMidFrame(t *testing.T) {
	cw := &countWriter{}
	renderOut = cw

	defer func() { renderOut = nil }()

	BeginBuffer()
	WriteString("prompt")
	FlushBuffer() // the prompt must now be on screen before a cursor query

	if cw.writes != 1 || cw.buf.String() != "prompt" {
		t.Fatalf("FlushBuffer did not flush partial frame: writes=%d content=%q", cw.writes, cw.buf.String())
	}

	WriteString("line")
	EndBuffer()

	if cw.writes != 2 {
		t.Fatalf("expected 2 writes (mid-flush + end), got %d", cw.writes)
	}

	if got := cw.buf.String(); got != "promptline" {
		t.Fatalf("content = %q", got)
	}
}
