package keymap

import (
	"github.com/reeflective/readline/internal/common"
	"github.com/reeflective/readline/internal/display"
	"github.com/xo/inputrc"
)

// Modes is used to manage the main and local keymaps for the shell.
type Modes struct {
	local Mode
	main  Mode

	keys    *common.Keys
	display *display.Engine
	opts    *inputrc.Config
}

// NewModes is a required constructor for the keymap modes manager.
// It initializes the keymaps to their defaults or configured values.
func NewModes(keys *common.Keys, dis *display.Engine, opts *inputrc.Config) *Modes {
	modes := &Modes{
		main:    Emacs,
		keys:    keys,
		display: dis,
		opts:    opts,
	}

	switch modes.opts.GetString("editing-mode") {
	case "emacs":
		modes.main = Emacs
		modes.display.UpdateCursor(display.Emacs)
	case "vi":
		modes.main = ViIns
		// rl.viInsertMode()
	}

	return modes
}

// Register adds commands to the list of available commands.
func (m *Modes) Register(commands map[string]func()) {
}

// SetMain sets the main keymap of the shell.
func (m *Modes) SetMain(keymap Mode) {
}

// SetLocal sets the local keymap of the shell.
func (m *Modes) SetLocal(keymap Mode) {
}

// ResetLocal empties the local keymap of the shell.
func (m *Modes) ResetLocal() {}

// Match matches the current input keys against the command map.
// If main is true, the main keymap is used, otherwise we use the local one.
func (m *Modes) Match(main bool) (command func(), prefixed bool) {
	return
}

// UpdateCursor reprints the cursor corresponding to the current keymaps.
func (m *Modes) UpdateCursor() {
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
