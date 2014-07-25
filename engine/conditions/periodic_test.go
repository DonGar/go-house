package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

func setupPeriodicRule(c *check.C, time string) *periodicCondition {
	body := &status.Status{}
	e := body.Set("status://", map[string]interface{}{"interval": time}, 0)
	c.Assert(e, check.IsNil)

	cond, e := newPeriodicCondition(body)
	c.Assert(e, check.IsNil)

	return cond
}

func (suite *MySuite) TestPeriodicStartStop(c *check.C) {
	cond := setupPeriodicRule(c, "1s")
	cond.Stop()
}

func (suite *MySuite) TestPeriodicFire(c *check.C) {

	start := time.Now()

	cond := setupPeriodicRule(c, "2ms")

	validateChannelRead(c, cond, true)
	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	// We've seen the end of a 2 ms periodic cycle.
	duration := time.Since(start)
	c.Check(duration < 3*time.Millisecond, check.Equals, true)

	validateChannelRead(c, cond, true)
	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	// We've seen the end of a second 2 ms periodic cycle.
	duration = time.Since(start)
	c.Check(duration < 5*time.Millisecond, check.Equals, true)

	cond.Stop()
}
