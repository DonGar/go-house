package adapter

import (
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestMgrBaseAdapters(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s := &status.Status{}
	actionsMgr := actions.NewManager()

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
	mgr, e = NewManager(s, actionsMgr)
	c.Assert(e, check.IsNil)

	// We created the right number of adapters.
	c.Check(len(mgr.adapters), check.Equals, 2)

	// We verify their contents.
	c.Check(mgr.adapters["TestBase"], check.NotNil)
	c.Check(mgr.adapters["TestBase2"], check.NotNil)
}

func (suite *MySuite) TestMgrAllAdaptersStop(c *check.C) {
	s := setupTestStatus(c)
	actionsMgr := actions.NewManager()

	// Create the manager.
	var mgr *Manager
	mgr, e := NewManager(s, actionsMgr)
	c.Assert(e, check.IsNil)

	// We created the right number of adapters.
	c.Check(len(mgr.adapters), check.Equals, 6)

	mgr.Stop()

	// We removed all adapters.
	c.Check(len(mgr.adapters), check.Equals, 0)

	// Verify Status Contents
	adapterUrls := []string{
		"status://TestBase", "status://TestFile", "status://TestFileSpecified",
		"status://TestParticle", "status://TestWeb",
	}

	for _, url := range adapterUrls {
		v, _, e := s.Get(url)
		c.Check(e, check.IsNil)
		c.Check(v, check.DeepEquals, interface{}(nil))
	}
}
