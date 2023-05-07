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

var rxRcvCursorPos = regexp.MustCompile(`\x1b\[([0-9]+);([0-9]+)R`)

// Keys is used read, manage and use keys input by the shell user.
type Keys struct {
	buf        []byte // Keys read and waiting to be used.
	matched    []rune
	matchedLen int // Number of keys matched against commands.
	macroKeys  []rune
	waiting    bool         // Currently waiting for keys on stdin.
	reading    bool         // Currently reading keys out of the main loop.
	keysOnce   chan []byte  // Passing keys from the main routine.
	cursor     chan []byte  // Cursor coordinates has been read on stdin.
	mutex      sync.RWMutex // Concurrency safety
}

// PopKey is used to pop a key off the key stack without
// yet marking this key as having matched a command.
func PopKey(keys *Keys) (key byte, empty bool) {
	switch {
	case len(keys.buf) > 0:
		key = keys.buf[0]
		keys.buf = keys.buf[1:]
	case len(keys.macroKeys) > 0:
		key = byte(keys.macroKeys[0])
		keys.macroKeys = keys.macroKeys[1:]
	default:
		return byte(0), true
	}

	return key, false
}

// MatchedKeys is used to indicate how many keys have been evaluated
// against the shell commands in the dispatching process (regardless of
// if a command was matched or not).
// This function should normally not be used by external users of the library.
func MatchedKeys(keys *Keys, matched []byte, args ...byte) {
	if len(matched) > 0 {
		keys.matched = append(keys.matched, []rune(string(matched))...)
		keys.matchedLen += len(matched)
	}

	if len(args) > 0 {
		keys.buf = append(args, keys.buf...)
	}
}

// WaitInput waits until an input key is either read from standard input,
// or directly returns if the key stack still/already has available keys.
func (k *Keys) WaitInput() {
	if len(k.buf) > 0 && k.matchedLen == 0 {
		return
	}

	// The macro engine might have fed some keys
	if len(k.macroKeys) > 0 {
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

// ReadKey reads keys from stdin like Read(), but immediately
// returns them instead of storing them in the stack, along with
// an indication on whether this key is an escape/abort one.
func (k *Keys) ReadKey() (key rune, isAbort bool) {
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
	case len(k.macroKeys) > 0:
		key = k.macroKeys[0]
		k.macroKeys = k.macroKeys[1:]

	case k.waiting:
		buf := <-k.keysOnce
		key = []rune(string(buf))[0]
	default:
		buf, _ := k.readInputFiltered()
		key = []rune(string(buf))[0]
	}

	// Always mark those keys as matched, so that
	// if the macro engine is recording, it will
	// capture them
	k.matched = append(k.matched, key)
	k.matchedLen++

	return key, key == inputrc.Esc
}

// Pop removes the first byte in the key stack (first read) and returns it.
// It returns either a key and the empty boolean set to false, or if no keys
// are present, returns a zero rune and empty set to true.
func (k *Keys) Pop() (key byte, empty bool) {
	switch {
	case len(k.buf) > 0:
		key = k.buf[0]
		k.buf = k.buf[1:]
	case len(k.macroKeys) > 0:
		key = byte(k.macroKeys[0])
		k.macroKeys = k.macroKeys[1:]
	default:
		return byte(0), true
	}

	k.matched = append(k.matched, rune(key))
	k.matchedLen++

	return key, false
}

// Caller returns the key that has matched the command currently being ran.
func (k *Keys) Caller() (keys []rune) {
	return k.matched
}

// Feed can be used to directly add keys to the stack.
// If begin is true, the keys are added on the top of
// the stack, otherwise they are being appended to it.
func (k *Keys) Feed(begin bool, keys ...rune) {
	if len(keys) == 0 {
		return
	}

	keyBuf := []rune(string(keys))

	k.mutex.Lock()
	defer k.mutex.Unlock()

	if begin {
		k.macroKeys = append(keyBuf, k.macroKeys...)
		// k.buf = append(keyBuf, k.buf...)
	} else {
		k.macroKeys = append(k.macroKeys, keyBuf...)
		// k.buf = append(k.buf, keyBuf...)
	}
}

// FlushUsed drops the keys that have matched a given command.
func (k *Keys) FlushUsed() {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	defer func() {
		k.matchedLen = 0
		k.matched = nil
	}()

	if k.matchedLen == 0 {
		return
	}

	k.matched = nil
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
	var match [][]string

	// Echo the query and wait for the main key
	// reading routine to send us the response back.
	fmt.Print("\x1b[6n")

	// In order not to get stuck with an input that might be user-one
	// (like when the user typed before the shell is fully started, and yet not having
	// queried cursor yet), we keep reading from stdin until we find the cursor response.
	// Everything else is passed back as user input.
	for {
		switch {
		case k.waiting, k.reading:
			cursor = <-k.cursor
		default:
			buf := make([]byte, keyScanBufSize)

			read, err := os.Stdin.Read(buf)
			if err != nil {
				return disable()
			}

			cursor = buf[:read]
		}

		// We have read (or have been passed) something.
		if len(cursor) == 0 {
			return disable()
		}

		// Attempt to locate cursor response in it.
		match = rxRcvCursorPos.FindAllStringSubmatch(string(cursor), 1)

		// If there is something but not cursor answer, its user input.
		if len(match) == 0 && len(cursor) > 0 {
			k.mutex.RLock()
			k.buf = append(k.buf, cursor...)
			k.mutex.RUnlock()

			continue
		}

		// And if empty, then we should abort.
		if len(match) == 0 {
			return disable()
		}

		break
	}

	// We know that we have a cursor answer, process it.
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
