package conditions

import (
	"github.com/DonGar/go-house/status"
)

type watchCondition struct {
	base

	status    *status.Status
	watchChan <-chan status.UrlMatches
}

func newWatchCondition(s *status.Status, body *status.Status) (*watchCondition, error) {

	watchUrl, e := body.GetString("status://watch")
	if e != nil {
		return nil, e
	}

	// Start watching for updates.
	watchChan, e := s.WatchForUpdate(watchUrl)
	if e != nil {
		return nil, e
	}

	// Throw away the initial event that's always sent.
	<-watchChan

	// Create our condition.
	c := &watchCondition{base{make(chan bool), make(chan bool)}, s, watchChan}

	// Start it's goroutine.
	go c.handleWatch()

	return c, nil
}

func (c *watchCondition) handleWatch() {

	for {
		select {
		case <-c.watchChan:
			// We got an update... notify our rule!
			c.result <- true
			c.result <- false

		case <-c.stop:
			c.status.ReleaseWatch(c.watchChan)
			c.stop <- true
			return
		}
	}
}
