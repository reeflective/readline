package core

import (
	"errors"
	"io"
	"testing"

	"github.com/reeflective/readline/inputrc"
)

type stubReadCloser struct {
	read func([]byte) (int, error)
}

func (s stubReadCloser) Read(p []byte) (int, error) {
	return s.read(p)
}

func (s stubReadCloser) Close() error {
	return nil
}

func TestWaitAvailableKeysMarksEOF(t *testing.T) {
	original := Stdin
	Stdin = stubReadCloser{
		read: func([]byte) (int, error) {
			return 0, io.EOF
		},
	}
	defer func() { Stdin = original }()

	keys := &Keys{}
	WaitAvailableKeys(keys, &inputrc.Config{})

	if !keys.IsEOF() {
		t.Fatal("expected EOF to be recorded")
	}
	if err := keys.ReadError(); err != nil {
		t.Fatalf("expected no non-EOF read error, got %v", err)
	}
}

func TestWaitAvailableKeysRecordsNonEOFReadError(t *testing.T) {
	want := errors.New("read failed")

	original := Stdin
	Stdin = stubReadCloser{
		read: func([]byte) (int, error) {
			return 0, want
		},
	}
	defer func() { Stdin = original }()

	keys := &Keys{}
	WaitAvailableKeys(keys, &inputrc.Config{})

	if keys.IsEOF() {
		t.Fatal("expected non-EOF read failure not to be marked as EOF")
	}
	if err := keys.ReadError(); !errors.Is(err, want) {
		t.Fatalf("expected read error %v, got %v", want, err)
	}
}
