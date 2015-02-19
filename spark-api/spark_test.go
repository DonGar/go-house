package sparkapi

import (
	"flag"
	"gopkg.in/check.v1"
	"testing"
)

// go test -network (in spark-api dir, only)
var network = flag.Bool("network", false, "Include networking tests")

var TEST_USER string = "house@bgb.cc"
var TEST_PASS string = "c4K4bJS&r4*o"

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (suite *MySuite) TestInterfaceCompliance(c *check.C) {
	var sa *SparkApi = nil
	var i SparkApiInterface

	// Compile time interface conformance test.
	i = sa
	_ = i
}

func (suite *MySuite) TestStartStop(c *check.C) {
	// Start and stop right away, without waiting for results.
	sa := NewSparkApi(TEST_USER, TEST_PASS)
	sa.Stop()
}

func (suite *MySuite) TestUpdates(c *check.C) {
	if !*network {
		c.Skip("-network tests not enabled.")
	}

	// Start and stop right away, without waiting for results.
	sa := NewSparkApi(TEST_USER, TEST_PASS)

	println("Blocking on devices.")
	devicesChan, eventChan := sa.Updates()
	c.Check(devicesChan, check.NotNil)
	c.Check(eventChan, check.NotNil)

	devices := <-devicesChan
	c.Check(devices, check.NotNil)

	// Can't wait on eventChan, since there might not be any.

	sa.Stop()
}

func (suite *MySuite) TestUpdatesStop(c *check.C) {
	if !*network {
		c.Skip("-network tests not enabled.")
	}
	// Start and stop right away, without waiting for results.
	sa := NewSparkApi(TEST_USER, TEST_PASS)
	_, _ = sa.Updates()
	sa.Stop()
}

func (suite *MySuite) TestCallFunctionBadDevice(c *check.C) {
	if !*network {
		c.Skip("-network tests not enabled.")
	}

	sa := NewSparkApi(TEST_USER, TEST_PASS)

	// Invoke a function on a non-exitant device.
	_, err := sa.CallFunction("device_name", "function", "argument")
	c.Check(err, check.NotNil)

	sa.Stop()
}
