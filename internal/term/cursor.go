package term

import (
	"os"
	"regexp"
	"strconv"
)

var (
	rxRcvCursorPos  = regexp.MustCompile(`^\x1b\[([0-9]+);([0-9]+)R$`)
	cursorPosBufLen = 64
)

// GetCursorPos queries the terminal for the current cursor
// coordinates. Returns -1, -1 if it could not determine them.
func GetCursorPos() (x int, y int) {
	// if !rl.EnableGetCursorPos {
	// 	return -1, -1
	// }

	disable := func() (int, int) {
		os.Stderr.WriteString("\r\ngetCursorPos() not supported by terminal emulator, disabling....\r\n")
		// rl.hint = []rune(seqFgRed + "getCursorPos() not supported by terminal emulator, disabling...")
		// rl.EnableGetCursorPos = false
		return -1, -1
	}

	print(getCursorPos)

	buf := make([]byte, cursorPosBufLen)

	read, err := os.Stdin.Read(buf)
	if err != nil {
		return disable()
	}

	if !rxRcvCursorPos.Match(buf[:read]) {
		return disable()
	}

	match := rxRcvCursorPos.FindAllStringSubmatch(string(buf[:read]), 1)

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
