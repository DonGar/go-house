package httpclient

import (
	"io"
	"io/ioutil"
	"net/http"
)

//
// Helper methods to make it easier to use HttpClient.
//

func UrlToReadCloser(hc HttpClientInterface, requestUrl string) (body io.ReadCloser, err error) {
	request, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}

	return hc.RequestToReadCloser(request)
}

func UrlToBytes(hc HttpClientInterface, requestUrl string) ([]byte, error) {
	body, err := UrlToReadCloser(hc, requestUrl)

	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(body)
}
