package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
)

// Mock condition for mocking inner conditions.

type mockCondition struct {
	result chan bool
}

func (m *mockCondition) Result() <-chan bool {
	return m.result
}

func (m *mockCondition) Stop() {
}

// Actual tests.

func validateParseStartStop(c *check.C, conditionsJson string, trueStart bool) {
	s := &status.Status{}

	body := &status.Status{}
	e := body.Set("status://test", "and", 0)
	e = body.SetJson("status://conditions", []byte(conditionsJson), 1)

	cond, e := NewCondition(s, body)
	c.Assert(e, check.IsNil)

	if trueStart {
		validateChannelRead(c, cond, true)
	}
	validateChannelEmpty(c, cond)

	cond.Stop()
}

func (suite *MySuite) TestAndStartStop(c *check.C) {
	validateParseStartStop(c, "[]", true)
	validateParseStartStop(c, `[
		  {
		  	"test": "watch",
        "trigger": "1",
        "watch": "status://iogear/iogear/desktop/active"
		  }
		]`, false)
	validateParseStartStop(c, `[
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

	conditionValues := []conditionValue{conditionValue{mockCond, false}}

	cond := &andCondition{base{s, make(chan bool), make(chan bool)}, false, conditionValues}
	cond.start()

	validateChannelEmpty(c, cond)

	mockCond.result <- true

	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	mockCond.result <- false

	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	cond.Stop()
}

func (suite *MySuite) TestAndThreeMock(c *check.C) {
	s := &status.Status{}
	mockCond := []*mockCondition{
		&mockCondition{make(chan bool)},
		&mockCondition{make(chan bool)},
		&mockCondition{make(chan bool)},
	}

	conditionValues := make([]conditionValue, len(mockCond))
	for i := range mockCond {
		conditionValues[i] = conditionValue{mockCond[i], false}
	}

	cond := &andCondition{base{s, make(chan bool), make(chan bool)}, false, conditionValues}
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
