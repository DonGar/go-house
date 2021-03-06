package conditions

import (
	"github.com/DonGar/go-house/status"
	"reflect"
)

type watchCondition struct {
	base

	hasTrigger bool
	trigger    interface{}

	watchChan <-chan status.UrlMatches
}

func newWatchCondition(s *status.Status, body *status.Status) (*watchCondition, error) {

	watchUrl, _, e := body.GetString("status://watch")
	if e != nil {
		return nil, e
	}

	// Look up the trigger value, if present.
	trigger, _, e := body.Get("status://trigger")
	hasTrigger := e == nil

	// Start watching for updates.
	watchChan, e := s.WatchForUpdate(watchUrl)
	if e != nil {
		return nil, e
	}

	// Throw away the initial event that's always sent.
	if !hasTrigger {
		<-watchChan
	}

	// Create our condition.
	c := &watchCondition{newBase(s), hasTrigger, trigger, watchChan}

	// Start it's goroutine.
	go c.Handler()

	return c, nil
}

func (c *watchCondition) triggerInMatches(matches status.UrlMatches) bool {
	for _, match := range matches {
		if reflect.DeepEqual(match.Value, c.trigger) {
			return true
		}
	}

	return false
}

func (c *watchCondition) Handler() {

	// If there is no trigger, then always send an initial update.
	if !c.hasTrigger {
		c.sendResult(false)
	}

	for {
		select {
		case matches := <-c.watchChan:
			if c.hasTrigger {
				c.sendResult(c.triggerInMatches(matches))
			} else {
				// We got an update... notify our rule!
				c.sendResult(true)
				c.sendResult(false)
			}

		case <-c.StopChan:
			c.status.ReleaseWatch(c.watchChan)
			c.StopChan <- true
			return
		}
	}
}
