package conditions

import (
	"fmt"
	"github.com/DonGar/go-house/status"
	"github.com/DonGar/go-timeglob/timeglob"
	"time"
)

type timeCondition struct {
	base

	timeglob *timeglob.TimeGlob // Time of day at which to fire.
	duration time.Duration      // Duration for which to fire.
}

func newTimeCondition(s *status.Status, body *status.Status) (*timeCondition, error) {
	timeDescription, _, e := body.GetString("status://time")
	if e != nil {
		return nil, e
	}

	// Parse time values.
	tg, e := timeglob.Parse(timeDescription)
	if e != nil {
		return nil, e
	}

	timeDescription, _, e = body.GetString("status://duration")
	if e != nil {
		timeDescription = "0s"
	}

	// Parse time values.
	duration, e := time.ParseDuration(timeDescription)
	if e != nil {
		return nil, e
	}

	c := &timeCondition{newBase(s), tg, duration}

	// Start it's goroutine.
	go c.Handler()

	return c, nil
}

func (c *timeCondition) findActiveState(now time.Time) (active bool, remainingActive time.Duration) {
	prev := c.timeglob.Prev(now)
	activeDuration := now.Sub(prev)
	active = prev != timeglob.UNKNOWN && activeDuration < c.duration

	if active {
		remainingActive = c.duration - activeDuration
	}

	return active, remainingActive
}

func (c *timeCondition) Handler() {
	start := c.timeglob.Ticker()
	stop := time.NewTimer(0) // Stop right away to force initial evaluation.

	handleActive := func() {
		active, remainingActive := c.findActiveState(time.Now())
		c.sendResult(active)
		if active {
			stop.Reset(remainingActive)
		} else {
			stop.Stop()
		}
	}

	for {
		select {
		case <-start.C:
			fmt.Printf("Start event.\n")
			// Force a true to be sent when ticker fires. Below might not since
			// now can be slightly after tick time when we are reached.
			c.sendResult(true)
			handleActive()

		case <-stop.C:
			fmt.Printf("Stop event.\n")
			handleActive()

		case <-c.StopChan:
			start.Stop()
			stop.Stop()
			c.StopChan <- true
			return
		}
	}
}
