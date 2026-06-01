package ui

import (
	"strings"
	"sync"

	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/core"
	"github.com/reeflective/readline/internal/strutil"
	"github.com/reeflective/readline/internal/term"
)

// Hint is in charge of printing the usage messages below the input line.
// Various other UI components have access to it so that they can feed
// specialized usage messages to it, like completions.
//
// The hint area is composed of several independent lanes, rendered top to
// bottom in the following order:
//
//   - persistent: status set by internal commands (Vim registers, iterations).
//   - provided:   passive hint computed from the current input line by a
//     user-registered hint provider (see Shell.SetHintProvider).
//   - transient:  one-shot status messages, typically pushed from an
//     asynchronous producer (see Shell.SetTransientHint). These survive
//     keystrokes that drive the completion/isearch lane.
//   - text:       hints owned by the completion and incremental-search engines,
//     plus temporary command messages.
//
// Keeping these in distinct lanes lets passive hinting, async status reporting
// and completion hints coexist without clobbering one another.
//
// All fields are guarded by mu: the transient lane in particular may be written
// from another goroutine while the main loop renders, so every accessor takes
// the lock.
type Hint struct {
	mu         sync.RWMutex
	text       []rune
	persistent []rune
	provided   []rune
	transient  []rune
	cleanup    bool
	temp       bool
	set        bool

	// provider, when set, computes the passive (provided) hint from the current
	// input line and cursor. It is re-evaluated on every refresh by the display
	// engine through UpdateProvided.
	provider func(line []rune, cursor int) []rune

	// refresh, when set, wakes the render loop after an async lane change (e.g.
	// SetTransient from another goroutine). Wired by the shell to the input wake
	// primitive; integrators do not call it directly.
	refresh func()
}

// NewHint creates a hint area wired to wake the render loop on asynchronous lane
// changes (such as SetTransient called from another goroutine), through the
// keys' input wake primitive.
func NewHint(keys *core.Keys) *Hint {
	hint := &Hint{}
	if keys != nil {
		hint.refresh = keys.RequestRefresh
	}

	return hint
}

// Set sets the hint message to the given text.
// Generally, this hint message will persist until either a command
// or the completion system overwrites it, or if hint.Reset() is called.
func (h *Hint) Set(hint string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.text = []rune(hint)
	h.set = true
}

// SetTemporary sets a hint message that will be cleared at the next keypress
// or command being run, which generally coincides with the next redisplay.
func (h *Hint) SetTemporary(hint string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.text = []rune(hint)
	h.set = true
	h.temp = true
}

// Persist adds a hint message to be persistently
// displayed until hint.ResetPersist() is called.
func (h *Hint) Persist(hint string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.persistent = []rune(hint)
}

// SetProvided sets the passive (provider) hint lane directly. It is normally
// recomputed from the current input line on every refresh (see SetProvider);
// an empty string clears the lane. This lane renders below the persistent lane
// and above the transient and completion lanes.
func (h *Hint) SetProvided(hint string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.provided = []rune(hint)
}

// SetProvider registers a function that computes the passive (provided) hint
// from the current input line and cursor position. It returns the hint text to
// show (return nil/empty for no hint). The provider is re-evaluated on every
// refresh, so the hint tracks the input as it changes.
//
// This is the "passive/background hinting" lane: it renders above the transient
// (async status) lane and the completion hints. Pass nil to remove the provider.
func (h *Hint) SetProvider(provider func(line []rune, cursor int) []rune) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.provider = provider

	if provider == nil {
		h.provided = make([]rune, 0)
	}
}

// UpdateProvided re-evaluates the registered provider (if any) against the given
// line and cursor and stores the result in the passive (provided) lane. It is
// called by the display engine on each refresh. The provider runs outside the
// lock so it may safely query shell state.
func (h *Hint) UpdateProvided(line []rune, cursor int) {
	h.mu.RLock()
	provider := h.provider
	h.mu.RUnlock()

	if provider == nil {
		return
	}

	h.SetProvided(string(provider(line, cursor)))
}

// SetTransient sets the transient (async status) hint lane. It is safe to call
// from any goroutine, and wakes an idle render loop so the message appears
// immediately. The message persists until ClearTransient is called or it is
// replaced; crucially it is NOT cleared by the completion or incremental search
// engines. It renders above the completion lane and below the provider lane.
func (h *Hint) SetTransient(hint string) {
	h.mu.Lock()
	h.transient = []rune(hint)
	refresh := h.refresh
	h.mu.Unlock()

	if refresh != nil {
		refresh()
	}
}

