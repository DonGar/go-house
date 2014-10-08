package sparkapi

import (
	"flag"
	"gopkg.in/check.v1"
	"testing"
)

var network = flag.Bool("network", false, "Include networking tests")

var TEST_USER string = "house@bgb.cc"
var TEST_PASS string = "c4K4bJS&r4*o"
var TEST_TOKEN string = "8e909b47c7911138b9891e47f8af75557058f410"

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
	sa := NewSparkApi(TEST_USER, TEST_PASS, TEST_TOKEN)
	sa.Stop()
}

func (suite *MySuite) TestDeviceUpdates(c *check.C) {
	if !*network {
		c.Skip("-network tests not enabled.")
	}

	// Start and stop right away, without waiting for results.
	sa := NewSparkApi(TEST_USER, TEST_PASS, TEST_TOKEN)

	println("Blocking on devices.")
	devices := <-sa.DeviceUpdates()
	c.Check(devices, check.NotNil)

	sa.Stop()
}

func (suite *MySuite) TestDeviceUpdatesStop(c *check.C) {
	// Start and stop right away, without waiting for results.
	sa := NewSparkApi(TEST_USER, TEST_PASS, TEST_TOKEN)
	_ = sa.DeviceUpdates()
	sa.Stop()
}

// Not useful, because we have no way to know if there will
// be an event.

// func (suite *MySuite) TestEvents(c *check.C) {
// 	// Start and stop right away, without waiting for results.
// 	sa := NewSparkApi(TEST_USER, TEST_PASS, TEST_TOKEN)

// 	println("Blocking on event.")
// 	event := <-sa.Events()
// 	c.Check(event, check.NotNil)

// 	sa.Stop()
// }

func (suite *MySuite) TestEventsStop(c *check.C) {
	// Start and stop right away, without waiting for results.
	sa := NewSparkApi(TEST_USER, TEST_PASS, TEST_TOKEN)
	_ = sa.Events()
	sa.Stop()
}
