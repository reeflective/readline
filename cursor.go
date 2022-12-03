package readline

import (
	"os"
	"regexp"
	"strconv"
)

const (
	cursorBlock             = "\x1b[2 q"
	cursorUnderline         = "\x1b[4 q"
	cursorBeam              = "\x1b[6 q"
	cursorBlinkingBlock     = "\x1b[1 q"
	cursorBlinkingUnderline = "\x1b[3 q"
	cursorBlinkingBeam      = "\x1b[5 q"
	cursorUserDefault       = "\x1b[0 q"
)

func (rl *Instance) updateCursor() {
	// The local keymap most of the time has priority
	switch rl.local {
	case viopp:
		print(cursorUnderline)
		return
	case visual:
		print(cursorBlock)
		return
	default:
	}

	// But if not, we check for the global keymap
	switch rl.main {
	case emacs:
		print(cursorBlinkingBlock)
	case viins:
		print(cursorBlinkingBeam)
	case vicmd:
		print(cursorBlinkingBlock)
	}
}

// findAndMoveCursor finds a specified character in the line, either forward
// or backward, a specified number of times, and moves the cursor to it.
func (rl *Instance) findAndMoveCursor(key string, count int, forward, skip bool) {
	if key == "" {
		return
	}

	cursor := rl.pos

	for {
		// Move the cursor in the specified direction and within bounds.
		if forward {
			cursor++
			if cursor > len(rl.line)-1 {
				break
			}
		} else {
			cursor--
			if cursor < 0 {
				break
			}
		}

		// Check if character matches
		if string(rl.line[cursor]) == key {
			count--
		}

		// When the count is 0, we matched the character count times
		if count == 0 {
			break
		}
	}

	if count > 0 {
		return
	}

	if skip {
		if forward {
			cursor--
		} else {
			cursor++
		}
	}

	// TODO: Should we return it instead of assigning it ?
	rl.pos = cursor
}

// Lmorg code
// -------------------------------------------------------------------------------

func leftMost() []byte {
	fd := int(os.Stdout.Fd())
	w, _, err := GetSize(fd)
	if err != nil {
		return []byte{'\r', '\n'}
	}

	b := make([]byte, w+1)
	for i := 0; i < w; i++ {
		b[i] = ' '
	}
	b[w] = '\r'

	return b
}

var rxRcvCursorPos = regexp.MustCompile("^\x1b([0-9]+);([0-9]+)R$")

func (rl *Instance) getCursorPos() (x int, y int) {
	if !rl.EnableGetCursorPos {
		return -1, -1
	}

	disable := func() (int, int) {
		os.Stderr.WriteString("\r\ngetCursorPos() not supported by terminal emulator, disabling....\r\n")
		rl.EnableGetCursorPos = false
		return -1, -1
	}

	print(seqGetCursorPos)
	b := make([]byte, 64)
	i, err := os.Stdin.Read(b)
	if err != nil {
		return disable()
	}

	if !rxRcvCursorPos.Match(b[:i]) {
		return disable()
	}

	match := rxRcvCursorPos.FindAllStringSubmatch(string(b[:i]), 1)
	y, err = strconv.Atoi(match[0][1])
	if err != nil {
		return disable()
	}

	x, err = strconv.Atoi(match[0][2])
	if err != nil {
		return disable()
	}

	return x, y
}

// DISPLAY ------------------------------------------------------------
// All cursorMoveFunctions move the cursor as it is seen by the user.
// This means that they are not used to keep any reference point when
// when we internally move around clearning and printing things

func moveCursorUp(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dA", i)
}

func moveCursorDown(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dB", i)
}

func moveCursorForwards(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dC", i)
}

func moveCursorBackwards(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dD", i)
}

func (rl *Instance) backspace() {
	if len(rl.line) == 0 || rl.pos == 0 {
		return
	}

	rl.deleteBackspace()
}

func (rl *Instance) getCursorStyle(mode string) (style string) {
	switch mode {
	}
	return
}

func (rl *Instance) moveCursorByAdjust(adjust int) {
	switch {
	case adjust > 0:
		rl.pos += adjust
	case adjust < 0:
		rl.pos += adjust
	}

	// The position can never be negative
	if rl.pos < 0 {
		rl.pos = 0
	}

	// The cursor can never be longer than the line
	if rl.pos > len(rl.line) {
		rl.pos = len(rl.line)
	}

	// If we are at the end of line, and not in Insert mode, move back one.
	if rl.main == vicmd && (rl.pos == len(rl.line)) && len(rl.line) > 0 {
		rl.pos--
	} else if rl.main == viins && rl.searchMode == HistoryFind && rl.modeAutoFind {
		rl.pos--
	}
}
