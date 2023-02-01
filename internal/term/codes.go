package term

// Terminal control sequences.
const (
	ClearLineAfter   = "\x1b[0K"
	ClearLineBefore  = "\x1b[1K"
	ClearLine        = "\x1b[2K"
	ClearScreenBelow = "\x1b[0J"
	ClearScreen      = "\x1b[2J" // Clears screen, preserving scroll buffer
	ClearDisplay     = "\x1b[3J" // Clears screen fully, wipes the scroll buffer
	CursorTopLeft    = "\x1b[H"

	getCursorPos     = "\x1b[6n" // response: "\x1b{Line};{Column}R"
	SaveCursorPos    = "\x1b7"
	RestoreCursorPos = "\x1b8"
	// SaveCursorPos    = "\x1b[s"
	// RestorCursorPos = "\x1b[u".
)
