package core

import (
	"io"
	"testing"
)

type readStep struct {
	data []byte
	err  error
}

type stubReadCloser struct {
	steps []readStep
	index int
}

func (s *stubReadCloser) Read(p []byte) (int, error) {
	if s.index >= len(s.steps) {
		return 0, io.EOF
	}

	step := s.steps[s.index]
	s.index++

	copy(p, step.data)

	return len(step.data), step.err
}

func (s *stubReadCloser) Close() error {
	return nil
}

func TestKeysReadKeySkipsEmptyReads(t *testing.T) {
	originalStdin := Stdin
	Stdin = &stubReadCloser{
		steps: []readStep{
			{},
			{data: []byte("x")},
		},
	}
	t.Cleanup(func() {
		Stdin = originalStdin
	})

	keys := &Keys{}

	key, isAbort := keys.ReadKey()

	if isAbort {
		t.Fatal("expected ReadKey to continue after an empty read")
	}

	if key != 'x' {
		t.Fatalf("expected key %q, got %q", 'x', key)
	}
}

func TestKeysReadKeyReturnsAbortOnEOF(t *testing.T) {
	originalStdin := Stdin
	Stdin = &stubReadCloser{
		steps: []readStep{
			{err: io.EOF},
		},
	}
	t.Cleanup(func() {
		Stdin = originalStdin
	})

	keys := &Keys{}

	key, isAbort := keys.ReadKey()

	if !isAbort {
		t.Fatal("expected ReadKey to abort on EOF")
	}

	if key != 0 {
		t.Fatalf("expected zero key on EOF, got %q", key)
	}
}
