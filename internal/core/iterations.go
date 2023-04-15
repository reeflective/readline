package core

import (
	"strconv"
	"strings"
)

// Iterations manages iterations for commands.
type Iterations struct {
	times  string
	active bool // Are we currently setting the iterations.
	set    bool // Has the last command been an iteration one.
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
	defer i.Reset()

	times, err := strconv.Atoi(i.times)

	// Any invalid value is still one time.
	if err != nil && times == -1 {
		times = 1
	}

	// At least one iteration
	if times == 0 {
		times++
	}

	return times
}

// IsSet returns true if an iteration/numeric argument is active.
func (i *Iterations) IsSet() bool {
	return i.active
}

// Reset resets the iterations if the last command was not one to set them.
func (i *Iterations) Reset() {
	if i.set {
		i.set = false
		return
	}

	i.times = ""
	i.active = false
}
