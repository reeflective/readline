package readline

import (
	"fmt"
	"strings"
)

// Log prints a formatted string below the current line and redisplays the prompt
// and input line (and possibly completions/hints if active) below the logged string.
// A newline is added to the message so that the prompt is correctly refreshed below.
func (rl *Instance) Log(msg string, args ...interface{}) {
	// First go back to the last line of the input line,
	// and clear everything below (hints and completions).
	rl.clearHelpers()
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.posY)
	moveCursorDown(rl.fullY + 1)
	print(seqClearLine)
	print(seqClearScreenBelow)
	rl.posY = 0

	// Skip a line, and print the formatted message.
	fmt.Printf(msg+"\n", args...)

	// Reprints the prompt, input line, and any active helpers.
	rl.Prompt.init(rl)
	enablePos := rl.EnableGetCursorPos
	rl.EnableGetCursorPos = false
	rl.renderHelpers()
	rl.EnableGetCursorPos = enablePos
}

// LogTransient prints a formatted string in place of the current prompt and input
// line, and then refreshes, or "pushes" the prompt/line below this printed message.
func (rl *Instance) LogTransient(msg string, args ...interface{}) {
	// First go back to the beginning of the line/prompt, and
	// clear everything below (prompt/line/hints/completions).
	if rl.Prompt.stillOnRefresh {
		moveCursorUp(1)
	}
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.posY)
	promptLines := strings.Count(rl.Prompt.primary, "\n")
	moveCursorUp(promptLines)
	print(seqClearLine)
	print(seqClearScreenBelow)
	rl.posY = 0

	// Print the logged message.
	fmt.Printf(msg+"\n", args...)

	// And redisplay the prompt, input line and any active helpers.
	rl.Prompt.init(rl)
	enablePos := rl.EnableGetCursorPos
	rl.EnableGetCursorPos = false
	rl.renderHelpers()
	rl.EnableGetCursorPos = enablePos
}
