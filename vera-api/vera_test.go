package veraapi

import (
	"flag"
	"gopkg.in/check.v1"
	"testing"
)

// go test -network (in spark-api dir, only)
var network = flag.Bool("network", false, "Include networking tests")

var TEST_HOST string = "foo_host"

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (suite *MySuite) TestInterfaceCompliance(c *check.C) {
	var sa *VeraApi = nil
	var i VeraApiInterface

	// Compile time interface conformance test.
	i = sa
	_ = i
}

func (suite *MySuite) TestStartStop(c *check.C) {
	// Start and stop right away, without waiting for results.
	sa := NewVeraApi(TEST_HOST)
	sa.Stop()
}

func (suite *MySuite) TestStartStopVera(c *check.C) {
	// Start and stop right away, without waiting for results.
	sa := NewVeraApi("vera")
	sa.Stop()
}
