package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

// Helpers for all condition test code.
func validateChannelRead(c *check.C, cond Condition, expected bool) {
	result, ok := <-cond.Result()

	c.Check(ok, check.Equals, true)
	c.Check(result, check.Equals, expected)
}

func validateChannelEmpty(c *check.C, cond Condition) {
	select {
	case result := <-cond.Result():
		c.Error("Got unexpected result: ", result)
	default:
	}
}

// Base condition tests.
func (suite *MySuite) TestBaseConditionStartStop(c *check.C) {
	s := &status.Status{}

	cBody := &status.Status{}
	e := cBody.Set(
		"status://",
		map[string]interface{}{"test": "base"},
		0)
	c.Assert(e, check.IsNil)

	cond, e := NewCondition(s, cBody)
	c.Check(e, check.IsNil)
	c.Check(cond.Result(), check.NotNil)

	cond.Stop()
}

func (suite *MySuite) TestBadConditionTest(c *check.C) {
	s := &status.Status{}

	cBody := &status.Status{}
	e := cBody.Set(
		"status://",
		map[string]interface{}{"test": "bogus"},
		0)
	c.Assert(e, check.IsNil)

	cond, e := NewCondition(s, cBody)
	c.Check(cond, check.IsNil)
	c.Check(e, check.NotNil)
}
