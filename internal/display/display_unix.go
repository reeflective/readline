//go:build unix
// +build unix

package display

import (
	"os"
	"os/signal"
	"syscall"
)

// WatchResize redisplays the interface on terminal resize events.
func WatchResize(eng *Engine) chan<- bool {
	done := make(chan bool, 1)

	resizeChannel := make(chan os.Signal, 1)
	signal.Notify(resizeChannel, syscall.SIGWINCH)

	go func() {
		for {
			select {
			case <-resizeChannel:
				// Route the regeneration + repaint through the input wake so
				// they run on the Readline goroutine, instead of mutating the
				// completion/display state and writing to stdout from here
				// (which races with the main loop).
				eng.completer.RequestRegen()
				eng.keys.RequestRefresh()
			case <-done:
				return
			}
		}
	}()

	return done
}
