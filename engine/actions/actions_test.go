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

func (m *mockActionResults) success(r ActionRegistrar, s *status.Status, action *status.Status) error {
	m.successCalls += 1
	return nil
}

func (m *mockActionResults) fail(r ActionRegistrar, s *status.Status, action *status.Status) error {
	m.failCalls += 1
	return fmt.Errorf(MOCK_FAILURE_MSG)
}

func (m *mockActionResults) fetch(r ActionRegistrar, s *status.Status, action *status.Status) error {
	url, e := action.GetString("status://url")
	if e != nil {
		return e
	}

	m.httpCalls += 1
	m.httpUrls = append(m.httpUrls, url)
	return nil
}

func (m *mockActionResults) registrar() ActionRegistrar {
	return ActionRegistrar{
		"success": m.success,
		"fail":    m.fail,
		"fetch":   m.fetch,
	}
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

func (suite *MySuite) TestFireActionNil(c *check.C) {
	r, s, a := setupTestActionEnv(c)

	e := FireAction(s, r.registrar(), a)

	c.Check(e, check.ErrorMatches, "Action: Can't perform .*")
	c.Check(r.successCalls, check.Equals, 0)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionStringRedirect(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", "status://action/actionSuccess", 0)

	e := FireAction(s, r.registrar(), a)

	c.Check(e, check.IsNil)
	c.Check(r.successCalls, check.Equals, 1)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionStringRedirectHttp(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", "http://foo/", 0)

	e := FireAction(s, r.registrar(), a)

	c.Check(e, check.IsNil)
	c.Check(r.successCalls, check.Equals, 0)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 1)
	c.Check(r.httpUrls, check.DeepEquals, []string{"http://foo/"})
}

func (suite *MySuite) TestFireActionStringRedirectBadLocation(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", "status://bogus/redirect", 0)

	e := FireAction(s, r.registrar(), a)

	c.Check(e, check.ErrorMatches, "Status: Node status://bogus of status://bogus/redirect does not exist.")
	c.Check(r.successCalls, check.Equals, 0)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionArray(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", []interface{}{
		"status://action/actionSuccess",
		"status://action/actionSuccess"}, 0)

	e := FireAction(s, r.registrar(), a)

	c.Check(e, check.IsNil)
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

	e := FireAction(s, r.registrar(), a)

	c.Check(e, check.ErrorMatches, MOCK_FAILURE_MSG)
	c.Check(r.successCalls, check.Equals, 1)
	c.Check(r.failCalls, check.Equals, 1)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionDict(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", map[string]interface{}{"action": "success"}, 0)

	e := FireAction(s, r.registrar(), a)

	c.Check(e, check.IsNil)
	c.Check(r.successCalls, check.Equals, 1)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionDictFail(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", map[string]interface{}{"action": "fail"}, 0)

	e := FireAction(s, r.registrar(), a)

	c.Check(e, check.ErrorMatches, MOCK_FAILURE_MSG)
	c.Check(r.successCalls, check.Equals, 0)
	c.Check(r.failCalls, check.Equals, 1)
	c.Check(r.httpCalls, check.Equals, 0)
}

func (suite *MySuite) TestFireActionRegistrarUnknown(c *check.C) {
	r, s, a := setupTestActionEnv(c)
	a.Set("status://", map[string]interface{}{"action": "unknown"}, 0)

	e := FireAction(s, r.registrar(), a)

	c.Check(e, check.ErrorMatches, "Action: No registered action: unknown")
	c.Check(r.successCalls, check.Equals, 0)
	c.Check(r.failCalls, check.Equals, 0)
	c.Check(r.httpCalls, check.Equals, 0)
}
