package readline

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// validRegisterKeys - All valid register IDs (keys) for read/write Vim registers.
var validRegisterKeys = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/-\""

// registers - Contains all memory registers resulting from delete/paste/search
// or other operations in the command line input.
type registers struct {
	unnamed            []rune            // Unnamed register, used by default
	num                map[int][]rune    // numbered registers (0-9)
	alpha              map[string][]rune // lettered registers ( a-z )
	ro                 map[string][]rune // read-only registers ( . % : )
	registerSelectWait bool              // The user wants to use a still unidentified register
	onRegister         bool              // We have identified the register, and acting on it.
	currentRegister    rune              // Any of the read/write registers ("/num/alpha)
	mutex              *sync.Mutex
}

func (rl *Instance) initRegisters() {
	rl.registers = &registers{
		num:   make(map[int][]rune, 10),
		alpha: make(map[string][]rune, 52),
		ro:    map[string][]rune{},
		mutex: &sync.Mutex{},
	}
}

// saveToRegister - Passing a function that will move around the line in the desired way, we get
// the number of Vim iterations and we save the resulting string to the appropriate buffer.
// It's the same as saveToRegisterTokenize, but without the need to generate tokenized &
// cursor-pos-actualized versions of the input line.
func (rl *Instance) saveToRegister(adjust int) {
	// Get the current cursor position and go the length specified.
	begin := rl.pos
	end := rl.pos
	end += adjust
	if end > len(rl.line)-1 {
		end = len(rl.line)
	} else if end < 0 {
		end = 0
	}

	var buffer []rune
	if end < begin {
		buffer = rl.line[end:begin]
	} else {
		buffer = rl.line[begin:end]
	}

	buf := string(buffer)

	rl.saveBufToRegister([]rune(buf))
}

// saveToRegisterTokenize - Passing a function that will move around the line in the desired way, we get
// the number of Vim iterations and we save the resulting string to the appropriate buffer. Because we
// need the cursor position to be really moved around between calls to the jumper, we also need the tokeniser.
func (rl *Instance) saveToRegisterTokenize(tokeniser tokeniser, jumper func(tokeniser) int, vii int) {
	// The register is going to have to heavily manipulate the cursor position.
	// Remember the original one first, for the end.
	beginPos := rl.pos

	// Get the current cursor position and go the length specified.
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(jumper(tokeniser))
	}
	end := rl.pos
	rl.pos = beginPos

	if end > len(rl.line)-1 {
		end = len(rl.line)
	} else if end < 0 {
		end = 0
	}

	var buffer []rune
	if end < rl.pos {
		buffer = rl.line[end:rl.pos]
	} else {
		buffer = rl.line[rl.pos:end]
	}

	buf := string(buffer)

	rl.saveBufToRegister([]rune(buf))
}

// saveBufToRegister - Instead of computing the buffer ourselves based on an adjust,
// let the caller pass directly this buffer, yet relying on the register system to
// determine which register will store the buffer.
func (rl *Instance) saveBufToRegister(buffer []rune) {
	buf := string(buffer)

	// When exiting this function the currently selected register is dropped,
	defer rl.registers.resetRegister()

	if len(buffer) == 0 || buf == "" {
		return
	}

	// Put the buffer in the appropriate registers.
	rl.registers.unnamed = []rune(buf)

	// If there is an active register, directly give it the buffer.
	// Check if its a numbered or lettered register, and put it in.
	if rl.registers.onRegister {
		num, err := strconv.Atoi(string(rl.registers.currentRegister))
		if err == nil && num < 10 {
			rl.registers.writeNumberedRegister(num, []rune(buf))
		} else if err != nil {
			rl.registers.writeAlphaRegister([]rune(buf))
		}
	} else {
		// Or, if no active register and if there is room on the numbered ones,
		rl.registers.writeNumberedRegister(-1, []rune(buf))
	}
}

// The user asked to paste a buffer onto the line, so we check from which register
// we are supposed to select the buffer, and return it to the caller for insertion.
func (rl *Instance) pasteFromRegister() (buffer []rune) {
	defer rl.registers.resetRegister()

	// If no actively selected register, return the unnamed buffer
	if !rl.registers.registerSelectWait && !rl.registers.onRegister {
		return rl.registers.unnamed
	}
	activeRegister := string(rl.registers.currentRegister)

	// Else find the active register, and return its content.
	num, err := strconv.Atoi(activeRegister)

	if err == nil {
		buf, found := rl.registers.num[num]
		if found {
			return buf
		}
		return
	}

	buf, found := rl.registers.alpha[activeRegister]
	if found {
		return buf
	}

	buf, found = rl.registers.ro[activeRegister]
	if found {
		return buf
	}

	return
}

