package main

import (
	"github.com/DonGar/go-house/options"
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
	e = s.Set("status://server/config", tempDir, 0)
	c.Assert(e, check.IsNil)

	o, e := options.NewOptions(s)
	c.Assert(e, check.IsNil)

	e = loadServerConfig(o, s)
	c.Assert(e, check.IsNil)

	serverValueJson, _, e := s.GetJson("status://server")

	// And the value we pulled out of the Status matches the file contents.
	c.Assert(string(serverValueJson), check.Equals, `{"foo":"bar"}`)
}
