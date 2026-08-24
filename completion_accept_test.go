package readline

import "testing"

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
