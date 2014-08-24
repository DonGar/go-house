package adapter

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestMgrBaseAdapters(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s := &status.Status{}
	e := s.SetJson("status://",
		[]byte(`
    {
      "server": {
      	"config": "./testdata",
        "adapters": {
          "base": {
            "TestBase": {
            },
            "TestBase2": {
            }
          }
        }
      }
    }`),
		0)
	c.Assert(e, check.IsNil)

	// Create the manager.
	var mgr *Manager
	mgr, e = NewManager(s)
	c.Assert(e, check.IsNil)

	// We created the right number of adapters.
	c.Check(len(mgr.adapters), check.Equals, 2)

	// We verify their contents.
	c.Check(mgr.adapters["TestBase"], check.NotNil)
	c.Check(mgr.adapters["TestBase2"], check.NotNil)
}

func (suite *MySuite) TestMgrAllAdaptersStop(c *check.C) {
	s := setupTestStatus(c)

	// Create the manager.
	var mgr *Manager
	mgr, e := NewManager(s)
	c.Assert(e, check.IsNil)

	// We created the right number of adapters.
	c.Check(len(mgr.adapters), check.Equals, 5)

	mgr.Stop()

	// We removed all adapters.
	c.Check(len(mgr.adapters), check.Equals, 0)

	// Verify Status Contents
	adapterUrls := []string{
		"status://TestBase", "status://TestFile", "status://TestFileSpecified",
		"status://TestSpark", "status://TestWeb",
	}

	for _, url := range adapterUrls {
		v, _, e := s.Get(url)
		c.Check(e, check.IsNil)
		c.Check(v, check.DeepEquals, interface{}(nil))
	}
}
