package readline

import "regexp"

// FindMode defines how the autocomplete suggestions display
type FindMode int

const (
	// HistoryFind - Searching through history
	HistoryFind = iota

	// CompletionFind - Searching through completion items
	CompletionFind
)

func (rl *Instance) backspaceTabFind() {
	if len(rl.tfLine) > 0 {
		rl.tfLine = rl.tfLine[:len(rl.tfLine)-1]
	}
	rl.updateTabFind([]rune{})
}

func (rl *Instance) updateTabFind(r []rune) {

	// Depending on search type, we give different hints
	rl.tfLine = append(rl.tfLine, r...)
	switch rl.regexpMode {
	case HistoryFind:
		rl.hintText = append([]rune("History search: "), rl.tfLine...)
	case CompletionFind:
		rl.hintText = append([]rune("Completion search: "), rl.tfLine...)
	}

	defer func() {
		rl.clearHelpers()
		rl.getTabCompletion()
		rl.renderHelpers()
	}()

	if len(rl.tfLine) == 0 {
		rl.tfSuggestions = append(rl.tcSuggestions, []string{}...)
		return
	}

	rx, err := regexp.Compile("(?i)" + string(rl.tfLine))
	if err != nil {
		rl.tfSuggestions = []string{err.Error()}
		return
	}

	rl.tfSuggestions = make([]string, 0)
	for i := range rl.tcSuggestions {
		if rx.MatchString(rl.tcSuggestions[i]) {
			rl.tfSuggestions = append(rl.tfSuggestions, rl.tcSuggestions[i])

		} else if rl.tcDisplayType == TabDisplayList && rx.MatchString(rl.tcDescriptions[rl.tcSuggestions[i]]) {
			// this is a list so lets also check the descriptions
			rl.tfSuggestions = append(rl.tfSuggestions, rl.tcSuggestions[i])
		}
	}
}

func (rl *Instance) resetTabFind() {
	rl.modeTabFind = false
	rl.tfLine = []rune{}
	if rl.modeAutoFind {
		rl.hintText = []rune{}
	} else {
		rl.hintText = []rune("Cancelled regexp suggestion find.")
	}

	rl.clearHelpers()
	rl.getTabCompletion()
	rl.renderHelpers()
}
