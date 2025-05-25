package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/reeflective/readline"
)

func main() {
	shell := readline.NewShell()
	shell.SetStdin(os.Stdin)
	shell.SetStdout(os.Stdout)
	shell.SetStderr(os.Stderr)
	shell.Prompt.Primary(func() string {
		return "example > "
	})

	for {
		line, err := shell.Readline()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Line: %s\n", line)
	}
}
