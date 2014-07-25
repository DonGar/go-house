package engine

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

// This creates standard status/options objects used by most Adapters tests.
func setupTestStatus(c *check.C) (s *status.Status, e error) {

	s = &status.Status{}
	e = s.SetJson("status://",
		[]byte(`
    {
      "server": {
        "adapters": {
        }
      },
      "testAdapter": {
      	"rule": {
    			"BasicOne": {
    				"condition": {
    					"test": "base"
    				},
    				"action": null
					},
    			"BasicTwo": {
    				"condition": {
    					"test": "base"
    				},
    				"action": null
					}
      	}
      }
    }`),
		0)
	c.Assert(e, check.IsNil)

	return s, nil
}

func (suite *MySuite) TestMgrStartStopEmpty(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s := &status.Status{}

	// Create the manager.
	var mgr *Manager

	mgr, e := NewManager(s)
	c.Assert(e, check.IsNil)

	// We give the rules manager a little time for a delayed update.
	// We verify that we have the expected number of rules
	time.Sleep(2 * time.Millisecond)
	c.Check(len(mgr.rules), check.Equals, 0)

	// Stop it.
	e = mgr.Stop()
	c.Check(e, check.IsNil)

	// We verify that there are no rules.
	c.Check(len(mgr.rules), check.Equals, 0)
}

func (suite *MySuite) TestMgrAddRule(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s := &status.Status{}

	ruleUrl := "status://rule/Test"
	ruleMatch := status.UrlMatch{
		23,
		map[string]interface{}{
			"condition": map[string]interface{}{
				"test": "base",
			},
			"action": nil,
		},
	}

	// Create the manager.
	var mgr *Manager

	mgr, e := NewManager(s)
	c.Assert(e, check.IsNil)

	// We give the rules manager a little time for a delayed update.
	// We verify that we have the expected number of rules
	time.Sleep(2 * time.Millisecond)
	c.Check(len(mgr.rules), check.Equals, 0)

	// Force add a new rule.
	e = mgr.AddRule(ruleUrl, ruleMatch)
	c.Assert(e, check.IsNil)
	c.Check(len(mgr.rules), check.Equals, 1)

	// Stop it.
	e = mgr.Stop()
	c.Check(e, check.IsNil)

	// We verify that there are no rules.
	c.Check(len(mgr.rules), check.Equals, 0)
}

func (suite *MySuite) TestMgrStartEditStop(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s, e := setupTestStatus(c)
	c.Assert(e, check.IsNil)

	// Create the manager.
	var mgr *Manager

	mgr, e = NewManager(s)
	c.Assert(e, check.IsNil)

	// We give the rules manager a little time for a delayed update.
	// We verify that we have the expected number of rules
	time.Sleep(2 * time.Millisecond)
	c.Check(len(mgr.rules), check.Equals, 2)

	e = s.SetJson("status://testAdapter/rule/BasicThree", []byte(`{
		"condition":{
			"test": "base"
		},
    "action": null}`), 1)
	c.Assert(e, check.IsNil)

	// We give the rules manager a little time to finish initializing.
	// We verify that we have the expected number of rules
	time.Sleep(2 * time.Millisecond)
	c.Check(len(mgr.rules), check.Equals, 3)

	e = s.Remove("status://testAdapter/rule/BasicTwo", 2)
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
