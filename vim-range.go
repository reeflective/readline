package readline

// 	// Out loop:
// 	// If last key is i or a, read one more key
// 	surroundMatcher, _ := regexp.Compile(`[ia]`)
// 	for surroundMatcher.MatchString(string(key[len(key)-1])) {
// 		b, i, _ := rl.readInput()
// 		key += string(b[:i])
// 	}

// TODO: This should return a boolean to notify if a range was actually selected.
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
// func matchRangeAction(keys string) (command, count, navKey string) {
// 	// All matchers
// 	changeAroundWord, _ := regexp.Compile(`^c([1-9][0-9]*)?[ia][wW]$`)
// 	aroundWordEnd, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?[ia][eE]$`)
// 	changeForwardWord, _ := regexp.Compile(`^c([1-9][0-9]*)?w$`)
// 	changeForwardBlankWord, _ := regexp.Compile(`^c([1-9][0-9]*)?W$`)
// 	changeForwardWordEnd, _ := regexp.Compile(`^c([1-9][0-9]*)?e$`)
// 	changeForwardBlankWordEnd, _ := regexp.Compile(`^c([1-9][0-9]*)?E$`)
// 	backwardWord, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?[bB]$`)
// 	charNextMatch, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?([FT].?)$`)
// 	downLine, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?j$`)
// 	upLine, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?j$`)
// 	backwardChar, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?h$`)
// 	forwardChar, _ := regexp.Compile(`^[cdy]([1-9][0-9]*)?l$`)
// 	repeat, _ := regexp.Compile(`^.([1-9][0-9]*)?([^0-9]+)$`)
//
// 	// Note that in all cases, the submatch of interest is the counter:
// 	// The navigation action is known by us automatically if we have a match.
// 	if match := changeAroundWord.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 		navKey = keys[len(keys)-2:]
// 	} else if match := aroundWordEnd.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 	} else if match := changeForwardWord.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 		navKey = "e"
// 	} else if match := changeForwardBlankWord.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 		navKey = "E"
// 	} else if match := changeForwardWordEnd.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 		navKey = "e"
// 	} else if match := changeForwardBlankWordEnd.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 		navKey = "E"
// 	} else if match := backwardWord.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 		navKey = keys[len(keys)-1:]
// 	} else if match := charNextMatch.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 		navKey = match[2]
// 	} else if match := downLine.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 	} else if match := upLine.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 	} else if match := backwardChar.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 		navKey = "h"
// 	} else if match := forwardChar.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 		// TODO HERE: substract 1 from count
// 		navKey = count + "l"
// 	} else if match := repeat.FindStringSubmatch(keys); len(match) > 0 {
// 		command = string(keys[0])
// 		count = match[1]
// 		navKey = match[2]
// 	}
//
// 	// We perform at least one action.
// 	if count == "" {
// 		count = "1"
// 	}
//
// 	return
// }
