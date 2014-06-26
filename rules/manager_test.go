package rules

import (
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestMgrStartStop(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	o, s, e := setupTestStatusOptions(c)
	c.Assert(e, check.IsNil)

	// Create the manager.
	var mgr *Manager

	mgr, e = NewManager(o, s)
	c.Assert(e, check.IsNil)

	// Stop it.
	e = mgr.Stop()
	c.Check(e, check.IsNil)
}
