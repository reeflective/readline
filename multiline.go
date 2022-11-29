package readline

import (
	"bytes"
	"os"
	"regexp"
)

var rxMultiline = regexp.MustCompile(`[\r\n]+`)

// initMultiline is ran once at the beginning of an instance start.
func (rl *Instance) initMultiline() (string, error) {
	r := []rune(rl.multilineSplit[0])
	rl.inputEditor(r)
	rl.carriageReturn()
	if len(rl.multilineSplit) > 1 {
		rl.multilineSplit = rl.multilineSplit[1:]
	} else {
		rl.multilineSplit = []string{}
	}

	return string(rl.line), nil
}

// processMultiline handles line input/editing when the last entered key was a carriage return.
func (rl *Instance) processMultiline(r []rune, b []byte, i int) (done, ret bool, val string, err error) {
	rl.multilineBuffer = append(rl.multilineBuffer, b[:i]...)

	if i == len(b) {
		done = true

		return
	}

	if !rl.allowMultiline(rl.multilineBuffer) {
		rl.multilineBuffer = []byte{}
		done = true

		return
	}

	s := string(rl.multilineBuffer)
	rl.multilineSplit = rxMultiline.Split(s, -1)

	r = []rune(rl.multilineSplit[0])
	rl.modeViMode = vimInsert
	rl.inputEditor(r)
	rl.carriageReturn()
	rl.multilineBuffer = []byte{}
	if len(rl.multilineSplit) > 1 {
		rl.multilineSplit = rl.multilineSplit[1:]
	} else {
		rl.multilineSplit = []string{}
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

		// NOTE: Where should we print the prompt more properly ?
		switch s {
		case "y", "Y":
			print("\r\n" + rl.Prompt.primary)
			return true

		case "n", "N":
			print("\r\n" + rl.Prompt.primary)
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
