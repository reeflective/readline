package completion

import (
	"sort"
	"testing"
)

// referenceSort is the original sorting path (sort.Stable + the Less method),
// kept here so we can assert sortStable produces an identical ordering.
func referenceSort(c RawValues) RawValues {
	out := make(RawValues, len(c))
	copy(out, c)
	sort.Stable(out)

	return out
}

func TestSortStableMatchesReference(t *testing.T) {
	cases := [][]string{
		{"Banana", "apple", "Cherry", "banana", "APPLE", "cherry"},
		{"z", "a", "m", "A", "Z", "M"},
		{"git push", "git pull", "git Push", "GIT", "git"},
		{},
		{"only"},
	}

	for _, values := range cases {
		raw := make(RawValues, len(values))
		for i, v := range values {
			raw[i] = Candidate{Value: v}
		}

		want := referenceSort(raw)

		got := make(RawValues, len(raw))
		copy(got, raw)
		got.sortStable()

		for i := range want {
			if got[i].Value != want[i].Value {
				t.Fatalf("ordering mismatch for %v at %d: got %q, want %q",
					values, i, got[i].Value, want[i].Value)
			}
		}
	}
}

func benchValues(n int) RawValues {
	base := make(RawValues, n)
	for i := range base {
		// Mixed case so the comparator does real case-folding work.
		base[i] = Candidate{Value: string(rune('A'+(i%26))) + string(rune('a'+(i*7%26)))}
	}

	return base
}

// BenchmarkSortStable vs BenchmarkReferenceSort: the keyed sort folds each
// value once (~n allocations) instead of twice per comparison (~n*log n).
func BenchmarkSortStable(b *testing.B) {
	base := benchValues(1000)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		work := make(RawValues, len(base))
		copy(work, base)
		work.sortStable()
	}
}

func BenchmarkReferenceSort(b *testing.B) {
	base := benchValues(1000)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		work := make(RawValues, len(base))
		copy(work, base)
		sort.Stable(work)
	}
}
