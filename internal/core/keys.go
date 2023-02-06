package core

import (
	"os"

	"github.com/xo/inputrc"
)

const (
	keyScanBufSize = 1024
)

// Keys is used read, manage and use keys input by the shell user.
type Keys struct {
	stack  []rune
	paused chan struct{}
}

// NewKeys is a required constructor for the readline key stack,
// as key reading might be paused/resumed and needs channel setup.
func NewKeys() *Keys {
	return &Keys{
		paused: make(chan struct{}, 1),
	}
}

// Read reads user input from stdin and stores the result in the key stack.
func (k *Keys) Read() {
	keys, _ := k.ReadArgument()
	k.stack = append(k.stack, keys...)
}

// ReadArgument reads keys from stdin like Read(), but immediately
// returns them instead of storing them in the stack, along with an
// indication on whether this key is an escape/abort one.
func (k *Keys) ReadArgument() (keys []rune, isAbort bool) {
	if k.paused == nil {
		k.paused = make(chan struct{}, 1)
	}

read:
	for {
		// Start reading from os.Stdin in the background.
		read, done := k.readInput()

		for {
			select {
			case keys = <-read:
				// We have read user input keys.
				if len(keys) == 1 && keys[0] == inputrc.Esc {
					isAbort = true
				}

				break read

			case <-k.paused:
				// We are asked to stop reading our keys for some time.
				// Close the reading goroutine, and wait for the caller
				// to notify us that we can start a new one.
				close(done)

				<-k.paused
				k.paused = make(chan struct{}, 1)

				continue
			}
		}
	}

	return keys, isAbort
}

// Pause temporarily pauses reading for input keys.
// This is used when the shell needs to query the terminal for its current state,
// which is output to stdout. Once done with your operation, close the channel to
// resume normal key input scan.
func (k *Keys) Pause() {
	// k.paused <- struct{}{}
	if k.paused == nil {
		k.paused = make(chan struct{}, 1)
	}
	// close(k.paused)
}

func (k *Keys) Resume() {
	// close(k.paused)
}

// Feed can be used to directly add keys to the stack.
// If begin is true, the keys are added on the top of
// the stack, otherwise they are being appended to it.
func (k *Keys) Feed(begin bool, keys ...rune) {
	if len(keys) == 0 {
		return
	}

	if begin {
		k.stack = append(keys, k.stack...)
	} else {
		k.stack = append(k.stack, keys...)
	}
}

// Pop pops (removes) the first key in the stack (last read) and returns it.
// It returns either a key and the empty boolean set to false, or if no keys
// are present, returns a zero rune and empty set to true.
func (k *Keys) Pop() (key rune, empty bool) {
	if len(k.stack) == 0 {
		return rune(0), true
	}

	key = k.stack[len(k.stack)-1]
	k.stack = k.stack[:len(k.stack)-1]

	return key, false
}

// Peek works like Pop(), except that it does not remove the key from the stack.
func (k *Keys) Peek() (key rune, empty bool) {
	if len(k.stack) == 0 {
		return rune(0), true
	}

	return k.stack[0], false
}

// PeekAll returns all the keys from the stack, without deleting them.
func (k *Keys) PeekAll() (keys []rune, empty bool) {
	if len(k.stack) == 0 {
		return k.stack, true
	}

	return k.stack, false
}

// Flush returns all keys stored in the stack and clears it.
func (k *Keys) Flush() []rune {
	keys := string(k.stack)
	k.stack = make([]rune, 0)

	return []rune(keys)
}

func (k *Keys) readInput() (keys chan []rune, done chan struct{}) {
	done = make(chan struct{})
	keys = make(chan []rune)

	go func() {
		buf := make([]byte, keyScanBufSize)

		read, err := os.Stdin.Read(buf)
		if err != nil {
			return
		}

		keys <- []rune(string(buf[:read]))
	}()

	return
}

// readOperator reads a key required by some (rare) widgets that directly read/need
// their argument/operator, without going though operator pending mode first.
// If all is true, we return all keys, including numbers (instead of adding them as iterations.)
// func (rl *Instance) readOperator(all bool) (key string, ret bool) {
// 	rl.enterVioppMode("")
// 	rl.updateCursor()
//
// 	defer func() {
// 		rl.exitVioppMode()
// 		rl.updateCursor()
// 	}()
//
// 	b, i, _ := rl.readInput()
// 	key = string(b[:i])
//
// 	// If the last key is a number, add to iterations instead,
// 	// and read another key input.
// 	if !all {
// 		numMatcher, _ := regexp.Compile(`^[1-9][0-9]*$`)
//
// 		for numMatcher.MatchString(string(key[len(key)-1])) {
// 			rl.iterations += string(key[len(key)-1])
//
// 			b, i, _ = rl.readInput()
// 			key = string(b[:i])
// 		}
// 	}
//
// 	// If the key is an escape key for the current mode.
// 	if len(key) == 1 &&
// 		(key[0] == charEscape) {
// 		ret = true
// 	}
//
// 	return
// }
