package conditions

import (
	"github.com/DonGar/go-house/status"
	"time"
)

type periodicCondition struct {
	period time.Duration
	result chan bool
	stop   chan bool
}

func newPeriodicCondition(body *status.Status) (*periodicCondition, error) {

	interval, e := body.GetString("status://interval")
	if e != nil {
		return nil, e
	}

	period, e := time.ParseDuration(interval)
	if e != nil {
		return nil, e
	}

	// Create the condition.
	c := &periodicCondition{period, make(chan bool), make(chan bool)}

	// Start it processing.
	go c.handleTicks()

	return c, nil
}

func (c *periodicCondition) handleTicks() {
	ticker := time.NewTicker(c.period)

	for {
		select {
		case <-ticker.C:
			c.result <- true
			c.result <- false
		case <-c.stop:
			ticker.Stop()
			c.stop <- true
			return
		}
	}

}

func (c *periodicCondition) Result() <-chan bool {
	return c.result
}

func (c *periodicCondition) Stop() {
	c.stop <- true
	<-c.stop
}
