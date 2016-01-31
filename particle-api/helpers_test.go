package particleapi

import (
	"github.com/DonGar/go-house/http-client"
	"gopkg.in/check.v1"
	"net/http"
)

func (suite *MySuite) TestUrlToResponse(c *check.C) {
	// Test starting out with an empty token. We should find
	// a new valid one, then succeed.

	if !*network {
		c.Skip("-network tests not enabled.")
	}

	sa := NewParticleApi(TEST_USER, TEST_PASS)

	// We failed with a BadRequest error.
	request, err := http.NewRequest("GET", DEVICES_URL, nil)
	c.Assert(err, check.IsNil)

	response, err := sa.requestToReadCloserWithToken(request)
	responseError, ok := err.(httpclient.ResponseError)
	c.Check(ok, check.Equals, true)
	c.Check(responseError.StatusCode, check.Equals, http.StatusBadRequest)
	c.Check(response, check.IsNil)

	// Do a token refresh.
	request, err = http.NewRequest("GET", DEVICES_URL, nil)
	c.Assert(err, check.IsNil)

	response, err = sa.requestToReadCloserWithTokenRefresh(request)
	c.Check(err, check.IsNil)
	c.Check(response, check.NotNil)
	response.Close()

	// We have a new token.
	c.Check(sa.token, check.Not(check.Equals), "")

	// Redo the original request, and it works.
	request, err = http.NewRequest("GET", DEVICES_URL, nil)
	c.Assert(err, check.IsNil)

	response, err = sa.requestToReadCloserWithToken(request)
	c.Check(err, check.IsNil)
	c.Check(response, check.NotNil)
	// response.Close()

	sa.Stop()
}

func (suite *MySuite) TestUrlLookup(c *check.C) {
	// Test starting out with an empty token. We should find
	// a new valid one, then succeed.

	if !*network {
		c.Skip("-network tests not enabled.")
	}

	hc := &httpclient.HttpClient{}

	// We test creating before looking up to ensure there is always something
	// to look up.

	// Verify that we can create a new token.
	token, err := refreshToken(hc, TEST_USER, TEST_PASS)
	c.Check(err, check.IsNil)
	c.Check(token, check.Not(check.Equals), "")

	// Verify that we can lookup ao test token.
	token, err = lookupToken(hc, TEST_USER, TEST_PASS)
	c.Check(err, check.IsNil)
	c.Check(token, check.Not(check.Equals), "")
}

func (suite *MySuite) TestUrlToResponseBadUser(c *check.C) {
	// Test starting out with a bad token, and bad user data.

	if !*network {
		c.Skip("-network tests not enabled.")
	}

	sa := NewParticleApi("", "")

	// Do a token refresh.
	response, err := sa.urlToReadCloserWithTokenRefresh(DEVICES_URL)
	responseError, ok := err.(httpclient.ResponseError)
	c.Check(ok, check.Equals, true)
	c.Check(responseError.StatusCode, check.Equals, http.StatusBadRequest)
	c.Check(response, check.IsNil)

	sa.Stop()
}
