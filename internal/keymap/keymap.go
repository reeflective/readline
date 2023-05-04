package keymap

import (
	"fmt"
	"sort"
	"strings"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/core"
)

var unescape = inputrc.Unescape

// Engine is used to manage the main and local keymaps for the shell.
type Engine struct {
	local    Mode
	main     Mode
	prefixed inputrc.Bind
	active   inputrc.Bind
	pending  []inputrc.Bind
	skip     bool
	isCaller bool

	keys       *core.Keys
	iterations *core.Iterations
	config     *inputrc.Config
	commands   map[string]func()
}

// NewEngine is a required constructor for the keymap modes manager.
// It initializes the keymaps to their defaults or configured values.
func NewEngine(keys *core.Keys, i *core.Iterations, opts ...inputrc.Option) (*Engine, *inputrc.Config) {
	modes := &Engine{
		main:       Emacs,
		keys:       keys,
		iterations: i,
		config:     inputrc.NewDefaultConfig(),
		commands:   make(map[string]func()),
	}

	// Load the inputrc configurations and set up related things.
	modes.ReloadConfig(opts...)

	return modes, modes.config
}

// Register adds commands to the list of available commands.
func (m *Engine) Register(commands map[string]func()) {
	for name, command := range commands {
		m.commands[name] = command
	}
}

// SetMain sets the main keymap of the shell.
func (m *Engine) SetMain(keymap Mode) {
	m.main = keymap
	m.UpdateCursor()
}

// Main returns the local keymap.
func (m *Engine) Main() Mode {
	return m.main
}

// SetLocal sets the local keymap of the shell.
func (m *Engine) SetLocal(keymap Mode) {
	m.local = keymap
	m.UpdateCursor()
}

// Local returns the local keymap.
func (m *Engine) Local() Mode {
	return m.local
}

// ResetLocal deactivates the local keymap of the shell.
func (m *Engine) ResetLocal() {
	m.local = ""
	m.UpdateCursor()
}

// UpdateCursor reprints the cursor corresponding to the current keymaps.
func (m *Engine) UpdateCursor() {
	switch m.local {
	case ViOpp:
		m.PrintCursor(ViOpp)
		return
	case Visual:
		m.PrintCursor(Visual)
		return
	}

	// But if not, we check for the global keymap
	switch m.main {
	case Emacs, EmacsStandard, EmacsMeta, EmacsCtrlX:
		m.PrintCursor(Emacs)
	case ViIns:
		m.PrintCursor(ViIns)
	case ViCmd:
		m.PrintCursor(ViCmd)
	}
}

// PendingCursor changes the cursor to pending mode,
// and returns a function to call once done with it.
func (m *Engine) PendingCursor() (restore func()) {
	m.PrintCursor(ViOpp)

	return func() {
		m.UpdateCursor()
	}
}

// IsEmacs returns true if the main keymap is one of the emacs modes.
func (m *Engine) IsEmacs() bool {
	switch m.main {
	case Emacs, EmacsStandard, EmacsMeta, EmacsCtrlX:
		return true
	default:
		return false
	}
}

// PrintBinds displays a list of currently bound commands (and their sequences)
// to the screen. If inputrcFormat is true, it displays it formatted such that
// the output can be reused in an .inputrc file.
func (m *Engine) PrintBinds(inputrcFormat bool) {
	var commands []string

	for command := range m.commands {
		commands = append(commands, command)
	}

	sort.Strings(commands)

	binds := m.config.Binds[string(m.Main())]

	// Make a list of all sequences bound to each command.
	allBinds := make(map[string][]string)

	for _, command := range commands {
		for key, bind := range binds {
			if bind.Action != command {
				continue
			}

			commandBinds := allBinds[command]
			commandBinds = append(commandBinds, inputrc.Escape(key))
			allBinds[command] = commandBinds
		}
	}

	if inputrcFormat {
		printBindsInputrc(commands, allBinds)
	} else {
		printBindsReadable(commands, allBinds)
	}
}

