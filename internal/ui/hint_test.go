package ui

import "testing"

// TestCoordinatesHintRowCount guards against a row-count regression in the hint
// area: CoordinatesHint must return exactly one row per non-empty lane (the same
// number of rows DisplayHint actually prints). A previous implementation split
// the rendered string on the per-line clear-sequence and double-counted every
// lane after the first (returning 2N-1 rows for N lanes), which made the helper
// repaint move the cursor up too far and drift the whole prompt upward by one
// row on every refresh that showed two or more hint lanes.
func TestCoordinatesHintRowCount(t *testing.T) {
	cases := []struct {
		name                                  string
		persistent, provided, transient, text string
		want                                  int
	}{
		{"none", "", "", "", "", 0},
		{"one-text", "", "", "", "hello", 1},
		{"one-provided", "", "passive hint", "", "", 1},
		{"provided-and-transient", "", "passive hint", "async status", "", 2},
		{"persistent-and-text", "register", "", "", "completion", 2},
		{"three-lanes", "register", "passive", "async", "", 3},
		{"all-four-lanes", "register", "passive", "async", "completion", 4},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := &Hint{}
			h.persistent = []rune(c.persistent)
			h.provided = []rune(c.provided)
			h.transient = []rune(c.transient)
			h.text = []rune(c.text)

			// Normalize empty strings to empty slices so len()==0 lanes are skipped.
			for _, p := range []*[]rune{&h.persistent, &h.provided, &h.transient, &h.text} {
				if len(*p) == 0 {
					*p = nil
				}
			}

			if got := CoordinatesHint(h); got != c.want {
				t.Fatalf("CoordinatesHint = %d, want %d", got, c.want)
			}
		})
	}
}
