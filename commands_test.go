package readline

import (
	"maps"
	"testing"
)

// allCommands aggregates every command set the shell registers with its keymap
// engine (see Shell.init in shell.go). The builders only construct maps of
// method values, so a zero Shell is enough to enumerate them.
func allCommands() commands {
	rl := new(Shell)

	all := make(commands)
	for _, set := range []commands{
		rl.standardCommands(),
		rl.viCommands(),
		rl.historyCommands(),
		rl.completionCommands(),
	} {
		maps.Copy(all, set)
	}

	return all
}

// TestArrowNavigationCommandsRegistered guards that every line/history
// navigation action bound by the library's own default keymaps resolves to a
// registered command. The vi-insert down-arrow binding (down-line-or-search)
// previously resolved to nothing and was a silent no-op; this is its
// regression guard, plus its siblings so the whole arrow-nav set stays wired.
func TestArrowNavigationCommandsRegistered(t *testing.T) {
	all := allCommands()

	for _, action := range []string{
		"up-line-or-search",
		"down-line-or-search",
		"up-line-or-history",
		"down-line-or-history",
		"down-line-or-select",
	} {
		if _, ok := all[action]; !ok {
			t.Errorf("navigation action %q is bound in the default keymaps but has no registered command", action)
		}
	}
}
