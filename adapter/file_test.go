package adapter

import (
	// "github.com/DonGar/go-house/options"
	// "github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestFileAdapterStartStopDefault(c *check.C) {
	o, s, e := setupTestStatusOptions(c)

	base := base{
		status:     s,
		options:    o,
		configUrl:  "status://server/adapters/TestFile",
		adapterUrl: "status://TestFile",
	}

	// Create a file adapter.
	fa, e := NewFileAdapter(base)
	c.Assert(e, check.IsNil)

	// Verify Status Contents.
	v, r, e := s.Get(base.adapterUrl)
	c.Check(v, check.DeepEquals, map[string]interface{}{"foo": "bar"})
	c.Check(r, check.Equals, 2)
	c.Check(e, check.IsNil)

	// Stop the file adapter.
	e = fa.Stop()
	c.Assert(e, check.IsNil)

	// Verify Status Contents.
	v, r, e = s.Get(base.adapterUrl)
	c.Check(v, check.DeepEquals, nil)
	c.Check(r, check.Equals, 3)
	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestFileAdapterStartStopFilename(c *check.C) {
	o, s, e := setupTestStatusOptions(c)

	base := base{
		status:     s,
		options:    o,
		configUrl:  "status://server/adapters/TestFileSpecified",
		adapterUrl: "status://TestFileSpecified",
	}

	// Create a file adapter.
	fa, e := NewFileAdapter(base)
	c.Assert(e, check.IsNil)

	// Verify Status Contents.
	v, r, e := s.Get(base.adapterUrl)
	c.Check(v, check.DeepEquals, map[string]interface{}{"foo": "bar"})
	c.Check(r, check.Equals, 2)
	c.Check(e, check.IsNil)

	// Stop the file adapter.
	e = fa.Stop()
	c.Assert(e, check.IsNil)

	// Verify Status Contents.
	v, r, e = s.Get(base.adapterUrl)
	c.Check(v, check.DeepEquals, nil)
	c.Check(r, check.Equals, 3)
	c.Check(e, check.IsNil)
}