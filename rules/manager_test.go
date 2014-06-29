package rules

import (
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

// This creates standard status/options objects used by most Adapters tests.
func setupTestStatusOptions(c *check.C) (o *options.Options, s *status.Status, e error) {

	s = &status.Status{}
	e = s.SetJson("status://",
		[]byte(`
    {
      "server": {
        "adapters": {
        }
      },
      "testAdapter": {
      	"rules": {
      		"base": {
      			"BasicOne": {
						},
      			"BasicTwo": {
						}
      		}
      	}
      }
    }`),
		0)
	c.Assert(e, check.IsNil)

	o, e = options.NewOptions(s)
	c.Assert(e, check.IsNil)

	return o, s, nil
}

func (suite *MySuite) TestMgrStartEditStop(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	o, s, e := setupTestStatusOptions(c)
	c.Assert(e, check.IsNil)

	// Create the manager.
	var mgr *Manager

	mgr, e = NewManager(o, s)
	c.Assert(e, check.IsNil)

	// We give the rules manager a little time for a delayed update.
	// We verify that we have the expected number of rules
	time.Sleep(2 * time.Millisecond)
	c.Check(len(mgr.rules), check.Equals, 2)

	e = s.SetJson("status://testAdapter/rules/base/BasicThree", []byte(`{}`), 1)
	c.Assert(e, check.IsNil)

	// We give the rules manager a little time to finish initializing.
	// We verify that we have the expected number of rules
	time.Sleep(2 * time.Millisecond)
	c.Check(len(mgr.rules), check.Equals, 3)

	e = s.Remove("status://testAdapter/rules/base/BasicTwo", 2)
	c.Assert(e, check.IsNil)

	// We give the rules manager a little time for a delayed update.
	// We verify that we have the expected number of rules
	time.Sleep(2 * time.Millisecond)
	c.Check(len(mgr.rules), check.Equals, 2)

	// Stop it.
	e = mgr.Stop()
	c.Check(e, check.IsNil)

	// We verify that there are no rules.
	c.Check(len(mgr.rules), check.Equals, 0)
}
