package sparkapi

import (
	"gopkg.in/check.v1"
	"net/http"
)

func (suite *MySuite) TestUrlToResponse(c *check.C) {
	// Test starting out with an empty token. We should find
	// a new valid one, then succeed.

	if !*network {
		c.Skip("-network tests not enabled.")
	}

	sa := NewSparkApi(TEST_USER, TEST_PASS)

	// We failed with a BadRequest error.
	response, err := sa.urlToResponseWithToken(DEVICES_URL)
	responseError, ok := err.(ResponseError)
	c.Check(ok, check.Equals, true)
	c.Check(responseError.StatusCode, check.Equals, http.StatusBadRequest)
	c.Check(response, check.IsNil)

	// Do a token refresh.
	response, err = sa.urlToResponseWithTokenRefresh(DEVICES_URL)
	c.Check(err, check.IsNil)
	c.Check(response, check.NotNil)
	response.Body.Close()

	// We have a new token.
	c.Check(sa.token, check.Not(check.Equals), "")

	// Redo the original request, and it works.
	response, err = sa.urlToResponseWithToken(DEVICES_URL)
	c.Check(err, check.IsNil)
	c.Check(response, check.NotNil)
	response.Body.Close()

	sa.Stop()
}

func (suite *MySuite) TestUrlLookup(c *check.C) {
	// Test starting out with an empty token. We should find
	// a new valid one, then succeed.

	if !*network {
		c.Skip("-network tests not enabled.")
	}

	// We test creating before looking up to ensure there is always something
	// to look up.

	// Verify that we can create a new token.
	token, err := refreshToken(TEST_USER, TEST_PASS)
	c.Check(err, check.IsNil)
	c.Check(token, check.Not(check.Equals), "")

	// Verify that we can lookup ao test token.
	token, err = lookupToken(TEST_USER, TEST_PASS)
	c.Check(err, check.IsNil)
	c.Check(token, check.Not(check.Equals), "")
}

func (suite *MySuite) TestUrlToResponseBadUser(c *check.C) {
	// Test starting out with a bad token, and bad user data.

	if !*network {
		c.Skip("-network tests not enabled.")
	}

	sa := NewSparkApi("", "")

	// Do a token refresh.
	response, err := sa.urlToResponseWithTokenRefresh(DEVICES_URL)
	responseError, ok := err.(ResponseError)
	c.Check(ok, check.Equals, true)
	c.Check(responseError.StatusCode, check.Equals, http.StatusBadRequest)
	c.Check(response, check.IsNil)

	sa.Stop()
}