// ClearTransient drops the transient (async status) hint lane. It is safe to
// call from any goroutine, and wakes an idle render loop so the change is shown.
func (h *Hint) ClearTransient() {
	h.mu.Lock()
	h.cleanup = h.cleanup || len(h.transient) > 0
	h.transient = make([]rune, 0)
	refresh := h.refresh
	h.mu.Unlock()

	if refresh != nil {
		refresh()
	}
}

// SetRefreshFunc registers a callback used to wake the render loop when an
// async hint lane changes (e.g. SetTransient from another goroutine). The shell
// wires this to the input wake primitive; integrators do not call it directly.
func (h *Hint) SetRefreshFunc(refresh func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.refresh = refresh
}

// Text returns the current hint text.
func (h *Hint) Text() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return string(h.text)
}

// Len returns the length of the current hint.
// This is generally used by consumers to know if there already
// is an active hint, in which case they might want to append to
// it instead of overwriting it altogether (like in isearch mode).
func (h *Hint) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.text)
}

// Reset removes the hint message. It only clears the completion/command lane
// (text); the persistent, provider and transient lanes are left untouched, so
// async status and passive hints survive completion/isearch activity.
func (h *Hint) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.text = make([]rune, 0)
	h.temp = false
	h.set = false
}

// ResetPersist drops the persistent hint section.
func (h *Hint) ResetPersist() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.cleanup = len(h.persistent) > 0
	h.persistent = make([]rune, 0)
}

// DisplayHint prints the hint (persistent, provided, transient and/or
// completion) sections.
func DisplayHint(hint *Hint) {
	hint.mu.Lock()
	defer hint.mu.Unlock()

	// The completion/command (text) lane supports one-shot temporary messages:
	// keep them for this render, clear them on the next.
	if hint.temp && hint.set {
		hint.set = false
	} else if hint.temp {
		hint.text = make([]rune, 0)
		hint.temp = false
		hint.set = false
	}

	if hint.empty() {
		if hint.cleanup {
			term.WriteString(term.ClearLineAfter)
		}

		hint.cleanup = false

		return
	}

	text := hint.renderLocked()

	if strutil.RealLength(text) == 0 {
		return
	}

	text += term.ClearLineAfter + color.Reset

	if len(text) > 0 {
		term.WriteString(text)
	}
}

// empty reports whether every hint lane is empty. The caller must hold the lock.
func (h *Hint) empty() bool {
	return len(h.text) == 0 &&
		len(h.persistent) == 0 &&
		len(h.provided) == 0 &&
		len(h.transient) == 0
}

// renderLocked builds the hint block from all lanes. The caller must hold the
// lock (read or write).
func (h *Hint) renderLocked() (text string) {
	// Top to bottom: status, passive provider hint, async transient status,
	// then completion/command hints.
	for _, lane := range [][]rune{h.persistent, h.provided, h.transient, h.text} {
		if len(lane) > 0 {
			text += string(lane) + term.NewlineReturn
		}
	}

	if strutil.RealLength(text) == 0 {
		return
	}

	// Ensure cross-platform, real display newline.
	text = strings.ReplaceAll(text, term.NewlineReturn, term.ClearLineAfter+term.NewlineReturn)

	return text
}

// CoordinatesHint returns the number of terminal rows used by the hint.
//
// Each non-empty lane occupies its own row (wrapping when wider than the
// terminal), matching exactly what DisplayHint prints (one NewlineReturn per
// lane). It is counted from the lanes directly rather than from the rendered
// string, so the inter-lane separators do not get miscounted as extra rows.
func CoordinatesHint(hint *Hint) int {
	hint.mu.RLock()
	defer hint.mu.RUnlock()

	width := term.GetWidth()
	if width <= 0 {
		width = 80
	}

	usedY := 0

	for _, lane := range [][]rune{hint.persistent, hint.provided, hint.transient, hint.text} {
		if len(lane) == 0 {
			continue
		}

		// A lane may itself contain embedded newlines; count each sub-line,
		// wrapping when it is wider than the terminal.
		for _, sub := range strings.Split(string(lane), "\n") {
			length := strutil.RealLength(sub)

			rows := length / width
			if length%width != 0 || rows == 0 {
				rows++
			}

			usedY += rows
		}
	}

	return usedY
}
