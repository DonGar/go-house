package server

import (
	"github.com/DonGar/go-house/adapter"
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"
)

func setupStatusHandler(c *check.C) (statusHandler *StatusHandler) {
	status := &status.Status{}

	adapterMgr, e := adapter.NewManager(status)
	c.Assert(e, check.IsNil)

	return &StatusHandler{
		status:     status,
		adapterMgr: adapterMgr,
	}
}

func setupStatusHandlerWithAdapter(c *check.C) (statusHandler *StatusHandler) {
	status := &status.Status{}

	// Add a web adapter.
	e := status.SetJson("status://",
		[]byte(`
    {
      "server": {
        "adapters": {
          "web": {
            "adapter": {}
          }
        }
      }
    }`),
		0)
	c.Assert(e, check.IsNil)

	adapterMgr, e := adapter.NewManager(status)
	c.Assert(e, check.IsNil)

	return &StatusHandler{
		status:     status,
		adapterMgr: adapterMgr,
	}
}

func (suite *MySuite) TestUnknownMethod(c *check.C) {
	statusHandler := setupStatusHandler(c)

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
		"Method FOO not supported\n")
}

//
// Test GET Requests
//

func (suite *MySuite) TestGetMinimal(c *check.C) {
	statusHandler := setupStatusHandler(c)

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
		"{\n  \"revision\": 0,\n  \"status\": null\n}\n")
}

func (suite *MySuite) TestGetRevisionMismatch(c *check.C) {
	statusHandler := setupStatusHandler(c)

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
		"{\n  \"revision\": 0,\n  \"status\": null\n}\n")
}

func (suite *MySuite) TestGetRevisionMatch(c *check.C) {
	statusHandler := setupStatusHandler(c)

	request, e := http.NewRequest("GET", "http://example.com/status/?revision=0", nil)
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Wake up after a while and cause a change to complete.
	go func() {
		time.Sleep(10 * time.Millisecond)
		statusHandler.status.Set("status://", 1, 0)
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
		"{\n  \"revision\": 1,\n  \"status\": 1\n}\n")
}

func (suite *MySuite) TestGetUnknownStatusPath(c *check.C) {
	statusHandler := setupStatusHandler(c)

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
	statusHandler := setupStatusHandler(c)

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
	statusHandler := setupStatusHandler(c)

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
		"{\n  \"revision\": 0,\n  \"status\": null\n}\n")
}

func (suite *MySuite) TestPostRevisionMismatch(c *check.C) {
	statusHandler := setupStatusHandler(c)

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
		"{\n  \"revision\": 0,\n  \"status\": null\n}\n")
}

func (suite *MySuite) TestPostUnknownStatusPath(c *check.C) {
	statusHandler := setupStatusHandler(c)

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
	statusHandler := setupStatusHandlerWithAdapter(c)

	request, e := http.NewRequest(
		"PUT", "http://example.com/status/adapter/", strings.NewReader(`{"foo": "bar"}`))
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
	statusHandler := setupStatusHandlerWithAdapter(c)

	request, e := http.NewRequest(
		"PUT", "http://example.com/status/adapter?revision=33", strings.NewReader(`{"foo": "bar"}`))
	c.Assert(e, check.IsNil)

	response := httptest.NewRecorder()

	// Perform the query
	statusHandler.ServeHTTP(response, request)

	// Validate the result.
	c.Check(response.Code, check.Equals, 400)
	c.Check(response.HeaderMap, check.DeepEquals, http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}})
	c.Check(response.Body.String(), check.Equals, "Status: Invalid Revision: 33 - status://adapter. Expected 2\n")
}

func (suite *MySuite) TestPutRevisionMatchUrl(c *check.C) {
	statusHandler := setupStatusHandlerWithAdapter(c)

	request, e := http.NewRequest(
		"PUT", "http://example.com/status/adapter?revision=2", strings.NewReader(`{"foo": "bar"}`))
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
	statusHandler := setupStatusHandlerWithAdapter(c)

	request, e := http.NewRequest(
		"PUT", "http://example.com/status/adapter/*/foo", strings.NewReader(`{"foo": "bar"}`))
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
		"Status: Wildcards not allowed here: status://adapter/*/foo\n")
}

func (suite *MySuite) TestPutNoneAdapterPath(c *check.C) {
	statusHandler := setupStatusHandlerWithAdapter(c)

	request, e := http.NewRequest(
		"PUT", "http://example.com/status/foo/", strings.NewReader(`{"foo": "bar"}`))
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
		"No adapter for status://foo/.\n")
}
