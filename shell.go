package readline

import (
	"os"
	"strings"

	"github.com/reeflective/readline/internal/common"
	"github.com/reeflective/readline/internal/completion"
	"github.com/reeflective/readline/internal/history"
	"github.com/reeflective/readline/internal/ui"
	"github.com/xo/inputrc"
)

// Shell is used to encapsulate the parameter group and run time of any given
// readline instance so that you can reuse the readline API for multiple entry
// captures without having to repeatedly unload configuration.
type Shell struct {
	// Core editor
	keys      *common.Keys        // Keys read user input keys and manage them.
	line      *common.Line        // The input line buffer and its management methods.
	cursor    *common.Cursor      // The cursor and its medhods.
	undo      *common.LineHistory // Line undo/redo history.
	selection *common.Selection   // The selection managees various visual/pending selections.
	buffers   *common.Buffers     // buffers (Vim registers) and methods use/manage/query them.

	// User interface
	opts      *inputrc.Config    // Contains all keymaps, binds and per-application settings.
	prompt    *ui.Prompt         // The prompt engine computes and renders prompt strings.
	hint      *ui.Hint           // Usage/hints for completion/isearch below the input line.
	completer *completion.Engine // Completions generation and display.
	histories *history.Sources   // All history sources and management methods.

	// Others
	keyMap     keymapMode // The current keymap mode ("emacs", "vi", "vi-command", etc).
	iterations string     // Numeric arguments to commands (eg. Vim iterations)
	prefixed   inputrc.Bind
	startAt    int

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

	// Core editor
	line := new(common.Line)
	cursor := common.NewCursor(line)
	keys := common.NewKeys()
	selection := common.NewSelection(line, cursor)

	shell.keys = keys
	shell.line = line
	shell.cursor = cursor
	shell.selection = selection
	shell.undo = new(common.LineHistory)
	shell.buffers = common.NewBuffers()

	// User interface
	opts := shell.newInputConfig()
	hint := new(ui.Hint)
	prompt := ui.NewPrompt(keys, line, cursor, opts)
	completer := completion.NewEngine(keys, line, cursor, hint, opts)

	shell.opts = opts
	shell.hint = hint
	shell.prompt = prompt
	shell.completer = completer

	// Others
	shell.histories = history.NewSources(line, cursor, hint)
	shell.initKeymap()

	return shell
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

// matchKeymap dispatches the keys to their commands.
func (rl *Shell) matchKeymap(km keymapMode) (cb widget, prefix bool) {
	keys, empty := rl.keys.PeekAll()
	if empty {
		return
	}

	// Commands
	binds := rl.opts.Binds[string(km)]
	if binds == nil {
		// Drop the key.
	}

	// Find binds matching by prefix or perfectly.
	match, prefixed := rl.matchCommand(keys, binds)

	// If the current keys have no matches but the previous
	// matching process found a prefix, use it with the keys.
	if match.Action == "" && len(prefixed) == 0 {
		return rl.resolveCommand(match), false
	}

	// Or several matches, in which case we must read another key.
	if match.Action != "" && len(prefixed) > 0 {
		rl.prefixed = match
		return nil, true
	}

	// Or no exact match and only prefixes
	if len(prefixed) > 0 {
		return nil, true
	}

	return rl.resolveCommand(match), false
}

func (rl *Shell) matchCommand(keys []rune, binds map[string]inputrc.Bind) (match inputrc.Bind, prefixed []inputrc.Bind) {
	for sequence, kbind := range binds {
		// If the keys are a prefix of the bind, keep the bind
		if len(keys) < len(sequence) && strings.HasPrefix(sequence, string(keys)) {
			prefixed = append(prefixed, kbind)
		}

		// Else if the match is perfect, keep the bind
		if string(keys) == sequence {
			match = kbind
		}
	}

	return
}

func (rl *Shell) resolveCommand(bind inputrc.Bind) widget {
	// If the bind is a macro, inject the keys back in our stack.
	if bind.Macro {
		return nil
	}

	if bind.Action == "" {
		return nil
	}

	// Standard widgets (all editing modes/styles)
	if wg, found := rl.standardWidgets()[bind.Action]; found && wg != nil {
		return wg
	}

	// Vim standard widgets don't return anything, wrap them in a simple call.
	if wg, found := rl.viWidgets()[bind.Action]; found && wg != nil {
		return wg
	}

	// History control widgets
	// if wg, found := rl.historyWidgets()[name]; found && wg != nil {
	// 	return wg
	// }

	// Completion & incremental search
	// if wg, found := rl.completionWidgets()[name]; found && wg != nil {
	// 	return wg
	// }

	return nil
}

func commandCallback(action string) EventCallback {
	return func(_ string, line []rune, pos int) EventReturn {
		event := EventReturn{
			Widget:  action,
			NewLine: line,
			NewPos:  pos,
		}

		return event
	}
}
