package completion

import (
	"testing"

	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/keymap"
	"github.com/reeflective/readline/internal/ui"
)

// newPrefixEngine builds a minimal engine whose menu already holds the given
// candidate values (spread across two groups to exercise multi-group walking),
// with the real line containing the already-typed prefix and the cursor at its
// end -- the state completeWord/menuComplete are in when the display-prefix
// option fires.
func newPrefixEngine(typed string, values ...string) (*Engine, *core.Line) {
	l := core.Line([]rune(typed))
	line := &l
	cur := core.NewCursor(line)
	cur.Set(line.Len())

	var groups []*group

	for i, v := range values {
		// Alternate candidates between two groups so commonPrefix must walk
		// more than one group and more than one row.
		gi := i % 2
		for len(groups) <= gi {
			groups = append(groups, &group{})
		}

		groups[gi].rows = append(groups[gi].rows, []Candidate{{Value: v}})
	}

	e := &Engine{
		line:   line,
		cursor: cur,
		prefix: typed,
		groups: groups,
	}

	return e, line
}

func TestCommonStringPrefix(t *testing.T) {
	cases := []struct{ a, b, want string }{
		{"foobar", "foobaz", "fooba"},
		{"foo", "foobar", "foo"},
		{"abc", "xyz", ""},
		{"", "foo", ""},
		{"café", "cafard", "caf"}, // rune-aligned, must not split the 'é'
	}

	for _, c := range cases {
		if got := commonStringPrefix(c.a, c.b); got != c.want {
			t.Errorf("commonStringPrefix(%q, %q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

func TestInsertCommonPrefix(t *testing.T) {
	t.Run("extends to shared prefix", func(t *testing.T) {
		e, line := newPrefixEngine("foo", "foobar", "foobaz", "foobat")
		e.InsertCommonPrefix()

		if got := string(*line); got != "fooba" {
			t.Fatalf("line = %q, want %q", got, "fooba")
		}

		if e.prefix != "fooba" {
			t.Fatalf("engine prefix = %q, want %q", e.prefix, "fooba")
		}

		if e.cursor.Pos() != 5 {
			t.Fatalf("cursor = %d, want 5 (end of extended word)", e.cursor.Pos())
		}
	})

	t.Run("no shared prefix beyond typed is a no-op", func(t *testing.T) {
		e, line := newPrefixEngine("f", "foo", "fbar")
		e.InsertCommonPrefix()

		// "foo" and "fbar" share only "f", which equals the typed prefix.
		if got := string(*line); got != "f" {
			t.Fatalf("line = %q, want unchanged %q", got, "f")
		}

		if e.prefix != "f" {
			t.Fatalf("engine prefix = %q, want unchanged %q", e.prefix, "f")
		}
	})

	t.Run("single candidate extends fully", func(t *testing.T) {
		e, line := newPrefixEngine("re", "readline")
		e.InsertCommonPrefix()

		if got := string(*line); got != "readline" {
			t.Fatalf("line = %q, want %q", got, "readline")
		}
	})

	t.Run("no candidates is a no-op", func(t *testing.T) {
		e, line := newPrefixEngine("foo")
		e.InsertCommonPrefix()

		if got := string(*line); got != "foo" {
			t.Fatalf("line = %q, want unchanged %q", got, "foo")
		}
	})
}

func TestAcceptCandidateCallsOnAcceptAfterInsertion(t *testing.T) {
	line := core.Line([]rune("rea"))
	cursor := core.NewCursor(&line)
	cursor.Set(line.Len())
	accepted := false
	grp := &group{
		rows: [][]Candidate{{{
			Value: "readline",
			OnAccept: func(line []rune, cursor int) ([]rune, int) {
				if got := string(line); got != "readline" {
					t.Fatalf("line during OnAccept = %q, want %q", got, "readline")
				}
				accepted = true
				return line, cursor
			},
		}}},
	}
	engine := &Engine{
		line:   &line,
		cursor: cursor,
		prefix: "rea",
		groups: []*group{grp},
	}

	engine.acceptCandidate()

	if !accepted {
		t.Fatal("OnAccept was not called")
	}
	if got := cursor.Pos(); got != len([]rune("readline")) {
		t.Fatalf("cursor = %d, want %d", got, len([]rune("readline")))
	}
}

func TestCancelCallsOnAcceptWhenVirtualCandidateBecomesReal(t *testing.T) {
	line := core.Line([]rune("rea"))
	cursor := core.NewCursor(&line)
	cursor.Set(line.Len())
	completed := core.Line([]rune("readline"))
	completedCursor := core.NewCursor(&completed)
	completedCursor.Set(completed.Len())
	accepted := false
	engine := &Engine{
		line:       &line,
		cursor:     cursor,
		compLine:   &completed,
		compCursor: completedCursor,
		selected: Candidate{Value: "readline", OnAccept: func(line []rune, cursor int) ([]rune, int) {
			if got := string(line); got != "readline" {
				t.Fatalf("line during OnAccept = %q, want %q", got, "readline")
			}
			accepted = true
			return line, cursor
		}},
	}

	engine.Cancel(false, false)

	if !accepted {
		t.Fatal("OnAccept was not called")
	}
	if got := cursor.Pos(); got != len([]rune("readline")) {
		t.Fatalf("cursor = %d, want %d", got, len([]rune("readline")))
	}
}

func TestResetAcceptsSelectedCandidateAfterMenuKeymapEnds(t *testing.T) {
	keys := new(core.Keys)
	iterations := new(core.Iterations)
	keymaps, config := keymap.NewEngine(keys, iterations)
	hint := ui.NewHint(keys)
	line := core.Line([]rune("rea"))
	cursor := core.NewCursor(&line)
	cursor.Set(line.Len())
	selection := core.NewSelection(&line, cursor)
	engine := NewEngine(hint, keymaps, config)
	Init(engine, keys, &line, cursor, selection, nil)
	completed := core.Line([]rune("readline"))
	completedCursor := core.NewCursor(&completed)
	completedCursor.Set(completed.Len())
	accepted := false
	engine.compLine = &completed
	engine.compCursor = completedCursor
	engine.selected = Candidate{Value: "readline", OnAccept: func(line []rune, cursor int) ([]rune, int) {
		accepted = true
		return []rune("accepted"), len([]rune("accepted"))
	}}

	engine.Reset()

	if !accepted {
		t.Fatal("OnAccept was not called")
	}
	if got := string(line); got != "accepted" {
		t.Fatalf("line = %q, want %q", got, "accepted")
	}
}
