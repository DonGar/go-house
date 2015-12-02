package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

func setupTimeCondition(c *check.C, time, duration string) *dailyCondition {
	s := &status.Status{}

	body := &status.Status{}
	err := body.Set("status://time", time, 0)
	c.Assert(err, check.IsNil)

	if duration != "" {
		err := body.Set("status://duration", duration, 1)
		c.Assert(err, check.IsNil)
	}

	cond, err := newDailyCondition(s, body)
	c.Assert(err, check.IsNil)

	return cond
}

func (suite *MySuite) TestDailyStartStop(c *check.C) {
	// The 0 duration, means the condition is always false.
	cond := setupTimeCondition(c, "12:00", "0s")
	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)
	cond.Stop()
}

func (suite *MySuite) TestDailyParseTime(c *check.C) {
	var timeOfDay time.Duration
	var err error

	timeOfDay, err = parseTime("foo")
	c.Check(timeOfDay, check.Equals, time.Duration(0))
	c.Check(err, check.NotNil)

	timeOfDay, err = parseTime("11:00:00AM")
	c.Check(timeOfDay, check.Equals, 11*time.Hour)
	c.Check(err, check.IsNil)

	timeOfDay, err = parseTime("12:34:56AM")
	c.Check(timeOfDay, check.Equals, time.Duration(34*time.Minute+56*time.Second))
	c.Check(err, check.IsNil)

	timeOfDay, err = parseTime("11:00AM")
	c.Check(timeOfDay, check.Equals, 11*time.Hour)
	c.Check(err, check.IsNil)

	timeOfDay, err = parseTime("23:00")
	c.Check(timeOfDay, check.Equals, 23*time.Hour)
	c.Check(err, check.IsNil)

	timeOfDay, err = parseTime("12:00AM")
	c.Check(timeOfDay, check.Equals, 0*time.Hour)
	c.Check(err, check.IsNil)

	timeOfDay, err = parseTime("11:00:00")
	c.Check(timeOfDay, check.Equals, 11*time.Hour)
	c.Check(err, check.IsNil)

	// Midnight
	timeOfDay, err = parseTime("00:00")
	c.Check(timeOfDay, check.Equals, 0*time.Hour)
	c.Check(err, check.IsNil)

	timeOfDay, err = parseTime("8:00")
	c.Check(timeOfDay, check.Equals, 8*time.Hour)
	c.Check(err, check.IsNil)

	// Noon
	timeOfDay, err = parseTime("12:00")
	c.Check(timeOfDay, check.Equals, 12*time.Hour)
	c.Check(err, check.IsNil)
}

func (suite *MySuite) TestFindFireState(c *check.C) {
	cond := setupTimeCondition(c, "11:00", "")

	// Verify that time/duration were parsed correctly.
	c.Check(cond.timeOfDay, check.Equals, 11*time.Hour)
	c.Check(cond.duration, check.Equals, 1*time.Minute)

	var now time.Time
	var active bool
	var delay time.Duration

	// Almost start time.
	now = time.Date(2014, time.June, 12, 10, 57, 12, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 2*time.Minute+48*time.Second)

	// Start time.
	now = time.Date(2014, time.June, 12, 11, 00, 00, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, true)
	c.Check(delay, check.Equals, 60*time.Second)

	// Almost end time.
	now = time.Date(2014, time.June, 12, 11, 00, 59, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, true)
	c.Check(delay, check.Equals, 1*time.Second)

	// End time.
	now = time.Date(2014, time.June, 12, 11, 01, 0, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 23*time.Hour+59*time.Minute)

	// After end time.
	now = time.Date(2014, time.June, 12, 11, 01, 2, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 23*time.Hour+58*time.Minute+58*time.Second)

	// When now is midnight.
	now = time.Date(2014, time.June, 12, 0, 0, 0, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 11*time.Hour)

	cond.Stop()
}

func (suite *MySuite) TestFindFireStateDuration(c *check.C) {
	cond := setupTimeCondition(c, "11:00", "12s")

	// Verify that time/duration were parsed correctly.
	c.Check(cond.timeOfDay, check.Equals, 11*time.Hour)
	c.Check(cond.duration, check.Equals, 12*time.Second)

	var now time.Time
	var active bool
	var delay time.Duration

	// Almost start time.
	now = time.Date(2014, time.June, 12, 10, 57, 12, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 2*time.Minute+48*time.Second)

	// Start time.
	now = time.Date(2014, time.June, 12, 11, 00, 00, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, true)
	c.Check(delay, check.Equals, 12*time.Second)

	// Almost end time.
	now = time.Date(2014, time.June, 12, 11, 00, 11, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, true)
	c.Check(delay, check.Equals, 1*time.Second)

	// End time.
	now = time.Date(2014, time.June, 12, 11, 00, 12, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 23*time.Hour+59*time.Minute+48*time.Second)

	// After end time.
	now = time.Date(2014, time.June, 12, 11, 01, 00, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 23*time.Hour+59*time.Minute)

	// When now is midnight.
	now = time.Date(2014, time.June, 12, 0, 0, 0, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 11*time.Hour)

	cond.Stop()
}

func (suite *MySuite) TestFindFireStateZeroDuration(c *check.C) {
	cond := setupTimeCondition(c, "11:00", "0s")

	// Verify that time/duration were parsed correctly.
	c.Check(cond.timeOfDay, check.Equals, 11*time.Hour)
	c.Check(cond.duration, check.Equals, 0*time.Second)

	var now time.Time
	var active bool
	var delay time.Duration

	// Even at start/end time, we never actually fire.
	now = time.Date(2014, time.June, 12, 11, 00, 00, 0, time.Local)
	active, delay = cond.findFireState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 24*time.Hour)

	cond.Stop()
}
