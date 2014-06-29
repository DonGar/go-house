package adapter

import (
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestFileAdapterStartStopDefault(c *check.C) {
	s, e := setupTestStatus(c)
	c.Assert(e, check.IsNil)

	config, _, e := s.GetSubStatus("status://server/adapters/file/TestFile")
	c.Assert(e, check.IsNil)

	base := base{
		status:     s,
		config:     config,
		adapterUrl: "status://TestFile",
	}

	// Create a file adapter.
	fa, e := newFileAdapter(nil, base)
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
	s, e := setupTestStatus(c)
	c.Assert(e, check.IsNil)

	config, _, e := s.GetSubStatus("status://server/adapters/file/TestFileSpecified")
	c.Assert(e, check.IsNil)

	base := base{
		status:     s,
		config:     config,
		adapterUrl: "status://TestFileSpecified",
	}

	// Create a file adapter.
	fa, e := newFileAdapter(nil, base)
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
