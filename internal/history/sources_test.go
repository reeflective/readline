package history

import (
	"testing"

	"github.com/reeflective/readline/internal/core"
)

// newTestSources builds a minimal Sources backed by a single in-memory source,
// enough to exercise Write's cap logic without a full shell.
func newTestSources(maxEntries int) (*Sources, *memory, *core.Line) {
	line := new(core.Line)
	mem := new(memory)

	src := &Sources{
		line:       line,
		list:       map[string]Source{defaultSourceName: mem},
		names:      []string{defaultSourceName},
		maxEntries: maxEntries,
	}

	return src, mem, line
}

// writeLine sets the working line and writes it to history.
func writeLine(src *Sources, line *core.Line, text string) {
	*line = core.Line([]rune(text))

	src.Write(false)
}

// TestWriteUnlimitedPersists guards the default (history-size unset -> -1):
// every distinct line must be written.
func TestWriteUnlimitedPersists(t *testing.T) {
	src, mem, line := newTestSources(-1)

	writeLine(src, line, "one")
	writeLine(src, line, "two")
	writeLine(src, line, "three")

	if mem.Len() != 3 {
		t.Fatalf("unlimited history wrote %d lines, want 3", mem.Len())
	}
}

// TestWriteUnderCapPersists is the direct regression guard for the inverted
// condition: with a positive cap and the source still under it, writes MUST
// land. The old `maxEntries >= Len()` skipped them all.
func TestWriteUnderCapPersists(t *testing.T) {
	src, mem, line := newTestSources(5)

	writeLine(src, line, "one")
	writeLine(src, line, "two")

	if mem.Len() != 2 {
		t.Fatalf("capped history under the limit wrote %d lines, want 2 (inverted-condition regression)", mem.Len())
	}
}

// TestWriteStopsAtCap verifies the cap is actually enforced: once the source
// holds maxEntries lines, further distinct lines are not appended.
func TestWriteStopsAtCap(t *testing.T) {
	src, mem, line := newTestSources(2)

	writeLine(src, line, "one")
	writeLine(src, line, "two")
	writeLine(src, line, "three") // should be refused: already at cap

	if mem.Len() != 2 {
		t.Fatalf("history grew to %d lines past its cap of 2", mem.Len())
	}
}

// TestWriteZeroMaxIsUnlimited documents that maxEntries == 0 is treated as
// unlimited (writes happen), matching the `> 0` guard.
func TestWriteZeroMaxIsUnlimited(t *testing.T) {
	src, mem, line := newTestSources(0)

	writeLine(src, line, "one")
	writeLine(src, line, "two")

	if mem.Len() != 2 {
		t.Fatalf("maxEntries==0 wrote %d lines, want 2 (should be unlimited)", mem.Len())
	}
}
