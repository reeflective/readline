package readline

import (
	"strings"
)

// vimDelete -
func (rl *Instance) viDelete(r rune) {

	// We are allowed to type iterations after a delete ('d') command.
	// in which case we don't exit the delete mode. The next thing typed
	// will thus be dispatched back here (like "2d4 then w).
	if !(r <= '9' && '0' <= r) {
		defer func() { rl.modeViMode = vimKeys }()
	}

	switch r {
	case 'b':
		vii := rl.getViIterations()
		rl.saveToRegister(rl.viJumpB(tokeniseLine), vii)
		for i := 1; i <= vii; i++ {
			rl.viDeleteByAdjust(rl.viJumpB(tokeniseLine))
		}

	case 'B':
		vii := rl.getViIterations()
		rl.saveToRegister(rl.viJumpB(tokeniseSplitSpaces), vii)
		for i := 1; i <= vii; i++ {
			rl.viDeleteByAdjust(rl.viJumpB(tokeniseSplitSpaces))
		}

	case 'd':
		rl.clearLine()
		rl.resetHelpers()
		rl.getHintText()

	case 'e':
		vii := rl.getViIterations()
		rl.saveToRegister(rl.viJumpE(tokeniseLine)+1, vii)
		for i := 1; i <= vii; i++ {
			rl.viDeleteByAdjust(rl.viJumpE(tokeniseLine) + 1)
		}

	case 'E':
		vii := rl.getViIterations()
		rl.saveToRegister(rl.viJumpE(tokeniseSplitSpaces)+1, vii)
		for i := 1; i <= vii; i++ {
			rl.viDeleteByAdjust(rl.viJumpE(tokeniseSplitSpaces) + 1)
		}

	case 'w':
		vii := rl.getViIterations()
		rl.saveToRegister(rl.viJumpW(tokeniseLine), vii)
		for i := 1; i <= vii; i++ {
			rl.viDeleteByAdjust(rl.viJumpW(tokeniseLine))
		}

	case 'W':
		vii := rl.getViIterations()
		rl.saveToRegister(rl.viJumpW(tokeniseSplitSpaces), vii)
		for i := 1; i <= vii; i++ {
			rl.viDeleteByAdjust(rl.viJumpW(tokeniseSplitSpaces))
		}

	case '%':
		rl.saveToRegister(rl.viJumpBracket(), 1)
		rl.viDeleteByAdjust(rl.viJumpBracket())

	case '$':
		rl.saveToRegister(len(rl.line)-rl.pos, 1)
		rl.viDeleteByAdjust(len(rl.line) - rl.pos)

	case '[':
		rl.saveToRegister(rl.viJumpPreviousBrace(), 1)
		rl.viDeleteByAdjust(rl.viJumpPreviousBrace())

	case ']':
		rl.saveToRegister(rl.viJumpNextBrace(), 1)
		rl.viDeleteByAdjust(rl.viJumpNextBrace())

	default:
		if r <= '9' && '0' <= r {
			rl.viIteration += string(r)
		}
		rl.viUndoSkipAppend = true
	}
}

