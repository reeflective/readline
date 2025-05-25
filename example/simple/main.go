package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/reeflective/readline"
)

func main() {
	shell := readline.NewShell()
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
