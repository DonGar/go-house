package httpclient

import (
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestFakeInterfaceCompliance(c *check.C) {
	var f *HttpClientFake = nil
	var i HttpClientInterface

	// Compile time interface conformance test.
	i = f
	_ = i
}

func (suite *MySuite) TestNoRequests(c *check.C) {
	// Mostly a demo of how to check that there were no requests.
	f := HttpClientFake{}
	c.Check(len(f.Recorded), check.Equals, 0)
}

func (suite *MySuite) TestSuccess(c *check.C) {
	f := &HttpClientFake{Default: SUCCESS}

	result, err := UrlToBytes(f, "http://request")
	c.Check(result, check.DeepEquals, []byte(SUCCESS.Result))
	c.Check(err, check.IsNil)
	c.Check(f.Recorded, check.DeepEquals, []string{"http://request"})
}

func (suite *MySuite) TestError(c *check.C) {
	f := &HttpClientFake{Default: NOT_FOUND}

	result, err := UrlToBytes(f, "http://request")
	c.Check(result, check.IsNil)
	c.Check(err, check.Equals, NOT_FOUND.Err)
	c.Check(f.Recorded, check.DeepEquals, []string{"http://request"})
}

func (suite *MySuite) TestMixedRequests(c *check.C) {
	f := &HttpClientFake{
		Results: ResultMap{"http://present": SUCCESS},
		Default: NOT_FOUND}

	result, err := UrlToBytes(f, "http://request")
	c.Check(result, check.IsNil)
	c.Check(err, check.Equals, NOT_FOUND.Err)

	result, err = UrlToBytes(f, "http://present")
	c.Check(result, check.DeepEquals, []byte(SUCCESS.Result))
	c.Check(err, check.IsNil)

	result, err = UrlToBytes(f, "http://request")
	c.Check(result, check.IsNil)
	c.Check(err, check.Equals, NOT_FOUND.Err)

	result, err = UrlToBytes(f, "http://present")
	c.Check(result, check.DeepEquals, []byte(SUCCESS.Result))
	c.Check(err, check.IsNil)

	c.Check(f.Recorded, check.DeepEquals, []string{
		"http://request", "http://present",
		"http://request", "http://present",
	})
}
