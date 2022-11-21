package readline

import (
	"fmt"
	"strings"

	ansi "github.com/acarl005/stripansi"
)

// SetPrompt will define the readline prompt string.
// It also calculates the runes in the string as well as any non-printable escape codes.
func (rl *Instance) SetPrompt(s string) {
	rl.prompt = s
}

// SetPromptRight sets the right-most prompt for the shell
func (rl *Instance) SetPromptRight(s string) {
	rl.promptRight = s
}

// SetPromptTransient sets a transient prompt for the shell
func (rl *Instance) SetPromptTransient(s string) {
	rl.promptTransient = s
}

// SetPromptSecondary sets the secondary prompt for the shell.
func (rl *Instance) SetPromptSecondary(s string) {
	rl.promptSecondary = s
}

// RefreshPromptLog - A simple function to print a string message (a log, or more broadly,
// an asynchronous event) without bothering the user, and by "pushing" the prompt below the message.
func (rl *Instance) RefreshPromptLog(log string) (err error) {
	// We adjust cursor movement, depending on which mode we're currently in.
	if !rl.modeTabCompletion {
		rl.tcUsedY = 1
		// Account for the hint line
	} else if rl.modeTabCompletion && rl.modeAutoFind {
		rl.tcUsedY = 0
	} else {
		rl.tcUsedY = 1
	}

	// Prompt offset
	if rl.isMultiline {
		rl.tcUsedY += 1
	} else {
		rl.tcUsedY += 0
	}

	// Clear the current prompt and everything below
	print(seqClearLine)
	if rl.stillOnRefresh {
		moveCursorUp(1)
	}
	rl.stillOnRefresh = true
	moveCursorUp(rl.hintY + rl.tcUsedY)
	moveCursorBackwards(GetTermWidth())
	print("\r\n" + seqClearScreenBelow)

	// Print the log
	fmt.Printf(log)

	// Add a new line between the message and the prompt, so not overloading the UI
	print("\n")

	// Print the prompt
	if rl.isMultiline {
		rl.tcUsedY += 3
		fmt.Println(rl.prompt)

	} else {
		rl.tcUsedY += 2
		fmt.Print(rl.prompt)
	}

	// Refresh the line
	rl.updateHelpers()

	return
}

// RefreshPromptInPlace - Refreshes the prompt in the very same place he is.
func (rl *Instance) RefreshPromptInPlace(prompt string) (err error) {
	// We adjust cursor movement, depending on which mode we're currently in.
	// Prompt data intependent
	if !rl.modeTabCompletion {
		rl.tcUsedY = 1
		// Account for the hint line
	} else if rl.modeTabCompletion && rl.modeAutoFind {
		rl.tcUsedY = 0
	} else {
		rl.tcUsedY = 1
	}

	// Update the prompt if a special has been passed.
	if prompt != "" {
		rl.prompt = prompt
	}

	if rl.isMultiline {
		rl.tcUsedY += 1
	}

	// Clear the input line and everything below
	print(seqClearLine)
	moveCursorUp(rl.hintY + rl.tcUsedY)
	moveCursorBackwards(GetTermWidth())
	print("\r\n" + seqClearScreenBelow)

	// Add a new line if needed
	if rl.isMultiline {
		fmt.Println(rl.prompt)
	} else {
		fmt.Print(rl.prompt)
	}

	// Refresh the line
	rl.updateHelpers()

	return
}

// printPrompt assumes we are at the very beginning of the line
// in which we will start printing the input buffer, and redraws
// the various prompts, taking care of offsetting its initial position.
func (rl *Instance) printPrompt() {
	// 1 - Get the number of lines which we must go up before printing
	var multilineOffset int
	multilineOffset += strings.Count(rl.prompt, "\n")
	multilineOffset += strings.Count(rl.promptRight, "\n")

	moveCursorUp(multilineOffset)

	// First draw the primary prompt. We are now where the input
	// zone starts, but we still might have to print the right prompt.
	print(rl.prompt)

	if rl.promptRight != "" {
		forwardOffset := GetTermWidth() - rl.inputAt - getRealLength(rl.promptRight)
		moveCursorForwards(forwardOffset)
		print(rl.promptRight)

		// Normally we are on a newline, since we just completed the previous.
		moveCursorUp(1)
		moveCursorForwards(rl.inputAt)
	}
}

// RefreshPromptCustom - Refresh the console prompt with custom values.
// @prompt      => If not nil (""), will use this prompt instead of the currently set prompt.
// @offset      => Used to set the number of lines to go upward, before reprinting. Set to 0 if not used.
// @clearLine   => If true, will clean the current input line on the next refresh.
func (rl *Instance) RefreshPromptCustom(prompt string, offset int, clearLine bool) (err error) {
	// We adjust cursor movement, depending on which mode we're currently in.
	if !rl.modeTabCompletion {
		rl.tcUsedY = 1
	} else if rl.modeTabCompletion && rl.modeAutoFind { // Account for the hint line
		rl.tcUsedY = 0
	} else {
		rl.tcUsedY = 1
	}

	// Add user-provided offset
	rl.tcUsedY += offset

	// Go back to prompt position, then up to the user provided offset.
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.posY)
	moveCursorUp(offset)

	// Then clear everything below our new position
	print(seqClearScreenBelow)

	// Update the prompt if a special has been passed.
	if prompt != "" {
		rl.prompt = prompt
	}

	// Add a new line if needed
	if rl.isMultiline && prompt == "" {
	} else if rl.isMultiline {
		fmt.Println(rl.prompt)
	} else {
		fmt.Print(rl.prompt)
	}

	// Refresh the line
	rl.updateHelpers()

	// If input line was empty, check that we clear it from detritus
	// The three lines are borrowed from clearLine(), we don't need more.
	if clearLine {
		rl.clearLine()
	}

	return
}

