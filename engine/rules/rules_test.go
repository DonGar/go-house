package rules

import (
	"fmt"
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/status"
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
	registrar         actions.ActionRegistrar
	successCount      int
	errorCount        int
	actionSuccessBody *status.Status
	actionErrorBody   *status.Status
}

func (m *mockActions) actionSuccess(r actions.ActionRegistrar, s *status.Status, action *status.Status) (e error) {
	m.successCount += 1
	return nil
}

func (m *mockActions) actionError(r actions.ActionRegistrar, s *status.Status, action *status.Status) (e error) {
	m.errorCount += 1
	return fmt.Errorf("Mock Error")
}

func newMockActions() *mockActions {
	results := &mockActions{}

	results.registrar = actions.ActionRegistrar{
		"success": results.actionSuccess,
		"error":   results.actionError}

	results.actionSuccessBody = &status.Status{}
	results.actionSuccessBody.Set("status://action", "success", 0)

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
			"action": null
		}`),
		status.UNCHECKED_REVISION)
	c.Assert(e, check.IsNil)

	rule, e := NewRule(mockStatus, mockActions.registrar, "Test Rule", ruleBody)
	c.Assert(e, check.IsNil)

	rule.Stop()

	c.Check(mockActions.successCount, check.Equals, 0)
}

func (suite *MySuite) TestRuleFireSingle(c *check.C) {
	s := &status.Status{}
	mockActions := newMockActions()
	mockCondition := &mockCondition{make(chan bool)}

	rule := &Rule{
		s,
		mockActions.registrar,
		"Test Rule Single",
		mockCondition,
		mockActions.actionSuccessBody,
		make(chan bool)}

	rule.start()

	mockCondition.result <- true

	rule.Stop()

	c.Check(mockActions.successCount, check.Equals, 1)
}

func (suite *MySuite) TestRuleFireRepeated(c *check.C) {
	s := &status.Status{}
	mockActions := newMockActions()
	mockCondition := &mockCondition{make(chan bool)}

	rule := &Rule{
		s,
		mockActions.registrar,
		"Test Rule Repeated",
		mockCondition,
		mockActions.actionSuccessBody,
		make(chan bool)}

	rule.start()

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

	c.Check(mockActions.successCount, check.Equals, 4)
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
		make(chan bool)}

	rule.start()

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

	c.Check(mockActions.errorCount, check.Equals, 4)
}
