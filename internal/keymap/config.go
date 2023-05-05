package keymap

import (
	"os"
	"os/user"

	"github.com/reeflective/readline/inputrc"
)

// ReloadConfig parses all valid .inputrc configurations and immediately
// updates/reloads all related settings (editing mode, variables behavior, etc.)
func (m *Engine) ReloadConfig(opts ...inputrc.Option) (err error) {
	// Builtin Go binds (in addition to default readline binds)
	m.loadBuiltinBinds()

	user, err := user.Current()

	// Parse library-specific configurations.
	//
	// This library implements various additional commands and keymaps.
	// Parse the configuration with a specific App name, ignoring errors.
	inputrc.UserDefault(user, m.config, inputrc.WithApp("go"))

	// Parse user configurations.
	//
	// Those default settings are the base options often needed
	// by /etc/inputrc on various Linux distros (for special keys).
	defaults := []inputrc.Option{
		inputrc.WithMode("emacs"),
		inputrc.WithTerm(os.Getenv("TERM")),
	}

	opts = append(defaults, opts...)

	// This will only overwrite binds that have been
	// set in those configs, and leave the default ones
	// (those just set above), so as to keep most of the
	// default functionality working out of the box.
	err = inputrc.UserDefault(user, m.config, opts...)
	if err != nil {
		return err
	}

	// Some configuration variables might have an
	// effect on our various keymaps and bindings.
	m.overrideBindsSpecial()

	defer m.UpdateCursor()

	// Startup editing mode
	switch m.config.GetString("editing-mode") {
	case "emacs":
		m.main = Emacs
	case "vi":
		m.main = ViIns
	}

	return nil
}

// loadBuiltinBinds adds additional command mappins that are not part
// of the standard C readline configuration: those binds therefore can
// reference commands or keymaps only implemented/used in this library.
func (m *Engine) loadBuiltinBinds() {
	// Emacs specials
	for seq, bind := range emacsKeys {
		m.config.Binds[string(Emacs)][seq] = bind
	}

	// Load default keymaps (main)
	for seq, bind := range vicmdKeys {
		m.config.Binds[string(ViCmd)][seq] = bind
		m.config.Binds[string(ViMove)][seq] = bind
		m.config.Binds[string(Vi)][seq] = bind
	}

	// Load default keymaps(local)
	m.config.Binds[string(Visual)] = visualKeys
	m.config.Binds[string(ViOpp)] = vioppKeys
	m.config.Binds[string(MenuSelect)] = menuselectKeys
	m.config.Binds[string(Isearch)] = isearchKeys

	// Default TTY binds
	for _, keymap := range m.config.Binds {
		keymap[inputrc.Unescape(`\C-C`)] = inputrc.Bind{Action: "abort"}
	}
}

// overrideBindsSpecial overwrites some binds as dictated by the configuration variables.
func (m *Engine) overrideBindsSpecial() {
	// Disable completion functions if required
	if m.config.GetBool("disable-completion") {
		for _, keymap := range m.config.Binds {
			for seq, bind := range keymap {
				switch bind.Action {
				case "complete", "menu-complete", "possible-completions":
					keymap[seq] = inputrc.Bind{Action: "self-insert"}
				}
			}
		}
	}
}