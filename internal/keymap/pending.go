package keymap

import "github.com/reeflective/readline/inputrc"

// action is represents the action of a widget, the number of times
// this widget needs to be run, and an optional operator argument.
// Most of the time we don't need this operator.
//
// Those actions are mostly used by widgets which make the shell enter
// the Vim operator pending mode, and thus require another key to be read.
type action struct {
	command inputrc.Bind
}

// AddPending registers a command as waiting for another command to run first,
// such as yank/delete/change actions, which accept/require a movement command.
func (m *Modes) Pending() {
	m.SetLocal(ViOpp)
	m.skip = true

	// Push the widget on the stack of widgets
	m.pending = append(m.pending, m.active)
}

// CancelPending is used by commands that have been registering themselves
// as waiting for a pending operator, but have actually been called twice
// in a row (eg. dd/yy in Vim mode). This removes those commands from queue.
func (m *Modes) CancelPending() {
	if len(m.pending) == 0 {
		return
	}

	m.pending = m.pending[:len(m.pending)-1]

	if len(m.pending) == 0 && m.Local() == ViOpp {
		m.SetLocal("")
	}
}

// IsPending returns true when invoked from within the command
// that also happens to be the next in line of pending commands.
func (m *Modes) IsPending() bool {
	if len(m.pending) == 0 {
		return false
	}

	return m.active.Action == m.pending[0].Action
}

// RunPending runs any command with pending execution.
func (m *Modes) RunPending() {
	if len(m.pending) == 0 {
		return
	}

	if m.skip {
		m.skip = false
		return
	}

	defer m.UpdateCursor()

	// Get the last registered action.
	pending := m.pending[len(m.pending)-1]
	m.pending = m.pending[:len(m.pending)-1]

	// The same command might be used twice in a row (dd/yy)
	if pending.Action == m.active.Action {
		m.isCaller = true
		defer func() { m.isCaller = false }()
	}

	if pending.Action == "" {
		return
	}

	// Resolve and run X times (iterations at pending time)
	command := m.resolveCommand(pending)

	// TODO: Handle returns from widgets.
	command()

	// And adapt the local keymap.
	if len(m.pending) == 0 && m.Local() == ViOpp {
		m.SetLocal("")
	}
}
