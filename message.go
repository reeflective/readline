package readline

import (
	"regexp"
	"sort"
	"strings"
)

type Messages struct {
	messages map[string]bool
}

func (m *Messages) init() {
	if m.messages == nil {
		m.messages = make(map[string]bool)
	}
}

func (m Messages) IsEmpty() bool {
	// TODO replacement for Action.skipCache - does this need to consider suppressed messages or is this fine?
	return len(m.messages) == 0
}

func (m *Messages) Add(s string) {
	m.init()
	m.messages[s] = true
}

func (m Messages) Get() []string {
	messages := make([]string, 0)
	for message := range m.messages {
		messages = append(messages, message)
	}
	sort.Strings(messages)
	return messages
}

func (m *Messages) Suppress(expr ...string) error {
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

func (m *Messages) Merge(other Messages) {
	if other.messages == nil {
		return
	}

	for key := range other.messages {
		m.Add(key)
	}
}

type SuffixMatcher struct {
	string
}

func (sm *SuffixMatcher) Add(suffixes ...rune) {
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
	sort.Sort(ByRune(unique))
	sm.string = string(unique)
}

func (sm *SuffixMatcher) Merge(other SuffixMatcher) {
	for _, r := range []rune(other.string) {
		sm.Add(r)
	}
}

func (sm SuffixMatcher) Matches(s string) bool {
	for _, r := range []rune(sm.string) {
		if r == '*' || strings.HasSuffix(s, string(r)) {
			return true
		}
	}
	return false
}

type ByRune []rune

func (r ByRune) Len() int           { return len(r) }
func (r ByRune) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ByRune) Less(i, j int) bool { return r[i] < r[j] }
