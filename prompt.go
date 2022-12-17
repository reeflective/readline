package readline

import (
	"strings"
	"unicode/utf8"

	ansi "github.com/acarl005/stripansi"
)

// prompt stores all prompt functions and strings,
// and is in charge of printing them as well as
// computing any resulting offsets.
type prompt struct {
	primary  string
	primaryF func() string

	right  string
	rightF func() string

	secondary  string
	secondaryF func() string

	transient  string
	transientF func() string

	tooltip  string
	tooltipF func(tip string) string

	// True if some logs have printed asynchronously
	// since last loop. Check refresh prompt funcs.
	stillOnRefresh bool

	// The offset used on the first line, where either
	// the full prompt (or the last line) is. Used for
	// correctly replacing the cursor.
	inputAt int
}

// Primary uses a function returning the string to use as the primary prompt
func (p *prompt) Primary(prompt func() string) {
	p.primaryF = prompt
}

// Right uses a function returning the string to use as the right prompt
func (p *prompt) Right(prompt func() string) {
	p.rightF = prompt
}

// Secondary uses a function returning the prompt to use as the secondary prompt.
func (p *prompt) Secondary(prompt func() string) {
	p.secondaryF = prompt
}

// Transient uses a function returning the prompt to use as a transcient prompt.
func (p *prompt) Transient(prompt func() string) {
	p.transientF = prompt
}

// Tooltip uses a function returning the prompt to use as a tooltip prompt.
func (p *prompt) Tooltip(prompt func(tip string) string) {
	p.tooltipF = prompt
}

// initPrompt is ran once at the beginning of an instance start.
func (p *prompt) init(rl *Instance) {
	// Generate the prompt strings for this run
	if p.primaryF != nil {
		p.primary = p.primaryF()
	}
	if p.rightF != nil {
		p.right = p.rightF()
	}

	// Print the primary prompt, potentially excluding the last line.
	print(p.getPrimary())
	p.stillOnRefresh = false

	// Compute some offsets needed by the last line.
	rl.Prompt.compute(rl)
}

func (p *prompt) update(rl *Instance) {
	var tooltipWord string

	shellWords := strings.Split(string(rl.line), " ")
	if len(shellWords) > 0 {
		tooltipWord = shellWords[0]
	}

	rl.Prompt.tooltip = rl.Prompt.tooltipF(tooltipWord)
}

func (p *prompt) printLast(rl *Instance) {
	// Either use RPROMPT or tooltip.
	var rprompt string
	if p.tooltip != "" {
		rprompt = p.tooltip
	} else {
		rprompt = p.right
	}

	// Print the primary prompt in any case.
	defer func() {
		primary := p.getPrimaryLastLine()
		print(primary)
		moveCursorBackwards(len(primary))
		moveCursorForwards(p.inputAt)
	}()

	if rprompt == "" {
		return
	}

	// Only print the right prompt if the input line
	// is shorter than the adjusted terminal width.
	lineFits := (rl.Prompt.inputAt + len(rl.line) +
		getRealLength(rprompt) + 1) < GetTermWidth()

	if !lineFits {
		return
	}

	// First go back to beginning of line, and clear everything
	moveCursorBackwards(GetTermWidth())
	print(seqClearLine)
	print(seqClearScreenBelow)

	// Go to where we must print the right prompt, print and go back
	forwardOffset := GetTermWidth() - getRealLength(rprompt) - 1
	moveCursorForwards(forwardOffset)
	print(rprompt)
	moveCursorBackwards(GetTermWidth())
}

// getPromptPrimary returns either the entire prompt if
// it's a single-line, or everything except the last line.
func (p *prompt) getPrimary() string {
	var primary string

	// Get the last line of the prompt to be printed.
	lastLineIndex := strings.LastIndex(p.primary, "\n")
	if lastLineIndex != -1 {
		primary = p.primary[:lastLineIndex+1]
	} else {
		primary = p.primary
	}

	return primary
}

// Get the last line of the prompt to be printed.
func (p *prompt) getPrimaryLastLine() string {
	var lastLine string
	lastLineIndex := strings.LastIndex(p.primary, "\n")
	if lastLineIndex != -1 {
		lastLine = p.primary[lastLineIndex+1:]
	} else {
		lastLine = p.primary
	}

	return lastLine
}

