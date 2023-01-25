package strutil

// MatchSurround returns the matching character of a rune that
// is either a bracket/brace/parenthesis, or a single/double quote.
func MatchSurround(r rune) (bchar, echar rune) {
	bchar = r
	echar = r

	switch bchar {
	case '{':
		echar = '}'
	case '(':
		echar = ')'
	case '[':
		echar = ']'
	case '<':
		echar = '>'
	case '}':
		bchar = '{'
		echar = '}'
	case ')':
		bchar = '('
		echar = ')'
	case ']':
		bchar = '['
		echar = ']'
	case '>':
		bchar = '<'
		echar = '>'
	case '"':
		bchar = '"'
		echar = '"'
	case '\'':
		bchar = '\''
		echar = '\''
	}

	return bchar, echar
}

// IsSurround returns true if the character is a quote or a bracket/brace, etc.
func IsSurround(r, e rune) bool {
	switch r {
	case '{':
		return e == '}'
	case '(':
		return e == ')'
	case '[':
		return e == ']'
	case '<':
		return e == '>'
	case '"':
		return e == '"'
	case '\'':
		return e == '\''
	}

	return e == r
}
