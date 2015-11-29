package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
)

func setupConstCondition(c *check.C, result bool) *constCondition {
	s := &status.Status{}
	body := &status.Status{}

	cond, e := newConstCondition(s, body, result)
	c.Assert(e, check.IsNil)

	return cond
}

func (suite *MySuite) TestConstStartStopTrue(c *check.C) {
	cond := setupConstCondition(c, true)
	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	cond.Stop()
}

func (suite *MySuite) TestConstStartStopFalse(c *check.C) {
	cond := setupConstCondition(c, false)
	validateChannelEmpty(c, cond)
	cond.Stop()
}
