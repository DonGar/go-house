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

	s, e := setupTestStatus(c)
	c.Assert(e, check.IsNil)

	// config, _, e := s.GetSubStatus("status://server/adapters/base/TestBase")
	// c.Assert(e, check.IsNil)

	// Create the manager.
	var mgr *Manager
	mgr, e = NewManager(s)
	c.Assert(e, check.IsNil)

	// We created the right number of adapters.
	c.Check(len(mgr.adapters), check.Equals, 4)

	mgr.Stop()

	// We removed all adapters.
	c.Check(len(mgr.adapters), check.Equals, 0)

	// Verify Status Contents
	v, _, e := s.Get("status://")
	c.Check(e, check.IsNil)
	c.Check(
		v,
		check.DeepEquals,
		map[string]interface{}{
			"server": map[string]interface{}{
				"config": "./testdata",
				"adapters": map[string]interface{}{
					"base": map[string]interface{}{
						"TestBase": map[string]interface{}{},
					},
					"file": map[string]interface{}{
						"TestFileSpecified": map[string]interface{}{
							"filename": "TestFile.json"},
						"TestFile": map[string]interface{}{},
					},
					"web": map[string]interface{}{
						"TestWeb": map[string]interface{}{},
					},
				},
			},
			"TestBase":          interface{}(nil),
			"TestFile":          interface{}(nil),
			"TestFileSpecified": interface{}(nil),
			"TestWeb":           interface{}(nil)})
}
