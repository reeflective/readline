package keymap

import (
	"strings"

	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/display"
	"github.com/xo/inputrc"
)

// Modes is used to manage the main and local keymaps for the shell.
type Modes struct {
	local    Mode
	main     Mode
	prefixed inputrc.Bind
	pending  []action

	keys       *core.Keys
	iterations *core.Iterations
	display    *display.Engine
	opts       *inputrc.Config
	commands   map[string]func()
}

// NewModes is a required constructor for the keymap modes manager.
// It initializes the keymaps to their defaults or configured values.
func NewModes(keys *core.Keys, i *core.Iterations, dis *display.Engine, opts *inputrc.Config) *Modes {
	modes := &Modes{
		main:       Emacs,
		keys:       keys,
		iterations: i,
		display:    dis,
		opts:       opts,
		commands:   make(map[string]func()),
	}

	switch modes.opts.GetString("editing-mode") {
	case "emacs":
		modes.main = Emacs
		modes.display.UpdateCursor(display.Emacs)
	case "vi":
		modes.main = ViIns
	}

	// Run the corresponding Vim mode widget to initialize.
	// rl.viInsertMode()

	return modes
}

// Register adds commands to the list of available commands.
func (m *Modes) Register(commands map[string]func()) {
	for name, command := range commands {
		m.commands[name] = command
	}
}

// SetMain sets the main keymap of the shell.
func (m *Modes) SetMain(keymap Mode) {
	m.main = keymap
	m.UpdateCursor()
}

// Main returns the local keymap.
func (m *Modes) Main() Mode {
	return m.main
}

// SetLocal sets the local keymap of the shell.
func (m *Modes) SetLocal(keymap Mode) {
	m.local = keymap
	m.UpdateCursor()
}

// Local returns the local keymap.
func (m *Modes) Local() Mode {
	return m.local
}

// ResetLocal deactivates the local keymap of the shell.
func (m *Modes) ResetLocal() {
	m.local = ""
	m.UpdateCursor()
}

// Match matches the current input keys against the command map.
// If main is true, the main keymap is used, otherwise we use the local one.
func (m *Modes) Match(main bool) (command func(), prefix bool) {
	var keymap Mode
	if main {
		keymap = m.main
	} else {
		keymap = m.local
	}

	keys, empty := m.keys.PeekAll()
	if empty {
		return
	}

	// Commands
	binds := m.opts.Binds[string(keymap)]
	if binds == nil {
		// Drop the key.
	}

	// Find binds matching by prefix or perfectly.
	match, prefixed := m.matchCommand(keys, binds)

	// If the current keys have no matches but the previous
	// matching process found a prefix, use it with the keys.
	if match.Action == "" && len(prefixed) == 0 {
		return m.resolveCommand(m.prefixed), false
	}

	// Or several matches, in which case we must read another key.
	if match.Action != "" && len(prefixed) > 0 {
		m.prefixed = match
		return nil, true
	}

	// Or no exact match and only prefixes
	if len(prefixed) > 0 {
		return nil, true
	}

	return m.resolveCommand(match), false
}

// UpdateCursor reprints the cursor corresponding to the current keymaps.
func (m *Modes) UpdateCursor() {
	switch m.local {
	case ViOpp:
		m.display.UpdateCursor(display.Viopp)
		return
	case Visual:
		m.display.UpdateCursor(display.Visual)
		return
	}

	// But if not, we check for the global keymap
	switch m.main {
	case Emacs, EmacsStandard, EmacsMeta, EmacsCtrlX:
		m.display.UpdateCursor(display.Emacs)
	case ViIns:
		m.display.UpdateCursor(display.Viins)
	case ViCmd:
		m.display.UpdateCursor(display.Vicmd)
	}
}

// IsEmacs returns true if the main keymap is one of the emacs modes.
func (m *Modes) IsEmacs() bool {
	switch m.main {
	case Emacs, EmacsStandard, EmacsMeta, EmacsCtrlX:
		return true
	default:
		return false
	}
}

func (m *Modes) matchCommand(keys []rune, binds map[string]inputrc.Bind) (inputrc.Bind, []inputrc.Bind) {
	var match inputrc.Bind
	var prefixed []inputrc.Bind

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

	return match, prefixed
}

func (m *Modes) resolveCommand(bind inputrc.Bind) func() {
	if bind.Macro {
		return nil
	}

	if bind.Action == "" {
		return nil
	}

	return m.commands[bind.Action]
}
