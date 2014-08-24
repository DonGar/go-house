package adapter

import (
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestSparkAdapterStartStop(c *check.C) {
	s, mgr, b := setupTestAdapter(c,
		"status://server/adapters/spark/TestSpark", "status://TestSpark")

	// Create a spark adapter.
	a, e := newSparkAdapter(mgr, b)
	c.Assert(e, check.IsNil)

	// Verify Status Contents.
	v, _, e := s.Get(b.adapterUrl)
	c.Check(
		v,
		check.DeepEquals,
		map[string]interface{}{"core": map[string]interface{}{}},
	)
	c.Check(e, check.IsNil)

	a.Stop()

	// Verify Status Contents.
	v, _, e = s.Get(b.adapterUrl)
	c.Check(v, check.DeepEquals, nil)
	c.Check(e, check.IsNil)
}
