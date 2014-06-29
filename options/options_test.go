package options

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (suite *MySuite) TestLoadServerConfig(c *check.C) {
	tempDir := c.MkDir()

	var e error

	// Write out a test server config file.
	serverFile := filepath.Join(tempDir, "server.json")
	e = ioutil.WriteFile(serverFile, []byte(`{"foo": "bar"}`), os.ModePerm)
	c.Assert(e, check.IsNil)

	s := &status.Status{}
	e = s.Set(CONFIG_DIR, tempDir, 0)
	c.Assert(e, check.IsNil)

	e = LoadServerConfig(s)
	c.Check(e, check.IsNil)

	serverValue, _, e := s.Get("status://server")
	c.Check(e, check.IsNil)
	c.Check(
		serverValue,
		check.DeepEquals,
		map[string]interface{}{
			"longitude": 0.0,
			"foo":       "bar",
			"config":    tempDir,
			"static":    "/home/dgarrett/Development/go-house/static",
			"downloads": "/tmp/Downloads",
			"latitude":  0.0,
		})
}

func (suite *MySuite) TestSetDefaultsEmptyStatus(c *check.C) {
	var e error
	var v interface{}

	s := &status.Status{}

	e = setDefaults(s, "config/dir")
	c.Check(e, check.IsNil)

	v, _, e = s.Get(CONFIG_DIR)
	c.Check(e, check.IsNil)
	c.Check(v, check.Equals, "config/dir")

	v, _, e = s.Get(STATIC_DIR)
	c.Check(e, check.IsNil)
	c.Check(v, check.Equals, "/home/dgarrett/Development/go-house/static")

	v, _, e = s.Get(DOWNLOADS_DIR)
	c.Check(e, check.IsNil)
	c.Check(v, check.Equals, "/tmp/Downloads")

	v, _, e = s.Get(LATITUDE)
	c.Check(e, check.IsNil)
	c.Check(v, check.Equals, 0.0)

	v, _, e = s.Get(LONGITUDE)
	c.Check(e, check.IsNil)
	c.Check(v, check.Equals, 0.0)
}

func (suite *MySuite) TestSetDefaultsPopulatedStatus(c *check.C) {
	var e error
	var v interface{}

	s := &status.Status{}
	e = s.SetJson("status://server",
		[]byte(`
		{
		  "config": "foo_config",
		  "static": "foo_static",
		  "downloads": "foo_downloads",
		  "latitude": 1.2,
		  "longitude": 3.4
		}`),
		0)
	c.Assert(e, check.IsNil)

	e = setDefaults(s, "config/dir")
	c.Check(e, check.IsNil)

	v, _, e = s.Get(CONFIG_DIR)
	c.Check(e, check.IsNil)
	c.Check(v, check.Equals, "foo_config")

	v, _, e = s.Get(STATIC_DIR)
	c.Check(e, check.IsNil)
	c.Check(v, check.Equals, "foo_static")

	v, _, e = s.Get(DOWNLOADS_DIR)
	c.Check(e, check.IsNil)
	c.Check(v, check.Equals, "foo_downloads")

	v, _, e = s.Get(LATITUDE)
	c.Check(e, check.IsNil)
	c.Check(v, check.Equals, 1.2)

	v, _, e = s.Get(LONGITUDE)
	c.Check(e, check.IsNil)
	c.Check(v, check.Equals, 3.4)
}

func (suite *MySuite) TestDefaultConfigDir(c *check.C) {
	v, e := defaultConfigDir()
	c.Check(e, check.IsNil)
	c.Check(v, check.Matches, ".*github.com/DonGar/go-house/options/_test")
}
