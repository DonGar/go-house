package rules

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

func setupPeriodicRule(c *check.C, time string) (rule, *mockActionHelper) {
	mockAh := &mockActionHelper{}

	s := &status.Status{}

	body := &status.Status{}
	e := body.Set("status://", map[string]interface{}{"interval": time}, 0)
	c.Assert(e, check.IsNil)

	b := base{s, mockAh.helper, "TestPeriodicRule", 3, body}

	r, e := newPeriodicRule(b)
	c.Assert(e, check.IsNil)

	return r, mockAh
}

func (suite *MySuite) TestPeriodicStartStop(c *check.C) {
	r, mockAh := setupPeriodicRule(c, "1s")

	e := r.Stop()
	c.Check(e, check.IsNil)

	// The rule shouldn't have fired.
	c.Check(mockAh.fireCount, check.Equals, 0)
}

func (suite *MySuite) TestPeriodicFire(c *check.C) {
	r, mockAh := setupPeriodicRule(c, "2ms")

	time.Sleep(3 * time.Millisecond)

	e := r.Stop()
	c.Check(e, check.IsNil)

	// The rule shouldn't have fired.
	c.Check(mockAh.fireCount, check.Equals, 1)
}
