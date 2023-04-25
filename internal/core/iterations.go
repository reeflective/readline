package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/reeflective/readline/internal/color"
)

// Iterations manages iterations for commands.
type Iterations struct {
	times   string // Stores iteration value
	active  bool   // Are we currently setting the iterations.
	pending bool   // Has the last command been an iteration one (vi-pending style)
}

// Add adds a string which might be a digit or a negative sign.
func (i *Iterations) Add(times string) {
	if times == "" {
		return
	}

	i.active = true
	i.pending = true

	if times == "-" || strings.HasPrefix(times, "-") {
		i.times = times
	} else {
		i.times += times
	}
}

// Get returns the number of iterations (possibly
// negative), and resets the iterations to 1.
func (i *Iterations) Get() int {
	times, err := strconv.Atoi(i.times)

	// Any invalid value is still one time.
	if err != nil && times == -1 {
		times = 1
	}

	// At least one iteration
	if times == 0 {
		times++
	}

	i.times = ""

	return times
}

// IsSet returns true if an iteration/numeric argument is active.
func (i *Iterations) IsSet() bool {
	return i.active
}

// IsPending returns true if the very last command executed was an
// iteration one. This is only meant for the main readline loop/run.
func (i *Iterations) IsPending() bool {
	return i.pending
}

// Reset resets the iterations (drops them).
func (i *Iterations) Reset() {
	i.times = ""
	i.active = false
	i.pending = false
}

// Reset resets the iterations if the last command was not one to set them.
// If the reset operated on active iterations, this function returns true.
func (i *Iterations) ResetPostCommand() (hint string) {
	if i.pending {
		hint = color.Dim + fmt.Sprintf("(arg: %s)", i.times)
	}

	if i.pending {
		i.pending = false

		return
	}

	i.active = false

	return
}
