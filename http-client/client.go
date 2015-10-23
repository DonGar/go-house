package httpclient

import (
	"io/ioutil"
	// "log"
	"net/http"
)

type HttpClientInterface interface {
	UrlToBytes(requestUrl string) ([]byte, error)
}

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

func (a *HttpClient) UrlToBytes(requestUrl string) ([]byte, error) {
	response, err := http.Get(requestUrl)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		// Read the full response, ignore a read error.
		bodyText, _ := ioutil.ReadAll(response.Body)
		return nil, ResponseError{response.Status, response.StatusCode, string(bodyText)}
	}

	bodyText, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return bodyText, nil
}
