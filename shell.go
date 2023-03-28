package readline

import (
	"os"

	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/display"
	"github.com/reeflective/readline/internal/editor"
	"github.com/reeflective/readline/internal/history"
	"github.com/reeflective/readline/internal/keymap"
	"github.com/reeflective/readline/internal/macro"
	"github.com/reeflective/readline/internal/ui"
	"github.com/xo/inputrc"
)

// Shell is used to encapsulate the parameter group and run time of any given
// readline instance so that you can reuse the readline API for multiple entry
// captures without having to repeatedly unload configuration.
type Shell struct {
	// Core editor
	keys       *core.Keys        // Keys read user input keys and manage them.
	line       *core.Line        // The input line buffer and its management methods.
	cursor     *core.Cursor      // The cursor and its medhods.
	undo       *core.LineHistory // Line undo/redo history.
	selection  *core.Selection   // The selection managees various visual/pending selections.
	iterations *core.Iterations  // Digit arguments for repeating commands.
	buffers    *editor.Buffers   // buffers (Vim registers) and methods use/manage/query them.

	// User interface
	opts      *inputrc.Config    // Contains all keymaps, binds and per-application settings.
	prompt    *ui.Prompt         // The prompt engine computes and renders prompt strings.
	hint      *ui.Hint           // Usage/hints for completion/isearch below the input line.
	completer *completion.Engine // Completions generation and display.
	histories *history.Sources   // All history sources and management methods.
	macros    *macro.Engine      // Record, use and display macros.
	display   *display.Engine    // Manages display refresh/update/clearing.

	// Others
	keymaps *keymap.Modes // Manages main/local keymaps and runs key matching.

	// User-provided functions
	//
	// AcceptMultiline enables the caller to decide if the shell should keep reading
	// for user input on a new line (therefore, with the secondary prompt), or if it
	// should return the current line at the end of the `rl.Readline()` call.
	// This function should return true if the line is deemed complete (thus asking
	// the shell to return from its Readline() loop), or false if the shell should
	// keep reading input.
	AcceptMultiline func(line []rune) (accept bool)

	// SyntaxHighlight is a helper function to provide syntax highlighting.
	// Once enabled, set to nil to disable again.
	SyntaxHighlighter func(line []rune) string

	// Completer is a function that produces completions.
	// It takes the readline line ([]rune) and cursor pos as parameters,
	// and returns completions with their associated metadata/settings.
	Completer func(line []rune, cursor int) Completions

	// SyntaxCompletion is used to autocomplete code syntax (like braces and
	// quotation marks). If you want to complete words or phrases then you might
	// be better off using the TabCompletion function.
	// SyntaxCompletion takes the line ([]rune) and cursor position, and returns
	// the new line and cursor position.
	SyntaxCompleter func(line []rune, cursor int) ([]rune, int)

	// HintText is a helper function which displays hint text the prompt.
	// HintText takes the line input from the promt and the cursor position.
	// It returns the hint text to display.
	HintText func(line []rune, cursor int) []rune
}

// NewShell returns a readline shell instance initialized with a default
// inputrc configuration and binds, and with an in-memory command history.
func NewShell() *Shell {
	shell := new(Shell)

	opts := shell.newInputConfig()

	// Core editor
	line := new(core.Line)
	cursor := core.NewCursor(line)
	keys := core.NewKeys()
	selection := core.NewSelection(line, cursor)
	iterations := new(core.Iterations)

	shell.keys = keys
	shell.line = line
	shell.cursor = cursor
	shell.selection = selection
	shell.undo = new(core.LineHistory)
	shell.buffers = editor.NewBuffers()
	shell.iterations = iterations

	// Keymaps and commands
	keymaps := keymap.NewModes(keys, iterations, opts)
	keymaps.Register(shell.standardWidgets())
	keymaps.Register(shell.viWidgets())
	keymaps.Register(shell.historyWidgets())
	keymaps.Register(shell.completionWidgets())

	shell.keymaps = keymaps

	// User interface
	hint := new(ui.Hint)
	prompt := ui.NewPrompt(keys, line, cursor, opts)
	macros := macro.NewEngine(keys, hint)
	completer := completion.NewEngine(keys, line, cursor, hint, keymaps, opts)
	history := history.NewSources(line, cursor, hint)
	display := display.NewEngine(selection, history, prompt, hint, completer, opts)

	shell.opts = opts
	shell.hint = hint
	shell.prompt = prompt
	shell.completer = completer
	shell.macros = macros
	shell.histories = history
	shell.display = display

	return shell
}

// init gathers all steps to perform at the beginning of readline loop.
func (rl *Shell) init() {
	// Some components need the last accepted line.
	rl.histories.Init()

	// Reset core editor components.
	rl.selection.Reset()
	rl.buffers.Reset()
	rl.undo.Reset()
	rl.keys.Flush()
	rl.cursor.ResetMark()
	rl.cursor.Set(0)
	rl.line.Set([]rune{}...) // TODO: Wrong; if line was inferred this resets it while it should not.

	// Reset/initialize user interface components.
	rl.hint.Reset()
	rl.completer.Reset(true, true)
	rl.display.Init(rl.SyntaxHighlighter)

	// Reset other components.
	rl.iterations.Reset()
}

func (rl *Shell) Prompt() *ui.Prompt {
	return rl.prompt
}

type History = history.Source

var inputrcConfigs = []string{
	"/etc/inputrc",
	os.Getenv("HOME") + "/.inputrc",
}

func (rl *Shell) newInputConfig() *inputrc.Config {
	opts := inputrc.NewDefaultConfig()

	// TODO: Before parsing user files, add default
	// binds for this readline library.

	// TODO: use parser instead of raw config and export its access.

	// Try to parse user/system inputrc.
	for _, path := range inputrcConfigs {
		if err := inputrc.ParseFile(path, opts); err != nil {
			println(err)
		}
	}

	return opts
}