// initPrompt is ran once at the beginning of an instance start.
func (rl *Instance) initPrompt() {
	// Here we have to either print prompt
	// and return new line (multiline)
	if rl.isMultiline {
		fmt.Println(rl.prompt)
	}
	rl.stillOnRefresh = false
	rl.computePrompt() // initialise the prompt for first print
}

// computePrompt - At any moment, returns an (1st or 2nd line) actualized prompt,
// considering all input mode parameters and prompt string values.
func (rl *Instance) computePrompt() (prompt []rune) {
	switch rl.InputMode {
	case Vim:
		rl.computePromptVim()
	case Emacs:
		rl.computePromptEmacs()
	}
	return
}

// computePromptAlt computes the correct lengths and offsets
// for all prompt components, but does not print any of them.
func (rl *Instance) computePromptAlt() {
	// The length of the prompt on the last line is where
	// the input line starts. Get this line and compute.
	lastLineIndex := strings.LastIndex(rl.prompt, "\n")
	if lastLineIndex != -1 {
		rl.inputAt = getRealLength(rl.prompt[lastLineIndex:])
	} else {
		rl.inputAt = getRealLength(rl.prompt[lastLineIndex:])
	}
}

func (rl *Instance) computePromptVim() {
	var vimStatus []rune // Here we use this as a temporary prompt string

	// Compute Vim status string first
	if rl.ShowVimMode {
		switch rl.modeViMode {
		case vimKeys:
			vimStatus = []rune(vimKeysStr)
		case vimInsert:
			vimStatus = []rune(vimInsertStr)
		case vimReplaceOnce:
			vimStatus = []rune(vimReplaceOnceStr)
		case vimReplaceMany:
			vimStatus = []rune(vimReplaceManyStr)
		case vimDelete:
			vimStatus = []rune(vimDeleteStr)
		}

		vimStatus = rl.colorizeVimPrompt(vimStatus)
	}

	// Append any optional prompts for multiline mode
	if rl.isMultiline {
		if rl.promptMultiline != "" {
			rl.realPrompt = append(vimStatus, []rune(rl.promptMultiline)...)
		} else {
			rl.realPrompt = vimStatus
			rl.realPrompt = append(rl.realPrompt, rl.defaultPrompt...)
		}
	}
	// Equivalent for non-multiline
	if !rl.isMultiline {
		if rl.prompt != "" {
			rl.realPrompt = append(vimStatus, []rune(" "+rl.prompt)...)
		} else {
			// Vim status might be empty, but we don't care
			rl.realPrompt = append(rl.realPrompt, vimStatus...)
		}
		// We add the multiline prompt anyway, because it might be empty and thus have
		// no effect on our user interface, or be specified and thus needed.
		// if rl.MultilinePrompt != "" {
		rl.realPrompt = append(rl.realPrompt, []rune(rl.promptMultiline)...)
		// } else {
		//         rl.realPrompt = append(rl.realPrompt, rl.defaultPrompt...)
		// }
	}

	// Strip color escapes
	rl.inputAt = getRealLength(string(rl.realPrompt))
}

func (rl *Instance) computePromptEmacs() {
	if rl.isMultiline {
		if rl.promptMultiline != "" {
			rl.realPrompt = []rune(rl.promptMultiline)
		} else {
			rl.realPrompt = rl.defaultPrompt
		}
	}
	if !rl.isMultiline {
		if rl.prompt != "" {
			rl.realPrompt = []rune(rl.prompt)
		}
		// We add the multiline prompt anyway, because it might be empty and thus have
		// no effect on our user interface, or be specified and thus needed.
		// if rl.MultilinePrompt != "" {
		rl.realPrompt = append(rl.realPrompt, []rune(rl.promptMultiline)...)
		// } else {
		//         rl.realPrompt = append(rl.realPrompt, rl.defaultPrompt...)
		// }
	}

	// Strip color escapes
	rl.inputAt = getRealLength(string(rl.realPrompt))
}

func (rl *Instance) colorizeVimPrompt(p []rune) (cp []rune) {
	if rl.VimModeColorize {
		return []rune(fmt.Sprintf("%s%s%s", BOLD, string(p), RESET))
	}

	return p
}

// getRealLength - Some strings will have ANSI escape codes, which might be wrongly
// interpreted as legitimate parts of the strings. This will bother if some prompt
// components depend on other's length, so we always pass the string in this for
// getting its real-printed length.
func getRealLength(s string) (l int) {
	return len(ansi.Strip(s))
}
