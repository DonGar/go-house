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
	temp_dir := c.MkDir()

	var e error

	// Write out a test server config file.
	server_file := filepath.Join(temp_dir, "server.json")
	e = ioutil.WriteFile(server_file, []byte(`{"foo": "bar"}`), os.ModePerm)
	c.Assert(e, check.IsNil)

	options := Options{config_dir: temp_dir}
	test_status := status.Status{}

	e = loadServerConfig(options, &test_status)
	c.Assert(e, check.IsNil)

	server_value_json, _, e := test_status.GetJson("status://server")

	// And the value we pulled out of the Status matches the file contents.
	c.Assert(string(server_value_json), check.Equals, `{"foo":"bar"}`)
}

func (suite *MySuite) TestFindOptions(c *check.C) {
	var options Options
	var e error

	options, e = findOptions()
	c.Check(e, check.IsNil)
	c.Check(
		options.config_dir,
		check.Matches,
		".*github.com/DonGar/go-house/_test")

}
