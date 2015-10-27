package veraapi

import (
	"flag"
	"github.com/DonGar/go-house/http-client"
	"github.com/DonGar/go-house/wait"
	"gopkg.in/check.v1"
	"testing"
	"time"
)

// go test -network (in spark-api dir, only)
var network = flag.Bool("network", false, "Include networking tests")

const TEST_HOST = "foo_host"

const (
	READ_DELAY  time.Duration = 100 * time.Millisecond
	EMPTY_DELAY time.Duration = 5 * time.Millisecond
)

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
	fhc := &httpclient.HttpClientFake{Default: httpclient.SUCCESS}

	a := NewVeraApiWithHttp("fake-vera-hostname", fhc)
	a.Stop()
}

func waitDevicesRead(c *check.C, a *VeraApi) []Device {
	var result []Device

	updateObserved := func() bool {
		select {
		case result = <-a.Updates():
			return true
		default:
			return false
		}
	}

	// Retry until there is a value on the channel.
	received := wait.Wait(READ_DELAY, updateObserved)

	c.Check(received, check.Equals, true)
	return result
}

func validateNoDevices(c *check.C, a *VeraApi) {
	// Sleep a little for things to see if extra results are generated.
	time.Sleep(EMPTY_DELAY)
	select {
	case result := <-a.Updates():
		c.Error("Unexpected result received: ", result)
	default:
	}
}

func (suite *MySuite) TestUpdateErrorResponse(c *check.C) {

	fhc := &httpclient.HttpClientFake{
		Results: httpclient.ResultMap{
			"http://fake-vera-hostname:3480/data_request?id=sdata": httpclient.NOT_FOUND},
		Default: httpclient.NOT_FOUND,
	}

	a := NewVeraApiWithHttp("fake-vera-hostname", fhc)

	// Since the API never gets data, it should never report devices.
	validateNoDevices(c, a)
	a.Stop()
}

func (suite *MySuite) TestUpdateFullResponse(c *check.C) {

	// The initial query gets a full, valid response.
	fullResponse := httpclient.FakeResult{string(readFile(c, FULL_JSON)), nil}
	fhc := &httpclient.HttpClientFake{
		Results: httpclient.ResultMap{
			"http://fake-vera-hostname:3480/data_request?id=sdata": fullResponse,
		},
		Default: httpclient.NOT_FOUND,
	}

	a := NewVeraApiWithHttp("fake-vera-hostname", fhc)

	// We should get told about the cool new devices!
	result := waitDevicesRead(c, a)
	c.Check(result, check.HasLen, 50)

	validateNoDevices(c, a)
	a.Stop()

	c.Check(fhc.Recorded, check.DeepEquals, []string{
		"http://fake-vera-hostname:3480/data_request?id=sdata",
	})
}
