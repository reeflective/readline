package term

// Terminal control sequences.
const (
	ClearLineAfter   = "\x1b[0k"
	ClearLineBefore  = "\x1b[1k"
	ClearLine        = "\x1b[2k"
	ClearScreenBelow = "\x1b[0J"
	ClearScreen      = "\x1b[2J" // Clears screen fully
	CursorTopLeft    = "\x1b[H"  // Clears screen and places cursor on top-left

	getCursorPos     = "\x1b[6n" // response: "\x1b{Line};{Column}R"
	SaveCursorPos    = "\x1b7"
	RestoreCursorPos = "\x1b8"
	// SaveCursorPos    = "\x1b[s"
	// RestorCursorPos = "\x1b[u".
)
