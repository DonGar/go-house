package actions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
)

func setupTestBuiltinActionEnv(c *check.C) (r *mockActionRegistrar, s *status.Status, a *status.Status) {
	r = &mockActionRegistrar{}
	s = &status.Status{}
	a = &status.Status{}

	// Load status with some predefined actions.
	e := s.SetJson("status://", []byte(`{
			"server": {
				"downloads": "/tmp/downloads",
				"email_address": "from@from.org",
				"relay_server":   "bogus_server:587",
				"relay_id_server":   "bogus_server",
				"relay_user":     "bogus_user",
				"relay_password": "bogus_password"
			},
			"adapter": {
				"host": {
					"hostA": {
						"mac": "00:11:22:33:44:55"
					},
					"hostB": {
					}
				}
			}
		}`), 0)
	c.Assert(e, check.IsNil)

	return r, s, a
}

func (suite *MySuite) TestSet(c *check.C) {
	r, s, a := setupTestBuiltinActionEnv(c)
	a.Set("status://", map[string]interface{}{
		"action":    "set",
		"component": "status://adapter/host/hostA",
		"dest":      "component_dest",
		"value":     "new_value",
	}, 0)

	e := ActionSet(r, s, a)

	c.Check(e, check.IsNil)

	v, _, _ := s.Get("status://")

	// See if status is properly updated.
	c.Check(v, check.DeepEquals, map[string]interface{}{
		"server": map[string]interface{}{
			"downloads":       "/tmp/downloads",
			"email_address":   "from@from.org",
			"relay_server":    "bogus_server:587",
			"relay_id_server": "bogus_server",
			"relay_user":      "bogus_user",
			"relay_password":  "bogus_password",
		},
		"adapter": map[string]interface{}{
			"host": map[string]interface{}{
				"hostA": map[string]interface{}{
					"mac":            "00:11:22:33:44:55",
					"component_dest": "new_value",
				},
				"hostB": map[string]interface{}{},
			}}})
}

func (suite *MySuite) TestSetWildcard(c *check.C) {
	r, s, a := setupTestBuiltinActionEnv(c)
	a.Set("status://", map[string]interface{}{
		"action":    "set",
		"component": "status://adapter/host/*",
		"dest":      "component_dest",
		"value":     "new_value",
	}, 0)

	e := ActionSet(r, s, a)

	c.Check(e, check.IsNil)

	v, _, _ := s.Get("status://")

	// See if status is properly updated.
	c.Check(v, check.DeepEquals, map[string]interface{}{
		"server": map[string]interface{}{
			"downloads":       "/tmp/downloads",
			"email_address":   "from@from.org",
			"relay_server":    "bogus_server:587",
			"relay_id_server": "bogus_server",
			"relay_user":      "bogus_user",
			"relay_password":  "bogus_password",
		},
		"adapter": map[string]interface{}{
			"host": map[string]interface{}{
				"hostA": map[string]interface{}{
					"mac":            "00:11:22:33:44:55",
					"component_dest": "new_value",
				},
				"hostB": map[string]interface{}{
					"component_dest": "new_value",
				},
			}}})
}

func (suite *MySuite) TestSetBadComponent(c *check.C) {
	r, s, a := setupTestBuiltinActionEnv(c)
	a.Set("status://", map[string]interface{}{
		"action":    "set",
		"component": "status://adapter/host/Nunya",
		"dest":      "component_dest",
		"value":     "new_value",
	}, 0)

	e := ActionSet(r, s, a)

	c.Check(e, check.IsNil)

	v, _, _ := s.Get("status://")

	// See if status is properly updated.
	c.Check(v, check.DeepEquals, map[string]interface{}{
		"server": map[string]interface{}{
			"downloads":       "/tmp/downloads",
			"email_address":   "from@from.org",
			"relay_server":    "bogus_server:587",
			"relay_id_server": "bogus_server",
			"relay_user":      "bogus_user",
			"relay_password":  "bogus_password",
		},
		"adapter": map[string]interface{}{
			"host": map[string]interface{}{
				"hostA": map[string]interface{}{
					"mac": "00:11:22:33:44:55",
				},
				"hostB": map[string]interface{}{},
			}}})
}

func (suite *MySuite) TestWol(c *check.C) {
	return

	r, s, a := setupTestBuiltinActionEnv(c)
	a.Set("status://", map[string]interface{}{
		"action":    "wol",
		"component": "status://adapter/host/*",
	}, 0)

	e := ActionWol(r, s, a)

	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestPing(c *check.C) {
	return

	r, s, a := setupTestBuiltinActionEnv(c)
	a.Set("status://", map[string]interface{}{
		"action":    "ping",
		"component": "status://adapter/host/*",
	}, 0)

	e := ActionPing(r, s, a)

	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestFetch(c *check.C) {
	return

	r, s, a := setupTestBuiltinActionEnv(c)
	a.Set("status://", map[string]interface{}{
		"action": "fetch",
		"url":    "http://www.google.com/",
	}, 0)

	e := ActionFetch(r, s, a)

	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestFetchDownload(c *check.C) {
	return

	r, s, a := setupTestBuiltinActionEnv(c)

	tempDir := c.MkDir()
	s.Set("status://server/downloads", tempDir, status.UNCHECKED_REVISION)

	a.Set("status://", map[string]interface{}{
		"action":        "fetch",
		"url":           "http://www.google.com/",
		"download_name": "foo.{time}.bar",
	}, 0)

	e := ActionFetch(r, s, a)

	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestExpandFileName(c *check.C) {
	s := &status.Status{}

	// Configure the downloads directory.
	e := s.Set("status://server/downloads", "/tmp/downloads", 0)
	c.Assert(e, check.IsNil)

	expanded := expandFileName(s, "")
	c.Check(expanded, check.Equals, "")

	expanded = expandFileName(s, "/foo")
	c.Check(expanded, check.Equals, "/foo")

	expanded = expandFileName(s, "foo")
	c.Check(expanded, check.Equals, "/tmp/downloads/foo")

	expanded = expandFileName(s, "foo.{time}.jpg")
	c.Check(expanded, check.Matches, "/tmp/downloads/foo..+.jpg")
}

func (suite *MySuite) TestEmail(c *check.C) {
	return

	r, s, a := setupTestBuiltinActionEnv(c)
	e := a.Set("status://", map[string]interface{}{
		"action":  "email",
		"to":      "dgarrett@acm.org",
		"subject": "Test Subject",
		"body":    "Test Body",
	}, 0)
	c.Assert(e, check.IsNil)

	e = ActionEmail(r, s, a)
	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestEmailAttached(c *check.C) {
	return

	r, s, a := setupTestBuiltinActionEnv(c)
	e := a.Set("status://", map[string]interface{}{
		"action":  "email",
		"to":      "dgarrett@acm.org",
		"subject": "Test Subject",
		"body":    "Test Body",
		"attachments": []interface{}{
			map[string]interface{}{
				"url":           "http://kitchen/snapshot.cgi?user=guest&pwd=",
				"download_name": "kitchen.jpg",
			},
			map[string]interface{}{
				"url":           "http://garage/snapshot.cgi?user=guest&pwd=",
				"download_name": "garage.jpg",
			},
		},
	}, 0)
	c.Assert(e, check.IsNil)

	e = ActionEmail(r, s, a)
	c.Check(e, check.IsNil)
}
