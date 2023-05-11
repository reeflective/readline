package completion

import (
	"fmt"
	"strings"

	"github.com/reeflective/readline/internal/color"
)

func (e *Engine) hintCompletions(comps Values) {
	hint := ""

	// First add the command/flag usage string if any,
	// and only if we don't have completions.
	if len(comps.values) == 0 || e.config.GetBool("usage-hint-always") {
		if comps.Usage != "" {
			hint += color.Dim + comps.Usage + "\n"
		}
	}

	// And all further messages
	for _, message := range comps.Messages.Get() {
		if message == "" {
			continue
		}

		hint += fmt.Sprintf("%s\n", message)
	}

	if e.Matches() == 0 && hint == "" && !e.auto {
		hint = e.hintNoMatches()
	}

	hint = strings.TrimSuffix(hint, "\n")
	if hint == "" {
		return
	}

	// Add the hint to the shell.
	e.hint.Set(hint)
}

func (e *Engine) hintNoMatches() string {
	noMatches := color.Dim + "no matching"

	var groups []string

	for _, group := range e.groups {
		if group.tag == "" {
			continue
		}

		groups = append(groups, group.tag)
	}

	if len(groups) > 0 {
		groupsStr := strings.Join(groups, ", ")
		noMatches += "'" + groupsStr + "'"
	}

	return noMatches + " completions"
}
