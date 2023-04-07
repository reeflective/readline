package history

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/ui"
)

// Sources manages and serves all history sources for the current shell.
type Sources struct {
	list       map[string]Source // Sources of history lines
	names      []string          // Names of histories stored in rl.histories
	sourcePos  int               // The index of the currently used history
	buf        string            // The current line saved when we are on another history line
	pos        int               // Index used for navigating the history lines with arrows/j/k
	infer      bool              // If the last command ran needs to infer the history line.
	accepted   bool              // The line has been accepted and must be returned.
	acceptHold bool              // Should we reuse the same accepted line on the next loop.
	acceptLine core.Line         // The line to return to the caller.
	acceptErr  error             // An error to return to the caller.

	// Shell parameters
	line   *core.Line
	cursor *core.Cursor
	hint   *ui.Hint
}

// NewSources is a required constructor for the history sources manager type.
func NewSources(line *core.Line, cur *core.Cursor, hint *ui.Hint) *Sources {
	sources := &Sources{
		list:   make(map[string]Source),
		line:   line,
		cursor: cur,
		hint:   hint,
	}

	sources.names = append(sources.names, defaultSourceName)
	sources.list[defaultSourceName] = new(memory)

	return sources
}

// Init initializes the history sources positions and buffers
// at the start of each readline loop. If the last command asked
// to infer a command line from the history, it is performed now.
func (h *Sources) Init() {
	h.sourcePos = 0
	h.accepted = false
	h.acceptErr = nil
	h.acceptLine = nil

	if !h.infer {
		h.pos = 0
		return
	}

	switch h.pos {
	case -1:
		h.pos = 0
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
	if len(h.list) == 1 && h.names[0] == defaultSourceName {
		delete(h.list, defaultSourceName)
		h.names = make([]string, 0)
	}

	h.names = append(h.names, name)
	h.list[name] = hist
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
		h.list = make(map[string]Source)
		h.names = make([]string, 0)

		return
	}

	for _, name := range sources {
		delete(h.list, name)

		for i, hname := range h.names {
			if hname == name {
				h.names = append(h.names[:i], h.names[i+1:]...)
				break
			}
		}
	}

	h.sourcePos = 0
	if !h.infer {
		h.pos = 0
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
	if (pos < 0 && h.pos == 0) || (pos > 0 && h.pos == history.Len()) {
		return
	}

	// Save the current line buffer if we are leaving it.
	if h.pos == 0 && (h.pos+pos) == 1 {
		h.buf = string(*h.line)
	}

	h.pos += pos

	switch {
	case h.pos > history.Len():
		h.pos = history.Len()
	case h.pos < 0:
		h.pos = 0
	case h.pos == 0:
		h.line.Set([]rune(h.buf)...)
		h.cursor.Set(h.line.Len())
	}

	if h.pos == 0 {
		return
	}

	// We now have the correct history index, fetch the line.
	next, err := history.GetLine(history.Len() - h.pos)
	if err != nil {
		h.hint.Set(color.FgRed + "history error: " + err.Error())
		return
	}

	h.line.Set([]rune(next)...)
	h.cursor.Set(h.line.Len())
}

// GetLast returns the last saved history line in the active history source.
func (h *Sources) GetLast() string {
	history := h.Current()

	if history == nil || history.Len() == 0 {
		return ""
	}

	last, err := history.GetLine(history.Len() - 1)
	if err != nil {
		return ""
	}

	return last
}

// Cycle checks for the next history source (if any) and makes it the active one.
// If next is false, the source cycles to the previous source.
func (h *Sources) Cycle(next bool) {
	switch next {
	case true:
		h.sourcePos++

		if h.sourcePos == len(h.names) {
			h.sourcePos = 0
		}
	case false:
		h.sourcePos--

		if h.sourcePos < 0 {
			h.sourcePos = len(h.names) - 1
		}
	}
}

// OnLastSource returns true if the currently active
// history source is the last one in the list.
func (h *Sources) OnLastSource() bool {
	return h.sourcePos == len(h.names)-1
}

// Current returns the current/active history source.
func (h *Sources) Current() Source {
	if len(h.list) == 0 {
		return nil
	}

	return h.list[h.names[h.sourcePos]]
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

	if len(strings.TrimSpace(line)) == 0 {
		return
	}

	for _, history := range h.list {
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
		h.pos, err = history.Write(line)
		if err != nil {
			h.hint.Set(color.FgRed + err.Error())
		}
	}
}

// Accept signals the line has been accepted by the user and must be
// returned to the readline caller. If hold is true, the line is preserved
// and redisplayed on the next loop. If infer, the line is not written to
// the history, but preserved as a line to match against on the next loop.
// If infer is false, the line is automatically written to active sources.
func (h *Sources) Accept(hold, infer bool, err error) {
	h.accepted = true
	h.acceptHold = hold
	h.acceptLine = *h.line
	h.acceptErr = err

	// Simply write the line to the history sources.
	h.Write(infer)
}

// LineAccepted returns true if the user has accepted the line, signaling
// that the shell must return from its loop. The error can be nil, but may
// indicate a CtrlC/CtrlD style error.
// If the input line contains any comments (as defined by the configured
// comment sign), they will be removed before returning the line. Those
// are nonetheless preserved when the line is saved to history sources.
func (h *Sources) LineAccepted() (bool, string, error) {
	if !h.accepted {
		return false, "", nil
	}

	// Remove all comments before returning the line to the caller.
	// TODO: Replace # with configured comment sign
	commentsMatch := regexp.MustCompile(`(^|\s)#.*`)
	line := commentsMatch.ReplaceAllString(string(h.acceptLine), "")

	return true, line, h.acceptErr
}

// InsertMatch replaces the line buffer with the first history line
// in the active source that matches the input line as a prefix.
func (h *Sources) InsertMatch(forward bool) {
	if len(h.list) == 0 {
		return
	}

	if h.Current() == nil {
		return
	}

	line, pos, found := h.matchFirst(forward)
	if !found {
		return
	}

	h.pos = pos
	h.buf = string(*h.line)
	h.line.Set([]rune(line)...)
	h.cursor.Set(h.line.Len())
}

// InferNext finds a line matching the current line in the history,
// finds the next line after it and, if any, inserts it.
func (h *Sources) InferNext() {
	if len(h.list) == 0 {
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
func (h *Sources) Suggest(line *core.Line) core.Line {
	if len(h.list) == 0 || len(*line) == 0 {
		return *line
	}

	if h.Current() == nil {
		return *line
	}

	// Don't autosuggest when the line is not
	// the current input line: we are completing.
	if line != h.line {
		return *line
	}
	// current := h.line
	// h.line = line
	// defer func() {
	// 	h.line = current
	// }()

	suggested, _, found := h.matchFirst(false)
	if !found {
		return *line
	}

	return core.Line([]rune(suggested))
}

// Complete returns completions with the current history source values.
// If forward is true, the completions are proposed from the most ancient
// line in the history source to the most recent. If filter is true,
// only lines that match the current input line as a prefix are given.
func (h *Sources) Complete(forward, filter bool) completion.Values {
	if len(h.list) == 0 {
		return completion.Values{}
	}

	history := h.Current()
	if history == nil {
		return completion.Values{}
	}

	h.hint.Set(color.Bold + color.FgCyanBright + h.names[h.sourcePos] + color.Reset)

	compLines := make([]completion.Candidate, 0)

	// Set up iteration clauses
	var histPos int
	var done func(i int) bool
	var move func(inc int) int

	if forward {
		histPos = -1
		done = func(i int) bool { return i < history.Len()-1 }
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
	comps.ListLong["*"] = true
	comps.PREFIX = string(*h.line)

	return comps
}

// Name returns the name of the currently active history source.
func (h *Sources) Name() string {
	return h.names[h.sourcePos]
}

func (h *Sources) matchFirst(forward bool) (line string, pos int, found bool) {
	if len(h.list) == 0 {
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
		if !strings.HasPrefix(histline, string(*h.line)) {
			// if !strings.HasPrefix(string(*h.line), histline) {
			continue
		}

		// Else we have our history match.
		return histline, histPos, true
	}

	// We should have returned a match from the loop.
	return "", 0, false
}
