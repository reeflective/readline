package core

import "github.com/reeflective/readline/internal/term"

const (
	// When the user enters a {, the shell will automatically insert { } (with a space in between).
	bracketAutocompleteNextPos = 2
)

// getTermWidth is used as a variable so that we can
// use specific terminal widths in our tests.
var getTermWidth = term.GetWidth
