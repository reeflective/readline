package core

import (
	"strconv"
	"sync"
	"unicode"
)

var (
	// ValidRegisterKeys - All valid register IDs (keys) for read/write Vim registers.
	ValidRegisterKeys = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/-\""

	numRegisters   = 10
	alphaRegisters = 52
)

// Buffers is a list of registers in which to put yanked/cut contents.
// These buffers technically are Vim registers with full functionality.
type Buffers struct {
	kill     []rune          // Kill buffer/register, used by default
	num      map[int][]rune  // numbered registers (0-9)
	alpha    map[rune][]rune // lettered registers ( a-z )
	ro       map[rune][]rune // read-only registers ( . % : )
	waiting  bool            // The user wants to use a still unidentified register
	selected bool            // We have identified the register, and acting on it.
	active   rune            // Any of the read/write registers ("/num/alpha)
	mutex    *sync.Mutex
}

// NewBuffers is a required constructor to set up all the buffers/registers
// for the shell, because it contains maps that must be correctly initialized.
func NewBuffers() *Buffers {
	return &Buffers{
		num:   make(map[int][]rune, numRegisters),
		alpha: make(map[rune][]rune, alphaRegisters),
		ro:    map[rune][]rune{},
		mutex: &sync.Mutex{},
	}
}

// SetActive sets the currently active register/buffer.
// Valid values are letters (lower/upper), digits (1-9),
// or read-only buffers ( . % : ).
func (reg *Buffers) SetActive(register rune) {
	defer func() {
		// We now have an active, identified register
		reg.waiting = false
		reg.selected = true
	}()

	// Numbered
	num, err := strconv.Atoi(string(register))
	if err == nil && num < 10 {
		reg.active = register
		return
	}
	// Read-only
	if _, found := reg.ro[register]; found {
		reg.active = register

		return
	}

	// Else, lettered
	reg.active = register
}

// Get returns the contents of a given register.
// If the rune is nil (rune(0)), it returns the value of the kill buffer (the " Vim register).
// If the register name is invalid, the function returns an empty rune slice.
func (reg *Buffers) Get(register rune) []rune {
	if register == rune(0) {
		return reg.kill
	}

	num, err := strconv.Atoi(string(reg.active))
	if err == nil {
		return reg.num[num]
	}

	if buf, found := reg.alpha[reg.active]; found {
		return buf
	}

	if buf, found := reg.ro[reg.active]; found {
		return buf
	}

	return nil
}

// Active returns the contents of the active buffer/register (or the kill
// buffer if no active register is active), and resets the active register.
func (reg *Buffers) Active() []rune {
	defer reg.Reset()

	if !reg.waiting && !reg.selected {
		return reg.kill
	}

	return reg.Get(reg.active)
}

// Pop rotates the kill ring and returns the new top.
func (reg *Buffers) Pop() []rune {
	if len(reg.num) == 0 {
		return reg.kill
	}

	// Reassign the kill buffer and
	// pop the first numbered register.
	reg.kill = []rune(reg.num[0])
	delete(reg.num, 0)

	return reg.kill
}

// GetKill returns the contents of the kill buffer.
func (reg *Buffers) GetKill() []rune {
	return reg.kill
}

// Write writes a slice to the currently active buffer, and/or to the kill one.
// After the operation, the buffers are reset, eg. none is considered active.
func (reg *Buffers) Write(content ...rune) {
	buf := string(content)

	defer reg.Reset()

	if len(content) == 0 || buf == "" {
		return
	}

	reg.kill = []rune(buf)

	// Either write to the active register, or add to numbered ones.
	if reg.selected {
		reg.WriteTo(reg.active, []rune(buf)...)
	} else {
		reg.writeNum(-1, []rune(buf))
	}
}

