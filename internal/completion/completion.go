package completion

// Completer is a function generating completions.
// This is generally used so that a given completer function
// (history, registers, etc) can be cached and reused by the engine.
type Completer func() Values

// Candidate represents a completion candidate.
type Candidate struct {
	Value       string // Value is the value of the completion as actually inserted in the line
	Display     string // When display is not nil, this string is used to display the completion in the menu.
	Description string // A description to display next to the completion candidate.
	Style       string // An arbitrary string of color/text effects to use when displaying the completion.
	Tag         string // All completions with the same tag are grouped together and displayed under the tag heading.

	// A list of runes that are automatically trimmed when a space or a non-nil character is
	// inserted immediately after the completion. This is used for slash-autoremoval in path
	// completions, comma-separated completions, etc.
	noSpace SuffixMatcher
}

// Values is used internally to hold all completion candidates and their associated data.
type Values struct {
	values   RawValues
	Messages Messages
	NoSpace  SuffixMatcher
	Usage    string
	ListLong map[string]bool
	NoSort   map[string]bool
	ListSep  map[string]string

	// Initially this will be set to the part of the current word
	// from the beginning of the word up to the position of the cursor;
	// it may be altered to give a common prefix for all matches.
	PREFIX string
}

// AddRaw adds completion values in bulk.
func AddRaw(values []Candidate) Values {
	return Values{
		values:   RawValues(values),
		ListLong: make(map[string]bool),
		NoSort:   make(map[string]bool),
		ListSep:  make(map[string]string),
	}
}

// updateVirtualComp - Either insert the current completion
// candidate virtually, or on the real line.
// func (rl *Instance) updateVirtualComp() {
// 	cur := rl.currentGroup()
// 	if cur == nil {
// 		return
// 	}
//
// 	completion := cur.selected().Value
// 	prefix := len(rl.tcPrefix)
//
// 	if rl.hasUniqueCandidate() {
// 		rl.insertCandidate()
// 		rl.undoSkipAppend = true
// 		rl.resetCompletion()
// 	} else {
// 		// Special case for the only special escape, which
// 		// if not handled, will make us insert the first
// 		// character of our actual rl.tcPrefix in the candidate.
// 		// TODO: This should be changed.
// 		if strings.HasPrefix(string(rl.tcPrefix), "%") {
// 			prefix++
// 		}
//
// 		if len(completion) >= prefix {
// 			rl.insertCandidateVirtual([]rune(completion[prefix:]))
// 		}
// 	}
// }
