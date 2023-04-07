package readline

import (
	"strings"

	"github.com/reeflective/readline/internal/history"
	"github.com/reeflective/readline/internal/strutil"
)

//
// API ----------------------------------------------------------------
//

// NewHistoryFromFile creates a new command history
// source writing to and reading from a file.
var NewHistoryFromFile = history.NewSourceFromFile

// AddHistoryFromFile adds a command history source from a file path.
// The name is used when using/searching the history source.
func (rl *Shell) AddHistoryFromFile(name, filepath string) {
	rl.histories.AddFromFile(name, filepath)
}

// Add adds a source of history lines bound to a given name (printed above
// this source when used). When only the default in-memory history is bound,
// it's replaced with the provided source. Following ones are added to the list.
func (rl *Shell) AddHistory(name string, source history.Source) {
	rl.histories.Add(name, source)
}

// Delete deletes one or more history source by name.
// If no arguments are passed, all currently bound sources are removed.
func (rl *Shell) DeleteHistory(sources ...string) {
	rl.histories.Delete(sources...)
}

// historyWidgets returns all history commands.
// Under each comment are gathered all commands related to the comment's
// subject. When there are two subgroups separated by an empty line, the
// second one comprises commands that are not legacy readline commands.
func (rl *Shell) historyWidgets() commands {
	widgets := map[string]func(){
		"accept-line":                            rl.acceptLine,
		"next-history":                           rl.downHistory, // down-history
		"previous-history":                       rl.upHistory,   // up-history
		"beginning-of-history":                   rl.beginningOfHistory,
		"end-of-history":                         rl.endOfHistory,
		"operate-and-get-next":                   rl.acceptLineAndDownHistory, // accept-line-and-down-history
		"fetch-history":                          rl.fetchHistory,
		"forward-search-history":                 rl.historyIncrementalSearchForward,  // history-incremental-search-forward
		"reverse-search-history":                 rl.historyIncrementalSearchBackward, // history-incremental-search-backward
		"non-incremental-forward-search-history": rl.nonIncrementalForwardSearchHistory,
		"non-incremental-reverse-search-history": rl.nonIncrementalReverseSearchHistory,
		"history-search-forward":                 rl.historySearchForward,
		"history-search-backward":                rl.historySearchBackward,
		"history-substring-search-forward":       rl.historySubstringSearchForward,
		"history-substring-search-backward":      rl.historySubstringSearchBackward,
		"yank-last-arg":                          rl.yankLastArg,
		"yank-nth-arg":                           rl.yankNthArg,

		"accept-and-hold":                   rl.acceptAndHold,
		"accept-and-infer-next-history":     rl.acceptAndInferNextHistory,
		"down-line-or-history":              rl.downLineOrHistory,
		"up-line-or-history":                rl.upLineOrHistory,
		"up-line-or-search":                 rl.upLineOrSearch,
		"down-line-or-search":               rl.downLineOrSearch,
		"infer-next-history":                rl.inferNextHistory,
		"beginning-of-buffer-or-history":    rl.beginningOfBufferOrHistory,
		"beginning-history-search-forward":  rl.beginningHistorySearchForward,
		"beginning-history-search-backward": rl.beginningHistorySearchBackward,
		"end-of-buffer-or-history":          rl.endOfBufferOrHistory,
		// "history-autosuggest-insert":          rl.historyAutosuggestInsert,
		"beginning-of-line-hist": rl.beginningOfLineHist,
		"end-of-line-hist":       rl.endOfLineHist,
	}

	return widgets
}

//
// Standard ----------------------------------------------------------------
//

func (rl *Shell) acceptLine() {
	rl.acceptLineWith(false, false)
}

func (rl *Shell) downHistory() {
	rl.undo.SkipSave()
	rl.histories.Walk(-1)
}

func (rl *Shell) upHistory() {
	rl.undo.SkipSave()
	rl.histories.Walk(1)
}

func (rl *Shell) beginningOfHistory() {
	rl.undo.SkipSave()

	history := rl.histories.Current()
	if history == nil {
		return
	}

	rl.histories.Walk(history.Len())
}

func (rl *Shell) endOfHistory() {
	history := rl.histories.Current()

	if history == nil {
		return
	}

	rl.histories.Walk(-history.Len() + 1)
}

func (rl *Shell) acceptLineAndDownHistory() {
	// rl.inferLine = true // The next loop will retrieve a line by histPos.
	// rl.acceptLine()
}

func (rl *Shell) fetchHistory() {}

func (rl *Shell) historyIncrementalSearchForward() {
	rl.undo.SkipSave()
	rl.historyCompletion(true, false)
}

func (rl *Shell) historyIncrementalSearchBackward() {
	rl.undo.SkipSave()
	rl.historyCompletion(false, false)
}

func (rl *Shell) nonIncrementalForwardSearchHistory() {}
func (rl *Shell) nonIncrementalReverseSearchHistory() {}

func (rl *Shell) historySearchForward() {
	rl.undo.SkipSave()

	// And either roll to the next history source, or
	// directly generate completions for the target history.
	//
	// Set the tab completion prefix as a filtering
	// mechanism here: will be updated by the comps anyway.
	// rl.historyCompletion(true, true)
}

func (rl *Shell) historySearchBackward() {
	rl.undo.SkipSave()

	// And either roll to the next history source, or
	// directly generate completions for the target history.
	//
	// Set the tab completion prefix as a filtering
	// mechanism here: will be updated by the comps anyway.
	// rl.historyCompletion(false, true)
}

