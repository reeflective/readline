package completion

import (
	"regexp"
	"sort"
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
		char, err := regexp.Compile(e)
		if err != nil {
			return err
		}

		for key := range m.messages {
			if char.MatchString(key) {
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
