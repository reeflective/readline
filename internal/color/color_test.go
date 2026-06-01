package color

import "testing"

func TestStrip(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "plain value", "plain value"},
		{"empty", "", ""},
		{"utf8 no escape", "café — résumé", "café — résumé"},
		{"esc colored", "\x1b[31mred\x1b[0m", "red"},
		{"esc surrounding", "a\x1b[1mb\x1b[0mc", "abc"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Strip(c.in); got != c.want {
				t.Fatalf("Strip(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestTrim(t *testing.T) {
	cases := []struct {
		name string
		in   string
		max  int
		want string
	}{
		{"shorter than max", "abc", 10, "abc"},
		{"exact trim", "abcdef", 3, "abc"},
		{"zero budget", "abcdef", 0, ""},
		// A negative budget is reachable from completion column math on a very
		// narrow terminal (maxDisplayWidth - trailingValueLen < 0). It must not
		// panic on input[:negative].
		{"negative budget", "abcdef", -4, ""},
		{"trim keeps escapes", "\x1b[31mabcdef\x1b[0m", 3, "\x1b[31mabc"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Trim(c.in, c.max); got != c.want {
				t.Fatalf("Trim(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
			}
		})
	}
}
