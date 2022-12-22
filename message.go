package readline

import (
	"regexp"
	"sort"
	"strings"
)

type messages struct {
	messages map[string]bool
}

func (m *messages) init() {
	if m.messages == nil {
		m.messages = make(map[string]bool)
	}
}

func (m messages) IsEmpty() bool {
	// TODO replacement for Action.skipCache - does this need to consider suppressed messages or is this fine?
	return len(m.messages) == 0
}

func (m *messages) Add(s string) {
	m.init()
	m.messages[s] = true
}

func (m messages) Get() []string {
	messages := make([]string, 0)
	for message := range m.messages {
		messages = append(messages, message)
	}
	sort.Strings(messages)
	return messages
}

func (m *messages) Suppress(expr ...string) error {
	m.init()

	for _, e := range expr {
		r, err := regexp.Compile(e)
		if err != nil {
			return err
		}

		for key := range m.messages {
			if r.MatchString(key) {
				delete(m.messages, key)
			}
		}
	}
	return nil
}

func (m *messages) Merge(other messages) {
	if other.messages == nil {
		return
	}

	for key := range other.messages {
		m.Add(key)
	}
}

type suffixMatcher struct {
	string
	pos int // Used to know if the saved suffix matcher is deprecated
}

func (sm *suffixMatcher) Add(suffixes ...rune) {
	if strings.Contains(sm.string, "*") || strings.Contains(string(suffixes), "*") {
		sm.string = "*"
		return
	}

	unique := []rune(sm.string)
	for _, r := range suffixes {
		if !strings.Contains(sm.string, string(r)) {
			unique = append(unique, r)
		}
	}
	sort.Sort(byRune(unique))
	sm.string = string(unique)
}

func (sm *suffixMatcher) Merge(other suffixMatcher) {
	for _, r := range []rune(other.string) {
		sm.Add(r)
	}
}

func (sm suffixMatcher) Matches(s string) bool {
	for _, r := range []rune(sm.string) {
		if r == '*' || strings.HasSuffix(s, string(r)) {
			return true
		}
	}
	return false
}

type byRune []rune

func (r byRune) Len() int           { return len(r) }
func (r byRune) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byRune) Less(i, j int) bool { return r[i] < r[j] }
