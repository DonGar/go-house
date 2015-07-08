package particleapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var OAUTH_URL string = PARTICLE_IO_URL + "oauth/token"
var TOKENS_URL string = PARTICLE_IO_URL + "v1/access_tokens"

//
// Type to encode HTTP Errors.
//
type ResponseError struct {
	Status     string
	StatusCode int
	BodyText   string
}

func (r ResponseError) Error() string {
	return "Request got code: " + r.Status + "\nBody:\n" + r.BodyText
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
		// Read the full response, ignore a read error.
		bodyText, _ := ioutil.ReadAll(response.Body)
		return nil, ResponseError{response.Status, response.StatusCode, string(bodyText)}
	}

	return response, err
}

func (a *ParticleApi) requestToResponseWithToken(request *http.Request) (*http.Response, error) {
	// Helper for performing an HTTP request, and getting back the body of
	// the response. Our access token is used for authorization.
	request.Header.Set("Authorization", "Bearer "+a.token)

	return requestToResponse(request)
}

func (a *ParticleApi) requestToResponseWithTokenRefresh(request *http.Request) (*http.Response, error) {
	// Helper for performing an HTTP request, and getting back the body of
	// the response. Our access token is used for authorization. Lookup/refresh
	// the token as needed.

	var err error

	// If we have a token, try to use it. It might fail if the token is
	// expired.
	if a.token != "" {
		response, err := a.requestToResponseWithToken(request)

		// If it worked, we are done.
		if err == nil {
			return response, err
		}

		// Exit if not a we are sure a token refresh won't help.
		responseError, ok := err.(ResponseError)
		if !ok || responseError.StatusCode != http.StatusBadRequest {
			return nil, err
		}
	}

	// Lookup for generate a new token. If we don't find one, try to create one.
	a.token, _ = lookupToken(a.username, a.password)
	if a.token == "" {
		a.token, err = refreshToken(a.username, a.password)
		if err != nil {
			return nil, err
		}
	}

	// Request with new token. If this fails, we are done.
	return a.requestToResponseWithToken(request)
}

func (a *ParticleApi) urlToResponseWithTokenRefresh(requestUrl string) (*http.Response, error) {
	// Perform requestToResponseWithTokenRefresh from a URL.

	request, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}

	return a.requestToResponseWithTokenRefresh(request)
}

func (a *ParticleApi) postToResponseWithTokenRefresh(postUrl string, postFormValues url.Values) (*http.Response, error) {
	// Perform requestToResponseWithTokenRefresh from a URL.

	request, err := http.NewRequest("POST", postUrl, strings.NewReader(postFormValues.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return a.requestToResponseWithTokenRefresh(request)
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

	// We only count tokens as valid if they still have > 12 hours left.
	valid := time.Now().Add(12 * time.Hour)

	// The server will return expired tokens in the response, so check dates for
	// a valid one.
	for _, token := range parsedResponse {

		tokenExpires, err := time.Parse(time.RFC3339, token.Expires_at)
		if err != nil {
			// If we can't parse the time on the token, we got bogus data from the
			// server. Report it up the chain.
			return "", err
		}

		if tokenExpires.After(valid) {
			// We found a valid token, use it!
			return token.Token, nil
		}
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
