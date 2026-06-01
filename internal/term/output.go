package term

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

// renderOut, when non-nil, overrides the destination for rendered terminal
// output. It is a test seam; in production it stays nil and output goes to
// os.Stdout, resolved at call time so that tests redirecting os.Stdout (the
// common capture pattern) keep working. Output goes to stdout — distinct from
// termFile (stderr), used only for terminal *queries* (size, cursor position).
var renderOut io.Writer

func out() io.Writer {
	if renderOut != nil {
		return renderOut
	}

	return os.Stdout
}

// Output buffering lets a whole frame (one Refresh) be coalesced into a single
// write to the terminal instead of dozens of small writes. This removes the
// inter-write flicker a terminal can show mid-frame and cuts write syscalls per
// keystroke from many to one.
//
// All rendering happens on the readline goroutine, so this state is effectively
// single-threaded; the mutex is a cheap guard against accidental concurrent use
// (e.g. a stray write from another goroutine) rather than a hot contention path.
var (
	outMu    sync.Mutex
	outDepth int
	outBuf   *bufio.Writer
)

// BeginBuffer starts buffering terminal writes. Calls nest: only the outermost
// BeginBuffer/EndBuffer pair allocates and flushes the buffer, so render
// entrypoints can be wrapped independently without coordinating.
func BeginBuffer() {
	outMu.Lock()
	defer outMu.Unlock()

	outDepth++
	if outDepth == 1 {
		outBuf = bufio.NewWriterSize(out(), 64*1024)
	}
}

// EndBuffer flushes and tears down the buffer when leaving the outermost frame.
// It is safe to call unmatched (it no-ops at depth 0), so it can be deferred
// even if BeginBuffer is on an early-return path.
func EndBuffer() {
	outMu.Lock()
	defer outMu.Unlock()

	if outDepth == 0 {
		return
	}

	outDepth--

	if outDepth == 0 && outBuf != nil {
		_ = outBuf.Flush()
		outBuf = nil
	}
}

// FlushBuffer pushes everything buffered so far to the terminal without ending
// the frame. It must be called before any operation that READS the terminal and
// depends on prior output being on screen — chiefly an ESC[6n cursor-position
// query (see GetCursorPos): otherwise the prompt would still sit in the buffer
// and the reported cursor position would be wrong.
func FlushBuffer() {
	outMu.Lock()
	defer outMu.Unlock()

	if outBuf != nil {
		_ = outBuf.Flush()
	}
}

// WriteString writes s to the terminal: into the frame buffer when one is
// active, otherwise straight to the output. Empty strings are dropped so callers
// need not guard.
func WriteString(s string) {
	if s == "" {
		return
	}

	outMu.Lock()
	defer outMu.Unlock()

	if outDepth > 0 && outBuf != nil {
		_, _ = io.WriteString(outBuf, s)
		return
	}

	_, _ = io.WriteString(out(), s)
}

// Printf formats and writes to the terminal, honouring the frame buffer.
func Printf(format string, a ...interface{}) {
	WriteString(fmt.Sprintf(format, a...))
}
