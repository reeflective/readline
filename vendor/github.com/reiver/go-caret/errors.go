package caret

import (
	"errors"
)

var (
	errInternalError = errors.New("caret: Internal Error")
	errNilReceiver   = errors.New("caret: Nil Receiver")
	errRuneError     = errors.New("caret: Rune Error")
	errNilWriter     = errors.New("caret: Nil Writer")
)
