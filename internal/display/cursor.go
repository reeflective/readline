package display

// CursorStyle is the style of the cursor
// in a given input mode/submode.
type CursorStyle string

// String - Implements fmt.Stringer.
func (c CursorStyle) String() string {
	cursor, found := cursors[c]
	if !found {
		return string(cursorUserDefault)
	}
	return cursor
}

const (
	cursorBlock             CursorStyle = "block"
	cursorUnderline         CursorStyle = "underline"
	cursorBeam              CursorStyle = "beam"
	cursorBlinkingBlock     CursorStyle = "blinking-block"
	cursorBlinkingUnderline CursorStyle = "blinking-underline"
	cursorBlinkingBeam      CursorStyle = "blinking-beam"
	cursorUserDefault       CursorStyle = "default"
)

var cursors = map[CursorStyle]string{
	cursorBlock:             "\x1b[2 q",
	cursorUnderline:         "\x1b[4 q",
	cursorBeam:              "\x1b[6 q",
	cursorBlinkingBlock:     "\x1b[1 q",
	cursorBlinkingUnderline: "\x1b[3 q",
	cursorBlinkingBeam:      "\x1b[5 q",
	cursorUserDefault:       "\x1b[0 q",
}

type mode string

// private equivalent of the keymaps, used to update cursor.
const (
	Emacs  mode = "emacs"
	Viins  mode = "vi-insert"
	Vicmd  mode = "vi-command"
	Visual mode = "visual"
	Viopp  mode = "vi-opp"
)

var defaultCursors = map[mode]CursorStyle{
	Viins:  cursorBlinkingBeam,
	Vicmd:  cursorBlinkingBlock,
	Viopp:  cursorBlinkingUnderline,
	Visual: cursorBlock,
	Emacs:  cursorBlinkingBlock,
}

// UpdateCursor prints the cursor for the given keymap mode,
// either default value or the one specified in inputrc file.
func (e *Engine) UpdateCursor(editMode mode) {
	var cursor CursorStyle

	// Check for a configured cursor in .inputrc file.
	modeSet := e.opts.GetString(string(editMode))
	if modeSet != "" {
		cursor = defaultCursors[mode(modeSet)]
	}

	// If the configured one was invalid, use default one.
	if cursor == "" {
		cursor = defaultCursors[editMode]
	}

	print(cursors[cursor])
}
