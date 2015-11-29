package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

func setupPeriodicRule(c *check.C, time string) *periodicCondition {
	s := &status.Status{}
	body := &status.Status{}
	e := body.Set("status://", map[string]interface{}{"interval": time}, 0)
	c.Assert(e, check.IsNil)

	cond, e := newPeriodicCondition(s, body)
	c.Assert(e, check.IsNil)

	return cond
}

func (suite *MySuite) TestPeriodicStartStop(c *check.C) {
	cond := setupPeriodicRule(c, "1s")
	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)
	cond.Stop()
}

func (suite *MySuite) TestPeriodicFire(c *check.C) {

	// 10 ms seems to be long enough to be safe when all tests are run in
	// parallel.
	period := 50 * time.Millisecond

	start := time.Now()

	cond := setupPeriodicRule(c, period.String())

	validateChannelRead(c, cond, false)
	validateChannelRead(c, cond, true)
	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	// We've seen the end of a periodic cycle.
	duration := time.Since(start)
	c.Check(duration < 2*period, check.Equals, true)

	validateChannelRead(c, cond, true)
	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	// We've seen the end of a second periodic cycle.
	duration = time.Since(start)
	c.Check(duration < 3*period, check.Equals, true)

	cond.Stop()
}
