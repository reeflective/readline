package keymap

import (
	"strings"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/strutil"
)

// MatchLocal incrementally attempts to match cached input keys against the local keymap.
// Returns the bind if matched, the corresponding command, and if we only matched by prefix.
func MatchLocal(eng *Engine) (bind inputrc.Bind, command func(), prefix bool) {
	if eng.local == "" {
		return
	}

	binds := eng.getContextBinds(false)
	if len(binds) == 0 {
		return
	}

	bind, command, prefix = eng.dispatch(binds)

	// Similarly to the MatchMain() function, give a special treatment to the escape key
	// (if it's alone): using escape in Viopp/menu-complete/isearch should cancel the
	// current mode, thus we return either a Vim movement-mode command, or nothing.
	if eng.isEscapeKey() && (prefix || command == nil) {
		bind, command, prefix = eng.handleEscape(false)
	}

	return
}

// MatchMain incrementally attempts to match cached input keys against the local keymap.
// Returns the bind if matched, the corresponding command, and if we only matched by prefix.
func MatchMain(eng *Engine) (bind inputrc.Bind, command func(), prefix bool) {
	if eng.main == "" {
		return
	}

	// Get relevant binds in the current context, possibly
	// restrained to a subset when non/incrementally-searching.
	binds := eng.getContextBinds(true)
	if len(binds) == 0 {
		return
	}

	bind, command, prefix = eng.dispatch(binds)

	// Non-incremental search mode should always insert the keys
	// if they did not exactly match one of the valid commands.
	if eng.nonIncSearch && (command == nil || prefix) {
		bind = inputrc.Bind{Action: "self-insert"}
		eng.active = bind
		command = eng.resolve(bind)
		prefix = false
	}

	// Adjusting for the ESC key: when convert-meta is enabled,
	// many binds will actually match ESC as a prefix. This makes
	// commands like vi-movement-mode unreachable, so if the bind
	// is vi-movement-mode, we return it to be ran regardless of
	// the other binds matching by prefix.
	if eng.isEscapeKey() && !eng.IsEmacs() && prefix {
		bind, command, prefix = eng.handleEscape(true)
	}

	return bind, command, prefix
}

func (m *Engine) dispatch(binds map[string]inputrc.Bind) (bind inputrc.Bind, cmd func(), prefix bool) {
	var keys, matched []byte

	for {
		key, empty := core.PopKey(m.keys)
		if empty {
			break
		}

		keys = append(keys, key)

		// Find binds (actions/macros)
		// matching by prefix or perfectly.
		match, prefixed := m.match(keys, binds)

		// If the current keys have no matches but the previous
		// matching process found a prefix, use it with the keys.
		if match.Action == "" && len(prefixed) == 0 {
			bind, cmd, prefix = m.exact(m.prefixed)
			break
		}

		// From here, there is at least one bind matched, by prefix
		// or exactly, so the key we popped is considered matched.
		matched = append(matched, key)

		// Handle different cases where we had more than one match.
		switch {
		case match.Action != "" && len(prefixed) > 0:
			prefix = true
			m.prefixed = match

			continue

		case len(prefixed) > 0:
			prefix = true
			continue
		}

		// Or an exact match, so drop any prefixed one.
		bind, cmd, prefix = m.exact(match)

		break
	}

	// We're done matching input against binds.
	if prefix {
		// If we matched by prefix, whether or not we have an exact
		// match amongst those or not, we should keep the keys for the
		// next dispatch run.
		core.MatchedPrefix(m.keys, keys...)
	} else {
		// Or mark the keys that FOR SURE matched against a command.
		// But if there are keys that have been tried but which didn't
		// match the filtered list of previous matches, we feed them back.
		keys = keys[len(matched):]
		core.MatchedKeys(m.keys, matched, keys...)
	}

	return bind, cmd, prefix
}

func (m *Engine) match(keys []byte, binds map[string]inputrc.Bind) (inputrc.Bind, []inputrc.Bind) {
	var match inputrc.Bind
	var prefixed []inputrc.Bind

	for sequence, kbind := range binds {
		// When convert-meta is on, any meta-prefixed bind should
		// be stripped and replaced with an escape meta instead.
		if m.config.GetBool("convert-meta") {
			sequence = strutil.ConvertMeta([]rune(sequence))
		}

		// If the keys are a prefix of the bind, keep the bind
		if len(string(keys)) < len(sequence) && strings.HasPrefix(sequence, string(keys)) {
			prefixed = append(prefixed, kbind)
		}

		// Else if the match is perfect, keep the bind
		if string(keys) == sequence {
			match = kbind
		}
	}

	return match, prefixed
}

func (m *Engine) resolve(bind inputrc.Bind) func() {
	if bind.Macro {
		return nil
	}

	if bind.Action == "" {
		return nil
	}

	return m.commands[bind.Action]
}

func (m *Engine) exact(match inputrc.Bind) (inputrc.Bind, func(), bool) {
	m.prefixed = inputrc.Bind{}
	m.active = match

	return m.active, m.resolve(match), false
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
		cmd = m.resolve(bind)
	}

	// Drop the escape key in the stack
	if main {
		m.keys.Pop()
	}

	return
}

func (m *Engine) isEscapeKey() bool {
	keys := m.keys.Caller()
	if len(keys) == 0 {
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
