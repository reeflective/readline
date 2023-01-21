package term

import (
	"fmt"
	"os"
	"regexp"
	"unicode/utf8"

	"golang.org/x/term"
)

// GetWidth returns the width of Stdout or 80 if the width cannot be established.
func GetWidth() (termWidth int) {
	var err error
	fd := int(os.Stdout.Fd())
	termWidth, _, err = GetSize(fd)
	if err != nil {
		termWidth = 80 // The defacto standard on older terms
	}

	return
}

// GetLength returns the length of the terminal
// (Y length), or 80 if it cannot be established.
func GetLength() int {
	width, _, err := term.GetSize(0)
	if err != nil || width == 0 {
		return 80
	}

	return width
}

func printf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	print(s)
}

func print(s string) {
	os.Stdout.WriteString(s)
}

var rxAnsiSgr = regexp.MustCompile("\x1b\\[[:;0-9]+m")

// Gets the number of runes in a string.
func strLen(s string) int {
	s = rxAnsiSgr.ReplaceAllString(s, "")
	return utf8.RuneCountInString(s)
}
