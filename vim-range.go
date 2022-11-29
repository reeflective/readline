package readline

import (
	"regexp"
)

// readRangeKeys recursively reads for input keys that have an effect on a range
// (iterations, movement keys, etc), and updates the current readline state accordingly.
// This handler is used with widgets accepting such range arguments, such as y,c, and d.
func (rl *Instance) readRangeKeys(key string, keymap keyMap) {
	// We always enter the operator pending mode when using ranges/movements.

	// We need at least 2 keys (one is the key parameter, the other is to be read here)

	// In loop:
	// If the last key is a number, add to iterations.

	// Out loop:
	// If last key is i or a, read one more key

	// Enter visual mode (this will not be noticed by the readline user)

	// Now we match the keys against some regular expressions.
	// This will basically split the keys in 3 parts: a caller key, a count, and a navigation key.

	// Finally, handle navigation

	// When no range was selected, we are done with visual mode (resets mark/cursor ranges)

	// Post-navigation handling.
}

// matchRangeAction tries to match a range expression against a series of regular expressions,
// returning the count (if any, or 1) and an optional navigation action key.
//
//  Selection Cases:
//
//  1. SAMPLE: `word1  word2  w`, CURSOR: at `w` of `word1`
//
//   c[we] -> `word1`
//   c2[we] -> `word1  word2`
//   ve -> `word1`
//   v2e -> `word1  word2`
//   vw -> `word1  w`
//   v2w -> `word1  word2  w`
//   [dy]e -> `word1`
//   [dy]2e -> `word1  word2`
//   [dy]w -> `word1  `
//   [dy]2w -> `word1  word2  `
//   [cdyv]iw -> `word1`
//   [cdyv]aw -> `word1  `
//   [cdyv]2iw -> `word1  `
//   [cdyv]2aw -> `word1  word2  `
//
//  2. SAMPLE: `a  bb  c  dd`, CURSOR: at `a`
//
//   cw -> `a`
//   c2w -> `a bb`
//   ce -> `a bb`
//   c2e -> `a bb c`
//
//  3. SAMPLE: ` .foo.  bar.  baz.`, CURSOR: at `f`
//
//   c[WE] -> `foo.`
//   c2[WE] -> `foo.  bar.`
//   vE -> `foo.`
//   v2E -> `foo.  bar.`
//   vW -> `foo.  b`
//   v2W -> `foo.  bar.  b`
//   d2W -> `foo.  bar.  b`
//   [dy]E -> `foo.`
//   [dy]2E -> `foo.  bar.`
//   [dy]W -> `foo.  `
//   [dy]2W -> `foo.  bar.  `
//   [cdyv]iW -> `.foo.`
//   [cdyv]aW -> `.foo.  `
//   [cdyv]2iW -> `.foo.  `
//   [cdyv]2aW -> `.foo.  bar.  `
//
//  4. SAMPLE: ` .foo.bar.baz.`, CURSOR: at `r`
//
//   [cdy]b -> `ba`
//   [cdy]B -> `.foo.ba`
//   vb -> `bar`
//   vB -> `.foo.bar`
//   vFf -> `foo.bar`
//   vTf -> `oo.bar`
//   [cdyv]fz -> `r.baz`
//   [cdy]Ff -> `foo.ba`
//   [cdyv]tz -> `r.ba`
//   [cdy]Tf -> `oo.ba`
//
func matchRangeAction(keys string) (count, navAction string) {
	count = "1"

	// All matchers
	changeAroundWord, _ := regexp.Compile(`^c([1-9][0-9]*)?[ia][wW]$`)
	aroundWordEnd, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?[ia][eE]$`)
	changeForwardWord, _ := regexp.Compile(`^c([1-9][0-9]*)?w$`)
	changeForwardBlanWord, _ := regexp.Compile(`^c([1-9][0-9]*)?W$`)
	changeForwardWordEnd, _ := regexp.Compile(`^c([1-9][0-9]*)?e$`)
	changeForwardBlankWordEnd, _ := regexp.Compile(`^c([1-9][0-9]*)?E$`)
	backwardWord, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?[bB]$`)
	charNextMatch, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?([FT].?)$`)
	nextLine, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?j$`)
	previousLine, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?h$`)
	forwardChar, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?l$`)
	repeat, _ := regexp.Compile(`^.([1-9][0-9]*)?([^0-9]+)$`)

	// Note that in all cases, the submatch of interest is the counter:
	// The navigation action is known by us automatically if we have a match.
	//
	// Note that when no numeric argument is found, match[1] will be empty.
	if match := changeAroundWord.FindStringSubmatch(keys); len(match) > 0 {
	} else if match := aroundWordEnd.FindStringSubmatch(keys); len(match) > 0 {
	} else if match := changeForwardWord.FindStringSubmatch(keys); len(match) > 0 {
	} else if match := changeForwardBlanWord.FindStringSubmatch(keys); len(match) > 0 {
	} else if match := changeForwardWordEnd.FindStringSubmatch(keys); len(match) > 0 {
	} else if match := changeForwardBlankWordEnd.FindStringSubmatch(keys); len(match) > 0 {
	} else if match := backwardWord.FindStringSubmatch(keys); len(match) > 0 {
	} else if match := charNextMatch.FindStringSubmatch(keys); len(match) > 0 {
	} else if match := nextLine.FindStringSubmatch(keys); len(match) > 0 {
	} else if match := previousLine.FindStringSubmatch(keys); len(match) > 0 {
	} else if match := forwardChar.FindStringSubmatch(keys); len(match) > 0 {
	} else if match := repeat.FindStringSubmatch(keys); len(match) > 0 {
	} else {
	}

	return
}
