package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

func (suite *MySuite) TestAfterStartStop(c *check.C) {
	good := []string{
		`{
			  "test": "after",
			  "delay": "2ms",
			  "condition": {
			    "test": "false"
			  }
		}`,
		`{
			  "test": "after",
			  "delay": "24h",
			  "condition": []
		}`,
	}

	for _, g := range good {
		validateConditionJson(c, "", g, false)
	}

	bad := []string{
		`{
			  "test": "after",
			  "delay": "2s"
		}`,
		`{
			  "test": "after",
			  "condition": []
		}`,
		`{
			  "test": "after",
			  "delay": "2",
			  "condition": {
			    "test": "true"
			  }
		}`,
		`{
			  "test": "after",
			  "delay": "24hour",
			  "condition": [{}]
		}`,
	}

	for _, b := range bad {
		validateConditionBadJson(c, b)
	}
}

func (suite *MySuite) TestAfterMock(c *check.C) {
	s := &status.Status{}
	mockCond := &mockCondition{make(chan bool)}

	cond := &afterCondition{
		newBase(s), mockCond, 10 * time.Millisecond}
	cond.start()

	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	// Set the source to true, we become true, after the delay.
	mockCond.result <- true

	// We should go true, but only after a delay.
	validateChannelEmptyInstant(c, cond)
	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	// Set the source to false, we become false (right away, but unproven)
	mockCond.result <- false

	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	// Setting the cond to true, then false before delay should yield no update.
	mockCond.result <- true
	mockCond.result <- false

	validateChannelEmpty(c, cond)

	// Set the source to true, we become true, after the delay.
	mockCond.result <- true

	validateChannelEmptyInstant(c, cond)
	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	cond.Stop()
}
