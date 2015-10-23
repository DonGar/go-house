package httpclient

import (
	"io"
	"io/ioutil"
	"net/http"
)

type HttpClientInterface interface {
	RequestToReadCloser(request *http.Request) (body io.ReadCloser, err error)
}

//
// Used to report Http Error code responses.
//

type ResponseError struct {
	Status     string
	StatusCode int
	BodyText   string
}

func (r ResponseError) Error() string {
	return "Request got code: " + r.Status + "\nBody:\n" + r.BodyText
}

//
// Define a real implementation.
//

type HttpClient struct {
}

func (a *HttpClient) RequestToReadCloser(request *http.Request) (body io.ReadCloser, err error) {
	// Given an http.Request object, perform the request, and return the
	// body of the response.

	client := &http.Client{}
	response, err := client.Do(request)

	if response != nil && response.StatusCode != http.StatusOK {
		// Read the full response, ignore a read error.
		bodyText, _ := ioutil.ReadAll(response.Body)
		return nil, ResponseError{response.Status, response.StatusCode, string(bodyText)}
	}

	if err != nil {
		return nil, err
	}

	return response.Body, err
}
