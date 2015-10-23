package httpclient

import (
	"flag"
	"gopkg.in/check.v1"
	"testing"
)

// go test -network (in spark-api dir, only)
var network = flag.Bool("network", false, "Include networking tests")

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (suite *MySuite) TestInterfaceCompliance(c *check.C) {
	var hc *HttpClient = nil
	var i HttpClientInterface

	// Compile time interface conformance test.
	i = hc
	_ = i
}

func (suite *MySuite) TestRealFetch(c *check.C) {
	if !*network {
		c.Skip("-network tests not enabled.")
	}

	var i HttpClientInterface = &HttpClient{}

	result, err := UrlToBytes(i, "http://www.google.com/")
	c.Check(result, check.NotNil)
	c.Check(err, check.IsNil)
}

func (suite *MySuite) TestFetchError(c *check.C) {
	if !*network {
		c.Skip("-network tests not enabled.")
	}

	var i HttpClientInterface = &HttpClient{}

	result, err := UrlToBytes(i, "http://www.google.com/nonexistent/babble/foo")
	c.Check(result, check.IsNil)
	c.Check(err, check.NotNil)
}

func (suite *MySuite) TestFetchBadUrl(c *check.C) {
	var i HttpClientInterface = &HttpClient{}

	result, err := UrlToBytes(i, "bogus://url/stuff")
	c.Check(result, check.IsNil)
	c.Check(err, check.NotNil)
}
