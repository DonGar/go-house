package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
)

// Actual tests.

func validateAndParseStartStop(c *check.C, conditionsJson string, expectedValue bool) {
	s := &status.Status{}

	body := &status.Status{}
	e := body.Set("status://test", "and", 0)
	e = body.SetJson("status://conditions", []byte(conditionsJson), 1)

	cond, e := NewCondition(s, body)
	c.Assert(e, check.IsNil)

	validateChannelRead(c, cond, expectedValue)
	validateChannelEmpty(c, cond)

	cond.Stop()
}

func (suite *MySuite) TestAndStartStop(c *check.C) {
	validateAndParseStartStop(c, "[]", true)
	validateAndParseStartStop(c, `[
		  {
		  	"test": "watch",
        "trigger": "1",
        "watch": "status://iogear/iogear/desktop/active"
		  }
		]`, false)
	validateAndParseStartStop(c, `[
		  {
		  	"test": "watch",
        "trigger": "1",
        "watch": "status://iogear/iogear/desktop/active"
		  },
		  {
		  	"test": "watch",
        "trigger": "2",
        "watch": "status://iogear/iogear/desktop/active"
		  }
		]`, false)
}

func (suite *MySuite) TestAndOneMock(c *check.C) {
	s := &status.Status{}
	mockCond := &mockCondition{make(chan bool)}

	conditionValues := []conditionValue{{mockCond, false}}

	cond := &andCondition{newBase(s), conditionValues}
	cond.start()

	validateChannelEmpty(c, cond)

	mockCond.result <- true

	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	mockCond.result <- false

	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	mockCond.result <- true

	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	cond.Stop()
}

func (suite *MySuite) TestAndThreeMock(c *check.C) {
	s := &status.Status{}
	mockCond := []*mockCondition{
		{make(chan bool)},
		{make(chan bool)},
		{make(chan bool)},
	}

	conditionValues := make([]conditionValue, len(mockCond))
	for i := range mockCond {
		conditionValues[i] = conditionValue{mockCond[i], false}
	}

	cond := &andCondition{newBase(s), conditionValues}
	cond.start()

	validateChannelEmpty(c, cond)

	mockCond[0].result <- true
	validateChannelEmpty(c, cond)

	mockCond[1].result <- true
	validateChannelEmpty(c, cond)

	mockCond[2].result <- true
	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	mockCond[1].result <- false
	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	mockCond[0].result <- false
	validateChannelEmpty(c, cond)

	mockCond[1].result <- true
	validateChannelEmpty(c, cond)

	mockCond[0].result <- true
	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	cond.Stop()
}
