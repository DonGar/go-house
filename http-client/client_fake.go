package httpclient

import (
	"gopkg.in/check.v1"
)

//
// Define a fake implementation.
//

type FakeResult struct {
	url, result string
	err         error
}

type HttpClientFake struct {
	c        *check.C
	Expected []FakeResult
	replay   int
}

func NewHttpClientFake(c *check.C, results []FakeResult) *HttpClientFake {
	return &HttpClientFake{c, results, 0}
}

func (a *HttpClientFake) UrlToBytes(requestUrl string) (result []byte, err error) {
	if a.replay >= len(a.Expected) {
		a.c.Fatal("Received more requests than expected.")
	}

	a.c.Assert(requestUrl, check.Equals, a.Expected[a.replay].url)

	if a.Expected[a.replay].err != nil {
		result, err = nil, a.Expected[a.replay].err
	} else {
		result, err = []byte(a.Expected[a.replay].result), nil
	}

	a.replay++
	return result, err
}

func (a *HttpClientFake) CheckConsumed() {
	a.c.Check(a.replay, check.Equals, len(a.Expected))
}
