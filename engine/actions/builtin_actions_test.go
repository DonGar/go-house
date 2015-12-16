package actions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
)

const INITIAL_ENV = `{
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
    }`

func setupTestBuiltinActionEnv(c *check.C) (s *status.Status, a *status.Status) {
	s = &status.Status{}
	a = &status.Status{}

	// Load status with some predefined actions.
	e := s.SetJson("status://", []byte(INITIAL_ENV), 0)
	c.Assert(e, check.IsNil)

	return s, a
}

func validateTestSet(c *check.C, actionJson, resultJson string) {
	s, a := setupTestBuiltinActionEnv(c)
	err := a.SetJson("status://", []byte(actionJson), 0)
	c.Assert(err, check.IsNil)

	// Perform action.
	err = actionSet(s, a)
	c.Assert(err, check.IsNil)

	// Validate Result.
	v, _, _ := s.GetJson("status://")
	c.Check(status.NormalizeJson(string(v)), check.DeepEquals, status.NormalizeJson(resultJson))
}

func (suite *MySuite) TestSet(c *check.C) {
	// String Value
	action := `{
    "action": "set",
    "component": "status://adapter/host/hostA",
    "dest": "component_dest",
    "value": "new_value"
  }`

	expected := `{
    "adapter": {
      "host": {
        "hostA": {
          "component_dest": "new_value",
          "mac": "00:11:22:33:44:55"
        },
        "hostB": {}
      }
    },
    "server": {
      "downloads": "/tmp/downloads",
      "email_address": "from@from.org",
      "relay_id_server": "bogus_server",
      "relay_password": "bogus_password",
      "relay_server": "bogus_server:587",
      "relay_user": "bogus_user"
    }
  }`

	validateTestSet(c, action, expected)

	// Int Value
	action = `{
    "action": "set",
    "component": "status://adapter/host/hostA",
    "dest": "component_dest",
    "value": 3
  }`

	expected = `{
    "adapter": {
      "host": {
        "hostA": {
          "component_dest": 3,
          "mac": "00:11:22:33:44:55"
        },
        "hostB": {}
      }
    },
    "server": {
      "downloads": "/tmp/downloads",
      "email_address": "from@from.org",
      "relay_id_server": "bogus_server",
      "relay_password": "bogus_password",
      "relay_server": "bogus_server:587",
      "relay_user": "bogus_user"
    }
  }`

	validateTestSet(c, action, expected)

	// Boolean Value
	action = `{
    "action": "set",
    "component": "status://adapter/host/hostA",
    "dest": "component_dest",
    "value": false
  }`

	expected = `{
    "adapter": {
      "host": {
        "hostA": {
          "component_dest": false,
          "mac": "00:11:22:33:44:55"
        },
        "hostB": {}
      }
    },
    "server": {
      "downloads": "/tmp/downloads",
      "email_address": "from@from.org",
      "relay_id_server": "bogus_server",
      "relay_password": "bogus_password",
      "relay_server": "bogus_server:587",
      "relay_user": "bogus_user"
    }
  }`

	validateTestSet(c, action, expected)

	// Wildcard Component
	action = `{
    "action": "set",
    "component": "status://adapter/host/*",
    "dest": "component_dest",
    "value": "new_value"
  }`

	expected = `{
    "adapter": {
      "host": {
        "hostA": {
          "component_dest": "new_value",
          "mac": "00:11:22:33:44:55"
        },
        "hostB": {
          "component_dest": "new_value"
        }
      }
    },
    "server": {
      "downloads": "/tmp/downloads",
      "email_address": "from@from.org",
      "relay_id_server": "bogus_server",
      "relay_password": "bogus_password",
      "relay_server": "bogus_server:587",
      "relay_user": "bogus_user"
    }
  }`

	validateTestSet(c, action, expected)

	// Unknown Component
	action = `{
    "action": "set",
    "component": "status://adapter/host/Nunya",
    "dest": "component_dest",
    "value": "new_value"
  }`

	expected = INITIAL_ENV

	validateTestSet(c, action, expected)
}

func (suite *MySuite) TestWol(c *check.C) {
	s, a := setupTestBuiltinActionEnv(c)
	a.Set("status://", map[string]interface{}{
		"action":    "wol",
		"component": "status://adapter/host/*",
	}, 0)

	e := actionWol(s, a)

	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestPing(c *check.C) {
	s, a := setupTestBuiltinActionEnv(c)
	a.Set("status://", map[string]interface{}{
		"action":    "ping",
		"component": "status://adapter/host/*",
	}, 0)

	e := actionPing(s, a)

	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestFetch(c *check.C) {
	s, a := setupTestBuiltinActionEnv(c)
	a.Set("status://", map[string]interface{}{
		"action": "fetch",
		"url":    "http://www.google.com/",
	}, 0)

	e := actionFetch(s, a)

	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestFetchDownload(c *check.C) {
	s, a := setupTestBuiltinActionEnv(c)

	tempDir := c.MkDir()
	s.Set("status://server/downloads", tempDir, status.UNCHECKED_REVISION)

	a.Set("status://", map[string]interface{}{
		"action":        "fetch",
		"url":           "http://www.google.com/",
		"download_name": "foo.{time}.bar",
	}, 0)

	e := actionFetch(s, a)

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
	s, a := setupTestBuiltinActionEnv(c)
	e := a.Set("status://", map[string]interface{}{
		"action":  "email",
		"to":      "dgarrett@acm.org",
		"subject": "Test Subject",
		"body":    "Test Body",
	}, 0)
	c.Assert(e, check.IsNil)

	e = actionEmail(s, a)
	c.Check(e, check.NotNil)
}

func (suite *MySuite) TestEmailAttached(c *check.C) {
	s, a := setupTestBuiltinActionEnv(c)
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

	e = actionEmail(s, a)
	c.Check(e, check.NotNil)
}
