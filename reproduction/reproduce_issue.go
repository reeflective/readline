package main

import (
	"strings"

	"github.com/reeflective/readline"
)

func main() {
	bashMultiline := func(line []rune) (accept bool) {
		if strings.HasSuffix(string(line), "\\") {
			return false
		}

		return true
	}

	shell := readline.NewShell()
	shell.AcceptMultiline = bashMultiline
	shell.Prompt.Primary(func() string { return ">" })
	// shell.Prompt.Primary(func() string { return "long >" })
	// shell.Config.Set("multiline-column", true)
	shell.Config.Set("multiline-column-numbered", true)

	shell.Readline()
}
