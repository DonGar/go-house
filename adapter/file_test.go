package adapter

import (
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestFileAdapterStartStopDefault(c *check.C) {
	s, mgr, b := setupTestAdapter(c,
		"status://server/adapters/file/TestFile", "status://TestFile")

	// Create a file adapter.
	fa, e := newFileAdapter(mgr, b)
	c.Assert(e, check.IsNil)

	// Verify Status Contents.
	v, r, e := s.Get(b.adapterUrl)
	c.Check(v, check.DeepEquals, map[string]interface{}{"foo": "bar"})
	c.Check(r, check.Equals, 2)
	c.Check(e, check.IsNil)

	// Stop the file adapter.
	fa.Stop()

	// Verify Status Contents.
	v, r, e = s.Get(b.adapterUrl)
	c.Check(v, check.DeepEquals, nil)
	c.Check(r, check.Equals, 3)
	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestFileAdapterStartStopFilename(c *check.C) {
	s, mgr, b := setupTestAdapter(c,
		"status://server/adapters/file/TestFileSpecified", "status://TestFileSpecified")

	// Create a file adapter.
	fa, e := newFileAdapter(mgr, b)
	c.Assert(e, check.IsNil)

	// Verify Status Contents.
	v, r, e := s.Get(b.adapterUrl)
	c.Check(v, check.DeepEquals, map[string]interface{}{"foo": "bar"})
	c.Check(r, check.Equals, 2)
	c.Check(e, check.IsNil)

	// Stop the file adapter.
	fa.Stop()

	// Verify Status Contents.
	v, r, e = s.Get(b.adapterUrl)
	c.Check(v, check.DeepEquals, nil)
	c.Check(r, check.Equals, 3)
	c.Check(e, check.IsNil)
}

// TODO: Write test case where the config file is modified to test file
// watching.
