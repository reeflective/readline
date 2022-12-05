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

	//
	// Keymaps ------------------------------------------------------------------------------------

	// InputMode - The shell can be used in Vim editing mode, or Emacs (classic).
	InputMode InputMode

	main          keymapMode // The main/global keymap, partially overridden by any local keymap.
	mainKeymap    keyMap     // All keys mapped to the name of their corresponding widgets/actions.
	local         keymapMode // The local keymap is used when completing menus, using Vim operators, etc.
	localKeymap   keyMap     // All keys mapped to the name of their corresponding widgets/actions.
	specialKeymap keyMap     // A keymap that is matched using regexp, (for things like digit arguments, etc.)

	// The shell maintains a list of all its keymaps, so that users can modify them, or add some.
	keymaps map[keymapMode]keyMap

	//
	// Vim Operating Parameters -------------------------------------------------------------------

	viIteration      string
	viUndoHistory    []undoItem
	viUndoSkipAppend bool
	visualLine       bool       // Is the visual mode VISUAL_LINE
	mark             int        // Visual selection mark. -1 when unactive
	activeRegion     bool       // Is a current range region active ?
	registers        *registers // All memory text registers, can be consulted with Alt"

	pendingIterations string // Iterations specific to viopp mode. (2y2w => "2"w)
	pendingActions    []action

	// Input Line ---------------------------------------------------------------------------------

	// readline operating parameters
	keys  string // Contains all keys (input by user) currently being processed by the shell.
	line  []rune // This is the input line, with entered text: full line = mlnPrompt + line
	pos   int    // Cursor position in the entire line.
	posX  int    // Cursor position X
	fullX int    // X coordinate of the full input line, including the prompt if needed.
	posY  int    // Cursor position Y (if multiple lines span)
	fullY int    // Y offset to the end of input line.

	// Buffer received from host programms
	multilineBuffer []byte
	multilineSplit  []string
	skipStdinRead   bool

	// SyntaxHighlight is a helper function to provide syntax highlighting.
	// Once enabled, set to nil to disable again.
	SyntaxHighlighter func([]rune) string

	// PasswordMask is what character to hide password entry behind.
	// Once enabled, set to 0 (zero) to disable the mask again.
	PasswordMask rune

	//
	// Completion ---------------------------------------------------------------------------------

	// TabCompleter is a simple function that offers completion suggestions.
	// It takes the readline line ([]rune) and cursor pos.
	// Returns a prefix string, and several completion groups with their items and description
	// Asynchronously add/refresh completions
	TabCompleter      func([]rune, int, DelayedTabContext) (string, []CompletionGroup)
	delayedTabContext DelayedTabContext

	// SyntaxCompletion is used to autocomplete code syntax (like braces and
	// quotation marks). If you want to complete words or phrases then you might
	// be better off using the TabCompletion function.
	// SyntaxCompletion takes the line ([]rune) and cursor position, and returns
	// the new line and cursor position.
	SyntaxCompleter func([]rune, int) ([]rune, int)

	// Asynchronously highlight/process the input line
	DelayedSyntaxWorker func([]rune) []rune
	delayedSyntaxCount  int64

	// MaxTabCompletionRows is the maximum number of rows to display in the tab
	// completion grid.
	MaxTabCompleterRows int // = 4

	// tab completion operating parameters
	tcGroups []*CompletionGroup // All of our suggestions tree is in here
	tcPrefix string             // The current tab completion prefix  against which to build candidates

	modeTabCompletion    bool
	compConfirmWait      bool // When too many completions, we ask the user to confirm with another Tab keypress.
	tabCompletionSelect  bool // We may have completions printed, but no selected candidate yet
	tabCompletionReverse bool // Groups sometimes use this indicator to know how they should handle their index
	tcUsedY              int  // Comprehensive offset of the currently built completions

	// Candidate /  virtual completion string / etc
	currentComp  []rune // The currently selected item, not yet a real part of the input line.
	lineComp     []rune // Same as rl.line, but with the currentComp inserted.
	lineRemain   []rune // When we complete in the middle of a line, we cut and keep the remain.
	compAddSpace bool   // If true, any candidate inserted into the real line is done with an added space.

	//
	// Completion Search  (Normal & History) -----------------------------------------------------

	modeTabFind  bool           // This does not change, because we will search in all options, no matter the group
	tfLine       []rune         // The current search pattern entered
	modeAutoFind bool           // for when invoked via ^R or ^F outside of [tab]
	searchMode   FindMode       // Used for varying hints, and underlying functions called
	regexSearch  *regexp.Regexp // Holds the current search regex match
	mainHist     bool           // Which history stdin do we want
	histHint     []rune         // We store a hist hint, for dual history sources

	//
	// History -----------------------------------------------------------------------------------

	// TODO: Should we store histories in a list or map, with name/bindkey as keys of the map ?
	// mainHistory - current mapped to CtrlR by default, with rl.SetHistoryCtrlR()
	mainHistory  History
	mainHistName string
	// altHistory is an alternative history input, in case a console user would
	// like to have two different history flows. Mapped to CtrlE by default, with rl.SetHistoryCtrlE()
	altHistory  History
	altHistName string

	// HistoryAutoWrite defines whether items automatically get written to
	// history.
	// Enabled by default. Set to false to disable.
	HistoryAutoWrite bool // = true

	// history operating params
	lineBuf    string
	histPos    int
	histNavIdx int // Used for quick history navigation.

	//
	// Hints -------------------------------------------------------------------------------------

	// HintText is a helper function which displays hint text the prompt.
	// HintText takes the line input from the promt and the cursor position.
	// It returns the hint text to display.
	HintText func([]rune, int) []rune

	// HintColor any ANSI escape codes you wish to use for hint formatting. By
	// default this will just be blue.
	// NOTE: Probably not needed either
	HintFormatting string

	hintText []rune // The actual hint text
	hintY    int    // Offset to hints, if it spans multiple lines

	//
	// Other -------------------------------------------------------------------------------------

	// TempDirectory is the path to write temporary files when editing a line in
	// $EDITOR. This will default to os.TempDir()
	TempDirectory string

	// GetMultiLine is a callback to your host program. Since multiline support
	// is handled by the application rather than readline itself, this callback
	// is required when calling $EDITOR. However if this function is not set
	// then readline will just use the current line.
	GetMultiLine func([]rune) []rune

	EnableGetCursorPos bool

	// User-defined keypress events
	evtKeyPress map[string]func(string, []rune, int) *EventReturn

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

	// Input Editing
	rl.InputMode = Emacs
	rl.initLine()

	// Keymaps
	rl.setBaseKeymap()

	// Completion
	rl.MaxTabCompleterRows = 50

	// History
	rl.mainHistory = new(ExampleHistory) // In-memory history by default.
	rl.HistoryAutoWrite = true

	// Others
	rl.HintFormatting = seqFgBlue
	rl.evtKeyPress = make(map[string]func(string, []rune, int) *EventReturn)
	rl.TempDirectory = os.TempDir()

	// Registers
	rl.initRegisters()

	return rl
}
