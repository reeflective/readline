package completion

import (
	"regexp"
	"sort"
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

func (m *Messages) Merge(other Messages) {
	if other.messages == nil {
		return
	}

	for key := range other.messages {
		m.Add(key)
	}
}
