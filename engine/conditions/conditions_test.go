package conditions

import (
	"fmt"
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

//
// Helpers for all condition test code.
//
func validateChannelRead(c *check.C, cond Condition, expected bool) {
	result, ok := <-cond.Result()

	c.Assert(ok, check.Equals, true)
	c.Assert(result, check.Equals, expected)
}

func validateChannelEmpty(c *check.C, cond Condition) {
	select {
	case result := <-cond.Result():
		c.Fatal("Result() channel not empty: ", result)
	default:
	}
}

//
// Condition Parsing Tests
//

func validateConditionJson(c *check.C, statusJson, condJson string, expectedValue bool) {
	fmt.Println("Running validateConditionJson: ", condJson)

	s := &status.Status{}
	if statusJson != "" {
		e := s.SetJson("status://", []byte(statusJson), 0)
		c.Assert(e, check.IsNil)
	}

	body := &status.Status{}
	e := body.SetJson("status://", []byte(condJson), 0)
	c.Assert(e, check.IsNil)

	cond, e := NewCondition(s, body)
	c.Check(e, check.IsNil)

	if expectedValue {
		validateChannelRead(c, cond, true)
	}
	validateChannelEmpty(c, cond)

	cond.Stop()
}

func validateConditionBadJson(c *check.C, condJson string) {
	fmt.Println("Running validateConditionBadJson: ", condJson)

	s := &status.Status{}

	body := &status.Status{}
	e := body.SetJson("status://", []byte(condJson), 0)
	c.Assert(e, check.IsNil)

	cond, e := NewCondition(s, body)

	c.Check(cond, check.IsNil)
	c.Check(e, check.NotNil)
}

func (suite *MySuite) TestBadConditionParsing(c *check.C) {
	validateConditionBadJson(c, `"foo"`)
	validateConditionBadJson(c, `"status://bogus"`)
	validateConditionBadJson(c, `["foo"]`)
	validateConditionBadJson(c, `[{}]`)
	validateConditionBadJson(c, `["status://bogus"]`)
	validateConditionBadJson(c, `{"test": "bogus"}`)
	validateConditionBadJson(c, `[{"test": "base"}, "status://bogus"]`)
	validateConditionBadJson(c, `{"foo": "bar"}`)
}

func (suite *MySuite) TestConditionStartStop(c *check.C) {
	statusJson := `{
		"base": { "test": "base" },
		"list": [{ "test": "base" }, { "test": "base" }],
		"rlist": ["status://base", "status://list", { "test": "base" }],
		"empty": []
	}`

	validateConditionJson(c, statusJson, `"status://base"`, false)
	validateConditionJson(c, statusJson, `"status://list"`, false)
	validateConditionJson(c, statusJson, `"status://rlist"`, false)
	validateConditionJson(c, statusJson, `"status://empty"`, true)

	validateConditionJson(c, statusJson, `[]`, true)
	validateConditionJson(c, statusJson, `[{ "test": "base" }]`, false)
	validateConditionJson(c, statusJson, `[{ "test": "base" }, { "test": "base" }, { "test": "base" }]`, false)

	validateConditionJson(c, statusJson, `{ "test": "base" }`, false)
}

//
// Base condition tests.
//

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
