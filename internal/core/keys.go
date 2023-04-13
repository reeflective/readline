package core

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/reeflective/readline/inputrc"
)

const (
	keyScanBufSize = 1024
)

var (
	rxRcvCursorPos  = regexp.MustCompile(`^\x1b\[([0-9]+);([0-9]+)R$`)
	cursorPosBufLen = 64
)

// Keys is used read, manage and use keys input by the shell user.
type Keys struct {
	stack       []rune      // Keys that have been read, not yet consumed.
	skipRead    bool        // Something fed the stack on its own, skip reading user input.
	macroCalled bool        // The last feeding is the macro engine, don't adjust cachedSkip yet.
	cachedSkip  int         // The number of keys fed for, to consume before reading input.
	matchedKeys int         // A call to keys.Matched() has happened, waiting for those to be dropped.
	paused      bool        // Reading the cursor position.
	reading     bool        // Reading user input.
	cursor      chan string // Passing cursor coordinates output.
	mutex       sync.RWMutex
}

// Read reads user input from stdin and stores the result in the key stack.
// If the macro engine has fed keys to match against, this will skip reading
// from standard input altogther.
func (k *Keys) Read() {
	if k.skipRead {
		return
	}

	keys, _ := k.ReadArgument()
	k.stack = append(k.stack, keys...)
}

// ReadArgument reads keys from stdin like Read(), but immediately
// returns them instead of storing them in the stack, along with an
// indication on whether this key is an escape/abort one.
func (k *Keys) ReadArgument() (keys []rune, isAbort bool) {
	k.reading = true
	k.cursor = make(chan string)

	defer func() {
		k.reading = false
	}()

	for {
		// Start reading from os.Stdin in the background.
		// We will either read keys from user, or an EOF
		// send by ourselves, because we pause reading.
		buf := make([]byte, keyScanBufSize)

		read, err := os.Stdin.Read(buf)
		if err != nil {
			return
		}

		keys = []rune(string(buf[:read]))

		if k.paused {
			k.cursor <- string(buf[:read])
			continue
		}

		// We have read user input keys.
		if len(keys) == 1 && keys[0] == inputrc.Esc {
			isAbort = true
		}

		break
	}

	return keys, isAbort
}

// Feed can be used to directly add keys to the stack.
// If begin is true, the keys are added on the top of the stack,
// otherwise they are being appended to it. If skipRead is true,
// the next call to keys.Read() will immediately return.
func (k *Keys) Feed(begin, skipRead bool, keys ...rune) {
	if len(keys) == 0 {
		return
	}

	// Until all those keys have been either matched
	// or dropped, don't read from stdin.
	k.skipRead = skipRead
	if k.skipRead {
		k.cachedSkip = len(keys)
	}

	k.macroCalled = true

	keys = []rune(string(keys))

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

	key = k.stack[0]
	k.stack = k.stack[1:]

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

// Matched performs two things feeds the provided keys back to the stack,
// and will drop them once the keys.FlushUsed function is called.
func (k *Keys) Matched(keys ...rune) {
	if len(keys) == 0 {
		return
	}

	k.matchedKeys = len(keys)

	keys = []rune(string(keys))
	k.stack = append(keys, k.stack...)
}

// FlushUsed drops the number of keys that have been fed with the last keys.Feed()
// call from the key stack. If the former call used skipRead = true, no keys are flushed.
func (k *Keys) FlushUsed() {
	if k.skipRead && !k.macroCalled {
		k.cachedSkip -= k.matchedKeys
		k.skipRead = k.cachedSkip > 0
	}

	k.macroCalled = false

	if k.matchedKeys == 0 {
		// println("No matched keys")
		return
	}

	// println("Keys before: " + strconv.QuoteToASCII(string(k.stack)))

	switch {
	case len(k.stack) < k.matchedKeys:
		k.stack = nil
	default:
		k.stack = k.stack[k.matchedKeys:]
	}

	// println("Keys after: " + strconv.QuoteToASCII(string(k.stack)))
	k.matchedKeys = 0

	// if k.skipRead {
	// 	os.Exit(0)
	// }
}

// Flush returns all keys stored in the stack and clears it.
func (k *Keys) Flush() []rune {
	keys := string(k.stack)
	k.stack = make([]rune, 0)

	return []rune(keys)
}

// GetCursorPos returns the current cursor position in the terminal.
// This is normally safe to call even if the shell is reading input.
func (k *Keys) GetCursorPos() (x, y int) {
	defer func() {
		k.paused = false
	}()

	disable := func() (int, int) {
		os.Stderr.WriteString("\r\ngetCursorPos() not supported by terminal emulator, disabling....\r\n")
		return -1, -1
	}

	var cursor string

	// Notify pause if we are reading.
	k.mutex.RLock()
	k.paused = k.reading
	k.mutex.RUnlock()

	fmt.Print("\x1b[6n")

	// Either read the output, or wait for
	// the main reading routine to catch it.
	if !k.reading {
		buf := make([]byte, cursorPosBufLen)

		read, err := os.Stdin.Read(buf)
		if err != nil {
			return disable()
		}
		cursor = string(buf[:read])
	} else {
		cursor = <-k.cursor
	}

	// Find it and return coordinates.
	if !rxRcvCursorPos.MatchString(cursor) {
		return disable()
	}

	match := rxRcvCursorPos.FindAllStringSubmatch(cursor, 1)

	y, err := strconv.Atoi(match[0][1])
	if err != nil {
		return disable()
	}

	x, err = strconv.Atoi(match[0][2])
	if err != nil {
		return disable()
	}

	return x, y
}
