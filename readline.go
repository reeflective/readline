package readline

import (
	"os"

	"github.com/reeflective/readline/internal/term"
)

// Readline displays the readline prompt and reads for user input.
// It can return from the call because of different several things:
//
// - When the user accepts the line (generally with Enter).
// - If a particular keystroke mapping returns an error
//   (like Ctrl-C, Ctrl-D, etc).
//
// In all cases, the current input line is returned along with any error,
// and it is up to the caller to decide what to do with the line result.
func (rl *Shell) Readline() (string, error) {
	descriptor := int(os.Stdin.Fd())

	state, err := term.MakeRaw(descriptor)
	if err != nil {
		return "", err
	}
	defer term.Restore(descriptor, state)

	rl.init()

	rl.prompt.PrimaryPrint()
	defer rl.prompt.TransientPrint()

	for {
		// Since we always update helpers after being asked to read
		// for user input again, we do it before actually reading it.
		rl.display.Refresh()

		// Block and wait for user input.
		rl.keys.Read()

		command, prefixed := rl.keymaps.Match(true)
		if command != nil {
			command()
		} else if prefixed {
			continue
		}

		// If the keys did not match any command either
		// exactly or by prefix, we flush the key stack.
		rl.keys.Flush()
	}
}
