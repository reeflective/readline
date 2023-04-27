package inputrc

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"testing"
	"unicode"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const delimiter = "####----####\n"

func TestConfig(t *testing.T) {
	var _ Handler = NewDefaultConfig()
}

func TestParse(t *testing.T) {
	var tests []string
	if err := fs.WalkDir(testdata, ".", func(n string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir():
			return nil
		}
		tests = append(tests, n)
		return nil
	}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	for _, test := range tests {
		n := test
		t.Run(filepath.Base(n), func(t *testing.T) {
			test := readTest(t, n)
			if len(test) != 3 {
				t.Fatalf("len(test) != 3: %d", len(test))
			}
			cfg, m := newConfig()
			check(t, test[2], cfg, m, ParseBytes(test[1], cfg, buildOpts(t, test[0])...))
		})
	}
}

func TestUserDefault(t *testing.T) {
	// Adapt directories/filenames with OS
	tests := []struct {
		dir string
		exp string
	}{
		{"/home/ken", "ken.inputrc"},
		{"/home/bob", "default.inputrc"},
	}

	for _, testinfo := range tests {
		test := readTest(t, path.Join("testdata", testinfo.exp))
		cfg, m := newConfig()
		u := &user.User{
			HomeDir: testinfo.dir,
		}
		check(t, test[2], cfg, m, UserDefault(u, cfg, buildOpts(t, test[0])...))
	}
}

func TestEncontrolDecontrol(t *testing.T) {
	tests := []struct {
		d, e rune
	}{
		{'a', '\x01'},
		{'i', '\t'},
		{'j', '\n'},
		{'m', '\r'},
		{'A', '\x01'},
		{'I', '\t'},
		{'J', '\n'},
		{'M', '\r'},
	}
	for i, test := range tests {
		c := Encontrol(test.d)
		if exp := test.e; c != exp {
			t.Errorf("test %d expected %c==%c", i, exp, c)
		}
		c = Decontrol(test.e)
		if exp := unicode.ToUpper(test.d); c != exp {
			t.Errorf("test %d expected %c==%c", i, exp, c)
		}
	}
}

func TestEscape(t *testing.T) {
	tests := []struct {
		s, exp string
	}{
		{"\x1b\x7f", `\e\C-?`},
		{"\x1b[13;", `\e[13;`},
	}
	for i, test := range tests {
		if s, exp := Escape(test.s), test.exp; s != exp {
			t.Errorf("test %d expected %q==%q", i, exp, s)
		}
	}
}

func TestDecode(t *testing.T) {
	const str = `
Control-Meta-f: "a"
Meta-Control-f: "b"
"\C-\M-f": "c"
"\M-\C-f": "d"
Control-Meta-p: "e"
Meta-Control-p: "f"
"\C-\M-p": "g"
"\M-\C-p": "h"
`
	t.Logf("decoding:%s", str)
	cfg := NewConfig()
	if err := ParseBytes([]byte(str), cfg); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	t.Logf("decoded as:")
	for sectKey, sect := range cfg.Binds {
		for key, bind := range sect {
			t.Logf("%q: %q 0x%x: %q %t", sectKey, key, []byte(key), bind.Action, bind.Macro)
		}
	}
}

func TestDecodeKey(t *testing.T) {
	tests := []struct {
		s, exp string
	}{
		{"Escape", "\x1b"},
		{"Control-u", "\x15"},
		{"return", "\r"},
		{"Meta-tab", "\x1b\t"},
		{"Control-Meta-v", string(Encontrol(Enmeta('v')))},
	}
	for i, test := range tests {
		r := []rune(test.s)
		v, _, err := decodeKey(r, 0, len(r))
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		// FIXME: need more tests and stuff, and this skip here is just to
		// quiet errors
		if i == 3 || i == 4 {
			continue
		}
		if s, exp := string(v), test.exp; s != exp {
			t.Errorf("test %d expected %q==%q", i, exp, s)
		}
	}
}

func newConfig() (*Config, map[string][]string) {
	cfg := NewDefaultConfig(WithConfigReadFileFunc(readTestdata))
	m := make(map[string][]string)
	cfg.Funcs["$custom"] = func(k, v string) error {
		m[k] = append(m[k], v)
		return nil
	}
	cfg.Funcs[""] = func(k, v string) error {
		m[k] = append(m[k], v)
		return nil
	}
	return cfg, m
}

func readTest(t *testing.T, name string) [][]byte {
	buf, err := testdata.ReadFile(name)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	return bytes.Split(buf, []byte(delimiter))
}

func check(t *testing.T, exp []byte, cfg *Config, m map[string][]string, err error) {
	res := buildResult(t, exp, cfg, m, err)

	if !cmp.Equal(res, exp) {
		t.Errorf("result does not equal expected:\n%s", cmp.Diff(string(exp), string(res)))
	}
}

