package readline

import (
	"os"
	"regexp"
)

// InputMode - The shell input mode
type InputMode string

const (
	// Vim - Vim editing mode
	Vim InputMode = "vim"
	// Emacs - Emacs (classic) editing mode
	Emacs InputMode = "emacs"
)

// readInput reads input from stdin and returns the result, length or an error.
func (rl *Instance) readInput() (b []byte, i int, err error) {
	rl.undoSkipAppend = false
	b = make([]byte, 1024)

	if !rl.skipStdinRead {
		i, err = os.Stdin.Read(b)
		if err != nil {
			return
		}
	}

	rl.skipStdinRead = false

	return
}

// readOperator reads a key required by some (rare) widgets that directly read/need
// their argument/operator, without going though operator pending mode first.
// If all is true, we return all keys, including numbers (instead of adding them as iterations.)
func (rl *Instance) readOperator(all bool) (key string, ret bool) {
	rl.enterVioppMode("")
	rl.updateCursor()

	defer func() {
		rl.exitVioppMode()
		rl.updateCursor()
	}()

	b, i, _ := rl.readInput()
	key = string(b[:i])

	// If the last key is a number, add to iterations instead,
	// and read another key input.
	if !all {
		numMatcher, _ := regexp.Compile(`^[1-9][0-9]*$`)

		for numMatcher.MatchString(string(key[len(key)-1])) {
			rl.iterations += string(key[len(key)-1])

			b, i, _ = rl.readInput()
			key = string(b[:i])
		}
	}

	// If the key is an escape key for the current mode.
	if len(key) == 1 &&
		(key[0] == charEscape) {
		ret = true
	}

	return
}
