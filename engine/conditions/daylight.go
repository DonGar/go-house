package conditions

import (
	// "fmt"
	"github.com/DonGar/go-house/status"
	"github.com/cpucycle/astrotime"
	"time"
)

type daylightCondition struct {
	base

	latitude  float64
	longitude float64
	day       bool
}

func newDaylightCondition(s *status.Status, body *status.Status, day bool) (*daylightCondition, error) {
	// Look up our Latitude and Longitude
	latitude := s.GetFloatWithDefault("status://server/latitude", 0.0)
	longitude := s.GetFloatWithDefault("status://server/longitude", 0.0)

	c := &daylightCondition{newBase(s), latitude, longitude, day}

	// Start it's goroutine.
	go c.Handler()

	return c, nil
}

// Less than 24 hours to help with daylight savings, and other weirdness.
const DAY_INCREMENT = 13 * time.Hour

func findNextSunrise(now time.Time, latitude, longitude float64) time.Time {
	// Start in the past, because astrotime sometimes returns tomorrow's result.
	calcNow := now.Add(-24 * time.Hour)

	for {
		next := astrotime.CalcSunrise(calcNow, latitude, longitude)
		if next.After(now) {
			return next
		}

		calcNow = calcNow.Add(DAY_INCREMENT)
	}
}

func findNextSunset(now time.Time, latitude, longitude float64) time.Time {
	// Start in the past, because astrotime sometimes returns tomorrow's result.
	calcNow := now.Add(-24 * time.Hour)

	for {
		next := astrotime.CalcSunset(calcNow, latitude, longitude)
		if next.After(now) {
			return next
		}

		calcNow = calcNow.Add(DAY_INCREMENT)
	}
}

func (c *daylightCondition) findIsDayAndNextChange(now time.Time) (isDay bool, changeTime time.Time) {
	nextSunrise := findNextSunrise(now, c.latitude, c.longitude)
	nextSunset := findNextSunset(now, c.latitude, c.longitude)

	if nextSunrise.Before(nextSunset) {
		// If the next transition is sunrise, it's night.
		isDay, changeTime = false, nextSunrise
	} else {
		// if the next transition is sunset, it's day.
		isDay, changeTime = true, nextSunset
	}

	// Adjust change time to work around time calculation float slop.
	// changeTime = changeTime.Add(5 * time.Minute)
	return isDay, changeTime
}

func (c *daylightCondition) Handler() {

	// Set the timer to fire immediately to send the initial state.
	timer := time.NewTimer(0)

	for {
		select {
		case <-timer.C:
			// Set timer for the next firing.
			now := time.Now()
			isDay, nextTransition := c.findIsDayAndNextChange(now)
			timer.Reset(nextTransition.Sub(now))

			// Announce new state.
			c.sendResult(isDay == c.day)

		case <-c.StopChan:
			timer.Stop()
			c.StopChan <- true
			return
		}
	}
}
