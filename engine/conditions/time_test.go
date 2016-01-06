package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

func setupTimeCondition(c *check.C, time, duration string) *timeCondition {
	s := &status.Status{}

	body := &status.Status{}
	err := body.Set("status://time", time, 0)
	c.Assert(err, check.IsNil)

	if duration != "" {
		err := body.Set("status://duration", duration, 1)
		c.Assert(err, check.IsNil)
	}

	cond, err := newTimeCondition(s, body)
	c.Assert(err, check.IsNil)

	return cond
}

func (suite *MySuite) TestTimeStartStop(c *check.C) {
	// The 0 duration, means the condition is always false.
	cond := setupTimeCondition(c, "2000/1/1 12:00", "0s")

	c.Check(cond.duration, check.Equals, 0*time.Second)

	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)
	cond.Stop()
}

func (suite *MySuite) TestTimeFindActiveState(c *check.C) {
	cond := setupTimeCondition(c, "11:00", "0s")

	var now time.Time
	var active bool
	var delay time.Duration

	// Almost start time.
	now = time.Date(2014, time.June, 12, 10, 57, 12, 0, time.Local)
	active, delay = cond.findActiveState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 0*time.Second)

	// Start time.
	now = time.Date(2014, time.June, 12, 11, 00, 00, 0, time.Local)
	active, delay = cond.findActiveState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 0*time.Second)

	// After time.
	now = time.Date(2014, time.June, 12, 11, 00, 0, 1, time.Local)
	active, delay = cond.findActiveState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 0*time.Second)

	// When now is midnight.
	now = time.Date(2014, time.June, 12, 0, 0, 0, 0, time.Local)
	active, delay = cond.findActiveState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 0*time.Second)

	cond.Stop()
}

func (suite *MySuite) TestTimeFindActiveStateDuration(c *check.C) {
	cond := setupTimeCondition(c, "11:00", "12s")

	// Verify that time/duration were parsed correctly.
	// c.Check(cond.timeOfDay, check.Equals, 11*time.Hour)
	c.Check(cond.duration, check.Equals, 12*time.Second)

	var now time.Time
	var active bool
	var delay time.Duration

	// Almost start time.
	now = time.Date(2014, time.June, 12, 10, 57, 12, 0, time.Local)
	active, delay = cond.findActiveState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 0*time.Second)

	// Start time.
	now = time.Date(2014, time.June, 12, 11, 00, 00, 0, time.Local)
	active, delay = cond.findActiveState(now)
	c.Check(active, check.Equals, true)
	c.Check(delay, check.Equals, 12*time.Second)

	// Almost end time.
	now = time.Date(2014, time.June, 12, 11, 00, 11, 0, time.Local)
	active, delay = cond.findActiveState(now)
	c.Check(active, check.Equals, true)
	c.Check(delay, check.Equals, 1*time.Second)

	// End time.
	now = time.Date(2014, time.June, 12, 11, 00, 12, 0, time.Local)
	active, delay = cond.findActiveState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 0*time.Second)

	// After end time.
	now = time.Date(2014, time.June, 12, 11, 01, 00, 0, time.Local)
	active, delay = cond.findActiveState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 0*time.Second)

	// When now is midnight.
	now = time.Date(2014, time.June, 12, 0, 0, 0, 0, time.Local)
	active, delay = cond.findActiveState(now)
	c.Check(active, check.Equals, false)
	c.Check(delay, check.Equals, 0*time.Second)

	cond.Stop()
}
