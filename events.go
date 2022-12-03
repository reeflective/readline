package readline

// EventReturn is a structure returned by the callback event function.
// This is used by readline to determine what state the API should
// return to after the readline event.
type EventReturn struct {
	ForwardKey    bool
	ClearHelpers  bool
	CloseReadline bool
	HintText      []rune
	NewLine       []rune
	NewPos        int
}

// AddEvent registers a new keypress handler
func (rl *Instance) AddEvent(keyPress string, callback func(string, []rune, int) *EventReturn) {
	rl.evtKeyPress[keyPress] = callback
}

// DelEvent deregisters an existing keypress handler
func (rl *Instance) DelEvent(keyPress string) {
	delete(rl.evtKeyPress, keyPress)
}

// TODO: This should be either removed or refactored into the new model.
// handleKeyPress is in charge of executing the handler that is register for a given keypress.
func (rl *Instance) handleKeyPress(s string) (done, mustReturn bool, val string, err error) {
	rl.clearHelpers()

	ret := rl.evtKeyPress[s](s, rl.line, rl.pos)

	rl.clearLine()
	rl.line = append(ret.NewLine, []rune{}...)
	rl.updateHelpers() // rl.echo
	rl.pos = ret.NewPos

	if ret.ClearHelpers {
		rl.resetHelpers()
	} else {
		rl.updateHelpers()
	}

	if len(ret.HintText) > 0 {
		rl.hintText = ret.HintText
		rl.clearHelpers()
		rl.renderHelpers()
	}
	if !ret.ForwardKey {
		done = true

		return
	}

	if ret.CloseReadline {
		rl.clearHelpers()
		mustReturn = true
		val = string(rl.line)

		return
	}

	return
}
