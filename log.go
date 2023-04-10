package readline

import (
	"fmt"

	"github.com/reeflective/readline/internal/term"
	"github.com/reeflective/readline/internal/ui"
)

// Prompt returns the Prompt type used by the readline shell.
// You can use this prompt to bind various prompt string handlers:
// primary, right, secondary, tooltip prompts, and a transient one.
func (rl *Shell) Prompt() *ui.Prompt {
	return rl.prompt
}

// Log prints a formatted string below the current line and redisplays the prompt
// and input line (and possibly completions/hints if active) below the logged string.
// A newline is added to the message so that the prompt is correctly refreshed below.
func (rl *Shell) Log(msg string, args ...interface{}) {
	// First go back to the last line of the input line,
	// and clear everything below (hints and completions).
	rl.display.CursorBelowLine()
	term.MoveCursorBackwards(term.GetWidth())
	print(term.ClearScreenBelow)

	// Skip a line, and print the formatted message.
	fmt.Printf(msg+"\n", args...)

	// Redisplay the prompt, input line and active helpers.
	rl.prompt.PrimaryPrint()
	rl.display.Refresh()
}

// LogTransient prints a formatted string in place of the current prompt and input
// line, and then refreshes, or "pushes" the prompt/line below this printed message.
func (rl *Shell) LogTransient(msg string, args ...interface{}) {
	// First go back to the beginning of the line/prompt, and
	// clear everything below (prompt/line/hints/completions).
	if rl.Prompt().Refreshing() {
		term.MoveCursorUp(1)
	}
	rl.display.CursorToLineStart()
	term.MoveCursorBackwards(term.GetWidth())

	term.MoveCursorUp(rl.Prompt().PrimaryUsed())
	print(term.ClearScreenBelow)

	// Print the logged message.
	fmt.Printf(msg+"\n", args...)

	// Redisplay the prompt, input line and active helpers.
	rl.prompt.PrimaryPrint()
	rl.display.Refresh()
}
