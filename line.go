package readline

import (
	"strings"

	"github.com/acarl005/stripansi"
)

// initLine is ran once at the beginning of an instance start.
func (rl *Instance) initLine() {
	// Line
	rl.line = []rune{}
	rl.currentComp = []rune{} // No virtual completion yet
	rl.lineComp = []rune{}    // So no virtual line either
	rl.modeViMode = vimInsert
	rl.pos = 0
	rl.posY = 0

	// Selection
	rl.mark = -1
	rl.activeRegion = false
}

// When the DelayedSyntaxWorker gives us a new line, we need to check if there
// is any processing to be made, that all lines match in terms of content.
func (rl *Instance) updateLine(line []rune) {
	if len(rl.currentComp) > 0 {
	} else {
		rl.line = line
	}

	rl.renderHelpers()
}

// getLine - In many places we need the current line input. We either return the real line,
// or the one that includes the current completion candidate, if there is any.
func (rl *Instance) getLine() []rune {
	if len(rl.currentComp) > 0 {
		return rl.lineComp
	}
	return rl.line
}

// computeLine computes the number of lines that the input line spans.
func (rl *Instance) computeLine() {
	var usedLines, usedX int

	var line string
	if len(rl.currentComp) > 0 {
		line = string(rl.lineComp)
	} else {
		line = string(rl.line)
	}

	// We split the input line on every newline first
	// We determine if each line alone spans more than one line.
	for i, line := range strings.Split(line, "\n") {

		lineLen := len(line)
		usedX += lineLen

		// Adjust for the first line that is printed after the prompt.
		if i == 0 {
			lineLen += rl.inputAt
		}

		usedLines += lineLen / GetTermWidth()
		remain := lineLen % GetTermWidth()
		if remain != 0 {
			usedLines++
		}

		// The last line gives us the full rest
		if i == len(strings.Split(line, "\n"))-1 {
			rl.fullX = remain
		}
	}

	rl.fullY = usedLines
}

// computeCursorPos determines the X and Y coordinates of the cursor.
func (rl *Instance) computeCursorPos() {
}

// printLine - refresh the current input line, either virtually completed or not.
// also renders the current completions and hints. To be noted, the updateReferences()
// function is only ever called once, and after having moved back to prompt position
// and having printed the line: this is so that at any moment, everyone has the good
// values for moving around, synchronized with the update input line.
func (rl *Instance) printLine() {
	// Then we print the prompt, and the line,
	switch {
	case rl.PasswordMask != 0:
	case rl.PasswordMask > 0:
		print(strings.Repeat(string(rl.PasswordMask), len(rl.line)) + " ")

	default:
		// Go back to prompt position, and clear everything below
		moveCursorBackwards(GetTermWidth())
		moveCursorUp(rl.posY)
		print(seqClearScreenBelow)

		// We are the very beginning of the line ON WHICH we are
		// going to write the input line, not higher, even if the
		// entire primary+right prompt span several lines.
		rl.Prompt.printLast(rl)

		// Assemble the line, taking virtual completions into account
		var line []rune
		if len(rl.currentComp) > 0 {
			line = rl.lineComp
		} else {
			line = rl.line
		}

		highlighted := string(line) + " "

		// Print the input line with optional syntax highlighting
		if rl.SyntaxHighlighter != nil {
			highlighted = rl.SyntaxHighlighter(line) + " "
			// print(rl.SyntaxHighlighter(line) + " ")
			// } else {
			// 	print(string(line) + " ")
		}

		// Adapt if there is a visual selection active
		highlighted = string(rl.highlightVisualLine([]rune(highlighted)))

		// And print
		print(highlighted)
	}

	// Update references with new coordinates only now, because
	// the new line may be longer/shorter than the previous one.
	rl.updateReferences()

	// Go back to the current cursor position, with new coordinates
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.fullY)
	moveCursorDown(rl.posY)
	moveCursorForwards(rl.posX)
}

