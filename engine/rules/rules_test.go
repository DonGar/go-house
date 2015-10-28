package rules

import (
	"fmt"
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/status"
	"github.com/DonGar/go-house/stoppable"
	"gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

//
// Action Mocks
//

type mockActions struct {
	registrar       *actions.Manager
	onCount         int
	offCount        int
	errorCount      int
	actionOnBody    *status.Status
	actionOffBody   *status.Status
	actionErrorBody *status.Status
}

func (m *mockActions) actionOn(s *status.Status, action *status.Status) (e error) {
	m.onCount += 1
	return nil
}

func (m *mockActions) actionOff(s *status.Status, action *status.Status) (e error) {
	m.offCount += 1
	return nil
}

func (m *mockActions) actionError(s *status.Status, action *status.Status) (e error) {
	m.errorCount += 1
	return fmt.Errorf("Mock Error")
}

func (m mockActions) verify(c *check.C, on, off, err int) {
	c.Check(m.onCount, check.Equals, on)
	c.Check(m.offCount, check.Equals, off)
	c.Check(m.errorCount, check.Equals, err)
}

func newMockActions() *mockActions {
	results := &mockActions{}

	results.registrar = actions.NewManager()
	results.registrar.RegisterAction("on", results.actionOn)
	results.registrar.RegisterAction("off", results.actionOff)
	results.registrar.RegisterAction("error", results.actionError)

	results.actionOnBody = &status.Status{}
	results.actionOnBody.Set("status://action", "on", 0)

	results.actionOffBody = &status.Status{}
	results.actionOffBody.Set("status://action", "off", 0)

	results.actionErrorBody = &status.Status{}
	results.actionErrorBody.Set("status://action", "error", 0)

	return results
}

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

func (suite *MySuite) TestRuleStartStop(c *check.C) {
	mockStatus := &status.Status{}
	mockActions := newMockActions()
	ruleBody := &status.Status{}

	e := ruleBody.SetJson(
		"status://",
		[]byte(`{
			"condition": {
				"test": "base"
			},
			"on": null
		}`),
		status.UNCHECKED_REVISION)
	c.Assert(e, check.IsNil)

	rule, e := NewRule(mockStatus, mockActions.registrar, "Test Rule", ruleBody)
	c.Assert(e, check.IsNil)

	rule.Stop()

	mockActions.verify(c, 0, 0, 0)
}

func (suite *MySuite) TestRuleOnFireSingle(c *check.C) {
	s := &status.Status{}
	mockActions := newMockActions()
	mockCondition := &mockCondition{make(chan bool)}

	rule := &Rule{
		s,
		mockActions.registrar,
		"Test Rule Single",
		mockCondition,
		mockActions.actionOnBody,
		mockActions.actionOffBody,
		stoppable.NewBase()}

	rule.start()

	mockCondition.result <- true
	mockCondition.result <- false

	rule.Stop()

	mockActions.verify(c, 1, 1, 0)
}

func (suite *MySuite) TestRuleOnOffFireRepeated(c *check.C) {
	s := &status.Status{}
	mockActions := newMockActions()
	mockCondition := &mockCondition{make(chan bool)}

	rule := &Rule{
		s,
		mockActions.registrar,
		"Test Rule Repeated",
		mockCondition,
		mockActions.actionOnBody,
		mockActions.actionOffBody,
		stoppable.NewBase()}

	rule.start()

	// Send several patterns of true/false since conditions are not required
	// to toggle rationally.
	mockCondition.result <- true
	mockCondition.result <- false
	mockCondition.result <- true
	mockCondition.result <- false

	rule.Stop()

	mockActions.verify(c, 2, 2, 0)
}

func (suite *MySuite) TestRuleOnActionOnly(c *check.C) {
	s := &status.Status{}
	mockActions := newMockActions()
	mockCondition := &mockCondition{make(chan bool)}

	rule := &Rule{
		s,
		mockActions.registrar,
		"Test Rule OnActionOnly",
		mockCondition,
		mockActions.actionOnBody,
		nil,
		stoppable.NewBase()}

	rule.start()

	// Send several patterns of true/false since conditions are not required
	// to toggle rationally.
	mockCondition.result <- true
	mockCondition.result <- false
	mockCondition.result <- true
	mockCondition.result <- false

	rule.Stop()

	mockActions.verify(c, 2, 0, 0)
}

func (suite *MySuite) TestRuleOffActionOnly(c *check.C) {
	s := &status.Status{}
	mockActions := newMockActions()
	mockCondition := &mockCondition{make(chan bool)}

	rule := &Rule{
		s,
		mockActions.registrar,
		"Test Rule OffActionOnly",
		mockCondition,
		nil,
		mockActions.actionOffBody,
		stoppable.NewBase()}

	rule.start()

	// Send several patterns of true/false since conditions are not required
	// to toggle rationally.
	mockCondition.result <- true
	mockCondition.result <- false
	mockCondition.result <- true
	mockCondition.result <- false

	rule.Stop()

	mockActions.verify(c, 0, 2, 0)
}

func (suite *MySuite) TestRuleFireError(c *check.C) {
	s := &status.Status{}
	mockActions := newMockActions()
	mockCondition := &mockCondition{make(chan bool)}

	rule := &Rule{
		s,
		mockActions.registrar,
		"Test Rule Error",
		mockCondition,
		mockActions.actionErrorBody,
		nil,
		stoppable.NewBase()}

	rule.start()

	// Send several actions to trigger errors.
	mockCondition.result <- true
	mockCondition.result <- false
	mockCondition.result <- true
	mockCondition.result <- false

	rule.Stop()

	mockActions.verify(c, 0, 0, 2)
}
