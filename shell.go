package readline

import (
	"os"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/display"
	"github.com/reeflective/readline/internal/editor"
	"github.com/reeflective/readline/internal/history"
	"github.com/reeflective/readline/internal/keymap"
	"github.com/reeflective/readline/internal/macro"
	"github.com/reeflective/readline/internal/ui"
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
	keymaps    *keymap.Modes     // Manages main/local keymaps and runs key matching.

	// User interface
	config    *inputrc.Config    // Contains all keymaps, binds and per-application settings.
	opts      []inputrc.Option   // Configuration parsing options (app/term/values, etc)
	prompt    *ui.Prompt         // The prompt engine computes and renders prompt strings.
	hint      *ui.Hint           // Usage/hints for completion/isearch below the input line.
	completer *completion.Engine // Completions generation and display.
	histories *history.Sources   // All history sources and management methods.
	macros    *macro.Engine      // Record, use and display macros.
	display   *display.Engine    // Manages display refresh/update/clearing.

	// User-provided functions

	// AcceptMultiline enables the caller to decide if the shell should keep reading
	// for user input on a new line (therefore, with the secondary prompt), or if it
	// should return the current line at the end of the `rl.Readline()` call.
	// This function should return true if the line is deemed complete (thus asking
	// the shell to return from its Readline() loop), or false if the shell should
	// keep reading input on a newline (thus, insert a newline and read).
	AcceptMultiline func(line []rune) (accept bool)

	// SyntaxHighlighter is a helper function to provide syntax highlighting.
	// Once enabled, set to nil to disable again.
	SyntaxHighlighter func(line []rune) string

	// Completer is a function that produces completions.
	// It takes the readline line ([]rune) and cursor pos as parameters,
	// and returns completions with their associated metadata/settings.
	Completer func(line []rune, cursor int) Completions

	// SyntaxCompletion is used to autocomplete code syntax (like braces and
	// quotation marks). If you want to complete words or phrases then you might
	// be better off using the Completer function.
	// SyntaxCompletion takes the line ([]rune) and cursor position, and returns
	// the new line and cursor position.
	SyntaxCompleter func(line []rune, cursor int) ([]rune, int)

	// HintText is a helper function which displays hint text below the line.
	// HintText takes the line input from the promt and the cursor position.
	// It returns the hint text to display.
	HintText func(line []rune, cursor int) []rune
}

// NewShell returns a readline shell instance initialized with a default
// inputrc configuration and binds, and with an in-memory command history.
// The constructor accepts an optional list of inputrc configuration options,
// which are used when parsing/loading and applying any inputrc configuration.
func NewShell(opts ...inputrc.Option) *Shell {
	shell := new(Shell)

	// Core editor
	keys := new(core.Keys)
	line := new(core.Line)
	cursor := core.NewCursor(line)
	undo := core.NewLineHistory(line, cursor)
	selection := core.NewSelection(line, cursor)
	iterations := new(core.Iterations)

	shell.keys = keys
	shell.line = line
	shell.cursor = cursor
	shell.selection = selection
	shell.undo = undo
	shell.buffers = editor.NewBuffers()
	shell.iterations = iterations

	// Keymaps and commands
	opts = append(opts, inputrc.WithTerm(os.Getenv("TERM")))

	keymaps, config := keymap.NewModes(keys, iterations, opts...)
	keymaps.Register(shell.standardCommands())
	keymaps.Register(shell.viCommands())
	keymaps.Register(shell.historyCommands())
	keymaps.Register(shell.completionCommands())

	shell.keymaps = keymaps
	shell.config = config
	shell.opts = opts

	// User interface
	hint := new(ui.Hint)
	prompt := ui.NewPrompt(keys, line, cursor, keymaps, config)
	macros := macro.NewEngine(keys, hint)
	completer := completion.NewEngine(keys, line, cursor, selection, hint, keymaps, config)
	history := history.NewSources(line, cursor, hint)
	display := display.NewEngine(keys, selection, history, prompt, hint, completer, config)

	completer.SetAutocompleter(shell.commandCompletion)

	shell.config = config
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
	// Reset core editor components.
	rl.selection.Reset()
	rl.buffers.Reset()
	rl.undo.Reset()
	rl.keys.Flush()
	rl.cursor.ResetMark()
	rl.cursor.Set(0)
	rl.line.Set([]rune{}...)
	rl.undo.Save()
	rl.iterations.Reset()

	// Some accept-* command must fetch a specific
	// line outright, or keep the accepted one.
	rl.histories.Init()

	// Reset/initialize user interface components.
	rl.hint.Reset()
	rl.completer.ResetForce()
	rl.display.Init(rl.SyntaxHighlighter)
}
