package history

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/common"
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/ui"
)

// Sources manages and serves all history sources for the current shell.
type Sources struct {
	histories        map[string]Source // Sources of history lines
	historyNames     []string          // Names of histories stored in rl.histories
	historySourcePos int               // The index of the currently used history
	lineBuf          string            // The current line saved when we are on another history line
	histPos          int               // Index used for navigating the history lines with arrows/j/k
	histHint         []rune            // We store a hist hint, for dual history sources
	histSuggested    []rune            // The last matching history line matching the current input.
	infer            bool              // If the last command ran needs to infer the history line.

	// Shell parameters
	line   *common.Line
	cursor *common.Cursor
	hint   *ui.Hint
}

// NewSources is a required constructor for the history sources manager type.
func NewSources(line *common.Line, cur *common.Cursor, hint *ui.Hint) *Sources {
	sources := &Sources{
		histories: make(map[string]Source),
		line:      line,
		cursor:    cur,
		hint:      hint,
	}

	sources.historyNames = append(sources.historyNames, defaultSourceName)
	sources.histories[defaultSourceName] = new(memory)

	return sources
}

// Init initializes the history sources positions and buffers
// at the start of each readline loop. If the last command asked
// to infer a command line from the history, it is performed now.
func (h *Sources) Init() {
	h.historySourcePos = 0

	if !h.infer {
		h.histPos = 0
		return
	}

	switch h.histPos {
	case -1:
		h.histPos = 0
	case 0:
		h.InferNext()
	default:
		h.Walk(-1)
	}

	h.infer = false
}

// Add adds a source of history lines bound to a given name (printed above
// this source when used). When only the default in-memory history is bound,
// it's replaced with the provided source. Following ones are added to the list.
func (h *Sources) Add(name string, hist Source) {
	if len(h.histories) == 1 && h.historyNames[0] == defaultSourceName {
		delete(h.histories, defaultSourceName)
		h.historyNames = make([]string, 0)
	}

	h.historyNames = append(h.historyNames, name)
	h.histories[name] = hist
}

// New creates a new History populated from, and writing to a file.
func (h *Sources) AddFromFile(name, file string) {
	hist := new(fileHistory)
	hist.file = file
	hist.lines, _ = openHist(file)
}

// Delete deletes one or more history source by name.
// If no arguments are passed, all currently bound sources are removed.
func (h *Sources) Delete(sources ...string) {
	if len(sources) == 0 {
		h.histories = make(map[string]Source)
		h.historyNames = make([]string, 0)

		return
	}

	for _, name := range sources {
		delete(h.histories, name)

		for i, hname := range h.historyNames {
			if hname == name {
				h.historyNames = append(h.historyNames[:i], h.historyNames[i+1:]...)
				break
			}
		}
	}

	h.historySourcePos = 0
	if !h.infer {
		h.histPos = 0
	}
}

// Walk goes to the next or previous history line in the active source.
// If at the end of the history, the last history line is kept.
// If going back to the beginning of it, the saved line buffer is restored.
func (h *Sources) Walk(pos int) {
	// We used to force using the main history.
	// h.historySourcePos = 0
	history := h.Current()

	if history == nil || history.Len() == 0 {
		return
	}

	// When we are on the last/first item, don't do anything,
	// as it would change things like cursor positions.
	if (pos < 0 && h.histPos == 0) || (pos > 0 && h.histPos == history.Len()) {
		return
	}

	// Save the current line buffer if we are leaving it.
	if h.histPos == 0 && (h.histPos+pos) == 1 {
		h.lineBuf = string(*h.line)
	}

	h.histPos += pos

	switch {
	case h.histPos > history.Len():
		h.histPos = history.Len()
	case h.histPos < 0:
		h.histPos = 0
	case h.histPos == 0:
		h.line.Set([]rune(h.lineBuf)...)
		h.cursor.Set(h.line.Len())
	}

	if h.histPos == 0 {
		return
	}

	// We now have the correct history index, fetch the line.
	next, err := history.GetLine(history.Len() - h.histPos)
	if err != nil {
		h.hint.Set(color.FgRed + "history error: " + err.Error())
		return
	}

	h.line.Set([]rune(next)...)
	h.cursor.Set(h.line.Len())
}

// Cycle checks for the next history source (if any) and makes it the active one.
// If next is false, the source cycles to the previous source.
func (h *Sources) Cycle(next bool) {
	switch next {
	case true:
		h.historySourcePos++

		if h.historySourcePos == len(h.historyNames) {
			h.historySourcePos = 0
		}
	case false:
		h.historySourcePos--

		if h.historySourcePos < 0 {
			h.historySourcePos = len(h.historyNames) - 1
		}
	}
}

// Current returns the current/active history source.
func (h *Sources) Current() Source {
	if len(h.histories) == 0 {
		return nil
	}

	return h.histories[h.historyNames[h.historySourcePos]]
}

