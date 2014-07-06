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

// End to end test that mixed config file settings, command line arguments, and
// default values.
func (suite *MySuite) TestIntializeServerConfig(c *check.C) {
	tempDir := c.MkDir()

	var e error

	configDir := filepath.Join(tempDir, "config")
	e = os.Mkdir(configDir, os.ModePerm)
	c.Assert(e, check.IsNil)

	// Write out a test server config file.
	serverFile := filepath.Join(configDir, "server.json")
	e = ioutil.WriteFile(serverFile, []byte(`
		{
			"foo": "bar",
			"port": 1701
		}`), os.ModePerm)
	c.Assert(e, check.IsNil)

	s := &status.Status{}
	e = s.Set(CONFIG_DIR, configDir, 0)
	c.Assert(e, check.IsNil)

	execName := filepath.Join(tempDir, "fake_exec")

	e = IntializeServerConfig(s, []string{
		execName, "--static_dir", "/args/static"})
	c.Check(e, check.IsNil)

	retrievedValues, _, e := s.Get("status://server")
	c.Check(e, check.IsNil)
	c.Check(
		retrievedValues,
		check.DeepEquals,
		map[string]interface{}{
			"port":      1701,
			"config":    configDir,
			"static":    "/args/static",
			"downloads": "/tmp/Downloads",
			"foo":       "bar",
		})
}

// Test parseFlags when all values are unspecified (pure defaults)
func (suite *MySuite) TestParseFlagsEmpty(c *check.C) {
	s := &status.Status{}

	configDir, e := parseFlags(s, []string{"/path/exec"})
	c.Check(e, check.IsNil)
	c.Check(configDir, check.Equals, "/path/config")

	retrievedValues, _, e := s.Get("status://server")
	c.Check(e, check.IsNil)
	c.Check(
		retrievedValues,
		check.DeepEquals,
		map[string]interface{}{
			"port":      80,
			"config":    "/path/config",
			"static":    "/path/static",
			"downloads": "/tmp/Downloads",
		})
}

// Test parseFlags when values are set, only in the command line.
func (suite *MySuite) TestParseFlagsArgs(c *check.C) {
	s := &status.Status{}

	configDir, e := parseFlags(s, []string{
		"/path/exec",
		"--port", "1702",
		"--config_dir", "/args/config",
		"--static_dir", "/args/static",
		"--downloads_dir", "/args/downloads",
	})
	c.Check(e, check.IsNil)
	c.Check(configDir, check.Equals, "/args/config")

	retrievedValues, _, e := s.Get("status://server")
	c.Check(e, check.IsNil)
	c.Check(
		retrievedValues,
		check.DeepEquals,
		map[string]interface{}{
			"port":      1702,
			"config":    "/args/config",
			"static":    "/args/static",
			"downloads": "/args/downloads",
		})
}

// Test parseFlags when values are set in the config, not the command line.
func (suite *MySuite) TestParseFlagsConfigSet(c *check.C) {
	s := &status.Status{}
	e := s.Set(
		"status://server",
		map[string]interface{}{
			"port":      2014,
			"static":    "/status/static",
			"downloads": "/status/downloads",
		},
		0)
	c.Check(e, check.IsNil)

	configDir, e := parseFlags(s, []string{"/path/exec"})
	c.Check(e, check.IsNil)
	c.Check(configDir, check.Equals, "/path/config")

	retrievedValues, _, e := s.Get("status://server")
	c.Check(e, check.IsNil)
	c.Check(
		retrievedValues,
		check.DeepEquals,
		map[string]interface{}{
			"port":      2014,
			"config":    "/path/config",
			"static":    "/status/static",
			"downloads": "/status/downloads",
		})
}

// Test parseFlags when values are set in both the config, and the command line.
func (suite *MySuite) TestParseFlagsConfigSetArgs(c *check.C) {
	s := &status.Status{}
	e := s.Set(
		"status://server",
		map[string]interface{}{
			"port":      2014,
			"static":    "/status/static",
			"downloads": "/status/downloads",
			"foo":       "bar",
		},
		0)
	c.Check(e, check.IsNil)

	configDir, e := parseFlags(s, []string{
		"/path/exec",
		"--port", "1702",
		"--config_dir", "/args/config",
		"--static_dir", "/args/static",
		"--downloads_dir", "/args/downloads",
	})
	c.Check(e, check.IsNil)
	c.Check(configDir, check.Equals, "/args/config")

	retrievedValues, _, e := s.Get("status://server")
	c.Check(e, check.IsNil)
	c.Check(
		retrievedValues,
		check.DeepEquals,
		map[string]interface{}{
			"port":      1702,
			"config":    "/args/config",
			"static":    "/args/static",
			"downloads": "/args/downloads",
			"foo":       "bar",
		})
}
