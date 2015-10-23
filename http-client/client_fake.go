package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type FakeResult struct {
	Result string
	Err    error
}

// A couple of handy standard results.
var SUCCESS FakeResult = FakeResult{"result", nil}
var NOT_FOUND FakeResult = FakeResult{"", ResponseError{"Not Found", 404, "Can't find it."}}

type ResultMap map[string]FakeResult

// Turn an io.Reader into an io.ReadCloser implementation.
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

//
// The Fake Interface Implementation.
//
type HttpClientFake struct {
	Results  ResultMap
	Default  FakeResult
	Recorded []string
}

func (a *HttpClientFake) RequestToReadCloser(request *http.Request) (body io.ReadCloser, err error) {
	if request.URL == nil {
		return nil, fmt.Errorf("No URL in request.")
	}

	requestUrl := request.URL.String()

	// Record the request.
	a.Recorded = append(a.Recorded, requestUrl)

	// Lookup result.
	result, ok := a.Results[requestUrl]
	if !ok {
		result = a.Default
	}

	// Return result.
	body = nopCloser{bytes.NewBufferString(result.Result)}
	return body, result.Err
}
