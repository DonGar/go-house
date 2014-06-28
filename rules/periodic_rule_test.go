package rules

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

func (suite *MySuite) TestPeriodicStartStop(c *check.C) {
	body := &status.Status{}
	e := body.SetJson("status://",
		[]byte(`
    {
			"interval": "1s"
    }`),
		0)
	c.Assert(e, check.IsNil)

	b := base{nil, "TestPeriodicRuleStartStop", 3, body}

	r, e := newPeriodicRule(b)
	c.Assert(e, check.IsNil)

	e = r.Stop()
	c.Assert(e, check.IsNil)
}

func (suite *MySuite) TestPeriodicFire(c *check.C) {
	body := &status.Status{}
	e := body.SetJson("status://",
		[]byte(`
    {
			"interval": "2ms"
    }`),
		0)
	c.Assert(e, check.IsNil)

	b := base{nil, "TestPeriodicRuleRepeating", 3, body}

	r, e := newPeriodicRule(b)
	c.Assert(e, check.IsNil)

	time.Sleep(3 * time.Millisecond)

	e = r.Stop()
	c.Assert(e, check.IsNil)
}
