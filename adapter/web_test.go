package adapter

import (
	"github.com/DonGar/go-house/stoppable"
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestWebAdapterStartStop(c *check.C) {
	s, e := setupTestStatus(c)
	c.Assert(e, check.IsNil)

	config, _, e := s.GetSubStatus("status://server/adapters/web/TestWeb")
	c.Assert(e, check.IsNil)

	base := base{stoppable.NewBase(), s, config, "status://TestWeb"}

	// We need just enough of a manager to let WebAdapter register itself.
	mgr := &Manager{webUrls: map[string]adapter{}}

	// Create a web adapter.
	a, e := newWebAdapter(mgr, base)
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
	v, r, e = s.Get(base.adapterUrl)
	c.Check(v, check.DeepEquals, nil)
	c.Check(r, check.Equals, 3)
	c.Check(e, check.IsNil)
}
