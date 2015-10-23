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
	f := NewHttpClientFake(c, nil)
	f.CheckConsumed()
}

func (suite *MySuite) TestSuccess(c *check.C) {
	expected := []FakeResult{
		{"http://request", "result", nil},
	}

	f := NewHttpClientFake(c, expected)

	result, err := f.UrlToBytes("http://request")
	c.Check(result, check.DeepEquals, []byte("result"))
	c.Check(err, check.IsNil)

	f.CheckConsumed()
}

func (suite *MySuite) TestError(c *check.C) {
	expectedErr := ResponseError{"Not Found", 404, "Can't find it."}
	expected := []FakeResult{
		{"http://request", "", expectedErr},
	}

	f := NewHttpClientFake(c, expected)

	result, err := f.UrlToBytes("http://request")
	c.Check(result, check.IsNil)
	c.Check(err, check.Equals, expectedErr)

	f.CheckConsumed()
}

func (suite *MySuite) TestSequence(c *check.C) {
	expectedErr := ResponseError{"Not Found", 404, "Ain't there."}
	expected := []FakeResult{
		{"http://request1", "result", nil},
		{"http://request2", "", expectedErr},
	}

	f := NewHttpClientFake(c, expected)

	result, err := f.UrlToBytes("http://request1")
	c.Check(result, check.DeepEquals, []byte("result"))
	c.Check(err, check.IsNil)

	result, err = f.UrlToBytes("http://request2")
	c.Check(result, check.IsNil)
	c.Check(err, check.Equals, expectedErr)

	f.CheckConsumed()
}
