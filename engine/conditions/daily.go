package conditions

import (
	"github.com/DonGar/go-house/status"
	"time"
)

type dailyCondition struct {
	base

	timeOfDay time.Duration // Time of day at which to fire.
	duration  time.Duration // Duration for which to fire.
}

func newDailyCondition(s *status.Status, body *status.Status) (*dailyCondition, error) {
	timeDescription, _, e := body.GetString("status://time")
	if e != nil {
		return nil, e
	}

	// Parse time values.
	startTime, e := parseTime(timeDescription)
	if e != nil {
		return nil, e
	}

	timeDescription, _, e = body.GetString("status://duration")
	if e != nil {
		timeDescription = "1m"
	}

	// Parse time values.
	duration, e := time.ParseDuration(timeDescription)
	if e != nil {
		return nil, e
	}

	c := &dailyCondition{newBase(s), startTime, duration}

	// Start it's goroutine.
	go c.Handler()

	return c, nil
}

func parseTime(timeDescription string) (timeOfDay time.Duration, e error) {
	// We accept a variety of formats for the time of day. Parse different
	// ways until one works.
	var fixedTime time.Time
	for _, format := range []string{"3:04:05PM", "3:04PM", "15:04:05", "15:04"} {
		fixedTime, e = time.Parse(format, timeDescription)
		if e == nil {
			break
		}
	}

	if e != nil {
		// None of the time formats worked.
		return 0, e
	}

	// Our parsed times have some very odd date information. Strip it out.
	hour, min, sec := fixedTime.Clock()
	timeOfDay = (time.Duration(hour)*time.Hour +
		time.Duration(min)*time.Minute +
		time.Duration(sec)*time.Second)

	return timeOfDay, nil
}

func (c *dailyCondition) findFireState(now time.Time) (active bool, fireTime time.Duration) {
	calcNow := now

	for {
		year, month, day := calcNow.Date()
		startOfDay := time.Date(year, month, day, 0, 0, 0, 0, calcNow.Location())
		startTime := startOfDay.Add(c.timeOfDay)
		endTime := startTime.Add(c.duration)

		// It's not time yet, so false, and wait for start time.
		if now.Before(startTime) {
			return false, startTime.Sub(now)
		}

		// startTime <= now <= endTime. So true, and wait for endTime.
		if now.Before(endTime) {
			return true, endTime.Sub(now)
		}

		// If we are already after today's fire window, look for tomorrow's. We only
		// move 13 hours to help deal with daylight savings, and other funkyness.
		calcNow = calcNow.Add(13 * time.Hour)
	}
}

func (c *dailyCondition) Handler() {
	// Set the timer to fire immediately to send the initial state.
	timer := time.NewTimer(0)

	for {
		select {
		case <-timer.C:
			// Lookup if we are active or not, and how long until next change.
			now := time.Now()
			active, fireDelay := c.findFireState(now)

			c.sendResult(active)
			timer.Reset(fireDelay)

		case <-c.StopChan:
			timer.Stop()
			c.StopChan <- true
			return
		}
	}
}
