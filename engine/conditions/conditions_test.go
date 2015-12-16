package conditions

import (
	"fmt"
	"github.com/DonGar/go-house/status"
	"github.com/DonGar/go-house/wait"
	"gopkg.in/check.v1"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

const (
	READ_DELAY  time.Duration = 100 * time.Millisecond
	EMPTY_DELAY time.Duration = 5 * time.Millisecond
)

//
// Helpers for all condition test code.
//
func channelRead(cond Condition) (result, received bool) {
	select {
	case result = <-cond.Result():
		return result, true
	default:
		return false, false
	}
}

func validateChannelRead(c *check.C, cond Condition, expected bool) {
	result, received := false, false

	updateObserved := func() bool {
		result, received = channelRead(cond)
		return received
	}

	// Retry until there is a value on the channel.
	wait.Wait(READ_DELAY, updateObserved)

	c.Check(received, check.Equals, true)
	c.Check(result, check.Equals, expected)
}

func validateChannelEmpty(c *check.C, cond Condition) {
	for {
		// Sleep a little for things to see if extra results are generated.
		time.Sleep(EMPTY_DELAY)
		result, received := channelRead(cond)
		if received {
			c.Error("Unexpected result received: ", result)
		} else {
			return
		}
	}
}

func validateChannelEmptyInstant(c *check.C, cond Condition) {
	// Validate that the channel is empty right now, but don't wait for things
	// to settle. Meant for testing things with timers.
	result, received := channelRead(cond)
	if received {
		c.Error("Channel has a value queued: ", result)
	}
}

// Mock condition for mocking inner conditions.
type mockCondition struct {
	result chan bool
}

func (m *mockCondition) Result() <-chan bool {
	return m.result
}

func (m *mockCondition) Stop() {
}

//
// Condition Parsing Tests
//

func constructConditionJson(c *check.C, statusJson, condJson string) (cond Condition, err error) {
	s := &status.Status{}
	if statusJson != "" {
		e := s.SetJson("status://", []byte(statusJson), 0)
		c.Assert(e, check.IsNil)
	}

	body := &status.Status{}
	e := body.SetJson("status://", []byte(condJson), 0)
	c.Assert(e, check.IsNil)

	return NewCondition(s, body)
}

func validateConditionJson(c *check.C, statusJson, condJson string, expected bool) {
	fmt.Println("Testing Good Json: ", condJson)
	cond, e := constructConditionJson(c, statusJson, condJson)
	c.Check(e, check.IsNil)

	validateChannelRead(c, cond, expected)
	validateChannelEmpty(c, cond)
	cond.Stop()
}

func validateConditionBadJson(c *check.C, condJson string) {
	fmt.Println("Testing Bad Json: ", condJson)
	cond, e := constructConditionJson(c, "{}", condJson)

	c.Check(cond, check.IsNil)
	c.Check(e, check.NotNil)
}

func (suite *MySuite) TestConditionParsingBad(c *check.C) {
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
    "true": { "test": "true" },
    "false": { "test": "false" },
		"list": [{ "test": "true" }, { "test": "true" }],
    "rlist": ["status://true", "status://list", { "test": "true" }],
    "flist": ["status://true", "status://rlist", { "test": "false" }],
		"empty": []
	}`

	validateConditionJson(c, statusJson, `{ "test": "true" }`, true)
	validateConditionJson(c, statusJson, `{ "test": "false" }`, false)

	validateConditionJson(c, statusJson, `[]`, true)
	validateConditionJson(c, statusJson, `[{ "test": "true" }]`, true)
	validateConditionJson(c, statusJson, `[{ "test": "false" }, { "test": "false" }, { "test": "false" }]`, false)

	validateConditionJson(c, statusJson, `"status://true"`, true)
	validateConditionJson(c, statusJson, `"status://false"`, false)
	validateConditionJson(c, statusJson, `"status://list"`, true)
	validateConditionJson(c, statusJson, `"status://rlist"`, true)
	validateConditionJson(c, statusJson, `"status://flist"`, false)
	validateConditionJson(c, statusJson, `"status://empty"`, true)
}

//
// Base condition tests.
//
