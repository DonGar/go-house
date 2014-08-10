package engine

import (
	"github.com/DonGar/go-house/status"
	"github.com/DonGar/go-house/stoppable"
	"log"
)

type newWatch func(url string, body *status.Status) (stoppable.Stoppable, error)

type watched struct {
	revision int
	value    stoppable.Stoppable
}

type watcher struct {
	status  *status.Status
	url     string
	factory newWatch
	active  map[string]watched
	stoppable.Base
}

func newWatcher(status *status.Status, url string, factory newWatch) *watcher {
	result := &watcher{status, url, factory, map[string]watched{}, stoppable.NewBase()}

	go result.Handler()

	return result
}

// This is our back ground process for noticing rules updates.
func (w *watcher) Handler() {
	watch, e := w.status.WatchForUpdate(w.url)
	if e != nil {
		panic("Failure watching: " + w.url)
	}

	for {
		select {
		case matches := <-watch:
			// First remove rules that were removed or updated.
			w.update(matches)
		case <-w.StopChan:
			// Stop watching for changes, remove all existing rules, and signal done.
			w.status.ReleaseWatch(watch)
			w.update(status.UrlMatches{})
			w.StopChan <- true
			return
		}
	}
}

// Remove any rules that have been removed, or updated.
func (w *watcher) update(matches status.UrlMatches) {

	// Remove all rules that no longer exist, or which have been updated.
	for url, active := range w.active {
		match, ok := matches[url]
		if !ok || match.Revision != active.revision {
			// It's no longer valid, remove it.
			active.value.Stop()
			delete(w.active, url)
		}
	}

	// Create all actives that don't exist in our manager.
	for url, match := range matches {
		// If the active already exists, leave it alone.
		if _, ok := w.active[url]; ok {
			continue
		}

		// Find it's body.
		activeBody := &status.Status{}
		e := activeBody.Set("status://", match.Value, 0)
		if e != nil {
			log.Panic(e) // This is supposed to be impossible.
		}

		// Create it.
		active, e := w.factory(url, activeBody)
		if e != nil {
			// Skip over invalid actives.
			log.Printf("INVALID: %s: %s", url, e.Error())
			continue
		}
		w.active[url] = watched{match.Revision, active}
	}
}