// func (rl *Instance) vimDelete(r []rune) {
//
//         // We are allowed to type iterations after a delete ('d') command.
//         // in which case we don't exit the delete mode. The next thing typed
//         // will thus be dispatched back here (like "2d4 then w).
//         if r[0] != 27 {
//                 defer func() { rl.modeViMode = vimKeys }()
//         }
//
//         vii := rl.getViIterations()
//
//         switch r[0] {
//         case 'b':
//                 rl.saveToRegister(rl.viJumpB(tokeniseLine))
//                 for i := 1; i <= vii; i++ {
//                         rl.viDeleteByAdjust(rl.viJumpB(tokeniseLine))
//                 }
//
//         case 'B':
//                 rl.saveToRegister(rl.viJumpB(tokeniseSplitSpaces))
//                 for i := 1; i <= vii; i++ {
//                         rl.viDeleteByAdjust(rl.viJumpB(tokeniseSplitSpaces))
//                 }
//
//         case 'd':
//                 rl.clearLine()
//                 rl.resetHelpers()
//                 rl.getHintText()
//
//         case 'e':
//                 rl.saveToRegister(rl.viJumpE(tokeniseLine) + 1)
//                 for i := 1; i <= vii; i++ {
//                         rl.viDeleteByAdjust(rl.viJumpE(tokeniseLine) + 1)
//                 }
//
//         case 'E':
//                 rl.saveToRegister(rl.viJumpE(tokeniseSplitSpaces) + 1)
//                 for i := 1; i <= vii; i++ {
//                         rl.viDeleteByAdjust(rl.viJumpE(tokeniseSplitSpaces) + 1)
//                 }
//
//         case 'w':
//                 rl.saveToRegister(rl.viJumpW(tokeniseLine))
//                 for i := 1; i <= vii; i++ {
//                         rl.viDeleteByAdjust(rl.viJumpW(tokeniseLine))
//                 }
//
//         case 'W':
//                 rl.saveToRegister(rl.viJumpW(tokeniseSplitSpaces))
//                 for i := 1; i <= vii; i++ {
//                         rl.viDeleteByAdjust(rl.viJumpW(tokeniseSplitSpaces))
//                 }
//
//         case '%':
//                 rl.saveToRegister(rl.viJumpBracket())
//                 rl.viDeleteByAdjust(rl.viJumpBracket())
//
//         case '$':
//                 rl.saveToRegister(len(rl.line) - rl.pos)
//                 rl.viDeleteByAdjust(len(rl.line) - rl.pos)
//
//         case '[':
//                 rl.saveToRegister(rl.viJumpPreviousBrace())
//                 rl.viDeleteByAdjust(rl.viJumpPreviousBrace())
//
//         case ']':
//                 rl.saveToRegister(rl.viJumpNextBrace())
//                 rl.viDeleteByAdjust(rl.viJumpNextBrace())
//
//         case 27:
//                 if len(r) > 1 && '1' <= r[1] && r[1] <= '9' {
//                         rl.viIteration += string(r)
//                         //         if rl.vimDeleteToken(r[1]) {
//                         //                 return
//                         //         }
//                 }
//                 fallthrough
//
//         default:
//                 if len(r) > 1 && r[0] <= '9' && '0' <= r[0] {
//                         fmt.Printf("test")
//                         rl.viIteration += string(r)
//                 }
//                 rl.viUndoSkipAppend = true
//         }
// }

func (rl *Instance) viDeleteByAdjust(adjust int) {
	var (
		newLine []rune
		backOne bool
	)

	// Avoid doing anything if input line is empty.
	if len(rl.line) == 0 {
		return
	}

	switch {
	case adjust == 0:
		rl.viUndoSkipAppend = true
		return
	case rl.pos+adjust == len(rl.line)-1:
		newLine = rl.line[:rl.pos]
		backOne = true
	case rl.pos+adjust == 0:
		newLine = rl.line[rl.pos:]
	case adjust < 0:
		newLine = append(rl.line[:rl.pos+adjust], rl.line[rl.pos:]...)
	default:
		newLine = append(rl.line[:rl.pos], rl.line[rl.pos+adjust:]...)
	}

	rl.line = newLine

	rl.updateHelpers()

	if adjust < 0 {
		rl.moveCursorByAdjust(adjust)
	}

	if backOne {
		rl.pos--
	}
}

func (rl *Instance) vimDeleteToken(r rune) bool {
	tokens, _, _ := tokeniseSplitSpaces(rl.line, 0)
	pos := int(r) - 48 // convert ASCII to integer
	if pos > len(tokens) {
		return false
	}

	s := string(rl.line)
	newLine := strings.Replace(s, tokens[pos-1], "", -1)
	if newLine == s {
		return false
	}

	rl.line = []rune(newLine)

	rl.updateHelpers()

	if rl.pos > len(rl.line) {
		rl.pos = len(rl.line) - 1
	}

	return true
}
