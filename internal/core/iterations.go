package core

import (
	"strconv"
	"strings"
)

// Iterations manages iterations for commands.
type Iterations string

// Add adds a string which might be a digit or a negative sign.
func (i *Iterations) Add(times string) {
	if times == "" {
		return
	}

	if times == "-" || strings.HasPrefix(times, "-") {
		*i = Iterations(times)
	} else {
		*i = Iterations(string(*i) + times)
	}
}

// Get returns the number of iterations (possibly
// negative), and resets the iterations to 1.
func (i *Iterations) Get() int {
	defer i.Reset()

	times, err := strconv.Atoi(string(*i))

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

// Reset resets the iterations.
func (i *Iterations) Reset() {
	*i = ""
}