// ConvertMeta recursively searches for metafied keys in a sequence,
// and replaces them with an esc prefix and their unmeta equivalent.
func (m *Engine) ConvertMeta(keys []rune) string {
	if len(keys) == 0 {
		return string(keys)
	}

	converted := make([]rune, 0)

	for i := 0; i < len(keys); i++ {
		char := keys[i]

		if !inputrc.IsMeta(char) {
			converted = append(converted, char)
			continue
		}

		// Replace the key with esc prefix and add the demetafied key.
		converted = append(converted, inputrc.Esc)
		converted = append(converted, inputrc.Demeta(char))
	}

	return string(converted)
}

// InputIsTerminator returns true when current input keys are one of
// the configured or builtin "terminators", which can be configured
// in .inputrc with the isearch-terminators variable.
func (m *Engine) InputIsTerminator() bool {
	// Make a special list of binds with a unique command.
	terminators := []string{
		inputrc.Unescape(`\C-G`),
		inputrc.Unescape(`C-]`),
	}

	binds := make(map[string]inputrc.Bind)

	for _, sequence := range terminators {
		binds[sequence] = inputrc.Bind{Action: "abort", Macro: false}
	}

	bind, cmd, _ := m.matchKeymap(binds)

	return bind.Action == "abort" && cmd != nil
}

// Commands returns the map of all command functions available to the shell.
// This includes the builtin commands (emacs/Vim/history/completion/etc), as
// well as any functions added by the user through Keymap.Register().
// The keys of this map are the names of each corresponding command function.
func (m *Engine) Commands() map[string]func() {
	return m.commands
}

// ActiveCommand returns the sequence/command currently being ran.
func (m *Engine) ActiveCommand() inputrc.Bind {
	return m.active
}

// MatchMain incrementally attempts to match cached input keys against the local keymap.
// Returns the bind if matched, the corresponding command, and if we only matched by prefix.
func (m *Engine) MatchMain() (bind inputrc.Bind, command func(), prefix bool) {
	if m.main == "" {
		return
	}

	binds := m.config.Binds[string(m.main)]
	if len(binds) == 0 {
		return
	}

	// When we are in incremental-search mode (local keymap),
	// the set of valid commands is restricted: filter out the
	// binds down to those.
	if m.Local() == Isearch {
		binds = m.IsearchCommands(m.Main())
	}

	bind, command, prefix = m.matchKeymap(binds)

	// Adjusting for the ESC key: when convert-meta is enabled,
	// many binds will actually match ESC as a prefix. This makes
	// commands like vi-movement-mode unreachable, so if the bind
	// is vi-movement-mode, we return it to be ran regardless of
	// the other binds matching by prefix.
	if m.isEscapeKey() && !m.IsEmacs() && prefix {
		bind, command, prefix = m.handleEscape(true)
	}

	return
}

// MatchMain incrementally attempts to match cached input keys against the main keymap.
// Returns the bind if matched, the corresponding command, and if we only matched by prefix.
func (m *Engine) MatchLocal() (bind inputrc.Bind, command func(), prefix bool) {
	if m.local == "" {
		return
	}

	binds := m.config.Binds[string(m.local)]
	if len(binds) == 0 {
		return
	}

	bind, command, prefix = m.matchKeymap(binds)

	// Similarly to the MatchMain() function, give a special treatment to the escape key
	// (if it's alone): using escape in Viopp/menu-complete/isearch should cancel the
	// current mode, thus we return either a Vim movement-mode command, or nothing.
	if m.isEscapeKey() && (prefix || command == nil) {
		bind, command, prefix = m.handleEscape(false)
	}

	return
}

func (m *Engine) matchKeymap(binds map[string]inputrc.Bind) (bind inputrc.Bind, cmd func(), prefix bool) {
	var keys []byte

	for {
		// Read keys one by one, and abort once exhausted.
		key, empty := m.keys.Pop()
		if empty {
			m.keys.MarkMatched(keys...)
			return
		}

		keys = append(keys, key)

		// Find binds (actions/macros) matching by prefix or perfectly.
		match, prefixed := m.matchCommand(keys, binds)

		// If the current keys have no matches but the previous
		// matching process found a prefix, use it with the keys.
		if match.Action == "" && len(prefixed) == 0 {
			prefix = false
			cmd = m.resolveCommand(m.prefixed)
			m.active = m.prefixed
			m.prefixed = inputrc.Bind{}

			m.keys.MarkMatched(keys...)
			return
		}

		// Or several matches, in which case we read another key.
		if match.Action != "" && len(prefixed) > 0 {
			prefix = true
			m.prefixed = match

			continue
		}

		// Or no exact match and only prefixes
		if len(prefixed) > 0 {
			prefix = true
			continue
		}

		// Or an exact match, so drop any prefixed one.
		m.active = match
		m.prefixed = inputrc.Bind{}

		m.keys.MarkMatched(keys...)
		return match, m.resolveCommand(match), false
	}
}

