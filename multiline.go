package readline

import (
	"bytes"
	"os"
)

// initMultiline is ran once at the beginning of an instance start.
func (rl *Instance) initMultiline() (string, error) {
	r := []rune(rl.multisplit[0])
	rl.editorInput(r)
	rl.carriageReturn()
	if len(rl.multisplit) > 1 {
		rl.multisplit = rl.multisplit[1:]
	} else {
		rl.multisplit = []string{}
	}

	return string(rl.line), nil
}

// processMultiline handles line input/editing when the last entered key was a carriage return.
func (rl *Instance) processMultiline(r []rune, b []byte, i int) (done, ret bool, val string, err error) {
	rl.multiline = append(rl.multiline, b[:i]...)

	if i == len(b) {
		done = true

		return
	}

	if !rl.allowMultiline(rl.multiline) {
		rl.multiline = []byte{}
		done = true

		return
	}

	s := string(rl.multiline)
	rl.multisplit = rxMultiline.Split(s, -1)

	r = []rune(rl.multisplit[0])
	rl.modeViMode = vimInsert
	rl.editorInput(r)
	rl.carriageReturn()
	rl.multiline = []byte{}
	if len(rl.multisplit) > 1 {
		rl.multisplit = rl.multisplit[1:]
	} else {
		rl.multisplit = []string{}
	}

	ret = true
	val = string(rl.line)

	return
}

func (rl *Instance) allowMultiline(data []byte) bool {
	rl.clearHelpers()
	printf("\r\nWARNING: %d bytes of multiline data was dumped into the shell!", len(data))
	for {
		print("\r\nDo you wish to proceed (yes|no|preview)? [y/n/p] ")

		b := make([]byte, 1024)

		i, err := os.Stdin.Read(b)
		if err != nil {
			return false
		}

		s := string(b[:i])
		print(s)

		switch s {
		case "y", "Y":
			print("\r\n" + rl.mainPrompt)
			return true

		case "n", "N":
			print("\r\n" + rl.mainPrompt)
			return false

		case "p", "P":
			preview := string(bytes.ReplaceAll(data, []byte{'\r'}, []byte{'\r', '\n'}))
			if rl.SyntaxHighlighter != nil {
				preview = rl.SyntaxHighlighter([]rune(preview))
			}
			print("\r\n" + preview)

		default:
			print("\r\nInvalid response. Please answer `y` (yes), `n` (no) or `p` (preview)")
		}
	}
}

func isMultiline(r []rune) bool {
	for i := range r {
		if (r[i] == '\r' || r[i] == '\n') && i != len(r)-1 {
			return true
		}
	}
	return false
}
