package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// Multiline code
	shell.AcceptMultiline = bashMultiline
	// shell.Prompt.Primary(func() string { return "long >" })
	shell.Config.Set("multiline-column", true)
	// shell.Config.Set("multiline-column-numbered", true)

	// Prompt code
	shell.Prompt.Primary(func() string {
		prompt := "\x1b[33mexample\x1b[0m [main] in \x1b[34m%s\x1b[0m\n> "
		wd, _ := os.Getwd()

		dir, err := filepath.Rel(os.Getenv("HOME"), wd)
		if err != nil {
			dir = filepath.Base(wd)
		}

		return fmt.Sprintf(prompt, dir)
	})

	shell.Prompt.Right(func() string {
		return "\x1b[1;30m" + time.Now().Format("03:04:05.000") + "\x1b[0m"
	})

	shell.Prompt.Transient(func() string { return "\x1b[1;30m" + ">> " + "\x1b[0m" })
	shell.Readline()
}
