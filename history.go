package readline

import (
	"fmt"
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

// AddHistorySource adds a source of history lines bound to a given name
// (printed above this source when used). All history sources can be used
// when completing history, and repeatedly using the completion key will
// cycle through them.
// When only the default in-memory history is bound, it is replaced with
// the provided ones. All following sources are added to the list.
func (rl *Instance) AddHistorySource(name string, history History) {
	if len(rl.histories) == 1 && rl.historyNames[0] == "local history" {
		delete(rl.histories, "local history")
		rl.historyNames = make([]string, 0)
	}

	rl.historyNames = append(rl.historyNames, name)
	rl.histories[name] = history
}

// defaultHistory is an example of a LineHistory interface:
type defaultHistory struct {
	items []string
}

// Write to history
func (h *defaultHistory) Write(s string) (int, error) {
	h.items = append(h.items, s)
	return len(h.items), nil
}

// GetLine returns a line from history
func (h *defaultHistory) GetLine(i int) (string, error) {
	return h.items[i], nil
}

// Len returns the number of lines in history
func (h *defaultHistory) Len() int {
	return len(h.items)
}

// Dump returns the entire history
func (h *defaultHistory) Dump() interface{} {
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
	rl.historySourcePos = 0
	rl.histPos = 0
	rl.undoHistory = []undoItem{{line: "", pos: 0}}
}

func (rl *Instance) nextHistorySource() {
	for i := range rl.historyNames {
		if i == rl.historySourcePos+1 {
			rl.historySourcePos = i
			break
		} else if i == len(rl.historyNames)-1 {
			rl.historySourcePos = 0
		}
	}
}

func (rl *Instance) prevHistorySource() {
	for i := range rl.historyNames {
		if i == rl.historySourcePos-1 {
			rl.historySourcePos = i
			break
		} else if i == len(rl.historyNames)-1 {
			rl.historySourcePos = len(rl.historyNames) - 1
		}
	}
}

func (rl *Instance) currentHistory() History {
	return rl.histories[rl.historyNames[rl.historySourcePos]]
}

// walkHistory - Browse historic lines
func (rl *Instance) walkHistory(i int) {
	var (
		new string
		err error
	)

	// Always use the main/first history.
	rl.historySourcePos = 0
	history := rl.currentHistory()

	if history == nil || history.Len() == 0 {
		return
	}

	// When we are exiting the current line buffer to move around
	// the history, we make buffer the current line
	if rl.histPos == 0 && (rl.histPos+i) == 1 {
		rl.lineBuf = string(rl.line)
	}

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
func (rl *Instance) completeHistory() Completions {
	if len(rl.histories) == 0 {
		return Completions{}
	}

	history := rl.currentHistory()
	if history == nil {
		return Completions{}
	}

	rl.histHint = []rune(rl.historyNames[rl.historySourcePos])

	// Set the hint line with everything
	rl.histHint = append([]rune(seqBold+seqFgCyanBright+string(rl.histHint)+seqReset), rl.tfLine...)
	rl.histHint = append(rl.histHint, []rune(seqReset)...)

	hist := &comps{
		maxLength: 10,
	}

	compLines := make([]Completion, 0)

	var (
		line string
		err  error
	)

NEXT_LINE:
	for i := history.Len() - 1; i > -1; i-- {
		line, err = history.GetLine(i)
		if err != nil {
			continue
		}

		if !strings.HasPrefix(line, string(rl.line)) || strings.TrimSpace(line) == "" {
			continue
		}

		line = strings.ReplaceAll(line, "\n", ` `)

		for _, row := range hist.values {
			if row[0].Display == line {
				continue NEXT_LINE
			}
		}

		// Proper pad for indexes
		indexStr := strconv.Itoa(i)
		pad := strings.Repeat(" ", len(strconv.Itoa(history.Len()))-len(indexStr))
		display := fmt.Sprintf("%s%s %s%s", seqDim, indexStr+pad, seqDimReset, line)

		value := Completion{
			Display: display,
			Value:   line,
		}

		compLines = append(compLines, value)
	}

	comps := CompleteRaw(compLines)
	comps.PREFIX = string(rl.line)

	return comps
}
