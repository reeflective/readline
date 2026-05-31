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
