package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/reeflective/readline/internal/color"
)

// Iterations manages iterations for commands.
type Iterations struct {
	times  string
	active bool // Are we currently setting the iterations.
	set    bool // Has the last command been an iteration one.
	reset  bool // Did a command reset the iterations (eg. vi-cmd-mode)
}

// Add adds a string which might be a digit or a negative sign.
func (i *Iterations) Add(times string) {
	if times == "" {
		return
	}

	i.active = true
	i.set = true

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
	i.reset = true

	return times
}

// IsSet returns true if an iteration/numeric argument is active.
func (i *Iterations) IsSet() bool {
	return i.active
}

// Reset resets the iterations (drops them).
func (i *Iterations) Reset() {
	i.times = ""
	i.active = false
	i.reset = true
	i.set = false
}

// Reset resets the iterations if the last command was not one to set them.
// If the reset operated on active iterations, this function returns true.
func (i *Iterations) ResetPostCommand() (hint string, wasActive bool) {
	if i.active {
		hint = color.Dim + fmt.Sprintf("(arg: %s)", i.times)
	}

	if i.set {
		i.set = false
		i.active = false

		return
	}

	wasActive = i.active || i.reset
	i.times = ""
	i.active = false
	i.reset = false

	return
}
