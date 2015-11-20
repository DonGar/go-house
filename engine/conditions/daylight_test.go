package conditions

import (
	"github.com/DonGar/go-house/status"
	"github.com/cpucycle/astrotime"
	"gopkg.in/check.v1"
	"time"
)

const (
	LATITUDE  = 37.3861
	LONGITUDE = 122.0839
)

func setupDaylightCondition(c *check.C, day bool) *daylightCondition {
	s := &status.Status{}

	e := s.Set("status://server/latitude", LATITUDE, 0)
	c.Assert(e, check.IsNil)

	e = s.Set("status://server/longitude", LONGITUDE, 1)
	c.Assert(e, check.IsNil)

	body := &status.Status{}

	cond, e := newDaylightCondition(s, body, day)
	c.Assert(e, check.IsNil)

	return cond
}

func (suite *MySuite) TestDayStartStop(c *check.C) {
	cond := setupDaylightCondition(c, true)
	cond.Stop()
}

func (suite *MySuite) TestFindIsDayAndChangeDelay(c *check.C) {
	cond := setupDaylightCondition(c, true)

	testDay := time.Date(2014, time.June, 13, 0, 0, 00, 0, time.UTC)

	// sunrise := time.Date(2014, time.June, 12, 12, 47, 00, 0, time.UTC)
	sunrise := astrotime.CalcSunrise(testDay, LATITUDE, LONGITUDE)
	sunset := astrotime.CalcSunset(testDay, LATITUDE, LONGITUDE)
	sunriseTomorrow := astrotime.CalcSunrise(testDay.Add(24*time.Hour), LATITUDE, LONGITUDE)

	// Test before sunrise.
	now := sunrise.Add(-5 * time.Minute)
	isDay, transitionTime := cond.findIsDayAndNextChange(now)
	c.Check(isDay, check.Equals, false)
	c.Check(transitionTime.Round(time.Minute), check.Equals, sunrise.Round(time.Minute))

	// Test after sunrise.
	now = sunrise.Add(5 * time.Minute)
	isDay, transitionTime = cond.findIsDayAndNextChange(now)
	c.Check(isDay, check.Equals, true)
	c.Check(transitionTime.Round(time.Minute), check.Equals, sunset.Round(time.Minute))

	// Test mid-day.
	now = sunrise.Add(7 * time.Hour)
	isDay, transitionTime = cond.findIsDayAndNextChange(now)
	c.Check(isDay, check.Equals, true)
	c.Check(transitionTime.Round(time.Minute), check.Equals, sunset.Round(time.Minute))

	// Test before sunset.
	now = sunset.Add(-5 * time.Minute)
	isDay, transitionTime = cond.findIsDayAndNextChange(now)
	c.Check(isDay, check.Equals, true)
	c.Check(transitionTime.Round(time.Minute), check.Equals, sunset.Round(time.Minute))

	// Test after sunset.
	now = sunset.Add(5 * time.Minute)
	isDay, transitionTime = cond.findIsDayAndNextChange(now)
	c.Check(isDay, check.Equals, false)
	c.Check(transitionTime.Round(time.Minute), check.Equals, sunriseTomorrow.Round(time.Minute))

	// Test mid-night.
	now = sunset.Add(6 * time.Hour)
	isDay, transitionTime = cond.findIsDayAndNextChange(now)
	c.Check(isDay, check.Equals, false)
	c.Check(transitionTime.Round(time.Minute), check.Equals, sunriseTomorrow.Round(time.Minute))

	cond.Stop()
}
