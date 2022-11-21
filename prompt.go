package readline

import (
	"fmt"
	"strings"

	ansi "github.com/acarl005/stripansi"
)

// Prompt uses a function returning the string to use as the primary prompt
func (rl *Instance) Prompt(prompt func() string) {
	rl.promptFunc = prompt
}

// PromptRight uses a function returning the string to use as the right prompt
func (rl *Instance) PromptRight(prompt func() string) {
	rl.promptRightFunc = prompt
}

// PromptTransient uses a function returning the prompt to use as a transcient prompt.
func (rl *Instance) PromptTransient(prompt func() string) {
	rl.promptTransientFunc = prompt
}

// PromptSecondary uses a function returning the prompt to use as the secondary prompt.
func (rl *Instance) PromptSecondary(prompt func() string) {
	rl.promptSecondaryFunc = prompt
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
	// Generate the prompt strings for this run
	rl.prompt = rl.promptFunc()
	rl.promptRight = rl.promptRightFunc()

	// Print the primary prompt, potentially excluding the last line.
	print(rl.getPromptPrimary())
	rl.stillOnRefresh = false

	// Compute some offsets needed by the last line.
	rl.computePrompt()
}

// computePromptAlt computes the correct lengths and offsets
// for all prompt components, but does not print any of them.
func (rl *Instance) computePrompt() {
	lastLineIndex := strings.LastIndex(rl.prompt, "\n")
	if lastLineIndex != -1 {
		rl.inputAt = len([]rune(ansi.Strip(rl.prompt[lastLineIndex:])))
	} else {
		rl.inputAt = len([]rune(ansi.Strip(rl.prompt)))
	}
}

// Prompt refresh offsets
// // 1 - Get the number of lines which we must go up before printing
// var multilineOffset int
// multilineOffset += strings.Count(prompt, "\n")
// multilineOffset += strings.Count(promptRight, "\n")
//
// moveCursorUp(multilineOffset)
// print(seqClearScreenBelow)

// printPrompt assumes we are at the very beginning of the line
// in which we will start printing the input buffer, and redraws
// the various prompts, taking care of offsetting its initial position.
func (rl *Instance) printPrompt() {
	// TODO: Here don't perform this if the line is longer than terminal
	if rl.promptRight != "" {
		// First go back to beginning of line, and clear everything
		moveCursorBackwards(GetTermWidth())
		print(seqClearLine)
		print(seqClearScreenBelow)

		forwardOffset := GetTermWidth() - getRealLength(rl.promptRight) - 1
		moveCursorForwards(forwardOffset)
		print(rl.promptRight)

		// Normally we are on a newline, since we just completed the previous.
		moveCursorBackwards(GetTermWidth())
	}

	print(rl.getPromptLastLine())
}

// getPromptPrimary returns either the entire prompt if
// it's a single-line, or everything except the last line.
func (rl *Instance) getPromptPrimary() string {
	var primary string

	// Get the last line of the prompt to be printed.
	lastLineIndex := strings.LastIndex(rl.prompt, "\n")
	if lastLineIndex != -1 {
		primary = rl.prompt[:lastLineIndex+1]
	} else {
		primary = rl.prompt
	}

	return primary
}

// Get the last line of the prompt to be printed.
func (rl *Instance) getPromptLastLine() string {
	var lastLine string
	lastLineIndex := strings.LastIndex(rl.prompt, "\n")
	if lastLineIndex != -1 {
		lastLine = rl.prompt[lastLineIndex+1:]
	} else {
		lastLine = rl.prompt
	}

	return lastLine
}

// computePrompt - At any moment, returns an (1st or 2nd line) actualized prompt,
// considering all input mode parameters and prompt string values.
func (rl *Instance) computePromptAlt() (prompt []rune) {
	switch rl.InputMode {
	case Vim:
		rl.computePromptVim()
	case Emacs:
		rl.computePromptEmacs()
	}
	return
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
