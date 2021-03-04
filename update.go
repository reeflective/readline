package readline

<<<<<<< HEAD
// updateHelpers is a key part of the whole refresh process:
// it should coordinate reprinting the input line, any hints and completions
// and manage to get back to the current (computed) cursor coordinates
func (rl *Instance) updateHelpers() {

	// Load all hints & completions before anything.
	rl.tcOffset = 0
	rl.getHintText()
	if rl.modeTabCompletion {
		rl.getTabCompletion()
	}
	// We clear everything
	rl.clearHelpers()

	// We are at the prompt line (with the latter
	// not printed yet), then reprint everything
	rl.renderHelpers()
}

// Update reference should be called only once in a "loop" (not Readline(), but key control loop)
func (rl *Instance) updateReferences() {

	// We always need to work with clean data,
	// since we will have incrementers all around
	rl.posX = 0
	rl.fullX = 0
	rl.posY = 0
	rl.fullY = 0

	var fullLine, cPosLine int
	if len(rl.currentComp) > 0 {
		fullLine = len(rl.lineComp)
		cPosLine = len(rl.lineComp[:rl.pos])
	} else {
		fullLine = len(rl.line)
		cPosLine = len(rl.line[:rl.pos])
	}

	// We need the X offset of the whole line
	toEndLine := rl.promptLen + fullLine
	fullOffset := toEndLine / GetTermWidth()
	rl.fullY = fullOffset
	fullRest := toEndLine % GetTermWidth()
	rl.fullX = fullRest

	// Use rl.pos value to get the offset to go TO/FROM the CURRENT POSITION
	lineToCursorPos := rl.promptLen + cPosLine
	offsetToCursor := lineToCursorPos / GetTermWidth()
	cPosRest := lineToCursorPos % GetTermWidth()

	// If we are at the end of line
	if fullLine == rl.pos {
		rl.posY = fullOffset

		if fullRest == 0 {
			rl.posX = 0
		} else if fullRest > 0 {
			rl.posX = fullRest
		}
	} else if rl.pos < fullLine {
		// If we are somewhere in the middle of the line
		rl.posY = offsetToCursor

		if cPosRest == 0 {
		} else if cPosRest > 0 {
			rl.posX = cPosRest
		}
	}
=======
import (
	"fmt"
	"strings"

	"github.com/lunixbochs/vtclean"
)

func moveCursorUp(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dA", i)
}

func moveCursorDown(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dB", i)
}

func moveCursorForwards(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dC", i)
}

func moveCursorBackwards(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dD", i)
}

// moveCursorToLinePos - Must calculate the length of the prompt, realtime
// and for all contexts/needs, and move the cursor appropriately
func moveCursorToLinePos(rl *Instance) {
	moveCursorForwards(rl.promptLen + rl.pos)
	return
}

func (rl *Instance) moveCursorByAdjust(adjust int) {
	switch {
	case adjust > 0:
		moveCursorForwards(adjust)
		rl.pos += adjust
	case adjust < 0:
		moveCursorBackwards(adjust * -1)
		rl.pos += adjust
	}

	if rl.modeViMode != vimInsert && rl.pos == len(rl.line) && len(rl.line) > 0 {
		moveCursorBackwards(1)
		rl.pos--
	}
}

func (rl *Instance) insert(r []rune) {
	for {
		// I don't really understand why `0` is creaping in at the end of the
		// array but it only happens with unicode characters.
		if len(r) > 1 && r[len(r)-1] == 0 {
			r = r[:len(r)-1]
			continue
		}
		break
	}

	switch {
	case len(rl.line) == 0:
		rl.line = r
	case rl.pos == 0:
		rl.line = append(r, rl.line...)
	case rl.pos < len(rl.line):
		r := append(r, rl.line[rl.pos:]...)
		rl.line = append(rl.line[:rl.pos], r...)
	default:
		rl.line = append(rl.line, r...)
	}

	rl.echo()

	rl.pos += len(r)
	moveCursorForwards(len(r) - 1)

	if rl.modeViMode == vimInsert {
		rl.updateHelpers()
	}
}

func (rl *Instance) backspace() {
	if len(rl.line) == 0 || rl.pos == 0 {
		return
	}

	moveCursorBackwards(1)
	rl.pos--
	rl.delete()
}

func (rl *Instance) delete() {
	switch {
	case len(rl.line) == 0:
		return
	case rl.pos == 0:
		rl.line = rl.line[1:]
		rl.echo()
		moveCursorBackwards(1)
	case rl.pos > len(rl.line):
		rl.backspace()
	case rl.pos == len(rl.line):
		rl.line = rl.line[:rl.pos]
		rl.echo()
		moveCursorBackwards(1)
	default:
		rl.line = append(rl.line[:rl.pos], rl.line[rl.pos+1:]...)
		rl.echo()
		moveCursorBackwards(1)
	}

	rl.updateHelpers()
}

func (rl *Instance) echo() {

	// We move the cursor back to the very beginning of the line:
	// prompt + cursor position
	moveCursorBackwards(rl.promptLen + rl.pos)

	switch {
	case rl.PasswordMask > 0:
		print(strings.Repeat(string(rl.PasswordMask), len(rl.line)) + " ")

	case rl.SyntaxHighlighter == nil:
		print(string(rl.mlnPrompt))

		// Depending on the presence of a virtually completed item,
		// print either the virtual line or the real one.
		if len(rl.currentComp) > 0 {
			line := rl.lineComp[:rl.pos]
			line = append(line, rl.lineRemain...)
			// print(string(line) + " ")
			rl.echoInputLine(line)
		} else {
			// print(string(rl.line) + " ")
			line := append(rl.line, []rune(" ")...)
			rl.echoInputLine(line)
			// moveCursorBackwards(len(rl.line) - rl.pos)
		}

	default:
		print(string(rl.mlnPrompt))

		// Depending on the presence of a virtually completed item,
		// print either the virtual line or the real one.
		if len(rl.currentComp) > 0 {
			line := rl.lineComp[:rl.pos]
			line = append(line, rl.lineRemain...)
			// print(rl.SyntaxHighlighter(line) + " ")
			line = []rune(rl.SyntaxHighlighter(line) + " ")
			rl.echoInputLine(line)
		} else {
			// print(rl.SyntaxHighlighter(rl.line) + " ")
			line := []rune(rl.SyntaxHighlighter(rl.line) + " ")
			rl.echoInputLine(line)
			// moveCursorBackwards(len(rl.line) - rl.pos)
		}
	}
}

// If the input line spans multiple lines, before we print hints, comps, etc.
func (rl *Instance) moveInputEnd() {

	// First go back to prompt
	moveCursorBackwards(rl.pos)
	moveCursorUp(rl.posY)

	// Then go back to end
	numlines := len(rl.line)/GetTermWidth() - 2
	rest := len(rl.line)%GetTermWidth() - 2

	if numlines > 0 {
		moveCursorDown(numlines)
	}
	if rest > 0 {
		moveCursorDown(1)
	}
}

// echoInputLine - The console considers various things at once (current term width,
// input line width, potentially any newline/tab tokens, etc), and renders the current
// input line appropriately, ensuring the cursor is always where it should be.
func (rl *Instance) echoInputLine(line []rune) {

	// First, clean the line from any color/terminal escape sequence it contains
	netLine := vtclean.Clean(string(line), false)

	// If there are any token to process it should be done here, because
	// we are going to need exact length/width values after that.

	// If, counting the prompt, our current input will not span multiple lines,
	// do not complicate: print and return
	firstLen := GetTermWidth() - len(rl.prompt)
	if len(netLine) <= firstLen {
		print(string(line))
		if len(rl.currentComp) <= 0 {
			moveCursorBackwards(len(rl.line) - rl.pos)
		}
		return
	}

	// We first go back to where the prompt is (either Vim status or "$"), and clear
	// everything below it. Completions, hinsts, etc will be reprinted anyway.
	numlines := len(line)/GetTermWidth() - 2
	rest := len(line)%GetTermWidth() - 2

	moveCursorBackwards(rest)
	moveCursorUp(numlines)
	print(seqClearScreenBelow)

	var lines [][]rune
	if numlines > 1 {
		for i := 0; i < numlines; i++ {
			ln := line[(i * GetTermWidth()):GetTermWidth()]
			lines = append(lines, ln)
		}

		if rest > 0 {
			ln := line[(numlines * GetTermWidth()) : GetTermWidth()-2]
			lines = append(lines, ln)
		}
	}

	if rest > 0 && numlines == 0 {
		lines = append(lines, line)
	}

	// For each line we may potentially adjust for any token we want
	// to interpret as newlines (or anything) that we might do here.
	// Only adjust numlines in the process, and overwrite carefully
	// the lines [][]rune variable.

	// Then we print each line
	for _, line := range lines {
		print(string(line))
	}

	// Move the cursor back to its current position
	moveCursorUp(numlines - rl.tcPosY)
	moveCursorToLinePos(rl)

	// Go back if we are not on the first line (no need to offsert prompt)
	if numlines > 0 {
		moveCursorBackwards(len(rl.prompt))
	}
}

func (rl *Instance) clearLine() {
	if len(rl.line) == 0 {
		return
	}

	var lineLen int
	if len(rl.lineComp) > len(rl.line) {
		lineLen = len(rl.lineComp)
	} else {
		lineLen = len(rl.line)
	}

	moveCursorBackwards(rl.pos)
	print(strings.Repeat(" ", lineLen))
	moveCursorBackwards(lineLen)

	// Real input line
	rl.line = []rune{}
	rl.pos = 0

	// Completions are also reset
	rl.clearVirtualComp()
>>>>>>> 611c6fb333d138b32958059c075a2d21c7ca09ae
}

func (rl *Instance) resetHelpers() {
	rl.modeAutoFind = false
<<<<<<< HEAD

	// Now reset all below-input helpers
=======
	rl.clearHelpers()
>>>>>>> 611c6fb333d138b32958059c075a2d21c7ca09ae
	rl.resetHintText()
	rl.resetTabCompletion()
}

<<<<<<< HEAD
// clearHelpers - Clears everything: prompt, input, hints & comps,
// and comes back at the prompt.
func (rl *Instance) clearHelpers() {

	// Now go down to the last line of input
	moveCursorDown(rl.fullY - rl.posY)
	moveCursorBackwards(rl.posX)
	moveCursorForwards(rl.fullX)

	// Clear everything below
	print(seqClearScreenBelow)

	// Go back to current cursor position
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.fullY - rl.posY)
	moveCursorForwards(rl.posX)
=======
func (rl *Instance) clearHelpers() {
	print("\r\n" + seqClearScreenBelow)
	moveCursorUp(1)
	moveCursorToLinePos(rl)

	// Reset some values
	rl.lineComp = []rune{}
	rl.currentComp = []rune{}
>>>>>>> 611c6fb333d138b32958059c075a2d21c7ca09ae
}

func (rl *Instance) renderHelpers() {

<<<<<<< HEAD
	// Optional, because neutral on placement
	rl.echo()

	// Go at beginning of first line after input remainder
	moveCursorDown(rl.fullY - rl.posY)
	moveCursorBackwards(GetTermWidth())

	// Print hints, check for any confirmation hint current.
	// (do not overwrite the confirmation question hint)
	if !rl.compConfirmWait {
		rl.getHintText()
		if len(rl.hintText) > 0 {
			print("\n")
			// moveCursorDown(1)
		}
		rl.writeHintText()
		moveCursorBackwards(GetTermWidth())

		// Print completions and go back to beginning of this line
		print("\n")
		// moveCursorDown(1)
		rl.writeTabCompletion()
		moveCursorBackwards(GetTermWidth())
		moveCursorUp(rl.tcUsedY - 1)
	}

	// If we are still waiting for the user to confirm too long completions
	if rl.compConfirmWait {
		print("\n")
		// moveCursorDown(1)
		rl.writeHintText()
		moveCursorBackwards(GetTermWidth())
		print("\n")
		// moveCursorDown(1)
	}

	// Anyway, compensate for hint printout
	if len(rl.hintText) > 0 {
		moveCursorUp(rl.hintY)
	} else {
		moveCursorUp(1)
	}

	// Go back to current cursor position
	moveCursorUp(rl.fullY - rl.posY)
	moveCursorForwards(rl.posX)
=======
	rl.echo()

	// If we are waiting for confirmation (too many comps),
	// do not overwrite the confirmation question hint.
	if !rl.compConfirmWait {
		// We also don't overwrite if in tab find mode, which has a special hint.
		if !rl.modeAutoFind {
			rl.getHintText()
		}
		// We write the hint anyway
		rl.writeHintText()
	}

	rl.writeTabCompletion()

	// If the length of completion is wider than the terminal length,
	// we refresh the prompt and the hints once again
	if rl.tcUsedY > GetTermLength() {
		if rl.Multiline {
			fmt.Println() // Completions don't add the last newline
			fmt.Println(rl.prompt)
		}
		rl.echo()
		if !rl.modeAutoFind {
			rl.getHintText()
		}
		// We write the hint again
		rl.writeHintText()

		// Very important, otherwise will reprint the loop
		rl.resetTabCompletion()
	} else {
		moveCursorUp(rl.tcUsedY)
	}

	if !rl.compConfirmWait {
		moveCursorUp(rl.hintY)
	}

	moveCursorBackwards(GetTermWidth())
	moveCursorToLinePos(rl)
}

func (rl *Instance) updateHelpers() {
	rl.tcOffset = 0
	rl.getHintText()
	if rl.modeTabCompletion {
		rl.getTabCompletion()
	}
	rl.clearHelpers()
	rl.renderHelpers()
>>>>>>> 611c6fb333d138b32958059c075a2d21c7ca09ae
}
