package strutil

import "strings"

// IsPunctuation returns true if the rune is non-blank word delimiter.
func IsPunctuation(r rune) bool {
	if (r >= 33 && 47 >= r) ||
		(r >= 58 && 64 >= r) ||
		(r >= 91 && 94 >= r) ||
		r == 96 ||
		(r >= 123 && 126 >= r) {
		return true
	}

	return false
}

// TrimWhiteSpaceRight trims all trailing whitespaces, tabs and newlines from a string.
func TrimWhiteSpaceRight(oldString string) (newString string) {
	return strings.TrimRightFunc(oldString, func(r rune) bool {
		if r == ' ' || r == '\t' || r == '\n' {
			return true
		}
		return false
	})
}

// IsBracket returns true if the character is an opening/closing bracket/brace/parenthesis.
func IsBracket(r rune) bool {
	if r == '(' ||
		r == ')' ||
		r == '{' ||
		r == '}' ||
		r == '[' ||
		r == ']' {
		return true
	}

	return false
}
