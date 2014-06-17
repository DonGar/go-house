package adapter

import (
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestWebAdapterStartStop(c *check.C) {
	o, s, e := setupTestStatusOptions(c)
	c.Assert(e, check.IsNil)

	config, _, e := s.GetSubStatus("status://server/adapters/TestWeb")
	c.Assert(e, check.IsNil)

	base := base{
		status:     s,
		options:    o,
		config:     config,
		adapterUrl: "status://TestWeb",
	}

	// We need just enough of a manager to let WebAdapter register itself.
	mgr := &AdapterManager{webUrls: map[string]Adapter{}}

	// Create a web adapter.
	a, e := NewWebAdapter(mgr, base)
	c.Assert(e, check.IsNil)

	// Verify Status Contents.
	v, r, e := s.Get(base.adapterUrl)
	c.Check(v, check.DeepEquals, map[string]interface{}{})
	c.Check(r, check.Equals, 2)
	c.Check(e, check.IsNil)

	// Verify Manager registration.
	c.Check(
		mgr.webUrls,
		check.DeepEquals,
		map[string]Adapter{"status://TestWeb": a})

	// Verify Manager WebAdapterStatusUrls (really a manager test, handy here)
	c.Check(
		mgr.WebAdapterStatusUrls(),
		check.DeepEquals,
		[]string{"status://TestWeb"})

	// Stop the web adapter.
	e = a.Stop()
	c.Assert(e, check.IsNil)

	// Verify Manager removal.
	c.Check(mgr.webUrls, check.DeepEquals, map[string]Adapter{})

	// Verify Status Contents.
	v, r, e = s.Get(base.adapterUrl)
	c.Check(v, check.DeepEquals, nil)
	c.Check(r, check.Equals, 3)
	c.Check(e, check.IsNil)
}
