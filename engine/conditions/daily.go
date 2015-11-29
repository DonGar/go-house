package conditions

import (
	"fmt"
	"github.com/DonGar/go-house/status"
	"github.com/cpucycle/astrotime"
	"time"
)

type timeType int

const (
	sunset timeType = iota
	sunrise
	fixed
)

type dailyCondition struct {
	base

	latitude    float64
	longitude   float64
	timeType    timeType      // sunset, sunrise, or fixed.
	fixedOffset time.Duration // If fixed, how long after midnight until we fire?
}

func newDailyCondition(s *status.Status, body *status.Status) (*dailyCondition, error) {
	// Look up our Latitude and Longitude
	latitude := s.GetFloatWithDefault("status://server/latitude", 0.0)
	longitude := s.GetFloatWithDefault("status://server/longitude", 0.0)

	timeDescription, _, e := body.GetString("status://time")
	if e != nil {
		return nil, e
	}

	// Parse time values.
	timeType, fixedOffset, e := parseTime(timeDescription)
	if e != nil {
		return nil, e
	}

	c := &dailyCondition{newBase(s), latitude, longitude, timeType, fixedOffset}

	// Start it's goroutine.
	go c.Handler()

	return c, nil
}

func parseTime(timeDescription string) (timeType timeType, fixedOffset time.Duration, e error) {
	switch timeDescription {
	case "sunrise":
		timeType = sunrise

	case "sunset":
		timeType = sunset

	default:
		timeType = fixed

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
			return 0, 0, e
		}

		// Our parsed times have some very odd date information. Strip it out.
		hour, min, sec := fixedTime.Clock()
		fixedOffset = (time.Duration(hour)*time.Hour +
			time.Duration(min)*time.Minute +
			time.Duration(sec)*time.Second)
	}

	return timeType, fixedOffset, nil
}

func (c *dailyCondition) findNextFireTime(now time.Time) (fireTime time.Time) {

	findFireTime := func(now time.Time) time.Time {
		switch c.timeType {
		case sunrise:
			// Push the time back by 5 minutes so rounding errors don't cause us to
			// fire more than once in a day.
			return astrotime.CalcSunrise(now, c.latitude, c.longitude)

		case sunset:
			// Push the time back by 5 minutes so rounding errors don't cause us to
			// fire more than once in a day.
			return astrotime.CalcSunset(now, c.latitude, c.longitude)

		case fixed:
			year, month, day := now.Date()
			startOfDay := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
			return startOfDay.Add(c.fixedOffset)

		default:
			panic(fmt.Errorf("Unknown timeType: %d", c.timeType))
		}
	}

	// If the time for today has already passed, look for the tomorrow. We move
	// forward by less than 24 hours to deal with daylight savings, and other edge
	// cases.
	findNextFrom := now
	for !fireTime.After(now) {
		fireTime = findFireTime(findNextFrom)
		findNextFrom = findNextFrom.Add(13 * time.Hour)
	}

	return fireTime
}

func (c *dailyCondition) Handler() {
	c.sendResult(false)

	now := time.Now()
	timer := time.NewTimer(c.findNextFireTime(now).Sub(now))

	for {
		select {
		case <-timer.C:
			// We turned true again.
			c.sendResult(true)
			c.sendResult(false)

			// Set timer for the next firing. Add 5 minutes to work around
			// sunrise/sunset calculation vagueness.
			now := time.Now()
			timer.Reset(c.findNextFireTime(now.Add(5 * time.Minute)).Sub(now))

		case <-c.StopChan:
			timer.Stop()
			c.StopChan <- true
			return
		}
	}
}
