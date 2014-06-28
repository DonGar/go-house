package rules

import (
	"fmt"
	"github.com/cpucycle/astrotime"
	"time"
)

type timeType int

const (
	sunset timeType = iota
	sunrise
	fixed
)

type dailyRule struct {
	base

	timeType    timeType      // sunset, sunrise, or fixed.
	fixedOffset time.Duration // If fixed, how long after midnight until we fire?
	stop        chan bool
}

func newDailyRule(base base) (rule, error) {

	timeDescription, e := base.body.GetString("status://time")
	if e != nil {
		return nil, e
	}

	// Create and populate the rule.
	r := &dailyRule{base: base, stop: make(chan bool)}
	r.timeType, r.fixedOffset, e = parseTime(timeDescription)
	if e != nil {
		return nil, e
	}

	// Start it's goroutine.
	go r.handleTimer()

	return r, nil
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

const lat = 37.3894
const long = 122.0819

func (r *dailyRule) findNextFireTime(now time.Time) (fireTime time.Time) {

	findFireTime := func(now time.Time) time.Time {
		switch r.timeType {
		case sunrise:
			// Push the time back by 5 minutes so rounding errors don't cause us to
			// fire more than once in a day.
			return astrotime.CalcSunrise(now, lat, long)

		case sunset:
			// Push the time back by 5 minutes so rounding errors don't cause us to
			// fire more than once in a day.
			return astrotime.CalcSunset(now, lat, long)

		case fixed:
			year, month, day := now.Date()
			startOfDay := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
			return startOfDay.Add(r.fixedOffset)

		default:
			panic(fmt.Errorf("Unknown timeType: %d", r.timeType))
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

func (r *dailyRule) handleTimer() {

	now := time.Now()
	timer := time.NewTimer(r.findNextFireTime(now).Sub(now))

	for {
		select {
		case <-timer.C:
			r.fire()

			// Set timer for the next firing. Add 5 minutes to work around
			// sunrise/sunset calculation vagueness.
			now := time.Now()
			timer.Reset(r.findNextFireTime(now.Add(5 * time.Minute)).Sub(now))

		case <-r.stop:
			timer.Stop()
			r.stop <- true
			return
		}
	}
}

func (r *dailyRule) Stop() error {
	r.stop <- true
	<-r.stop
	return r.base.Stop()
}
