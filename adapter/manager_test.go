package adapter

import (
	"github.com/DonGar/go-house/options"
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
        "adapters": {
          "TestBase": {
            "type": "base"
          },
          "TestBase2": {
            "type": "base"
          }
        }
      }
    }`),
		0)
	c.Assert(e, check.IsNil)

	o := &options.Options{
		ConfigDir: "./testdata",
	}

	// Create the manager.
	var adapterMgr *AdapterManager
	adapterMgr, e = NewAdapterManager(o, s)
	c.Assert(e, check.IsNil)

	// We created the right number of adapters.
	c.Check(len(adapterMgr.adapters), check.Equals, 2)

	// We verify their contents.
	c.Check(adapterMgr.adapters["TestBase"], check.DeepEquals, &base{
		options:    o,
		status:     s,
		configUrl:  "status://server/adapters/TestBase",
		adapterUrl: "status://TestBase",
	})
	c.Check(adapterMgr.adapters["TestBase2"], check.DeepEquals, &base{
		options:    o,
		status:     s,
		configUrl:  "status://server/adapters/TestBase2",
		adapterUrl: "status://TestBase2",
	})
}

func (suite *MySuite) TestMgrAllAdaptersStop(c *check.C) {

	o, s, e := setupTestStatusOptions(c)
	c.Assert(e, check.IsNil)

	// Create the manager.
	var adapterMgr *AdapterManager
	adapterMgr, e = NewAdapterManager(o, s)
	c.Assert(e, check.IsNil)

	// We created the right number of adapters.
	c.Check(len(adapterMgr.adapters), check.Equals, 3)

	// We verify their contents.
	c.Check(adapterMgr.adapters["TestBase"], check.DeepEquals, &base{
		options:    o,
		status:     s,
		configUrl:  "status://server/adapters/TestBase",
		adapterUrl: "status://TestBase",
	})

	e = adapterMgr.Stop()
	c.Check(e, check.IsNil)

	// We removed all adapters.
	c.Check(len(adapterMgr.adapters), check.Equals, 0)

	// Verify Status Contents
	v, _, e := s.Get("status://")
	c.Check(e, check.IsNil)
	c.Check(
		v,
		check.DeepEquals,
		map[string]interface{}{
			"server": map[string]interface{}{
				"adapters": map[string]interface{}{
					"TestBase": map[string]interface{}{"type": "base"},
					"TestFile": map[string]interface{}{"type": "file"},
					"TestFileSpecified": map[string]interface{}{
						"type": "file", "filename": "TestFile.json",
					},
				},
			},
			"TestBase":          interface{}(nil),
			"TestFile":          interface{}(nil),
			"TestFileSpecified": interface{}(nil),
		})
}