// setActiveRegister - The user has typed "<regiserID>, and we don't know yet
// if we are about to copy to/from it, so we just set as active, so that when
// the action to perform on it will be asked, we know which one to use.
func (r *registers) setActiveRegister(reg rune) {
	defer func() {
		// We now have an active, identified register
		r.registerSelectWait = false
		r.onRegister = true
	}()

	// Numbered
	num, err := strconv.Atoi(string(reg))
	if err == nil && num < 10 {
		r.currentRegister = reg
		return
	}
	// Read-only
	if _, found := r.ro[string(reg)]; found {
		r.currentRegister = reg

		return
	}

	// Else, lettered
	r.currentRegister = reg
}

// writeNumberedRegister - Add a buffer to one of the numbered registers
// Pass a number above 10 to indicate we just push it on the num stack.
func (r *registers) writeNumberedRegisterAlt(idx int, buf []rune, push bool) {
	// No numbered register above 10
	if len(r.num) > 10 {
		return
	}
	// No push to the stack if we are already using 9
	var max int
	if push {
		for i := range r.num {
			if i > max && string(r.num[i]) != string(buf) {
				max = i
			}
		}
		if max < 9 {
			r.num[max+1] = buf
		}
	} else {
		// Add to the stack with the specified register
		r.num[idx] = buf
	}
}

// writeNumberedRegister - Add a buffer to one of the numbered registers
// Pass a number above 10 to indicate we just push it on the num stack.
func (r *registers) writeNumberedRegister(idx int, buf []rune) {
	// No numbered register above 10
	if idx > 9 {
		return
	}

	// Add to the stack with the specified register
	if idx != -1 {
		r.num[idx] = buf

		return
	}

	// No push to the stack if we are already using 9
	for i := len(r.num); i > 0; i-- {
		if i == 10 {
			i--
		}
		r.num[i] = append([]rune{}, r.num[i-1]...)
	}

	r.num[0] = append([]rune{}, buf...)
}

// writeAlphaRegister - Either adds a buffer to a new/existing letterd register,
// or appends to this new/existing register if the currently active register is
// the uppercase letter for this register.
func (r *registers) writeAlphaRegister(buffer []rune) {
	appendRegs := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	appended := false
	for _, char := range appendRegs {
		if char == r.currentRegister {
			reg := strings.ToLower(string(r.currentRegister))
			_, exists := r.alpha[reg]
			if exists {
				r.alpha[reg] = append(r.alpha[reg], buffer...)
			} else {
				r.alpha[reg] = buffer
			}
			appended = true
		}
	}
	if !appended {
		r.alpha[string(r.currentRegister)] = buffer
	}
}

// resetRegister - there is no currently active register anymore,
// nor we are currently setting one as active.
func (r *registers) resetRegister() {
	r.currentRegister = ' '
	r.registerSelectWait = false
	r.onRegister = false
}

// The user can show registers completions and insert, no matter the cursor position.
func (rl *Instance) completeRegisters() Completions {
	comps := Message(seqFgBlue + "-- registers --" + seqReset)

	unnamed := Completion{
		Value:   string(rl.registers.unnamed),
		Display: seqDim + "\"\"" + seqDimReset + " " + string(rl.registers.unnamed),
	}
	comps.values = append(comps.values, unnamed)

	comps.values = append(comps.values, rawValues(rl.completeNumRegs())...)
	comps.values = append(comps.values, rawValues(rl.completeAlphaRegs())...)
	comps = comps.NoSort()

	return comps
}

func (rl *Instance) completeNumRegs() []Completion {
	regs := make([]Completion, 0)
	tag := seqDim + "num ([0-9])" + seqReset

	var nums []int
	for reg := range rl.registers.num {
		nums = append(nums, reg)
	}

	sort.Ints(nums)

	for _, reg := range nums {
		buf := rl.registers.num[reg]
		comp := Completion{
			Tag:     tag,
			Value:   string(buf),
			Display: fmt.Sprintf("%s\"%d%s %s", seqDim, reg, seqDimReset, string(buf)),
		}

		regs = append(regs, comp)
	}

	return regs
}

func (rl *Instance) completeAlphaRegs() []Completion {
	regs := make([]Completion, 0)
	tag := seqDim + "alpha ([a-z], [A-Z])" + seqReset

	var lett []string
	for reg := range rl.registers.alpha {
		lett = append(lett, reg)
	}
	sort.Strings(lett)

	for _, reg := range lett {
		buf := rl.registers.alpha[reg]
		comp := Completion{
			Tag:     tag,
			Value:   string(buf),
			Display: fmt.Sprintf("%s\"%s%s %s", seqDim, reg, seqDimReset, string(buf)),
		}

		regs = append(regs, comp)
	}

	return regs
}
