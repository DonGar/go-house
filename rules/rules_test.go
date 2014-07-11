package rules

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

//
// Test Mocks
//

type mockActionHelper struct {
	fireCount  int
	lastAction *status.Status
}

func (m *mockActionHelper) helper(action *status.Status) {
	m.fireCount += 1
	m.lastAction = action
}

type mockCondition struct {
	result chan bool
}

func (m *mockCondition) Result() <-chan bool {
	return m.result
}

func (m *mockCondition) Stop() {
}

// Tests

func (suite *MySuite) TestRuleStartStop(c *check.C) {
	mockAH := &mockActionHelper{}
	actionBody := &status.Status{}
	mockCondition := &mockCondition{}

	rule := newRule(mockAH.helper, "Test Rule", 23, actionBody, mockCondition)

	rule.Stop()

	c.Check(mockAH.fireCount, check.Equals, 0)
}

func (suite *MySuite) TestRuleFireSingle(c *check.C) {
	mockAH := &mockActionHelper{}
	actionBody := &status.Status{}
	mockCondition := &mockCondition{make(chan bool)}

	rule := newRule(mockAH.helper, "Test Rule", 23, actionBody, mockCondition)

	mockCondition.result <- true

	rule.Stop()

	c.Check(mockAH.fireCount, check.Equals, 1)
	c.Check(mockAH.lastAction, check.Equals, actionBody)
}

func (suite *MySuite) TestRuleFireRepeated(c *check.C) {
	mockAH := &mockActionHelper{}
	actionBody := &status.Status{}
	mockCondition := &mockCondition{make(chan bool)}

	rule := newRule(mockAH.helper, "Test Rule", 23, actionBody, mockCondition)

	// Send several patterns of true/false since conditions are not required
	// to toggle rationally.
	mockCondition.result <- true
	mockCondition.result <- false
	mockCondition.result <- false

	mockCondition.result <- true
	mockCondition.result <- true

	mockCondition.result <- true
	mockCondition.result <- false

	rule.Stop()

	c.Check(mockAH.fireCount, check.Equals, 4)
	c.Check(mockAH.lastAction, check.Equals, actionBody)
}
