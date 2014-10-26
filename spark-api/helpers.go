package sparkapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var OAUTH_URL string = SPARK_IO_URL + "oauth/token"
var TOKENS_URL string = SPARK_IO_URL + "v1/access_tokens"

//
// Type to encode HTTP Errors.
//
type ResponseError struct {
	Status     string
	StatusCode int
}

func (r ResponseError) Error() string {
	return "Request got code: " + r.Status
}

// Helper method that makes a request, and reads the result body into an []byte.
func requestToResponse(request *http.Request) (*http.Response, error) {
	// Given an http.Request object, perform the request, and return the
	// body of the response.

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, ResponseError{response.Status, response.StatusCode}
	}

	return response, err
}

func (s *SparkApi) urlToResponseWithToken(requestUrl string) (*http.Response, error) {
	// Helper for performing an HTTP request, and getting back the body of
	// the response. Our access token is used for authorization.

	request, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+s.token)

	return requestToResponse(request)
}

func (s *SparkApi) urlToResponseWithTokenRefresh(requestUrl string) (*http.Response, error) {
	// Helper for performing an HTTP request, and getting back the body of
	// the response. Our access token is used for authorization. Lookup/refresh
	// the token as needed.

	var err error

	// If we have a token, try to use it. It might fail if the token is
	// expired.
	if s.token != "" {
		bodyText, err := s.urlToResponseWithToken(requestUrl)

		// If it worked, we are done.
		if err == nil {
			return bodyText, err
		}

		// Exit if not a we are sure a token refresh won't help.
		responseError, ok := err.(ResponseError)
		if !ok || responseError.StatusCode != http.StatusBadRequest {
			return nil, err
		}
	}

	// Lookup for generate a new token. If we don't find one, try to create one.
	s.token, _ = lookupToken(s.username, s.password)
	if s.token == "" {
		s.token, err = refreshToken(s.username, s.password)
		if err != nil {
			return nil, err
		}
	}

	// Request with new token. If this fails, we are done.
	return s.urlToResponseWithToken(requestUrl)
}

func lookupToken(username, password string) (token string, err error) {
	// Look for an existing token we can use for all requests.

	request, err := http.NewRequest("GET", TOKENS_URL, nil)
	if err != nil {
		return "", err
	}
	request.SetBasicAuth(username, password)

	// Ask for the new token.
	response, err := requestToResponse(request)
	if err != nil {
		return "", err
	}

	// Read the full response.
	bodyText, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var parsedResponse []struct {
		Token      string
		Expires_at string
		Client     string
	}

	err = json.Unmarshal(bodyText, &parsedResponse)
	if err != nil {
		return "", err
	}

	// We found it.
	if len(parsedResponse) > 0 {
		return parsedResponse[0].Token, nil
	}

	// We didn't find one.
	return "", nil
}

func refreshToken(username, password string) (token string, err error) {
	// Generate a new token we can use for future requests.

	// Build our request.
	postFormValues := url.Values{
		"grant_type": {"password"},
		"username":   {username},
		"password":   {password},
	}

	request, err := http.NewRequest(
		"POST", OAUTH_URL, strings.NewReader(postFormValues.Encode()))
	if err != nil {
		return "", err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("spark", "spark")

	// Ask for the new token.
	response, err := requestToResponse(request)
	if err != nil {
		return "", err
	}

	// Read the full response.
	bodyText, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// Parse the response.
	var parsedResponse struct {
		Access_Token string
	}

	err = json.Unmarshal(bodyText, &parsedResponse)
	if err != nil {
		return "", err
	}

	// We found it.
	return parsedResponse.Access_Token, nil
}
