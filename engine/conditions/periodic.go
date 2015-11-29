package conditions

import (
	"github.com/DonGar/go-house/status"
	"time"
)

type periodicCondition struct {
	base

	period time.Duration
}

func newPeriodicCondition(s *status.Status, body *status.Status) (*periodicCondition, error) {

	interval, _, e := body.GetString("status://interval")
	if e != nil {
		return nil, e
	}

	period, e := time.ParseDuration(interval)
	if e != nil {
		return nil, e
	}

	// Create the condition.
	c := &periodicCondition{newBase(s), period}

	// Start it processing.
	go c.Handler()

	return c, nil
}

func (c *periodicCondition) Handler() {
	c.sendResult(false)
	ticker := time.NewTicker(c.period)

	for {
		select {
		case <-ticker.C:
			c.sendResult(true)
			c.sendResult(false)
		case <-c.StopChan:
			ticker.Stop()
			c.StopChan <- true
			return
		}
	}
}