// Write writes the accepted input line to all available sources.
// If infer is true, the next history initialization will automatically
// insert the next history line found after the first match of the line
// that has just been written (thus, normally, accepted/executed).
func (h *Sources) Write(infer bool) {
	if infer {
		h.infer = true
		return
	}

	line := string(*h.line)

	for _, history := range h.histories {
		if history == nil {
			continue
		}

		var err error

		// Don't write the line if it's identical to the last one.
		last, err := history.GetLine(0)
		if err == nil && last != "" && last == line {
			continue
		}

		// Save the line and notify through hints if an error raised.
		h.histPos, err = history.Write(line)
		if err != nil {
			h.hint.Set(color.FgRed + err.Error())
		}
	}
}

// InsertMatch replaces the line buffer with the first history line
// in the active source that matches the input line as a prefix.
func (h *Sources) InsertMatch(forward bool) {
	if len(h.histories) == 0 {
		return
	}

	history := h.Current()
	if history == nil {
		return
	}

	line, pos, found := h.matchFirst(forward)
	if !found {
		return
	}

	h.histPos = pos
	h.lineBuf = string(*h.line)
	h.line.Set([]rune(line)...)
	h.cursor.Set(h.line.Len())
}

// InferNext finds a line matching the current line in the history,
// finds the next line after it and, if any, inserts it.
func (h *Sources) InferNext() {
	if len(h.histories) == 0 {
		return
	}

	history := h.Current()
	if history == nil {
		return
	}

	_, pos, found := h.matchFirst(false)
	if !found {
		return
	}

	// If we have no match we return, or check for the next line.
	if history.Len() <= (history.Len()-pos)+1 {
		return
	}

	// Insert the next line
	line, err := history.GetLine(pos + 1)
	if err != nil {
		return
	}

	h.line.Set([]rune(line)...)
	h.cursor.Set(h.line.Len())
}

// Suggest returns the first line matching the current line buffer,
// so that caller can use for things like history autosuggestion.
// If no line matches the current line, it will return the latter.
func (h *Sources) Suggest() common.Line {
	if len(h.histories) == 0 {
		return *h.line
	}

	history := h.Current()
	if history == nil {
		return *h.line
	}

	line, _, found := h.matchFirst(false)
	if !found {
		return *h.line
	}

	return common.Line([]rune(line))
}

// Complete returns completions with the current history source values.
// If forward is true, the completions are proposed from the most ancient
// line in the history source to the most recent. If filter is true,
// only lines that match the current input line as a prefix are given.
func (h *Sources) Complete(forward, filter bool) completion.Values {
	if len(h.histories) == 0 {
		return completion.Values{}
	}

	history := h.Current()
	if history == nil {
		return completion.Values{}
	}

	h.hint.Set(color.Bold + color.FgCyanBright + h.historyNames[h.historySourcePos] + color.Reset)

	compLines := make([]completion.Candidate, 0)

	// Set up iteration clauses
	var histPos int
	var done func(i int) bool
	var move func(inc int) int

	if forward {
		histPos = -1
		done = func(i int) bool { return i < history.Len() }
		move = func(pos int) int { return pos + 1 }
	} else {
		histPos = history.Len()
		done = func(i int) bool { return i > 0 }
		move = func(pos int) int { return pos - 1 }
	}

	// And generate the completions.
nextLine:
	for done(histPos) {
		histPos = move(histPos)

		line, err := history.GetLine(histPos)
		if err != nil {
			continue
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		if filter && !strings.HasPrefix(line, string(*h.line)) {
			continue
		}

		display := strings.ReplaceAll(line, "\n", ``)

		for _, comp := range compLines {
			if comp.Display == line {
				continue nextLine
			}
		}

		// Proper pad for indexes
		indexStr := strconv.Itoa(histPos)
		pad := strings.Repeat(" ", len(strconv.Itoa(history.Len()))-len(indexStr))
		display = fmt.Sprintf("%s%s %s%s", color.Dim, indexStr+pad, color.DimReset, display)

		value := completion.Candidate{
			Display: display,
			Value:   line,
		}

		compLines = append(compLines, value)
	}

	comps := completion.AddRaw(compLines)
	comps.NoSort["*"] = true
	comps.PREFIX = string(*h.line)

	return comps
}

func (h *Sources) matchFirst(forward bool) (line string, pos int, found bool) {
	if len(h.histories) == 0 {
		return
	}

	history := h.Current()
	if history == nil {
		return
	}

	// Set up iteration clauses
	var histPos int
	var done func(i int) bool
	var move func(inc int) int

	if forward {
		histPos = -1
		done = func(i int) bool { return i < history.Len() }
		move = func(pos int) int { return pos + 1 }
	} else {
		histPos = history.Len()
		done = func(i int) bool { return i > 0 }
		move = func(pos int) int { return pos - 1 }
	}

	for done(histPos) {
		histPos = move(histPos)

		histline, err := history.GetLine(histPos)
		if err != nil {
			return
		}

		// If too short
		if len(histline) < h.line.Len() {
			continue
		}

		// Or if not fully matching
		if !strings.HasPrefix(string(*h.line), histline) {
			continue
		}

		// Else we have our history match.
		return histline, histPos, true
	}

	return
}