// WriteTo writes a slice directly to a target register.
// If the register name is invalid, nothing is written anywhere.
func (reg *Buffers) WriteTo(register rune, content ...rune) {
	buf := string(content)

	if len(content) == 0 || buf == "" {
		return
	}

	// If number register.
	num, err := strconv.Atoi(string(register))
	if num > 0 && num < 10 && err != nil {
		reg.writeNum(num, []rune(buf))

		return
	}

	// If lettered register.
	if unicode.IsLetter(register) {
		reg.writeAlpha(register, []rune(buf))

		return
	}
}

// Reset forgets any active/pending buffer/register, but does not delete its contents.
func (reg *Buffers) Reset() {
	reg.active = ' '
	reg.waiting = false
	reg.selected = false
}

// Complete returns the contents of all buffers as a structured list of completions.
func (reg *Buffers) Complete() {}

func (reg *Buffers) writeNum(register int, buf []rune) {
	// No numbered register above 10
	if register > numRegisters-1 {
		return
	}

	// Add to the stack with the specified register
	if register != -1 {
		reg.num[register] = buf

		return
	}

	// No push to the stack if we are already using 9
	for i := len(reg.num); i > 0; i-- {
		if i == numRegisters {
			i--
		}

		reg.num[i] = append([]rune{}, reg.num[i-1]...)
	}

	reg.num[0] = append([]rune{}, buf...)
}

func (reg *Buffers) writeAlpha(register rune, buf []rune) {
	appendRegs := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	appended := false

	for _, char := range appendRegs {
		if char == register {
			register = unicode.ToLower(reg.active)
			_, exists := reg.alpha[register]

			if exists {
				reg.alpha[register] = append(reg.alpha[register], buf...)
			} else {
				reg.alpha[register] = buf
			}

			appended = true
		}
	}

	if !appended {
		reg.alpha[register] = buf
	}
}

// The user can show registers completions and insert, no matter the cursor position.
// func (rl *Instance) completeRegisters() Completions {
// 	comps := Message(seqFgBlue + "-- registers --" + seqReset)
//
// 	display := strings.ReplaceAll(string(rl.registers.kill), "\n", ``)
// 	unnamed := Completion{
// 		Value:   string(rl.registers.kill),
// 		Display: seqDim + "\"\"" + seqDimReset + " " + display,
// 	}
// 	comps.values = append(comps.values, unnamed)
//
// 	comps.values = append(comps.values, rawValues(rl.completeNumRegs())...)
// 	comps.values = append(comps.values, rawValues(rl.completeAlphaRegs())...)
// 	comps = comps.NoSort()
//
// 	return comps
// }
//
// func (rl *Instance) completeNumRegs() []Completion {
// 	regs := make([]Completion, 0)
// 	tag := seqDim + "num ([0-9])" + seqReset
//
// 	var nums []int
// 	for reg := range rl.registers.num {
// 		nums = append(nums, reg)
// 	}
//
// 	sort.Ints(nums)
//
// 	for _, reg := range nums {
// 		buf := rl.registers.num[reg]
// 		display := strings.ReplaceAll(string(buf), "\n", ``)
//
// 		comp := Completion{
// 			Tag:     tag,
// 			Value:   string(buf),
// 			Display: fmt.Sprintf("%s\"%d%s %s", seqDim, reg, seqDimReset, display),
// 		}
//
// 		regs = append(regs, comp)
// 	}
//
// 	return regs
// }
//
// func (rl *Instance) completeAlphaRegs() []Completion {
// 	regs := make([]Completion, 0)
// 	tag := seqDim + "alpha ([a-z], [A-Z])" + seqReset
//
// 	var lett []string
// 	for reg := range rl.registers.alpha {
// 		lett = append(lett, reg)
// 	}
// 	sort.Strings(lett)
//
// 	for _, reg := range lett {
// 		buf := rl.registers.alpha[reg]
// 		display := strings.ReplaceAll(string(buf), "\n", ``)
//
// 		comp := Completion{
// 			Tag:     tag,
// 			Value:   string(buf),
// 			Display: fmt.Sprintf("%s\"%s%s %s", seqDim, reg, seqDimReset, display),
// 		}
//
// 		regs = append(regs, comp)
// 	}
//
// 	return regs
// }
