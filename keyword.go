package readline

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	// Max integer value on 64 bit architecture.
	maxInt = 9223372036854775807
)

// keywordSwitcher is a function modifying a given word, returning:
// @done => If true, the handler performed a change.
// @new  => The updated word.
// @bpos => Offset to begin position.
// @epos => Offset to end position.
type keywordSwitcher func(word string, increase bool) (done bool, new string, bpos, epos int)

// keywordSwitchers returns all keywordSwitchers of the shell.
func (rl *Instance) keywordSwitchers() []keywordSwitcher {
	return []keywordSwitcher{
		rl.switchNumber,
		rl.switchBoolean,
		rl.switchWeekday,
		rl.switchOperator,
	}
}

func (rl *Instance) switchNumber(word string, increase bool) (done bool, new string, bpos, epos int) {
	vii := rl.getViIterations()
	if !increase {
		vii = -vii
	}

	if done, new, bpos, epos = rl.switchHexa(word, vii); done {
		return
	}

	if done, new, bpos, epos = rl.switchBinary(word, vii); done {
		return
	}

	if done, new, bpos, epos = rl.switchDecimal(word, vii); done {
		return
	}

	return
}

// Hexadecimal cases:
//
// 1. Increment:
// 0xDe => 0xdf
// 0xdE => 0xDF
// 0xde0 => 0xddf
// 0xffffffffffffffff => 0x0000000000000000
// 0X9 => 0XA
// 0Xdf => 0Xe0
//
// 2. Decrement:
// 0xdE0 => 0xDDF
// 0xffFf0 => 0xfffef
// 0xfffF0 => 0xFFFEF
// 0x0 => 0xffffffffffffffff
// 0X0 => 0XFFFFFFFFFFFFFFFF
// 0Xf => 0Xe
func (rl *Instance) switchHexa(word string, inc int) (done bool, new string, bpos, epos int) {
	hexadecimal, _ := regexp.Compile(`[^0-9]?(0[xX][0-9a-fA-F]*)`)
	match := hexadecimal.FindString(word)
	if match == "" {
		return
	}

	done = true

	number := match
	prefix := match[:2]
	hexVal := number[len(prefix):]
	indexes := hexadecimal.FindStringIndex(word)
	mbegin, mend := indexes[0], indexes[1]
	bpos, epos = mbegin, mend

	// lower := true
	if match, _ := regexp.MatchString(`[A-Z][0-9]*$`, number); !match {
		hexVal = strings.ToUpper(hexVal)
	}

	num, err := strconv.ParseInt(hexVal, 16, 64)
	if err != nil {
		done = false
		return
	}

	max64Bit := big.NewInt(maxInt)
	bigNum := big.NewInt(num)
	bigInc := big.NewInt(int64(inc))
	sum := bigNum.Add(bigNum, bigInc)
	zero := big.NewInt(0)

	numBefore := num

	if sum.Cmp(zero) < 0 {
		offset := bigInc.Sub(max64Bit, sum.Abs(sum))
		if offset.IsInt64() {
			num = offset.Int64()
		} else {
			num = math.MaxInt64
		}
	} else if sum.CmpAbs(max64Bit) >= 0 {
		offset := bigInc.Sub(sum, max64Bit)
		if offset.IsInt64() {
			num = offset.Int64()
		} else {
			num = int64(inc) - (num - numBefore)
		}
	} else {
		num = sum.Int64()
	}

	hexVal = fmt.Sprintf("%x", num)
	new = prefix + hexVal

	return
}