// highlightVisualLine adds highlighting of the region if we are in a visual mode.
func (rl *Instance) highlightVisualLine(line []rune) string {
	if rl.local != visual || rl.mark == -1 {
		return string(line)
	}

	// Compute begin and end of region
	var start, end int
	if rl.mark < rl.pos {
		start = rl.mark
		end = rl.pos
	} else {
		start = rl.pos
		end = rl.mark
	}

	// Adjust if we are in visual line mode
	if rl.local == visual && rl.visualLine {
		end = len(line) - 1
	}

	// First strip the current line (potentially highlighted by user)
	// from its colors sequences, to get the correct indexes.
	stripped := []rune(stripansi.Strip(string(line)))

	// Make the highlighted region
	highlightedRegion := []rune(seqBgRed)
	highlightedRegion = append(highlightedRegion, stripped[start:end+1]...)
	highlightedRegion = append(highlightedRegion, []rune(seqReset)...)

	// And assemble it into the entire line
	visualLine := string(stripped[:start])
	visualLine += string(highlightedRegion)
	visualLine += string(stripped[end+1:])

	return visualLine
}

func (rl *Instance) clearLine() {
	if len(rl.line) == 0 {
		return
	}

	// We need to go back to prompt
	moveCursorUp(rl.posY)
	moveCursorBackwards(GetTermWidth())
	moveCursorForwards(rl.inputAt)

	// Clear everything after & below the cursor
	print(seqClearScreenBelow)

	// Real input line
	rl.line = []rune{}
	rl.lineComp = []rune{}
	rl.pos = 0
	rl.posX = 0
	rl.fullX = 0
	rl.posY = 0
	rl.fullY = 0

	// Completions are also reset
	rl.clearVirtualComp()
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

	// We can ONLY have three fondamentally different cases:
	switch {
	// The line is empty
	case len(rl.line) == 0:
		rl.line = r

	// We are inserting somewhere in the middle
	case rl.pos < len(rl.line):
		forwardLine := append(r, rl.line[rl.pos:]...)
		rl.line = append(rl.line[:rl.pos], forwardLine...)

	// We are at the end of the input line
	case rl.pos == len(rl.line):
		rl.line = append(rl.line, r...)
	}

	rl.pos += len(r)

	// This should also update the rl.pos
	rl.updateHelpers()
}

func (rl *Instance) carriageReturn() {
	rl.clearHelpers()
	print("\r\n")
	if rl.HistoryAutoWrite {
		var err error

		// Main history
		if rl.mainHistory != nil {
			rl.histPos, err = rl.mainHistory.Write(string(rl.line))
			if err != nil {
				print(err.Error() + "\r\n")
			}
		}
		// Alternative history
		if rl.altHistory != nil {
			rl.histPos, err = rl.altHistory.Write(string(rl.line))
			if err != nil {
				print(err.Error() + "\r\n")
			}
		}
	}
}

func (rl *Instance) deletex() {
	switch {
	case len(rl.line) == 0:
		return
	case rl.pos == 0:
		rl.line = rl.line[1:]
	case rl.pos > len(rl.line):
		rl.pos = len(rl.line)
	case rl.pos == len(rl.line):
		rl.pos--
		rl.line = rl.line[:rl.pos]
	default:
		rl.line = append(rl.line[:rl.pos], rl.line[rl.pos+1:]...)
	}

	rl.updateHelpers()
}

// TODO: Identical to deleteBackspace/
func (rl *Instance) deleteX() {
	switch {
	case len(rl.line) == 0:
		return
	case rl.pos == 0:
		return
	case rl.pos > len(rl.line):
		rl.pos = len(rl.line)
	case rl.pos == len(rl.line):
		rl.pos--
		rl.line = rl.line[:rl.pos]
	default:
		rl.pos--
		rl.line = append(rl.line[:rl.pos], rl.line[rl.pos+1:]...)
	}

	rl.updateHelpers()
}

func (rl *Instance) deleteBackspace() {
	switch {
	case len(rl.line) == 0:
		return
	case rl.pos == 0:
		return
		// if len(rl.line) > 0 {
		// 	rl.line = rl.line[1:]
		// }
	case rl.pos > len(rl.line):
		rl.backspace() // There is an infite loop going on here...
	case rl.pos == len(rl.line):
		rl.pos--
		rl.line = rl.line[:rl.pos]
	default:
		rl.pos--
		rl.line = append(rl.line[:rl.pos], rl.line[rl.pos+1:]...)
	}

	rl.updateHelpers()
}

func (rl *Instance) deleteToBeginning() {
	rl.resetVirtualComp(false)
	// Keep the line length up until the cursor
	rl.line = rl.line[rl.pos:]
	rl.pos = 0
}
