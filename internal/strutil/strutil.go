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

// GetQuotedWordStart returns the position of the outmost containing quote
// of the word (going backward from the end of the line), if the current word
// is a shell word that is not closed yet.
// Ex: `this 'quote contains "surrounded" words`. the outermost quote is the single one.
func GetQuotedWordStart(line []rune) (unclosed bool, pos int) {
	var (
		single, double bool
		spos, dpos     = -1, -1
	)

	for pos, char := range line {
		switch char {
		case '\'':
			single = !single
			spos = pos
		case '"':
			double = !double
			dpos = pos
		default:
			continue
		}
	}

	if single && double {
		unclosed = true

		if spos < dpos {
			pos = spos
		} else {
			pos = dpos
		}

		return
	}

	if single {
		unclosed = true
		pos = spos
	} else if double {
		unclosed = true
		pos = dpos
	}

	return unclosed, pos
}
