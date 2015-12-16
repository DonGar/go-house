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
	good := []string{
		`{
		      "test": "periodic",
		      "interval": "200ms"
		  }`,
		`{
        "test": "periodic",
        "interval": "24h"
    }`,
	}

	for _, g := range good {
		validateConditionJson(c, "", g, false)
	}

	bad := []string{
		`{
	       "test": "periodic"
	   }`,
		`{
	       "test": "periodic",
	       "interval": []
	   }`,
		`{
	       "test": "periodic",
	       "interval": "2"
	   }`,
	}

	for _, b := range bad {
		validateConditionBadJson(c, b)
	}
}

func (suite *MySuite) TestPeriodicFire(c *check.C) {

	// Seems to be long enough to be safe when all tests are run in parallel.
	period := 5 * EMPTY_DELAY

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
