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
	if len(comps.values) == 0 {
		hint += color.Dim + comps.Usage + "\n"
	}

	// And all further messages
	for _, message := range comps.Messages.Get() {
		if message == "" {
			continue
		}

		hint += fmt.Sprintf("%s\n", message)
	}

	hint = strings.TrimSuffix(hint, "\n")

	// Add the hint to the shell.
	e.hint.Set(hint)
}

func (e *Engine) hintNoMatches() {
	noMatches := color.Dim + "no matching"

	var groups []string

	for _, group := range e.groups {
		if group.tag == "" {
			continue
		}

		groups = append(groups, group.tag)
	}

	// History has no named group, so add it
	// if len(groups) == 0 && len(rl.histHint) > 0 {
	// 	groups = append(groups, rl.historyNames[rl.historySourcePos])
	// }

	if len(groups) > 0 {
		groupsStr := strings.Join(groups, ", ")
		noMatches += "'" + groupsStr + "'"
	}

	noMatches += " completions"

	e.hint.Set(noMatches)
}

func (e *Engine) hintIsearch() {
	var currentMode string

	if e.hint.Len() > 0 {
		currentMode = e.hint.Text() + color.FgCyan + " (isearch): "
	} else {
		currentMode = "isearch: "
	}

	hint := color.Bold + color.FgCyan + currentMode + color.Reset + color.BgDarkGray
	hint += string(*e.isearchBuf)

	if e.isearchRgx == nil && e.isearchBuf.Len() > 0 {
		hint += color.FgRed + " ! failed to compile search regexp"
	} else if e.noCompletions() && e.isearchBuf.Len() > 0 {
		hint += color.FgRed + " ! no matches"
	}

	// hint += color.Reset

	e.hint.Set(hint)
}