// computePromptAlt computes the correct lengths and offsets
// for all prompt components, but does not print any of them.
func (p *prompt) compute(rl *Instance) {
	prompt := rl.Prompt.primary

	lastLineIndex := strings.LastIndex(prompt, "\n")
	if lastLineIndex != -1 {
		rl.Prompt.inputAt = len([]rune(ansi.Strip(prompt[lastLineIndex:])))
	} else {
		rl.Prompt.inputAt = len([]rune(ansi.Strip(prompt)))
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

// RefreshPromptLog - A simple function to print a string message (a log, or more broadly,
// an asynchronous event) without bothering the user, and by "pushing" the prompt below the message.
// func (rl *Instance) RefreshPromptLog(log string) (err error) {
// 	// We adjust cursor movement, depending on which mode we're currently in.
// 	if !rl.modeTabCompletion {
// 		rl.tcUsedY = 1
// 		// Account for the hint line
// 	} else if rl.modeTabCompletion && rl.modeAutoFind {
// 		rl.tcUsedY = 0
// 	} else {
// 		rl.tcUsedY = 1
// 	}
//
// 	// Prompt offset
// 	if rl.isMultiline {
// 		rl.tcUsedY += 1
// 	} else {
// 		rl.tcUsedY += 0
// 	}
//
// 	// Clear the current prompt and everything below
// 	print(seqClearLine)
// 	if rl.stillOnRefresh {
// 		moveCursorUp(1)
// 	}
// 	rl.stillOnRefresh = true
// 	moveCursorUp(rl.hintY + rl.tcUsedY)
// 	moveCursorBackwards(GetTermWidth())
// 	print("\r\n" + seqClearScreenBelow)
//
// 	// Print the log
// 	fmt.Printf(log)
//
// 	// Add a new line between the message and the prompt, so not overloading the UI
// 	print("\n")
//
// 	// Print the prompt
// 	if rl.isMultiline {
// 		rl.tcUsedY += 3
// 		fmt.Println(rl.prompt)
//
// 	} else {
// 		rl.tcUsedY += 2
// 		fmt.Print(rl.prompt)
// 	}
//
// 	// Refresh the line
// 	rl.updateHelpers()
//
// 	return
// }
//
// // RefreshPromptInPlace - Refreshes the prompt in the very same place he is.
// func (rl *Instance) RefreshPromptInPlace(prompt string) (err error) {
// 	// We adjust cursor movement, depending on which mode we're currently in.
// 	// Prompt data intependent
// 	if !rl.modeTabCompletion {
// 		rl.tcUsedY = 1
// 		// Account for the hint line
// 	} else if rl.modeTabCompletion && rl.modeAutoFind {
// 		rl.tcUsedY = 0
// 	} else {
// 		rl.tcUsedY = 1
// 	}
//
// 	// Update the prompt if a special has been passed.
// 	if prompt != "" {
// 		rl.prompt = prompt
// 	}
//
// 	if rl.isMultiline {
// 		rl.tcUsedY += 1
// 	}
//
// 	// Clear the input line and everything below
// 	print(seqClearLine)
// 	moveCursorUp(rl.hintY + rl.tcUsedY)
// 	moveCursorBackwards(GetTermWidth())
// 	print("\r\n" + seqClearScreenBelow)
//
// 	// Add a new line if needed
// 	if rl.isMultiline {
// 		fmt.Println(rl.prompt)
// 	} else {
// 		fmt.Print(rl.prompt)
// 	}
//
// 	// Refresh the line
// 	rl.updateHelpers()
//
// 	return
// }
//
// // RefreshPromptCustom - Refresh the console prompt with custom values.
// // @prompt      => If not nil (""), will use this prompt instead of the currently set prompt.
// // @offset      => Used to set the number of lines to go upward, before reprinting. Set to 0 if not used.
// // @clearLine   => If true, will clean the current input line on the next refresh.
// func (rl *Instance) RefreshPromptCustom(prompt string, offset int, clearLine bool) (err error) {
// 	// We adjust cursor movement, depending on which mode we're currently in.
// 	if !rl.modeTabCompletion {
// 		rl.tcUsedY = 1
// 	} else if rl.modeTabCompletion && rl.modeAutoFind { // Account for the hint line
// 		rl.tcUsedY = 0
// 	} else {
// 		rl.tcUsedY = 1
// 	}
//
// 	// Add user-provided offset
// 	rl.tcUsedY += offset
//
// 	// Go back to prompt position, then up to the user provided offset.
// 	moveCursorBackwards(GetTermWidth())
// 	moveCursorUp(rl.posY)
// 	moveCursorUp(offset)
//
// 	// Then clear everything below our new position
// 	print(seqClearScreenBelow)
//
// 	// Update the prompt if a special has been passed.
// 	if prompt != "" {
// 		rl.prompt = prompt
// 	}
//
// 	// Add a new line if needed
// 	if rl.isMultiline && prompt == "" {
// 	} else if rl.isMultiline {
// 		fmt.Println(rl.prompt)
// 	} else {
// 		fmt.Print(rl.prompt)
// 	}
//
// 	// Refresh the line
// 	rl.updateHelpers()
//
// 	// If input line was empty, check that we clear it from detritus
// 	// The three lines are borrowed from clearLine(), we don't need more.
// 	if clearLine {
// 		rl.clearLine()
// 	}
//
// 	return
// }

// getRealLength - Some strings will have ANSI escape codes, which might be wrongly
// interpreted as legitimate parts of the strings. This will bother if some prompt
// components depend on other's length, so we always pass the string in this for
// getting its real-printed length.
func getRealLength(s string) (l int) {
	colorStripped := ansi.Strip(s)
	return utf8.RuneCountInString(colorStripped)
}
