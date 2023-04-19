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
func IsSurround(bchar, echar rune) bool {
	switch bchar {
	case '{':
		return echar == '}'
	case '(':
		return echar == ')'
	case '[':
		return echar == ']'
	case '<':
		return echar == '>'
	case '"':
		return echar == '"'
	case '\'':
		return echar == '\''
	}

	return echar == bchar
}

// AdjustSurroundQuotes returns the correct mark and cursor positions when
// we want to know where a shell word enclosed with quotes (and potentially
// having inner ones) starts and ends.
func AdjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos int) (mark, cpos int) {
	mark = -1
	cpos = -1

	if (sBpos == -1 || sEpos == -1) && (dBpos == -1 || dEpos == -1) {
		return
	}

	doubleFirstAndValid := (dBpos < sBpos && // Outtermost
		dBpos >= 0 && // Double found
		sBpos >= 0 && // compared with a found single
		dEpos > sEpos) // ensuring that we are not comparing unfound

	singleFirstAndValid := (sBpos < dBpos &&
		sBpos >= 0 &&
		dBpos >= 0 &&
		sEpos > dEpos)

	if (sBpos == -1 || sEpos == -1) || doubleFirstAndValid {
		mark = dBpos
		cpos = dEpos
	} else if (dBpos == -1 || dEpos == -1) || singleFirstAndValid {
		mark = sBpos
		cpos = sEpos
	}

	return
}
