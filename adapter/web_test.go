package adapter

import (
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestWebAdapterStartStop(c *check.C) {

	s, mgr, b := setupTestAdapter(c,
		"status://server/adapters/web/TestWeb", "status://TestWeb")

	// Create a web adapter.
	a, e := newWebAdapter(mgr, b)
	c.Assert(e, check.IsNil)

	// Verify Status Contents.
	v, r, e := s.Get(b.adapterUrl)
	c.Check(v, check.DeepEquals, map[string]interface{}{})
	c.Check(r, check.Equals, 2)
	c.Check(e, check.IsNil)

	// Verify Manager registration.
	c.Check(
		mgr.webUrls,
		check.DeepEquals,
		map[string]adapter{"status://TestWeb": a})

	// Verify Manager WebAdapterStatusUrls (really a manager test, handy here)
	c.Check(
		mgr.WebAdapterStatusUrls(),
		check.DeepEquals,
		[]string{"status://TestWeb"})

	// Stop the web adapter.
	a.Stop()

	// Verify Manager removal.
	c.Check(mgr.webUrls, check.DeepEquals, map[string]adapter{})

	// Verify Status Contents.
	v, r, e = s.Get(b.adapterUrl)
	c.Check(v, check.DeepEquals, nil)
	c.Check(r, check.Equals, 3)
	c.Check(e, check.IsNil)
}
