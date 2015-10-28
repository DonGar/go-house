package actions

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

type mockActionResults struct {
	successCalls int
	failCalls    int
	httpCalls    int
	httpUrls     []string
}

const MOCK_FAILURE_MSG = "Mock Action Failed."

func (m *mockActionResults) success(s *status.Status, action *status.Status) error {
	m.successCalls += 1
	return nil
}

func (m *mockActionResults) fail(s *status.Status, action *status.Status) error {
	m.failCalls += 1
	return fmt.Errorf(MOCK_FAILURE_MSG)
}

func (m *mockActionResults) fetch(s *status.Status, action *status.Status) error {
	url, _, e := action.GetString("status://url")
	if e != nil {
		return e
	}

	m.httpCalls += 1
	m.httpUrls = append(m.httpUrls, url)
	return nil
}

func (m *mockActionResults) mgr() *Manager {
	result := NewManager()
	result.RegisterAction("success", m.success)
	result.RegisterAction("fail", m.fail)
	result.RegisterAction("fetch", m.fetch)

	return result
}

func setupTestActionEnv(c *check.C) (r *mockActionResults, s *status.Status, a *status.Status) {
	r = &mockActionResults{}
	s = &status.Status{}
	a = &status.Status{}

	// Load status with some predefined actions.
	e := s.SetJson("status://action", []byte(`
		{
			"actionSuccess": {"action": "success"},
			"actionFail": {"action": "fail"},
			"actionUnknown": {"action": "unknown"}
		}
		`), 0)
	c.Assert(e, check.IsNil)

	return r, s, a
}

func (suite *MySuite) TestManagerRegisterUnRegister(c *check.C) {

	var testAction Action = func(s *status.Status, action *status.Status) error { return nil }

	mgr := NewManager()

	// Register a couple of actions.
	err := mgr.RegisterAction("foo", testAction)
	c.Check(err, check.IsNil)

	err = mgr.RegisterAction("bar", testAction)
	c.Check(err, check.IsNil)

	// Register one a second time.
	err = mgr.RegisterAction("foo", testAction)
	c.Check(err, check.ErrorMatches, "Action: Already Exists: .*")

	// Look one up.
	resultAction, err := mgr.lookupAction("foo")
	c.Check(err, check.IsNil)
	c.Check(resultAction, check.NotNil)

	// Remove it.
	err = mgr.UnRegisterAction("foo")
	c.Check(err, check.IsNil)

	// Make sure we can't look it up.
	resultAction, err = mgr.lookupAction("foo")
	c.Check(err, check.ErrorMatches, "Action: Not Registered: .*")
	c.Check(resultAction, check.IsNil)

	// Look up the other action to make sure it's still valid.
	resultAction, err = mgr.lookupAction("bar")
	c.Check(err, check.IsNil)
	c.Check(resultAction, check.NotNil)

	// Look up an unknown action.
	resultAction, err = mgr.lookupAction("unknown")
	c.Check(err, check.ErrorMatches, "Action: Not Registered: .*")
	c.Check(resultAction, check.IsNil)

	// Unregister up an unknown action.
	err = mgr.UnRegisterAction("unknown")
	c.Check(err, check.ErrorMatches, "Action: Not Registered: .*")
}

func (suite *MySuite) TestFireActionNil(c *check.C) {
	r, s, a := setupTestActionEnv(c)

	r.mgr().FireAction(s, a)

	c.Check(r.successCalls, check.Equals, 0)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionStringRedirect(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", "status://action/actionSuccess", 0)

	r.mgr().FireAction(s, a)

	c.Check(r.successCalls, check.Equals, 1)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionStringRedirectHttp(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", "http://foo/", 0)

	r.mgr().FireAction(s, a)

	c.Check(r.successCalls, check.Equals, 0)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 1)
	c.Check(r.httpUrls, check.DeepEquals, []string{"http://foo/"})
}

func (suite *MySuite) TestFireActionStringRedirectBadLocation(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", "status://bogus/redirect", 0)

	r.mgr().FireAction(s, a)

	c.Check(r.successCalls, check.Equals, 0)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionArray(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", []interface{}{
		"status://action/actionSuccess",
		"status://action/actionSuccess"}, 0)

	r.mgr().FireAction(s, a)

	c.Check(r.successCalls, check.Equals, 2)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionArrayFailure(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", []interface{}{
		"status://action/actionSuccess",
		"status://action/actionFail",
		"status://action/actionSuccess"}, 0)

	r.mgr().FireAction(s, a)

	c.Check(r.successCalls, check.Equals, 2)
	c.Check(r.failCalls, check.Equals, 1)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionDict(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", map[string]interface{}{"action": "success"}, 0)

	r.mgr().FireAction(s, a)

	c.Check(r.successCalls, check.Equals, 1)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionDictFail(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", map[string]interface{}{"action": "fail"}, 0)

	r.mgr().FireAction(s, a)

	c.Check(r.successCalls, check.Equals, 0)
	c.Check(r.failCalls, check.Equals, 1)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireManagerUnknown(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", map[string]interface{}{"action": "unknown"}, 0)

	r.mgr().FireAction(s, a)

	c.Check(r.successCalls, check.Equals, 0)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 0)
}
