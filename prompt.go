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

// Primary uses a function returning the string to use as the primary prompt.
func (p *prompt) Primary(prompt func() string) {
	p.primaryF = prompt
}

// Right uses a function returning the string to use as the right prompt.
func (p *prompt) Right(prompt func() string) {
	p.rightF = prompt
}

// Secondary uses a function returning the prompt to use as the secondary prompt.
func (p *prompt) Secondary(prompt func() string) {
	p.secondaryF = prompt
}

// Transient uses a function returning the prompt to use as a transient prompt.
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
	if p.transientF != nil {
		p.transient = p.transientF()
	}
	if p.secondaryF != nil {
		p.secondary = p.secondaryF()
	}

	// Compute some offsets needed by the last line.
	rl.Prompt.compute(rl)

	// Print the primary prompt, potentially excluding the last line.
	print(p.getPrimary())
	p.stillOnRefresh = false
}

// getPromptPrimary returns either the entire prompt if
// it's a single-line, or everything except the last line.
func (p *prompt) getPrimary() string {
	var primary string

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
	prompt := p.primary

	lastLineIndex := strings.LastIndex(prompt, "\n")
	if lastLineIndex != -1 {
		rl.Prompt.inputAt = getRealLength(prompt[lastLineIndex+1:])
	} else {
		rl.Prompt.inputAt = getRealLength(prompt)
	}
}

// update is called after each key/widget processing, and refreshes
// the prompts that need to be at these intervals.
func (p *prompt) update(rl *Instance) {
	if rl.Prompt.tooltipF == nil {
		return
	}

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
	defer print(p.getPrimaryLastLine())

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

func (p *prompt) printTransient(rl *Instance) {
	if p.transient == "" {
		return
	}

	// First offset the newlines returned by our widgets,
	// and clear everything below us.
	moveCursorBackwards(GetTermWidth())
	moveCursorUp(rl.fullY)
	promptLines := strings.Count(p.primary, "\n")
	moveCursorUp(promptLines)
	print(seqClearLine)
	print(seqClearScreenBelow)

	// And print both the prompt and the input line.
	print(p.transient)
	println(string(rl.line))
}

// getRealLength - Some strings will have ANSI escape codes, which might be wrongly
// interpreted as legitimate parts of the strings. This will bother if some prompt
// components depend on other's length, so we always pass the string in this for
// getting its real-printed length.
func getRealLength(s string) (l int) {
	colorStripped := ansi.Strip(s)
	return utf8.RuneCountInString(colorStripped)
}
