package display

import (
	"testing"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/core"
)

func TestInlineSuggestionAppliesOnlyAtLineEndWithPrefix(t *testing.T) {
	line := core.Line{}
	line.Set([]rune("sta")...)
	cursor := core.NewCursor(&line)
	cursor.Set(line.Len())

	engine := &Engine{line: &line, cursor: cursor}
	engine.SetInlineSuggestion("status")

	if !engine.inlineSuggestionApplies("sta") {
		t.Fatal("inline suggestion should apply when it extends the line at the cursor")
	}

	cursor.Set(1)
	if engine.inlineSuggestionApplies("sta") {
		t.Fatal("inline suggestion should not apply when the cursor is not at the end")
	}

	cursor.Set(line.Len())
	engine.SetInlineSuggestion("stop")
	if engine.inlineSuggestionApplies("sta") {
		t.Fatal("inline suggestion should not apply when it does not extend the current line")
	}
}

func TestCoordinatesLineUsesInlineSuggestion(t *testing.T) {
	line := core.Line{}
	line.Set([]rune("sta")...)
	cursor := core.NewCursor(&line)
	cursor.Set(line.Len())

	engine := &Engine{
		line:   &line,
		cursor: cursor,
		opts:   inputrc.NewConfig(),
	}
	engine.SetInlineSuggestion("status --verbose")

	if got := engine.coordinatesLine(true); string(*got) != "status --verbose" {
		t.Fatalf("coordinates line = %q, want inline suggestion", string(*got))
	}
}

func TestCoordinatesLinePrefersHistorySuggestion(t *testing.T) {
	line := core.Line{}
	line.Set([]rune("sta")...)
	cursor := core.NewCursor(&line)
	cursor.Set(line.Len())

	cfg := inputrc.NewConfig()
	_ = cfg.Set("history-autosuggest", true)

	engine := &Engine{
		line:   &line,
		cursor: cursor,
		opts:   cfg,
	}
	engine.suggested.Set([]rune("status-from-history")...)
	engine.SetInlineSuggestion("status-from-inline")

	if got := engine.coordinatesLine(true); string(*got) != "status-from-history" {
		t.Fatalf("coordinates line = %q, want history suggestion", string(*got))
	}
}
