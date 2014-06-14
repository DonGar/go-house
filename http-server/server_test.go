package server

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (suite *MySuite) TestUnknownMethod(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	request, e := http.NewRequest("FOO", "http://example.com/status/", nil)
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 405)
	c.Check(
		response.HeaderMap,
		check.DeepEquals,
		http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}})
	c.Check(
		response.Body.String(),
		check.Equals,
		"Method FOO no supported\n")
}

//
// Test GET Requests
//

func (suite *MySuite) TestGetMinimal(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	request, e := http.NewRequest("GET", "http://example.com/status/", nil)
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 200)
	c.Check(
		response.HeaderMap,
		check.DeepEquals,
		http.Header{"Content-Type": []string{"application/json"}})
	c.Check(
		response.Body.String(),
		check.Equals,
		`{"revision":0,"status":null}`)
}

func (suite *MySuite) TestGetRevisionMismatch(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	request, e := http.NewRequest("GET", "http://example.com/status/?revision=11", nil)
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 200)
	c.Check(
		response.HeaderMap,
		check.DeepEquals,
		http.Header{"Content-Type": []string{"application/json"}})
	c.Check(
		response.Body.String(),
		check.Equals,
		`{"revision":0,"status":null}`)
}

func (suite *MySuite) TestGetRevisionMatch(c *check.C) {
	status := status.Status{}
	statusHandler := StatusHandler{status: &status}

	request, e := http.NewRequest("GET", "http://example.com/status/?revision=0", nil)
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Wake up after a while and cause a change to complete.
	go func() {
		time.Sleep(10 * time.Millisecond)
		status.Set("status://", 1, 0)
	}()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate that we got the updated result.
	c.Check(response.Code, check.Equals, 200)
	c.Check(
		response.HeaderMap,
		check.DeepEquals,
		http.Header{"Content-Type": []string{"application/json"}})
	c.Check(
		response.Body.String(),
		check.Equals,
		`{"revision":1,"status":1}`)
}

func (suite *MySuite) TestGetUnknownStatusPath(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	request, e := http.NewRequest("GET", "http://example.com/status/foo", nil)
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 404)
	c.Check(
		response.HeaderMap,
		check.DeepEquals,
		http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}})
	c.Check(
		response.Body.String(),
		check.Equals,
		"Status url not found: status://foo\n")
}

func (suite *MySuite) TestGetWildcardStatusPath(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	request, e := http.NewRequest("GET", "http://example.com/status/foo/*", nil)
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 400)
	c.Check(
		response.HeaderMap,
		check.DeepEquals,
		http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}})
	c.Check(
		response.Body.String(),
		check.Equals,
		"Status: Wildcards not allowed here: status://foo/*\n")
}

//
// Test POST  Requests
//

func (suite *MySuite) TestPostMinimal(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	requestData := url.Values{}
	request, e := http.NewRequest("POST", "http://example.com/status/",
		strings.NewReader(requestData.Encode()))
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 200)
	c.Check(
		response.HeaderMap,
		check.DeepEquals,
		http.Header{"Content-Type": []string{"application/json"}})
	c.Check(
		response.Body.String(),
		check.Equals,
		`{"revision":0,"status":null}`)
}

func (suite *MySuite) TestPostRevisionMismatch(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	requestData := url.Values{"revision": []string{"11"}}
	request, e := http.NewRequest("POST", "http://example.com/status/",
		strings.NewReader(requestData.Encode()))
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 200)
	c.Check(
		response.HeaderMap,
		check.DeepEquals,
		http.Header{"Content-Type": []string{"application/json"}})
	c.Check(
		response.Body.String(),
		check.Equals,
		`{"revision":0,"status":null}`)
}

func (suite *MySuite) TestPostUnknownStatusPath(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	request, e := http.NewRequest("POST", "http://example.com/status/foo", nil)
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 404)
	c.Check(
		response.HeaderMap,
		check.DeepEquals,
		http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}})
	c.Check(
		response.Body.String(),
		check.Equals,
		"Status url not found: status://foo\n")
}

//
// Test PUT  Requests
//

func (suite *MySuite) TestPutMinimal(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	request, e := http.NewRequest(
		"PUT", "http://example.com/status/", strings.NewReader(`{"foo": "bar"}`))
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 200)
	c.Check(response.HeaderMap, check.DeepEquals, http.Header{})
	c.Check(response.Body.String(), check.Equals, "")
}

func (suite *MySuite) TestPutRevisionMismatch(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	request, e := http.NewRequest(
		"PUT", "http://example.com/status/?revision=3", strings.NewReader(`{"foo": "bar"}`))
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 400)
	c.Check(response.HeaderMap, check.DeepEquals, http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}})
	c.Check(response.Body.String(), check.Equals, "Status: Invalid Revision: 3 - status://. Expected 0\n")
}

func (suite *MySuite) TestPutRevisionMatchUrl(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	request, e := http.NewRequest(
		"PUT", "http://example.com/status/?revision=0", strings.NewReader(`{"foo": "bar"}`))
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 200)
	c.Check(response.HeaderMap, check.DeepEquals, http.Header{})
	c.Check(response.Body.String(), check.Equals, "")
}

func (suite *MySuite) TestPutWildcardStatusPath(c *check.C) {
	statusHandler := StatusHandler{status: &status.Status{}}

	request, e := http.NewRequest(
		"PUT", "http://example.com/status/*/foo", strings.NewReader(`{"foo": "bar"}`))
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 400)
	c.Check(
		response.HeaderMap,
		check.DeepEquals,
		http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}})
	c.Check(
		response.Body.String(),
		check.Equals,
		"Status: Wildcards not allowed here: status://*/foo\n")
}
