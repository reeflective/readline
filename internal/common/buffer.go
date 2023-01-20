package common

// Buffers is a list of registers in which to put yanked/cut contents.
// These buffers technically are Vim registers with full functionality.
type Buffers struct{}

// SetActive sets the currently active register/buffer.
func (reg *Buffers) SetActive(register rune) {}

// Get returns the contents of a given register.
// If the rune is nil, it returns the value of the kill buffer (the " Vim register).
// If the register name is invalid, the function returns an empty rune slice.
func (reg *Buffers) Get(register rune) []rune { return nil }

// Write writes a slice to the currently active buffer, and/or to the kill one.
func (reg *Buffers) Write(content ...rune) {}

// WriteTo writes a slice directly to a target register.
// If the register name is invalid, nothing is written anywhere.
func (reg *Buffers) WriteTo(register rune, content ...rune) {}

// Complete returns the contents of all buffers as a structured list of completions.
func (reg *Buffers) Complete() {}
