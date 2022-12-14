package readline

import (
	"strconv"
	"strings"
)

// History is an interface to allow you to write your own history logging
// tools. eg sqlite backend instead of a file system.
// By default readline will just use the dummyLineHistory interface which only
// logs the history to memory ([]string to be precise).
type History interface {
	// Append takes the line and returns an updated number of lines or an error
	Write(string) (int, error)

	// GetLine takes the historic line number and returns the line or an error
	GetLine(int) (string, error)

	// Len returns the number of history lines
	Len() int

	// Dump returns everything in readline. The return is an interface{} because
	// not all LineHistory implementations will want to structure the history in
	// the same way. And since Dump() is not actually used by the readline API
	// internally, this methods return can be structured in whichever way is most
	// convenient for your own applications (or even just create an empty
	// function which returns `nil` if you don't require Dump() either)
	Dump() interface{}
}

// SetHistoryCtrlR - Set the history source triggered with Ctrl-r combination
func (rl *Instance) SetHistoryCtrlR(name string, history History) {
	rl.mainHistName = name
	rl.mainHistory = history
}

// GetHistoryCtrlR - Returns the history source triggered by Ctrl-r
func (rl *Instance) GetHistoryCtrlR() History {
	return rl.mainHistory
}

// SetHistoryAltR - Set the history source triggered with Alt-r combination
func (rl *Instance) SetHistoryAltR(name string, history History) {
	rl.altHistName = name
	rl.altHistory = history
}

// GetHistoryAltR - Returns the history source triggered by Alt-r
func (rl *Instance) GetHistoryAltR() History {
	return rl.altHistory
}

// ExampleHistory is an example of a LineHistory interface:
type ExampleHistory struct {
	items []string
}

// Write to history
func (h *ExampleHistory) Write(s string) (int, error) {
	h.items = append(h.items, s)
	return len(h.items), nil
}

// GetLine returns a line from history
func (h *ExampleHistory) GetLine(i int) (string, error) {
	return h.items[i], nil
}

// Len returns the number of lines in history
func (h *ExampleHistory) Len() int {
	return len(h.items)
}

// Dump returns the entire history
func (h *ExampleHistory) Dump() interface{} {
	return h.items
}

// NullHistory is a null History interface for when you don't want line
// entries remembered eg password input.
type NullHistory struct{}

// Write to history
func (h *NullHistory) Write(s string) (int, error) {
	return 0, nil
}

// GetLine returns a line from history
func (h *NullHistory) GetLine(i int) (string, error) {
	return "", nil
}

// Len returns the number of lines in history
func (h *NullHistory) Len() int {
	return 0
}

// Dump returns the entire history
func (h *NullHistory) Dump() interface{} {
	return []string{}
}

// initHistory is ran once at the beginning of an instance start.
func (rl *Instance) initHistory() {
	// We need this set to the last command, so that we can access it quickly
	rl.histPos = 0
	rl.undoHistory = []undoItem{{line: "", pos: 0}}
}

// walkHistory - Browse historic lines
func (rl *Instance) walkHistory(i int) {
	var (
		new string
		err error
	)

	// Work with correct history source (depends on CtrlR/CtrlE)
	var history History
	if !rl.mainHist {
		history = rl.altHistory
	} else {
		history = rl.mainHistory
	}

	// Nothing happens if the history is nil or empty.
	if history == nil || history.Len() == 0 {
		return
	}

	// When we are exiting the current line buffer to move around
	// the history, we make buffer the current line
	if rl.histPos == 0 && (rl.histPos+i) == 1 {
		rl.lineBuf = string(rl.line)
	}

	// Move the history position first. It is caught below if out of bounds.
	rl.histPos += i

	switch {
	case rl.histPos > history.Len():
		// The history is greater than the length of history: maintain
		// it at the last index, to keep the same line in the buffer.
		rl.histPos = history.Len()
	case rl.histPos < 0:
		// We can never go lower than the last history line, which is our current line.
		rl.histPos = 0
	case rl.histPos == 0:
		// The 0 index is our current line
		rl.line = []rune(rl.lineBuf)
		rl.pos = len(rl.lineBuf)
	}

	// We now have the correct history index. Use it to find the history line.
	// If the history position is not zero, we need to use a history line.
	if rl.histPos > 0 {
		new, err = history.GetLine(history.Len() - rl.histPos)
		if err != nil {
			rl.resetHelpers()
			print("\r\n" + err.Error() + "\r\n")
			// NOTE: Same here ? Should we print the prompt more properly ?
			print(rl.Prompt.primary)
			return
		}

		rl.clearLine()
		rl.line = []rune(new)
		rl.pos = len(rl.line)
	}
}

// completeHistory - Populates a CompletionGroup with history and returns it the shell
// we populate only one group, so as to pass it to the main completion engine.
func (rl *Instance) completeHistory() (hist *CompletionGroup, notEmpty bool) {
	hist = &CompletionGroup{
		DisplayType: TabDisplayMap,
		MaxLength:   10,
	}

	// Switch to completion flux first
	var history History
	if !rl.mainHist {
		if rl.altHistory == nil {
			return
		}
		history = rl.altHistory
		rl.histHint = []rune(rl.altHistName + ": ")
	} else {
		if rl.mainHistory == nil {
			return
		}
		history = rl.mainHistory
		rl.histHint = []rune(rl.mainHistName + ": ")
	}

	if history.Len() > 0 {
		notEmpty = true
	}

	hist.init(rl)

	var (
		line string
		err  error
	)

	// rl.tcPrefix = string(rl.line) // We use the current full line for filtering

NEXT_LINE:
	for i := history.Len() - 1; i > -1; i-- {
		line, err = history.GetLine(i)
		if err != nil {
			continue
		}

		if !strings.HasPrefix(line, rl.tcPrefix) || strings.TrimSpace(line) == "" {
			continue
		}

		line = strings.ReplaceAll(line, "\n", ` `)

		for _, val := range hist.Values {
			if val.Display == line {
				continue NEXT_LINE
			}
		}

		value := CompletionValue{
			Display:     line,
			Value:       line,
			Description: DIM + strconv.Itoa(i) + seqReset,
		}
		hist.Values = append(hist.Values, value)
	}

	return
}
