package diff

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// ReloadMsg is sent to the Bubble Tea program when a watched file changes.
type ReloadMsg struct{}

// WatchFiles starts a goroutine that watches the given paths and sends
// ReloadMsg to the program when any of them change. Debounced at 200ms.
// Call cancel() to stop the watcher.
func WatchFiles(p *tea.Program, paths []string) (cancel func(), err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		watcher.Add(path)
	}
	done := make(chan struct{})
	go func() {
		defer watcher.Close()
		timer := time.NewTimer(0)
		timer.Stop()
		for {
			select {
			case <-done:
				return
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}
				timer.Reset(200 * time.Millisecond)
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			case <-timer.C:
				p.Send(ReloadMsg{})
			}
		}
	}()
	return func() { close(done) }, nil
}
