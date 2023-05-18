//go:build windows
// +build windows

package core

import (
	"unsafe"

	"github.com/reeflective/readline/inputrc"
)

const (
	VK_CANCEL   = 0x03
	VK_BACK     = 0x08
	VK_TAB      = 0x09
	VK_RETURN   = 0x0D
	VK_SHIFT    = 0x10
	VK_CONTROL  = 0x11
	VK_MENU     = 0x12
	VK_ESCAPE   = 0x1B
	VK_LEFT     = 0x25
	VK_UP       = 0x26
	VK_RIGHT    = 0x27
	VK_DOWN     = 0x28
	VK_DELETE   = 0x2E
	VK_LSHIFT   = 0xA0
	VK_RSHIFT   = 0xA1
	VK_LCONTROL = 0xA2
	VK_RCONTROL = 0xA3
	VK_SNAPSHOT = 0x2C
	VK_INSERT   = 0x2D
	VK_HOME     = 0x24
	VK_END      = 0x23
	VK_PRIOR    = 0x21
	VK_NEXT     = 0x22
)

const (
	CharTab       = 9
	CharBackspace = 127
)

func init() {
	Stdin = NewRawReader()
}

// rawReader translates Windows input to ANSI sequences,
// to provide the same behavior as Unix terminals.
type rawReader struct {
	ctrlKey  bool
	altKey   bool
	shiftKey bool
}

// NewRawReader returns a new rawReader for Windows.
func NewRawReader() *rawReader {
	r := new(rawReader)
	return r
}

// Read reads input record from stdin on Windows.
// It keeps reading until it gets a key event.
func (r *rawReader) Read(buf []byte) (int, error) {
	ir := new(_INPUT_RECORD)
	var read int
	var err error

next:
	// ReadConsoleInputW reads input record from stdin.
	err = kernel.ReadConsoleInputW(stdin,
		uintptr(unsafe.Pointer(ir)),
		1,
		uintptr(unsafe.Pointer(&read)),
	)
	if err != nil {
		return 0, err
	}
	if ir.EventType != EVENT_KEY {
		goto next
	}

	// Reset modifiers if key is released.
	ker := (*_KEY_EVENT_RECORD)(unsafe.Pointer(&ir.Event[0]))
	if ker.bKeyDown == 0 { // keyup
		if r.ctrlKey || r.altKey || r.shiftKey {
			switch ker.wVirtualKeyCode {
			case VK_RCONTROL, VK_LCONTROL, VK_CONTROL:
				r.ctrlKey = false
			case VK_MENU: // alt
				r.altKey = false
			case VK_SHIFT, VK_LSHIFT, VK_RSHIFT:
				r.shiftKey = false
			}
		}
		goto next
	}

	// Keypad, special and arrow keys.
	if ker.unicodeChar == 0 {
		if modifiers, target := r.translateSeq(ker); target != 0 {
			return r.writeEsc(buf, append(modifiers, target)...)
		}
		goto next
	}

	char := rune(ker.unicodeChar)

	// Encode keys with modifiers.
	if r.ctrlKey {
		char = inputrc.Encontrol(char)
	} else if r.altKey {
		char = inputrc.Enmeta(char)
	} else if r.shiftKey && char == 9 { // shift + tab
		return r.writeEsc(buf, 91, 90)
	}

	// Else, the key is a normal character.
	return r.write(buf, char)
}

// Close is a stub to satisfy io.Closer.
func (r *rawReader) Close() error {
	return nil
}

func (r *rawReader) writeEsc(b []byte, char ...rune) (int, error) {
	b[0] = byte(inputrc.Esc)
	n := copy(b[1:], []byte(string(char)))
	return n + 1, nil
}

func (r *rawReader) write(b []byte, char ...rune) (int, error) {
	n := copy(b, []byte(string(char)))
	return n, nil
}

func (r *rawReader) translateSeq(ker *_KEY_EVENT_RECORD) (modifiers []rune, target rune) {
	// Encode keys with modifiers by default,
	// unless the modifier is pressed alone.
	modifiers = append(modifiers, 91)

	// Modifiers add a default sequence, which is the good sequence for arrow keys by default.
	// The first rune is this sequence might be modified below, if the target is a special key
	// but not an arrow key.
	switch ker.wVirtualKeyCode {
	case VK_RCONTROL, VK_LCONTROL, VK_CONTROL:
		r.ctrlKey = true
	case VK_MENU: // alt
		r.altKey = true
	case VK_SHIFT, VK_LSHIFT, VK_RSHIFT:
		r.shiftKey = true
	}

	switch {
	case r.ctrlKey:
		modifiers = append(modifiers, 49, 59, 53)
	case r.altKey:
		modifiers = append(modifiers, 49, 59, 51)
	case r.shiftKey:
		modifiers = append(modifiers, 49, 59, 50)
	}

	changeModifiers := func(swap rune, pos int) {
		if len(modifiers) > pos-1 && pos > 0 {
			modifiers[pos] = swap
		} else {
			modifiers = append(modifiers, swap)
		}
	}

	// Now we handle the target key.
	switch ker.wVirtualKeyCode {
	// Keypad & arrow keys
	case VK_LEFT:
		target = 68
	case VK_RIGHT:
		target = 67
	case VK_UP:
		target = 65
	case VK_DOWN:
		target = 66
	case VK_HOME:
		target = 72
	case VK_END:
		target = 70

	// Other special keys, with effects on modifiers.
	case VK_SNAPSHOT:
	case VK_INSERT:
		changeModifiers(50, 2)
		target = 126
	case VK_DELETE:
		changeModifiers(51, 2)
		target = 126
	case VK_PRIOR:
		changeModifiers(53, 2)
		target = 126
	case VK_NEXT:
		changeModifiers(54, 2)
		target = 126
	}

	return
}
