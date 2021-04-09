package readline

import (
	"sort"
	"strconv"

	"github.com/evilsocket/islazy/tui"
)

// registers - Contains all memory registers resulting from delete/paste/search
// or other operations in the command line input.
type registers struct {
	unnamed            []rune            // Unnamed register, used by default
	num                map[int][]rune    // numbered registers (0-9)
	alpha              map[string][]rune // lettered registers ( a-z )
	ro                 map[string][]rune // read-only registers ( . % : )
	registerSelectWait bool              // The user wants to use a still unidentified register
	onRegister         bool              // We have identified the register, and acting on it.
	currentRegister    rune              // Any of the numbered registers
}

func (rl *Instance) initRegisters() {
	rl.registers = &registers{
		num:   make(map[int][]rune, 10),
		alpha: make(map[string][]rune, 52),
		ro:    map[string][]rune{},
	}
}

// saveToRegister - Passing a function that will move around the line in the desired way, we get
// the number of Vim iterations adn we save the resulting string to the appropriate buffer.
func (rl *Instance) saveToRegister(adjust int) {

	// When exiting this function the currently selected register is dropped,
	defer rl.registers.resetRegister()

	// Get the current cursor position and go the length specified.
	begin := rl.pos
	vii := rl.getViIterations()
	for i := 1; i <= vii; i++ {
		rl.moveCursorByAdjust(adjust)
	}
	end := rl.pos

	// Get a copy of the text subset, and immediately replace cursor
	buffer := rl.line[begin:end]
	rl.pos = begin

	// Put the buffer in the appropriate registers.
	// By default, always in the unnamed one first.
	rl.registers.unnamed = buffer

	// Or additionally on a specific one.
	// Check if its a numbered of lettered register, and put it in.
	if rl.registers.registerSelectWait {
		num, err := strconv.Atoi(string(rl.registers.currentRegister))
		if err == nil && num < 10 {
			rl.registers.writeNumberedRegister(num, buffer)
		}
		if err != nil {
			rl.registers.alpha[string(rl.registers.currentRegister)] = buffer
		}
	}
}

// saveBufToRegister - Instead of computing the buffer ourselves based on an adjust,
// let the caller pass directly this buffer, yet relying on the register system to
// determine which register will store the buffer.
func (rl *Instance) saveBufToRegister(buffer []rune) {
	// When exiting this function the currently selected register is dropped,
	defer rl.registers.resetRegister()

	// Put the buffer in the appropriate registers.
	// By default, always in the unnamed one first.
	rl.registers.unnamed = buffer

	// Or additionally on a specific one.
	// Check if its a numbered of lettered register, and put it in.
	if rl.registers.registerSelectWait {
		num, err := strconv.Atoi(string(rl.registers.currentRegister))
		if err == nil && num < 10 {
			rl.registers.writeNumberedRegister(num, buffer)
		}
		if err != nil {
			rl.registers.alpha[string(rl.registers.currentRegister)] = buffer
		}
	}
}

// The user asked to paste a buffer onto the line, so we check from which register
// we are supposed to select the buffer, and return it to the caller for insertion.
func (rl *Instance) pasteFromRegister() (buffer []rune) {

	// When exiting this function the currently selected register is dropped,
	defer rl.registers.resetRegister()

	// If no actively selected register, return the unnamed buffer
	if !rl.registers.registerSelectWait {
		return rl.registers.unnamed
	}
	activeRegister := string(rl.registers.currentRegister)

	// Else find the active register, and return its content.
	num, err := strconv.Atoi(activeRegister)

	// Either from the numbered ones.
	if err == nil {
		buf, found := rl.registers.num[num]
		if found {
			return buf
		}
		return
	}
	// or the lettered ones
	buf, found := rl.registers.alpha[activeRegister]
	if found {
		return buf
	}
	// Or the read-only ones
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
	// Numbered
	num, err := strconv.Atoi(string(reg))
	if err == nil && num < 10 {
		r.currentRegister = reg
		return
	}
	// Read-only
	_, found := r.ro[string(reg)]
	if found {
		r.currentRegister = reg
		return
	}

	// Else, lettered
	r.currentRegister = reg

	// We now have an active, identified register
	r.registerSelectWait = false
	r.onRegister = true
}

func (r *registers) resetRegister() {
	r.currentRegister = ' '
	r.registerSelectWait = false
	r.onRegister = false
}

// writeNumberedRegister - Add a buffer to one of the numbered registers
func (r *registers) writeNumberedRegister(idx int, buf []rune) {
	if len(r.num) > 10 {
		return
	}
	r.num[idx] = buf
}

// The user can show registers completions and insert, no matter the cursor position.
func (rl *Instance) completeRegisters() []*CompletionGroup {

	// We set the hint exceptionally
	hint := YELLOW + " :registers" + RESET
	rl.hintText = []rune(hint)

	// Make the groups
	regs := &CompletionGroup{
		Name:         tui.DIM + "([0-9], [a-z], [A-Z])" + tui.RESET,
		DisplayType:  TabDisplayMap,
		MaxLength:    20,
		Descriptions: map[string]string{},
	}

	// Unnamed
	regs.Suggestions = append(regs.Suggestions, "\"")
	regs.Descriptions["\""] = string(rl.registers.unnamed)

	// Numbered registers
	var nums []int
	for reg := range rl.registers.num {
		nums = append(nums, reg)
	}
	sort.Ints(nums)
	for _, reg := range nums {
		buf := rl.registers.num[reg]
		regs.Suggestions = append(regs.Suggestions, string(buf))
		regs.Descriptions[string(buf)] = "\033[38;5;237m" + strconv.Itoa(reg) + RESET
	}

	// Letter registers
	var lett []string
	for reg := range rl.registers.alpha {
		lett = append(lett, reg)
	}
	sort.Strings(lett)
	for _, reg := range lett {
		buf := rl.registers.alpha[reg]
		regs.Suggestions = append(regs.Suggestions, string(buf))
		regs.Descriptions[string(buf)] = "\033[38;5;237m" + reg + RESET
	}

	return []*CompletionGroup{regs}
}
