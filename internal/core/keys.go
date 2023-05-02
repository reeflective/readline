package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/reeflective/readline/inputrc"
)

const (
	keyScanBufSize = 1024
)

var rxRcvCursorPos = regexp.MustCompile(`^\x1b\[([0-9]+);([0-9]+)R$`)

// Keys is used read, manage and use keys input by the shell user.
type Keys struct {
	buf      []byte       // Keys read and waiting to be used.
	matched  int          // Number of keys matched against commands.
	waiting  bool         // Currently waiting for keys on stdin.
	reading  bool         // Currently reading keys out of the main loop.
	keysOnce chan []byte  // Passing keys from the main routine.
	cursor   chan []byte  // Cursor coordinates has been read on stdin.
	mutex    sync.RWMutex // Concurrency safety
}

// WaitInput waits until an input key is either read from standard input,
// or directly returns if the key stack still/already has available keys.
func (k *Keys) WaitInput() {
	if len(k.buf) > 0 && k.matched == 0 {
		return
	}

	k.mutex.Lock()
	k.waiting = true
	k.mutex.Unlock()

	defer func() {
		k.mutex.Lock()
		k.waiting = false
		k.mutex.Unlock()
	}()

	for {
		// Start reading from os.Stdin in the background.
		// We will either read keys from user, or an EOF
		// send by ourselves, because we pause reading.
		keys, err := k.readInputFiltered()
		if err != nil && errors.Is(err, io.EOF) {
			return
		}

		if len(keys) == 0 {
			continue
		}

		switch {
		case k.reading:
			k.keysOnce <- keys
			continue

		default:
			k.mutex.RLock()
			k.buf = append(k.buf, keys...)
			k.mutex.RUnlock()
		}

		return
	}
}

func (k *Keys) readInputFiltered() (keys []byte, err error) {
	// Start reading from os.Stdin in the background.
	// We will either read keys from user, or an EOF
	// send by ourselves, because we pause reading.
	buf := make([]byte, keyScanBufSize)

	read, err := os.Stdin.Read(buf)
	if err != nil && errors.Is(err, io.EOF) {
		return
	}

	// Always attempt to extract cursor position info.
	// If found, strip it and keep the remaining keys.
	cursor, keys := k.extractCursorPos(buf[:read])

	if len(cursor) > 0 {
		k.cursor <- cursor
	}

	return keys, nil
}

// ReadKey reads keys from stdin like Read(), but immediately
// returns them instead of storing them in the stack, along with
// an indication on whether this key is an escape/abort one.
func (k *Keys) ReadKey() (keys []rune, isAbort bool) {
	k.mutex.RLock()
	k.keysOnce = make(chan []byte)
	k.reading = true
	k.mutex.RUnlock()

	defer func() {
		k.mutex.RLock()
		k.reading = false
		k.mutex.RUnlock()
	}()

	switch {
	case k.waiting:
		buf := <-k.keysOnce
		keys = []rune(string(buf))
	default:
		buf, _ := k.readInputFiltered()
		keys = []rune(string(buf))
	}

	isAbort = len(keys) == 1 && keys[0] == inputrc.Esc

	return
}

// Pop removes the first byte in the key stack (first read) and returns it.
// It returns either a key and the empty boolean set to false, or if no keys
// are present, returns a zero rune and empty set to true.
func (k *Keys) Pop() (key byte, empty bool) {
	if len(k.buf) == 0 {
		return byte(0), true
	}

	key = k.buf[0]
	k.buf = k.buf[1:]

	return key, false
}

// Peek returns the most ancient rune in the current key stack.
// Since the keys are stored as a list of bytes, and that some
// runes are multibyte code points, the returned rune might be
// longer than one byte.
func (k *Keys) Peek() (key rune, empty bool) {
	if len(k.buf) == 0 {
		return rune(0), true
	}

	keys := []rune(string(k.buf))

	return keys[0], false
}

// PeekAll returns all the keys from the stack, without deleting them.
func (k *Keys) PeekAll() (keys []rune, empty bool) {
	if len(k.buf) == 0 {
		return []rune(string(k.buf)), true
	}

	return []rune(string(k.buf)), false
}

// Feed can be used to directly add keys to the stack.
// If begin is true, the keys are added on the top of
// the stack, otherwise they are being appended to it.
func (k *Keys) Feed(begin bool, keys ...rune) {
	if len(keys) == 0 {
		return
	}

	keyBuf := []byte(string(keys))

	k.mutex.Lock()
	defer k.mutex.Unlock()

	if begin {
		k.buf = append(keyBuf, k.buf...)
	} else {
		k.buf = append(k.buf, keyBuf...)
	}
}

// MarkMatched is used to indicate how many keys have been evaluated
// against the shell commands in the dispatching process (regardless of
// if a command was matched or not).
// This function should normally not be used by external users of the library.
func (k *Keys) MarkMatched(keys ...byte) {
	if len(keys) == 0 {
		return
	}

	k.matched = len(keys)
	k.buf = append(keys, k.buf...)
}

// FlushUsed drops the keys that have matched a given command.
func (k *Keys) FlushUsed() {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	defer func() {
		k.matched = 0
	}()

	if k.matched == 0 {
		return
	}

	switch {
	case len(k.buf) < k.matched:
		k.buf = nil
	default:
		k.buf = k.buf[k.matched:]
	}
}

// Flush returns all keys stored in the stack and clears it.
func (k *Keys) Flush() []rune {
	keys := string(k.buf)
	k.buf = make([]byte, 0)

	return []rune(keys)
}

// GetCursorPos returns the current cursor position in the terminal.
// It is safe to call this function even if the shell is reading input.
func (k *Keys) GetCursorPos() (x, y int) {
	k.cursor = make(chan []byte)

	disable := func() (int, int) {
		os.Stderr.WriteString("\r\ngetCursorPos() not supported by terminal emulator, disabling....\r\n")
		return -1, -1
	}

	var cursor []byte

	// Echo the query and wait for the main key
	// reading routine to send us the response back.
	fmt.Print("\x1b[6n")

	switch {
	case k.waiting, k.reading:
		cursor = <-k.cursor
	default:
		buf := make([]byte, keyScanBufSize)

		read, err := os.Stdin.Read(buf)
		if err != nil && errors.Is(err, io.EOF) {
			return disable()
		}

		cursor = buf[:read]
	}

	if len(cursor) == 0 {
		return disable()
	}

	match := rxRcvCursorPos.FindAllStringSubmatch(string(cursor), 1)
	if len(match) == 0 {
		return disable()
	}

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

func (k *Keys) extractCursorPos(keys []byte) (cursor, remain []byte) {
	if !rxRcvCursorPos.Match(keys) {
		return cursor, keys
	}

	allCursors := rxRcvCursorPos.FindAll(keys, -1)
	cursor = allCursors[len(allCursors)-1]
	remain = rxRcvCursorPos.ReplaceAll(keys, nil)

	return
}
