package term

import (
	"fmt"
	"os"
)

// StdoutTerm is the terminal file descriptor for stdout,
// it can be overwritten to use other file descriptors or custom io.Writers.
var StdoutTerm *os.File = os.Stdout

// fallback terminal width when we can't get it through query.
var defaultTermWidth = 80

// GetWidth returns the width of Stdout or 80 if the width cannot be established.
func GetWidth() (termWidth int) {
	var err error
	fd := int(StdoutTerm.Fd())
	termWidth, _, err = GetSize(fd)

	if err != nil || termWidth == 0 {
		termWidth = defaultTermWidth
	}

	return
}

// GetLength returns the length of the terminal
// (Y length), or 80 if it cannot be established.
func GetLength() int {
	termFd := int(StdoutTerm.Fd())

	_, length, err := GetSize(termFd)
	if err != nil || length == 0 {
		return defaultTermWidth
	}

	return length
}

func printf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	fmt.Print(s)
}