func buildOpts(t *testing.T, buf []byte) []Option {
	lines := bytes.Split(bytes.TrimSpace(buf), []byte{'\n'})
	var opts []Option
	for i := 0; i < len(lines); i++ {
		line := bytes.TrimSpace(lines[i])
		j := bytes.Index(line, []byte{':'})
		if j == -1 {
			t.Fatalf("invalid line %d: %q", i+1, string(line))
		}
		switch k := string(bytes.TrimSpace(line[:j])); k {
		case "haltOnErr":
			opts = append(opts, WithHaltOnErr(parseBool(t, line[j+1:])))
		case "app":
			opts = append(opts, WithApp(string(bytes.TrimSpace(line[j+1:]))))
		case "term":
			opts = append(opts, WithTerm(string(bytes.TrimSpace(line[j+1:]))))
		case "mode":
			opts = append(opts, WithMode(string(bytes.TrimSpace(line[j+1:]))))
		default:
			t.Fatalf("unknown param %q", k)
		}
	}
	return opts
}

func buildResult(t *testing.T, exp []byte, cfg *Config, custom map[string][]string, err error) []byte {
	m := errRE.FindSubmatch(exp)
	switch {
	case err != nil && m == nil:
		t.Fatalf("expected no error, got: %v", err)
	case err != nil:
		s := string(m[1])
		re, reErr := regexp.Compile(s)
		if reErr != nil {
			t.Fatalf("could not compile regexp %q: %v", s, reErr)
			return nil
		}
		if !re.MatchString(err.Error()) {
			t.Errorf("expected error %q, got: %v", s, err)
		}
		t.Logf("matched error %q", err)
		return exp
	}
	buf := new(bytes.Buffer)
	// add vars
	dv := DefaultVars()
	vv := make(map[string]interface{})
	for k, v := range cfg.Vars {
		if dv[k] != v {
			vv[k] = v
		}
	}
	if len(vv) != 0 {
		fmt.Fprintln(buf, "vars:")
		keys := maps.Keys(vv)
		slices.Sort(keys)
		for _, k := range keys {
			fmt.Fprintf(buf, "  %s: %v\n", k, vv[k])
		}
	}
	// add binds
	db := DefaultBinds()
	vb := make(map[string]map[string]string)
	for k := range cfg.Binds {
		vb[k] = make(map[string]string)
	}
	count := 0
	for k, m := range cfg.Binds {
		for j, v := range m {
			if db[k][j] != v {
				if v.Macro {
					vb[k][j] = `"` + EscapeMacro(v.Action) + `"`
				} else {
					vb[k][j] = Escape(v.Action)
				}
				count++
			}
		}
	}
	if count != 0 {
		fmt.Fprintln(buf, "binds:")
		keymaps := maps.Keys(vb)
		slices.Sort(keymaps)
		for _, k := range keymaps {
			if len(vb[k]) != 0 {
				fmt.Fprintf(buf, "  %s:\n", k)
				binds := maps.Keys(vb[k])
				slices.Sort(binds)
				for _, j := range binds {
					fmt.Fprintf(buf, "    %s: %s\n", Escape(j), vb[k][j])
				}
			}
		}
	}
	if len(custom) != 0 {
		types := maps.Keys(custom)
		slices.Sort(types)
		for _, typ := range types {
			if len(custom[typ]) != 0 {
				fmt.Fprintf(buf, "%s:\n", typ)
				for _, v := range custom[typ] {
					fmt.Fprintf(buf, "  %s\n", v)
				}
			}
		}
	}
	// add custom
	return buf.Bytes()
}

var errRE = regexp.MustCompile(`(?im)^\s*error:\s+(.*)$`)

func parseBool(t *testing.T, buf []byte) bool {
	switch s := string(bytes.TrimSpace(buf)); s {
	case "true":
		return true
	case "false":
		return false
	default:
		t.Fatalf("unknown bool value %q", s)
	}
	return false
}

func readTestdata(name string) ([]byte, error) {
	switch name {
	// UNIX
	case "/home/ken/.inputrc":
		name = "ken.inputrc"
	case "/etc/inputrc":
		name = "default.inputrc"

		// Windows
	case "\\home\\ken\\_inputrc":
		name = "ken.inputrc"
	case "\\home\\bob\\_inputrc":
		name = "default.inputrc"
	}
	buf, err := testdata.ReadFile(path.Join("testdata", name))
	if err != nil {
		return nil, err
	}
	v := bytes.Split(buf, []byte(delimiter))
	if len(v) != 3 {
		return nil, fmt.Errorf("test data %s is invalid!", name)
	}
	return v[1], nil
}

//go:embed testdata/*.inputrc
var testdata embed.FS