// Binary cases:
//
// 1. Increment:
// 0b1 => 0b10
// 0x1111111111111111111111111111111111111111111111111111111111111111 =>
// 0x0000000000000000000000000000000000000000000000000000000000000000
// 0B0 => 0B1
//
// 2. Decrement:
// 0b1 => 0b0
// 0b100 => 0b011
// 0B010 => 0B001
// 0b0 =>
// 0x1111111111111111111111111111111111111111111111111111111111111111
func (rl *Instance) switchBinary(word string, inc int) (done bool, new string, bpos, epos int) {
	binary, _ := regexp.Compile(`[^0-9]?(0[bB][01]*)`)
	match := binary.FindString(word)
	if match == "" {
		return
	}

	done = true

	number := match
	prefix := match[:2]
	binVal := number[len(prefix):]
	indexes := binary.FindStringIndex(word)
	mbegin, mend := indexes[0], indexes[1]
	bpos, epos = mbegin, mend

	num, err := strconv.ParseInt(binVal, 2, 64)
	if err != nil {
		done = false
		return
	}

	max64Bit := big.NewInt(maxInt)
	bigNum := big.NewInt(num)
	bigInc := big.NewInt(int64(inc))
	sum := bigNum.Add(bigNum, bigInc)
	zero := big.NewInt(0)

	numBefore := num

	if sum.Cmp(zero) < 0 {
		offset := bigInc.Sub(max64Bit, sum.Abs(sum))
		if offset.IsInt64() {
			num = offset.Int64()
		} else {
			num = math.MaxInt64
		}
	} else if sum.CmpAbs(max64Bit) >= 0 {
		offset := bigInc.Sub(sum, max64Bit)
		if offset.IsInt64() {
			num = offset.Int64()
		} else {
			num = int64(inc) - (num - numBefore)
		}
	} else {
		num = sum.Int64()
	}

	binVal = fmt.Sprintf("%b", num)
	new = prefix + binVal
	return
}

// Decimal cases:
//
// 1. Increment:
// 0 => 1
// 99 => 100
//
// 2. Decrement:
// 0 => -1
// 10 => 9
// aa1230xa => aa1231xa // NOT WORKING => MATCHED BY HEXA
// aa1230bb => aa1231bb
// aa123a0bb => aa124a0bb
func (rl *Instance) switchDecimal(word string, inc int) (done bool, new string, bpos, epos int) {
	decimal, _ := regexp.Compile(`([-+]?[0-9]+)`)
	match := decimal.FindString(word)
	if match == "" {
		return
	}

	done = true

	indexes := decimal.FindStringIndex(word)
	mbegin, mend := indexes[0], indexes[1]
	bpos, epos = mbegin, mend

	num, _ := strconv.Atoi(match)
	// numBefore := num
	num += inc
	// if num < 0 {
	// 	num = math.MaxInt64
	// } else if num == math.MaxInt64 {
	// 	num = inc - (num - numBefore)
	// }

	new = strconv.Itoa(num)

	// Add prefix if needed
	if word[0] == '+' {
		new = "+" + new
	}

	// Don't consider anything done if result is empty.
	if new == "" {
		done = false
	}

	return
}

func (rl *Instance) switchBoolean(word string, increase bool) (done bool, new string, bpos, epos int) {
	epos = len(word)

	option, _ := regexp.Compile(`(^[+-]{0,2})`)
	if match := option.FindString(word); match != "" {
		indexes := option.FindStringIndex(word)
		bpos = indexes[1]
		word = word[bpos:]
	}

	booleans := map[string]string{
		"true":  "false",
		"false": "true",
		"t":     "f",
		"f":     "t",
		"yes":   "no",
		"no":    "yes",
		"y":     "n",
		"n":     "y",
		"on":    "off",
		"off":   "on",
	}

	new, done = booleans[strings.ToLower(word)]
	if !done {
		return
	}

	done = true

	// Transform case
	if match, _ := regexp.MatchString(`^[A-Z]+$`, word); match {
		new = strings.ToLower(new)
	} else if match, _ := regexp.MatchString(`^[A-Z]`, word); match {
		letter := new[0]
		upper := unicode.ToUpper(rune(letter))
		new = string(upper) + new[1:]
	}

	return
}

func (rl *Instance) switchWeekday(word string, increase bool) (done bool, new string, bpos, epos int) {
	return
}

func (rl *Instance) switchOperator(word string, increase bool) (done bool, new string, bpos, epos int) {
	epos = len(word)

	operators := map[string]string{
		"&&":  "||",
		"||":  "&&",
		"++":  "--",
		"--":  "++",
		"==":  "!=",
		"!=":  "==",
		"===": "!==",
		"!==": "===",
		"+":   "-",
		"-":   "*",
		"*":   "/",
		"/":   "+",
		"and": "or",
		"or":  "and",
	}

	new, done = operators[strings.ToLower(word)]
	if !done {
		return
	}

	done = true

	// Transform case
	if match, _ := regexp.MatchString(`^[A-Z]+$`, word); match {
		new = strings.ToLower(new)
	} else if match, _ := regexp.MatchString(`^[A-Z]`, word); match {
		letter := new[0]
		upper := unicode.ToUpper(rune(letter))
		new = string(upper) + new[1:]
	}

	return
}
