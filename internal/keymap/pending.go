package keymap

// action is represents the action of a widget, the number of times
// this widget needs to be run, and an optional operator argument.
// Most of the time we don't need this operator.
//
// Those actions are mostly used by widgets which make the shell enter
// the Vim operator pending mode, and thus require another key to be read.
type action struct {
	command    string
	iterations int
}

// AddPending registers a command as waiting for another command to run first,
// such as yank/delete/change actions, which accept/require a movement command.
func (m *Modes) AddPending(command string) {
	if command == "" {
		return
	}

	m.SetLocal(ViOpp)

	act := action{
		command:    command,
		iterations: m.iterations.Get(),
	}

	// Push the widget on the stack of widgets
	m.pending = append(m.pending, act)
}

// When the same widget is called twice in a row (like `yy` or `dd`),
// it will call this function to avoid running itself a third time.
func (m *Modes) ReleasePending(command string) {
	if len(m.pending) == 0 {
		return
	}

	// Just pop the widget without using it.
	if m.pending[0].command == command {
		// rl.getPendingWidget()
	}
}
