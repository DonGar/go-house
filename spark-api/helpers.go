package sparkapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

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
	request.Header.Set("Authorization", "Bearer "+s.accessToken)

	return requestToResponse(request)
}

func (s *SparkApi) urlToResponseWithTokenRefresh(requestUrl string) (*http.Response, error) {
	// Helper for performing an HTTP request, and getting back the body of
	// the response. Our access token is used for authorization. Refresh
	// the token if it's not valid for any reason.

	bodyText, err := s.urlToResponseWithToken(requestUrl)

	// If it worked, we are done.
	if err == nil {
		return bodyText, err
	}

	// Exit if not a candidate for a bad/expired token.
	responseError, ok := err.(ResponseError)
	if !ok || responseError.StatusCode != http.StatusBadRequest {
		return nil, err
	}

	// Get a new token.
	err = s.refreshToken()
	if err != nil {
		return nil, err
	}

	// Request with new token.
	return s.urlToResponseWithToken(requestUrl)
}

func (s *SparkApi) refreshToken() (e error) {
	// Ask for a new token, and store it in out accessToken member.

	// Build our request.
	postFormValues := url.Values{
		"grant_type": {"password"},
		"username":   {s.username},
		"password":   {s.password},
	}

	request, err := http.NewRequest(
		"POST", OAUTH_URL, strings.NewReader(postFormValues.Encode()))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("spark", "spark")

	// Ask for the new token.
	response, err := requestToResponse(request)
	if err != nil {
		return err
	}

	// Read the full response.
	bodyText, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// Parse the response.
	var parsedResponse struct {
		Access_Token string
	}

	err = json.Unmarshal(bodyText, &parsedResponse)
	if err != nil {
		return err
	}

	// Save it on this adaptor.
	s.accessToken = parsedResponse.Access_Token
	return nil
}