func (m *Engine) matchCommand(keys []byte, binds map[string]inputrc.Bind) (inputrc.Bind, []inputrc.Bind) {
	var match inputrc.Bind
	var prefixed []inputrc.Bind

	for sequence, kbind := range binds {
		// When convert-meta is on, any meta-prefixed bind should
		// be stripped and replaced with an escape meta instead.
		if m.config.GetBool("convert-meta") {
			sequence = m.ConvertMeta([]rune(sequence))
		}

		// If the keys are a prefix of the bind, keep the bind
		if len(string(keys)) < len(sequence) && strings.HasPrefix(sequence, string(keys)) {
			prefixed = append(prefixed, kbind)
		}

		// Else if the match is perfect, keep the bind
		if string(keys) == sequence {
			// if inputrc.Unescape(string(keys)) == sequence {
			match = kbind
		}
	}

	return match, prefixed
}

func (m *Engine) resolveCommand(bind inputrc.Bind) func() {
	if bind.Macro {
		return nil
	}

	if bind.Action == "" {
		return nil
	}

	return m.commands[bind.Action]
}

func (m *Engine) isEscapeKey() bool {
	keys, empty := m.keys.PeekAll()
	if empty || len(keys) == 0 {
		return false
	}

	if len(keys) > 1 {
		return false
	}

	if keys[0] != inputrc.Esc {
		return false
	}

	return true
}

// handleEscape is used to override or change the matched command when the escape key has
// been pressed: it might exit completion/isearch menus, use the vi-movement-mode, etc.
func (m *Engine) handleEscape(main bool) (bind inputrc.Bind, cmd func(), pref bool) {
	switch {
	case m.prefixed.Action == "vi-movement-mode":
		// The vi-movement-mode command always has precedence over
		// other binds when we are currently using the main keymap.
		bind = m.prefixed
	case !main:
		// When using the local keymap, we simply drop any prefixed
		// or matched bind, so that the key will be matched against
		// the main keymap: between both, completion/isearch menus
		// will likely be cancelled.
		bind = inputrc.Bind{}
	}

	// Drop what needs to, and resolve the command.
	m.prefixed = inputrc.Bind{}

	if bind.Action != "" && !bind.Macro {
		cmd = m.resolveCommand(bind)
	}

	return
}

func printBindsReadable(commands []string, all map[string][]string) {
	for _, command := range commands {
		commandBinds := all[command]
		sort.Strings(commandBinds)

		switch {
		case len(commandBinds) == 0:
			fmt.Printf("%s is not bound to any keys\n", command)

		case len(commandBinds) > 5:
			var firstBinds []string

			for i := 0; i < 5; i++ {
				firstBinds = append(firstBinds, "\""+commandBinds[i]+"\"")
			}

			bindsStr := strings.Join(firstBinds, ", ")
			fmt.Printf("%s can be found on %s ...\n", command, bindsStr)

		default:
			var firstBinds []string

			for _, bind := range commandBinds {
				firstBinds = append(firstBinds, "\""+bind+"\"")
			}

			bindsStr := strings.Join(firstBinds, ", ")
			fmt.Printf("%s can be found on %s\n", command, bindsStr)
		}
	}
}

func printBindsInputrc(commands []string, all map[string][]string) {
	for _, command := range commands {
		commandBinds := all[command]
		sort.Strings(commandBinds)

		switch {
		case len(commandBinds) == 0:
			fmt.Printf("# %s (not bound)\n", command)
		default:
			for _, bind := range commandBinds {
				fmt.Printf("\"%s\": %s\n", bind, command)
			}
		}
	}
}