func (rl *Shell) historySubstringSearchForward()  {}
func (rl *Shell) historySubstringSearchBackward() {}

func (rl *Shell) yankLastArg() {
	// Get the last history line.
	last := rl.histories.GetLast()
	if last == "" {
		return
	}

	// Split it into words, and get the last one.
	words, err := strutil.Split(last)
	if err != nil || len(words) == 0 {
		return
	}

	// Get the last word, and quote it if it contains spaces.
	lastArg := words[len(words)-1]
	if strings.ContainsAny(lastArg, " \t") {
		if strings.Contains(lastArg, "\"") {
			lastArg = "'" + lastArg + "'"
		} else {
			lastArg = "\"" + lastArg + "\""
		}
	}

	// And append it to the end of the line.
	rl.line.Insert(rl.cursor.Pos(), []rune(lastArg)...)
	rl.cursor.Move(len(lastArg))
}

func (rl *Shell) yankNthArg() {
	// Get the last history line.
	last := rl.histories.GetLast()
	if last == "" {
		return
	}

	// Split it into words, and get the last one.
	words, err := strutil.Split(last)
	if err != nil || len(words) == 0 {
		return
	}

	var lastArg string

	// Abort if the required position is out of bounds.
	argNth := rl.iterations.Get()
	if len(words) < argNth {
		return
	} else {
		lastArg = words[argNth-1]
	}

	// Quote if required.
	if strings.ContainsAny(lastArg, " \t") {
		if strings.Contains(lastArg, "\"") {
			lastArg = "'" + lastArg + "'"
		} else {
			lastArg = "\"" + lastArg + "\""
		}
	}

	// And append it to the end of the line.
	rl.line.Insert(rl.line.Len(), []rune(lastArg)...)
	rl.cursor.Move(len(lastArg))
}

//
// Added -------------------------------------------------------------------
//

func (rl *Shell) acceptAndHold() {
	rl.acceptLineWith(true, false)
}

func (rl *Shell) acceptAndInferNextHistory() {
	// rl.inferLine = true // The next loop will retrieve a line.
	// rl.histPos = 0      // And will find it by trying to match one.
	// rl.acceptLine()
}

func (rl *Shell) downLineOrHistory() {
	rl.undo.SkipSave()

	times := rl.iterations.Get()
	linesDown := rl.line.Lines() - rl.cursor.Line()

	// If we can go down some lines out of
	// the available iterations, use them.
	if linesDown > 0 {
		rl.cursor.LineMove(times)
		times -= linesDown
	}

	if times > 0 {
		rl.histories.Walk(times * -1)
	}
}

func (rl *Shell) upLineOrHistory() {
	rl.undo.SkipSave()

	times := rl.iterations.Get()
	linesUp := rl.cursor.Line()

	// If we can go down some lines out of
	// the available iterations, use them.
	if linesUp > 0 {
		rl.cursor.LineMove(times * -1)
		times -= linesUp
	}

	if times > 0 {
		rl.histories.Walk(times)
	}
}

func (rl *Shell) upLineOrSearch() {
	rl.undo.SkipSave()
	switch {
	case rl.cursor.Line() > 0:
		rl.cursor.LineMove(-1)
	default:
		rl.historySearchBackward()
	}
}

func (rl *Shell) downLineOrSearch() {
	rl.undo.SkipSave()
	switch {
	case rl.cursor.Line() < rl.line.Lines():
		rl.cursor.LineMove(1)
	default:
		rl.historySearchForward()
	}
}

func (rl *Shell) inferNextHistory() {
	rl.undo.SkipSave()
	rl.histories.InferNext()
}

func (rl *Shell) beginningOfBufferOrHistory() {
	rl.undo.SkipSave()

	if rl.cursor.Pos() > 0 {
		rl.cursor.Set(0)
		return
	}

	rl.beginningOfHistory()
}

func (rl *Shell) endOfBufferOrHistory() {
	rl.undo.SkipSave()

	if rl.cursor.Pos() < rl.line.Len()-1 {
		rl.cursor.Set(rl.line.Len())
		return
	}

	rl.endOfHistory()
}

func (rl *Shell) beginningOfLineHist() {
	rl.undo.SkipSave()

	switch {
	// case rl.pos <= 0:
	// 	rl.beginningOfLine()
	default:
		rl.histories.Walk(1)
	}
}

func (rl *Shell) endOfLineHist() {
	rl.undo.SkipSave()

	switch {
	// case rl.cursor.Pos() < len(rl.line)-1:
	// 	rl.endOfLine()
	default:
		rl.histories.Walk(-1)
	}
}

func (rl *Shell) beginningHistorySearchBackward() {
	// rl.historySearchLine(false)
}

func (rl *Shell) beginningHistorySearchForward() {
	// rl.historySearchLine(true)
}

func (rl *Shell) acceptLineWith(infer, hold bool) {
	// Without multiline support, we always return the line.
	if rl.AcceptMultiline == nil {
		rl.display.AcceptLine()
		rl.histories.Accept(hold, infer, nil)
		return
	}

	// Ask the caller if the line should be accepted
	// as is, save the command line and accept it.
	if rl.AcceptMultiline(*rl.line) {
		rl.display.AcceptLine()
		rl.histories.Accept(hold, infer, nil)
		return
	}

	// If not, we should start editing another line,
	// and insert a newline where our cursor value is.
	// This has the nice advantage of being able to work
	// in multiline mode even in the middle of the buffer.
	rl.line.Insert(rl.cursor.Pos(), '\n')
	rl.cursor.Inc()
}
