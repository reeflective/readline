package term

import (
	"fmt"
	"os"
)

// termFile is the file descriptor used for all low-level terminal queries (size,
// cursor position) and escape sequences. We deliberately use stderr rather than
// stdout: stdout is the stream most likely to be redirected (e.g. `app | other`),
// whereas stderr usually stays attached to the controlling terminal, giving a
// reliable terminal size even when the program's output is piped.
var termFile = os.Stderr

// fallback terminal width when we can't get it through query.
var defaultTermWidth = 80

// GetWidth returns the width of the terminal or 80 if it cannot be established.
func GetWidth() (termWidth int) {
	var err error
	fd := int(termFile.Fd())
	termWidth, _, err = GetSize(fd)

	if err != nil || termWidth == 0 {
		termWidth = defaultTermWidth
	}

	return
}

// GetLength returns the length of the terminal
// (Y length), or 80 if it cannot be established.
func GetLength() int {
	termFd := int(termFile.Fd())

	_, length, err := GetSize(termFd)
	if err != nil || length == 0 {
		return defaultTermWidth
	}

	return length
}

func printf(format string, a ...interface{}) {
	WriteString(fmt.Sprintf(format, a...))
}

// EnableBracketedPaste enables bracketed paste mode.
func EnableBracketedPaste() {
	WriteString(BracketedPasteStart)
}

// DisableBracketedPaste disables bracketed paste mode.
func DisableBracketedPaste() {
	WriteString(BracketedPasteEnd)
}
