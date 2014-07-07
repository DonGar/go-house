package logging

import (
	"strings"
	"sync"
)

type CachedLogging struct {
	revision  int
	logs      []string
	notifiers []chan CachedLoggingUpdate
	lock      sync.Mutex
}

type CachedLoggingUpdate struct {
	Revision int
	Logs     []string
}

func (c *CachedLogging) Write(p []byte) (n int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// TODO: Currently we assume each Write is exactly one line. That's not a
	// valid assumption.

	// Append the new log line.
	c.logs = append(c.logs, strings.TrimSpace(string(p)))

	// Keep at most 100 log lines.
	if len(c.logs) >= 100 {
		c.logs = c.logs[1:len(c.logs)]
	}

	// Update our revision, and notify any listeners.
	c.revision += 1

	// Signal to all of our notifiers that we were updated.
	for _, n := range c.notifiers {
		c.notifyWatch(n)
	}

	// We always write the full blob, return success.
	return len(p), nil
}

func (c *CachedLogging) notifyWatch(watch chan CachedLoggingUpdate) {
	select {
	case <-watch: // Clear the channel if it hasn't been read from.
	default: // Read nothing if it's empty.
	}

	update := CachedLoggingUpdate{Revision: c.revision}
	update.Logs = append(update.Logs, c.logs...)

	watch <- update
}

func (c *CachedLogging) WatchForUpdate() (watch <-chan CachedLoggingUpdate) {
	c.lock.Lock()
	defer c.lock.Unlock()

	newWatch := make(chan CachedLoggingUpdate, 1)
	c.notifiers = append(c.notifiers, newWatch)
	c.notifyWatch(newWatch)

	return newWatch
}

func (c *CachedLogging) ReleaseWatch(watch <-chan CachedLoggingUpdate) {
	c.lock.Lock()
	defer c.lock.Unlock()

	trimmedNotifiers := make([]chan CachedLoggingUpdate, 0, len(c.notifiers))

	for _, w := range c.notifiers {
		if w != watch {
			trimmedNotifiers = append(trimmedNotifiers, w)
		}
	}

	c.notifiers = trimmedNotifiers
}
