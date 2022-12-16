package readline

import (
	"regexp"
)

func (rl *Instance) backspaceTabFind() {
	if len(rl.tfLine) > 0 {
		rl.tfLine = rl.tfLine[:len(rl.tfLine)-1]
	}
	rl.updateTabFind([]rune{})
}

// Filter and refresh (print) a list of completions. The caller should have reset
// the virtual completion system before, so that should not clash with this.
func (rl *Instance) updateTabFind(r []rune) {
	rl.tfLine = append(rl.tfLine, r...)

	// The search regex is common to all search modes
	var err error
	rl.regexSearch, err = regexp.Compile("(?i)" + string(rl.tfLine))
	if err != nil {
		rl.hintText = []rune(Red("Failed to match search regexp"))
	}

	// We update and print
	rl.clearHelpers()
	// rl.getTabCompletion()
	rl.renderHelpers()
}

func (rl *Instance) resetTabFind() {
	rl.tfLine = []rune{}

	rl.clearHelpers()
	rl.resetTabCompletion()

	// If we were browsing history, we don't load the completions again
	// if rl.searchMode != HistoryFind {
	// rl.getTabCompletion()
	// }

	rl.renderHelpers()
}
