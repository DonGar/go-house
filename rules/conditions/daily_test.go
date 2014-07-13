package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

func setupTimeCondition(c *check.C, time string) *dailyCondition {
	s := &status.Status{}

	e := s.Set("status://server/latitude", 37.3861, 0)
	c.Assert(e, check.IsNil)

	e = s.Set("status://server/longitude", 122.0839, 1)
	c.Assert(e, check.IsNil)

	body := &status.Status{}
	e = body.Set("status://", map[string]interface{}{"time": time}, 0)
	c.Assert(e, check.IsNil)

	cond, e := newDailyCondition(s, body)
	c.Assert(e, check.IsNil)

	return cond
}

func (suite *MySuite) TestDailyStartStop(c *check.C) {
	cond := setupTimeCondition(c, "12:00")
	cond.Stop()
}

func (suite *MySuite) TestDailyParseTime(c *check.C) {
	var timeType timeType
	var fixedOffset time.Duration
	var e error

	timeType, fixedOffset, e = parseTime("foo")
	c.Check(timeType, check.Equals, sunset)
	c.Check(fixedOffset, check.Equals, time.Duration(0))
	c.Check(e, check.NotNil)

	timeType, fixedOffset, e = parseTime("sunset")
	c.Check(timeType, check.Equals, sunset)
	c.Check(fixedOffset, check.Equals, time.Duration(0))
	c.Check(e, check.IsNil)

	timeType, fixedOffset, e = parseTime("sunrise")
	c.Check(timeType, check.Equals, sunrise)
	c.Check(fixedOffset, check.Equals, time.Duration(0))
	c.Check(e, check.IsNil)

	timeType, fixedOffset, e = parseTime("11:00:00AM")
	c.Check(timeType, check.Equals, fixed)
	c.Check(fixedOffset, check.Equals, 11*time.Hour)
	c.Check(e, check.IsNil)

	timeType, fixedOffset, e = parseTime("12:34:56AM")
	c.Check(timeType, check.Equals, fixed)
	c.Check(fixedOffset, check.Equals, time.Duration(34*time.Minute+56*time.Second))
	c.Check(e, check.IsNil)

	timeType, fixedOffset, e = parseTime("11:00AM")
	c.Check(timeType, check.Equals, fixed)
	c.Check(fixedOffset, check.Equals, 11*time.Hour)
	c.Check(e, check.IsNil)

	timeType, fixedOffset, e = parseTime("12:00AM")
	c.Check(timeType, check.Equals, fixed)
	c.Check(fixedOffset, check.Equals, 0*time.Hour)
	c.Check(e, check.IsNil)

	timeType, fixedOffset, e = parseTime("11:00:00")
	c.Check(timeType, check.Equals, fixed)
	c.Check(fixedOffset, check.Equals, 11*time.Hour)
	c.Check(e, check.IsNil)

	// Midnight
	timeType, fixedOffset, e = parseTime("00:00")
	c.Check(timeType, check.Equals, fixed)
	c.Check(fixedOffset, check.Equals, 0*time.Hour)
	c.Check(e, check.IsNil)

	timeType, fixedOffset, e = parseTime("8:00")
	c.Check(timeType, check.Equals, fixed)
	c.Check(fixedOffset, check.Equals, 8*time.Hour)
	c.Check(e, check.IsNil)

	// Noon
	timeType, fixedOffset, e = parseTime("12:00")
	c.Check(timeType, check.Equals, fixed)
	c.Check(fixedOffset, check.Equals, 12*time.Hour)
	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestFindFireTimeForDateSunrise(c *check.C) {
	cond := setupTimeCondition(c, "sunrise")

	sunriseToday := time.Date(2014, time.June, 12, 05, 47, 01, 0, time.Local)
	sunriseTomorrow := time.Date(2014, time.June, 13, 05, 46, 59, 0, time.Local)

	// Test before sunrise.
	now := time.Date(2014, time.June, 12, 2, 57, 12, 0, time.Local)
	fireTime := cond.findNextFireTime(now)
	c.Check(fireTime, check.Equals, sunriseToday)

	// Test 5 min after sunrise.
	now = sunriseToday.Add(5 * time.Minute)
	fireTime = cond.findNextFireTime(now)
	c.Check(fireTime, check.Equals, sunriseTomorrow)

	// Test after sunrise.
	now = time.Date(2014, time.June, 12, 10, 57, 12, 0, time.Local)
	fireTime = cond.findNextFireTime(now)
	c.Check(fireTime, check.Equals, sunriseTomorrow)

	// Test way after sunrise.
	now = time.Date(2014, time.June, 12, 23, 57, 12, 0, time.Local)
	fireTime = cond.findNextFireTime(now)
	c.Check(fireTime, check.Equals, sunriseTomorrow)

	cond.Stop()
}

func (suite *MySuite) TestFindFireTimeForDateSunset(c *check.C) {
	cond := setupTimeCondition(c, "sunset")

	sunsetToday := time.Date(2014, time.June, 12, 20, 29, 33, 0, time.Local)
	sunsetTomorrow := time.Date(2014, time.June, 13, 20, 30, 00, 0, time.Local)

	// Test before sunset.
	now := time.Date(2014, time.June, 12, 2, 57, 12, 0, time.Local)
	fireTime := cond.findNextFireTime(now)
	c.Check(fireTime, check.Equals, sunsetToday)

	// Test 5 min after sunset. Round since time of sunset is approximate.
	now = sunsetToday.Add(5 * time.Minute)
	fireTime = cond.findNextFireTime(now)
	c.Check(fireTime.Round(time.Minute), check.Equals, sunsetTomorrow)

	// Test after sunset.
	now = time.Date(2014, time.June, 12, 22, 57, 12, 0, time.Local)
	fireTime = cond.findNextFireTime(now)
	c.Check(fireTime.Round(time.Minute), check.Equals, sunsetTomorrow)

	cond.Stop()
}

func (suite *MySuite) TestFixedFindFireDelay(c *check.C) {
	cond := setupTimeCondition(c, "11:00")

	fixedToday := time.Date(2014, time.June, 12, 11, 00, 00, 0, time.Local)
	fixedTomorrow := time.Date(2014, time.June, 13, 11, 00, 00, 0, time.Local)

	// Typical short deay.
	now := time.Date(2014, time.June, 12, 10, 57, 12, 0, time.Local)
	fireTime := cond.findNextFireTime(now)
	c.Check(fireTime, check.Equals, fixedToday)

	// When now is midnight.
	now = time.Date(2014, time.June, 12, 0, 0, 0, 0, time.Local)
	fireTime = cond.findNextFireTime(now)
	c.Check(fireTime, check.Equals, fixedToday)

	// Very, very short fireTime.
	now = time.Date(2014, time.June, 12, 10, 59, 59, 200, time.Local)
	fireTime = cond.findNextFireTime(now)
	c.Check(fireTime, check.Equals, fixedToday)

	// Delay is zero.
	now = time.Date(2014, time.June, 12, 11, 0, 0, 0, time.Local)
	fireTime = cond.findNextFireTime(now)
	c.Check(fireTime, check.Equals, fixedTomorrow)

	// Just past deadline.
	now = time.Date(2014, time.June, 12, 11, 59, 0, 0, time.Local)
	fireTime = cond.findNextFireTime(now)
	c.Check(fireTime, check.Equals, fixedTomorrow)

	cond.Stop()
}

// TODO: Figure out how to mock time.Now() and test rule firing.