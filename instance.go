package readline

import (
	"os"
	"regexp"
	"sync"
)

// Instance is used to encapsulate the parameter group and run time of any given
// readline instance so that you can reuse the readline API for multiple entry
// captures without having to repeatedly unload configuration.
type Instance struct {
	// The prompt supports all oh-my-posh prompt types (primary/rprompt/secondary/transient/tooltip)
	// In addition, the shell offers some functions to refresh the prompt on demand, with varying
	// behavior options (refresh below a message, or in place, etc)
	Prompt *prompt

	// Configuration stores all keymaps, prompt styles and other completion/helper settings.
	config *config

	//
	// Keymaps ------------------------------------------------------------------------------------

	main    keymapMode             // The main/global keymap, partially overridden by any local keymap.
	local   keymapMode             // The local keymap is used when completing menus, using Vim operators, etc.
	widgets map[keymapMode]widgets // Widgets wrapped into EventCallbacks at bind time

	// widgetPrefixMatched is a widget that perfectly matched a given input key, but was also
	// found along other widgets matching the key only as prefix. This is used so that when reading
	// the next key, if no match is found, the key is used by this widget.
	widgetPrefixMatched EventCallback

	// interruptHandlers are all special handlers being called when the shell receives an interrupt
	// signal key, like CtrlC/CtrlD. These are not directly assigned in the various keymaps, and are
	// matched against input keys before any other keymap.
	interruptHandlers map[rune]func() error

	//
	// Vim Operating Parameters -------------------------------------------------------------------

	iterations        string     // Global iterations.
	negativeArg       bool       // Emacs supports negative iterations.
	registers         *registers // All memory text registers, can be consulted with Alt"
	registersComplete bool       // When the completer is for registers, used to reset
	isViopp           bool       // Keeps track of vi operator pending mode BEFORE trying to match the current key.
	pendingActions    []action   // Widgets that have registered themselves as waiting for another action to be ran.
	viinsEnterPos     int        // The index at which insert mode was entered

	// Input Line ---------------------------------------------------------------------------------

	// IsMultiline enables the caller to decide if the shell should keep reading for user input
	// on a new line (therefore, with the secondary prompt), or if it should return the current
	// line at the end of the `rl.Readline()` call.
	//
	// The `line` parameter is the entire, compounded buffer: for example, if you already returned
	// `false` with this function, the shell has 3 lines buffered, and one current. The `line` here
	// is the aggregate of the tree buffered, and the current.
	// As well, `Readline()` will return this same aggregate.
	IsMultiline func(line []rune) (accept bool)

	// EnableGetCursorPos will allow the shell to send a special sequence
	// to the the terminal to get the current cursor X and Y coordinates.
	// This is true by default, to enable smart completion estate use.
	EnableGetCursorPos bool

	// SyntaxHighlight is a helper function to provide syntax highlighting.
	// Once enabled, set to nil to disable again.
	SyntaxHighlighter func([]rune) string

	// PasswordMask is what character to hide password entry behind.
	// Once enabled, set to 0 (zero) to disable the mask again.
	PasswordMask rune

	// Buffer & line
	keys      string // Contains all keys (input by user) not yet consumed by the shell widgets.
	line      []rune // This is the input line, with entered text: full line = mlnPrompt + line
	accepted  bool   // Set by 'accept-line' widget, to notify return the line to the caller
	err       error  // Errors returned by interrupt signal handlers
	inferLine bool   // When a "accept-line-and-down-history" widget wants to immediately retrieve/use a line.

	// Buffer received from host programs
	multilineSplit []string
	skipStdinRead  bool

	// selection management
	visualLine bool         // Is the visual mode VISUAL_LINE
	marks      []*selection // Visual/surround selections areas, often highlighted.

	// Cursor
	pos   int // Cursor position in the entire line.
	hpos  int // The line on which the cursor is (differs from posY, which accounts for wraps)
	posX  int // Cursor position X
	posY  int // Cursor position Y (if multiple lines span)
	fullX int // X coordinate of the full input line, including the prompt if needed.
	fullY int // Y offset to the end of input line.

	//
	// Completion ---------------------------------------------------------------------------------

	// Completer is a function that produces completions.
	// It takes the readline line ([]rune) and cursor pos.
	// It return a type holding all completions and their associated settings.
	Completer func(line []rune, cursor int) Completions

	// SyntaxCompletion is used to autocomplete code syntax (like braces and
	// quotation marks). If you want to complete words or phrases then you might
	// be better off using the TabCompletion function.
	// SyntaxCompletion takes the line ([]rune) and cursor position, and returns
	// the new line and cursor position.
	SyntaxCompleter func(line []rune, cursor int) ([]rune, int)

	// Asynchronously highlight/process the input line
	DelayedSyntaxWorker func([]rune) []rune
	delayedSyntaxCount  int64

	// The current completer to use to produce completions: normal/history/registers
	// Used so that autocomplete can use the correct completer all along.
	completer func()

	// tab completion operating parameters
	tcGroups   []*comps       // All of our suggestions tree is in here
	tcPrefix   string         // The current tab completion prefix  against which to build candidates
	tcUsedY    int            // Comprehensive offset of the currently built completions
	comp       []rune         // The currently selected item, not yet a real part of the input line.
	compSuffix suffixMatcher  // The suffix matcher is kept for removal after actually inserting the candidate.
	compLine   []rune         // Same as rl.line, but with the currentComp inserted.
	tfLine     []rune         // The current search pattern entered
	tfPos      int            // Cursor position in the isearch buffer
	isearch    *regexp.Regexp // Holds the current search regex match

	//
	// History -----------------------------------------------------------------------------------

	// Current line undo/redo history.
	undoHistory    []undoItem
	undoPos        int
	isUndoing      bool
	undoSkipAppend bool

	// Past history
	histories        map[string]History // Sources of history lines
	historyNames     []string           // Names of histories stored in rl.histories
	historySourcePos int                // The index of the currently used history
	lineBuf          string             // The current line saved when we are on another history line
	histPos          int                // Index used for navigating the history lines with arrows/j/k
	histHint         []rune             // We store a hist hint, for dual history sources
	histSuggested    []rune             // The last matching history line matching the current input.

	//
	// Hints -------------------------------------------------------------------------------------

	// HintText is a helper function which displays hint text the prompt.
	// HintText takes the line input from the promt and the cursor position.
	// It returns the hint text to display.
	HintText func([]rune, int) []rune

	hint  []rune // The actual hint text
	hintY int    // Offset to hints, if it spans multiple lines

	//
	// Other -------------------------------------------------------------------------------------

	// TempDirectory is the path to write temporary files when
	// editing a line in $EDITOR. This will default to os.TempDir().
	TempDirectory string

	// concurency
	mutex sync.Mutex
}

// NewInstance is used to create a readline instance and initialise it with sane defaults.
func NewInstance() *Instance {
	rl := new(Instance)

	// Prompt
	rl.Prompt = &prompt{
		primary: "$ ",
	}
	rl.Prompt.compute(rl)

	// Keymaps and configuration
	rl.loadDefaultConfig()
	rl.bindWidgets()
	rl.loadInterruptHandlers()

	// Line
	rl.lineInit()
	rl.initRegisters()

	// History
	rl.historyNames = append(rl.historyNames, "local history")
	rl.histories = make(map[string]History)
	rl.histories["local history"] = new(MemoryHistory)

	// Others
	rl.TempDirectory = os.TempDir()
	rl.EnableGetCursorPos = true

	return rl
}
