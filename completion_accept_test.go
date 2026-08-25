package readline

import (
	"testing"
	"time"
)

func TestResetNotifiesSelectedPublicCompletion(t *testing.T) {
	shell := NewShell()
	accepted := ""
	shell.Completer = func(_ []rune, _ int) Completions {
		result := CompleteRaw([]Completion{
			{Value: "sha256", Display: "sha256", Description: "hmac", Tag: "function", OnAccept: func(_ []rune, _ int) ([]rune, int) {
				accepted = "hmac"
				return []rune("import hmac"), len([]rune("import hmac"))
			}},
			{Value: "sha256", Display: "sha256", Description: "hash", Tag: "function", OnAccept: func(_ []rune, _ int) ([]rune, int) {
				accepted = "hash"
				return []rune("import hash"), len([]rune("import hash"))
			}},
		}).JustifyDescriptions()
		result.PREFIX = "sha"
		return result
	}
	if err := shell.Config.Set("menu-complete-display-prefix", true); err != nil {
		t.Fatal(err)
	}
	shell.Line().Set([]rune("sha")...)
	shell.Cursor().Set(3)

	shell.completeWord()
	shell.completeWord()
	shell.completer.Reset()

	if accepted == "" {
		t.Fatal("selected completion did not call OnAccept")
	}
	if got := string(*shell.Line()); got != "import "+accepted {
		t.Fatalf("line = %q, want transformed %s completion", got, accepted)
	}
}

func TestConfirmationRequiredCompletionCanBeCancelledOrCommittedWithoutSubmit(t *testing.T) {
	shell := NewShell()
	accepted := false
	shell.Completer = func(_ []rune, _ int) Completions {
		result := CompleteRaw([]Completion{{
			Value: "math", Display: "math", Description: "trb/std/math", Tag: "module",
			RequireConfirmation: true,
			OnAccept: func(_ []rune, _ int) ([]rune, int) {
				accepted = true
				line := []rune("import trb/std/math")
				return line, len(line)
			},
		}})
		result.PREFIX = "mat"
		return result
	}
	if err := shell.Config.Set("menu-complete-display-prefix", true); err != nil {
		t.Fatal(err)
	}
	shell.Line().Set([]rune("mat")...)
	shell.Cursor().Set(3)

	shell.completeWord()
	if got := string(*shell.Line()); got != "mat" {
		t.Fatalf("real line after completion = %q, want original input", got)
	}
	if !shell.completer.IsInserting() {
		t.Fatal("confirmation-required completion was not kept virtual")
	}
	shell.abort()
	if got := string(*shell.Line()); got != "mat" {
		t.Fatalf("line after cancellation = %q, want original input", got)
	}
	if accepted {
		t.Fatal("cancelled completion called OnAccept")
	}

	shell.completeWord()
	shell.acceptLine()
	if !accepted {
		t.Fatal("confirmed completion did not call OnAccept")
	}
	if got := string(*shell.Line()); got != "import trb/std/math" {
		t.Fatalf("line after confirmation = %q, want import", got)
	}
	if lineAccepted, _, _ := shell.History.LineAccepted(); lineAccepted {
		t.Fatal("confirmation submitted the line")
	}

	shell.acceptLine()
	lineAccepted, line, err := shell.History.LineAccepted()
	if err != nil {
		t.Fatal(err)
	}
	if !lineAccepted || line != "import trb/std/math" {
		t.Fatalf("second accept = (%v, %q), want submitted import", lineAccepted, line)
	}
}

func TestConfirmationRequiredCompletionConsumesFirstEnter(t *testing.T) {
	shell := confirmationTestShell(t)
	accepted := make(chan struct{}, 1)
	shell.Completer = func(_ []rune, _ int) Completions {
		result := CompleteRaw([]Completion{{
			Value: "math", Display: "math", Description: "trb/std/math", Tag: "module",
			RequireConfirmation: true,
			OnAccept: func(_ []rune, _ int) ([]rune, int) {
				accepted <- struct{}{}
				line := []rune("import trb/std/math")
				return line, len(line)
			},
		}})
		result.PREFIX = "mat"
		return result
	}

	shell.Keys.Feed(false, []rune("mat\t\r")...)
	result := make(chan string, 1)
	go func() {
		line, _ := shell.Readline()
		result <- line
	}()

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("first Enter did not confirm the completion")
	}
	select {
	case line := <-result:
		t.Fatalf("first Enter submitted %q instead of keeping it editable", line)
	case <-time.After(50 * time.Millisecond):
	}

	shell.Keys.Feed(false, '\r')
	shell.Keys.RequestRefresh()
	select {
	case line := <-result:
		if line != "import trb/std/math" {
			t.Fatalf("second Enter submitted %q, want import", line)
		}
	case <-time.After(time.Second):
		t.Fatal("second Enter did not submit the confirmed import")
	}
}

func TestConfirmationRequiredCompletionBackspaceCancelsSelection(t *testing.T) {
	shell := confirmationTestShell(t)
	shell.Completer = func(_ []rune, _ int) Completions {
		result := CompleteRaw([]Completion{{
			Value: "math", Display: "math", Description: "trb/std/math", Tag: "module",
			RequireConfirmation: true,
			OnAccept: func(_ []rune, _ int) ([]rune, int) {
				t.Fatal("cancelled completion called OnAccept")
				return nil, 0
			},
		}})
		result.PREFIX = "mat"
		return result
	}

	shell.Keys.Feed(false, []rune("mat\t\x7f\r")...)
	line, err := shell.Readline()
	if err != nil {
		t.Fatal(err)
	}
	if line != "mat" {
		t.Fatalf("line after Backspace cancellation = %q, want original input", line)
	}
}

func confirmationTestShell(t *testing.T) *Shell {
	t.Helper()
	shell := NewShell()
	for name, value := range map[string]bool{
		"cursor-position-probe":        false,
		"enable-bracketed-paste":       false,
		"menu-complete-display-prefix": true,
	} {
		if err := shell.Config.Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	return shell
}
