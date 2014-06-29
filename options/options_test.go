package options

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (suite *MySuite) TestOptionsDefaults(c *check.C) {
	var options *Options
	var e error

	s := &status.Status{}
	options, e = NewOptions(s)
	c.Check(e, check.IsNil)

	// ConfigDir should be based on executable location.
	cd, e := options.ConfigDir()
	c.Check(e, check.IsNil)
	c.Check(cd, check.Matches, ".*github.com/DonGar/go-house/options/_test")

	// Currently hard coded.
	sd, e := options.StaticDir()
	c.Check(e, check.IsNil)
	c.Check(sd, check.Equals, "/home/dgarrett/Development/go-house/static")

	// The remaining values don't have defaults.
	la, e := options.Latitude()
	c.Check(e, check.NotNil)
	c.Check(la, check.Equals, 0.0)

	lo, e := options.Longitude()
	c.Check(e, check.NotNil)
	c.Check(lo, check.Equals, 0.0)

	dd, e := options.DownloadDir()
	c.Check(e, check.NotNil)
	c.Check(dd, check.Equals, "")
}

func (suite *MySuite) TestOptionsExplict(c *check.C) {
	var options *Options
	var e error

	s := &status.Status{}
	e = s.SetJson("status://server",
		[]byte(`
		{
		  "port": 8081,

		  "config": "/tmp/Configs",
		  "downloads": "/tmp/Downloads",

		  "latitude": 37.3861,
		  "longitude": 122.0839
		}`),
		0)
	c.Assert(e, check.IsNil)

	options, e = NewOptions(s)
	c.Check(e, check.IsNil)

	// ConfigDir should be based on executable location.
	cd, e := options.ConfigDir()
	c.Check(e, check.IsNil)
	c.Check(cd, check.Matches, "/tmp/Configs")

	// Currently hard coded.
	sd, e := options.StaticDir()
	c.Check(e, check.IsNil)
	c.Check(sd, check.Equals, "/home/dgarrett/Development/go-house/static")

	// The remaining values don't have defaults.
	la, e := options.Latitude()
	c.Check(e, check.IsNil)
	c.Check(la, check.Equals, 37.3861)

	lo, e := options.Longitude()
	c.Check(e, check.IsNil)
	c.Check(lo, check.Equals, 122.0839)

	dd, e := options.DownloadDir()
	c.Check(e, check.IsNil)
	c.Check(dd, check.Equals, "/tmp/Downloads")
}
