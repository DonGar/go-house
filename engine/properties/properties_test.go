package properties

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

//
// Condition Mocks
//

type mockCondition struct {
	result chan bool
}

func (m *mockCondition) Result() <-chan bool {
	return m.result
}

func (m *mockCondition) Stop() {
}

// Tests

func propertyParseStartStop(c *check.C, bodyStr string) error {
	s := &status.Status{}
	body := &status.Status{}

	e := body.SetJson("status://", []byte(bodyStr), status.UNCHECKED_REVISION)
	c.Assert(e, check.IsNil)

	rule, e := NewProperty(s, "Test Property", body)
	if e != nil {
		return e
	}

	rule.Stop()
	return nil
}

func (suite *MySuite) TestPropertyStartStop(c *check.C) {

	goodBodies := []string{
		`{
			"target": "status://target",
			"values": []
		}`,
		`{
			"target": "status://target",
			"values": [],
			"default": "bar"
		}`,
		`{
			"target": "status://target",
			"values": [
			  { "condition": { "test": "base" }, "value": "foo"	}
			],
			"default": "bar"
		}`,
		`{
			"target": "status://target",
			"values": [
			  { "condition": { "test": "base" }, "value": "foo"	},
			  { "condition": { "test": "base" }, "value": "bar"	},
			  { "condition": { "test": "base" }, "value": "baz"	}
			]
		}`,
	}

	for _, body := range goodBodies {
		e := propertyParseStartStop(c, body)
		c.Check(e, check.IsNil)
	}

	badBodies := []string{
		// No target.
		`{}`,
		// No values.
		`{
			"target": "status://target"
 			}`,
		// Bad values.
		`{
			"target": "status://target",
			"values": "not an array"
			}`,
		// No condition in values.
		`{
			"target": "status://target",
			"values": [
			  { "value": "foo"	}
			]
			}`,
		// No value in values.
		`{
			"target": "status://target",
			"values": [
			  { "condition": { "test": "base" }	}
			]
			}`,
		// Invalid condition in values.
		`{
			"target": "status://target",
			"values": [
			  { "condition": { "test": "unknown" }, "value": "foo"	}
			]
			}`,
	}

	for _, body := range badBodies {
		e := propertyParseStartStop(c, body)
		c.Check(e, check.NotNil)
	}
}

func validateTarget(c *check.C, s *status.Status, expected string) {
	// We give the watcher a little time for a delayed update.
	time.Sleep(1 * time.Millisecond)
	target, e := s.GetString("status://target")
	c.Check(e, check.IsNil)
	c.Check(target, check.Equals, expected)
}

func (suite *MySuite) TestPropertyFireDefault(c *check.C) {
	s := &status.Status{}

	p := &Property{
		s,
		"Test Property Default",
		"status://target",
		[]conditionValue{},
		true,
		"default set",
		make(chan bool),
	}

	p.start()
	validateTarget(c, s, "default set")

	p.Stop()
}

func (suite *MySuite) TestPropertyFireCondition(c *check.C) {
	s := &status.Status{}
	mockCond := &mockCondition{make(chan bool)}

	p := &Property{
		s,
		"Test Property Default",
		"status://target",
		[]conditionValue{conditionValue{mockCond, false, "condition"}},
		true,
		"default set",
		make(chan bool),
	}

	p.start()
	validateTarget(c, s, "default set")

	mockCond.result <- true
	validateTarget(c, s, "condition")

	mockCond.result <- false
	validateTarget(c, s, "default set")

	p.Stop()
}

func (suite *MySuite) TestPropertyFireConditionMultiple(c *check.C) {
	s := &status.Status{}
	mockCond := []*mockCondition{
		&mockCondition{make(chan bool)},
		&mockCondition{make(chan bool)},
		&mockCondition{make(chan bool)}}

	p := &Property{
		s,
		"Test Property Default",
		"status://target",
		[]conditionValue{
			conditionValue{mockCond[0], false, "condition 0"},
			conditionValue{mockCond[1], false, "condition 1"},
			conditionValue{mockCond[2], false, "condition 2"},
		},
		true,
		"default set",
		make(chan bool),
	}

	p.start()
	validateTarget(c, s, "default set")

	mockCond[0].result <- true
	validateTarget(c, s, "condition 0")

	mockCond[2].result <- true
	validateTarget(c, s, "condition 0")

	mockCond[1].result <- true
	validateTarget(c, s, "condition 0")

	mockCond[0].result <- false
	validateTarget(c, s, "condition 1")

	mockCond[2].result <- false
	validateTarget(c, s, "condition 1")

	mockCond[1].result <- false
	validateTarget(c, s, "default set")

	mockCond[2].result <- true
	validateTarget(c, s, "condition 2")

	p.Stop()
}

func (suite *MySuite) TestPropertyFireConditionMultipleNoDefault(c *check.C) {
	s := &status.Status{}
	mockCond := []*mockCondition{
		&mockCondition{make(chan bool)},
		&mockCondition{make(chan bool)},
		&mockCondition{make(chan bool)}}

	p := &Property{
		s,
		"Test Property Default",
		"status://target",
		[]conditionValue{
			conditionValue{mockCond[0], false, "condition 0"},
			conditionValue{mockCond[1], false, "condition 1"},
		},
		false,
		"",
		make(chan bool),
	}

	p.start()

	// Make sure default is not created at start up.
	_, e := s.GetString("status://target")
	c.Check(e, check.NotNil)

	mockCond[1].result <- true
	validateTarget(c, s, "condition 1")

	mockCond[0].result <- true
	validateTarget(c, s, "condition 0")

	mockCond[0].result <- false
	validateTarget(c, s, "condition 1")

	// Make sure turning off all conditions leaves target at last value.
	mockCond[1].result <- false
	validateTarget(c, s, "condition 1")

	p.Stop()
}
