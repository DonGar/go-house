package main

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

	options := Options{ConfigDir: tempDir}
	testStatus := status.Status{}

	e = loadServerConfig(options, &testStatus)
	c.Assert(e, check.IsNil)

	serverValueJson, _, e := testStatus.GetJson("status://server")

	// And the value we pulled out of the Status matches the file contents.
	c.Assert(string(serverValueJson), check.Equals, `{"foo":"bar"}`)
}

func (suite *MySuite) TestFindOptions(c *check.C) {
	var options Options
	var e error

	options, e = findOptions()
	c.Check(e, check.IsNil)
	c.Check(
		options.ConfigDir,
		check.Matches,
		".*github.com/DonGar/go-house/_test")

}
