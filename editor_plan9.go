//go:build plan9
// +build plan9

package readline

import "errors"

func (rl *Instance) StartEditorWithBuffer(multiline []rune, filename string) ([]rune, error) {
	return rl.line, errors.New("Not currently supported on Plan 9")
}
