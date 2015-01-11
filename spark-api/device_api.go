package sparkapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var DEVICES_URL string = SPARK_IO_URL + "v1/devices"

func (s *SparkApi) discoverDevices() ([]Device, error) {
	// Lookup the list of devices, and discoverDeviceDetails on each.

	// Do the device lookup.
	requestUrl := DEVICES_URL
	response, err := s.urlToResponseWithTokenRefresh(requestUrl)
	if err != nil {
		return nil, err
	}

	// Read the full response.
	bodyText, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Parse the response.
	var parsedResponse []struct {
		Id         string
		Name       string
		Last_heard string
		Connected  bool
	}

	err = json.Unmarshal(bodyText, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Can't unmarshel device: %s\n%s", err, string(bodyText))
	}

	devices := make([]Device, len(parsedResponse))

	// Fill out device details.
	for i, r := range parsedResponse {
		devices[i].Id = r.Id
		devices[i].Name = r.Name
		devices[i].LastHeard = r.Last_heard
		devices[i].Connected = r.Connected

		// Look up more details.
		err = s.discoverDeviceDetails(&devices[i])
		if err != nil {
			return nil, err
		}
	}

	return devices, nil
}

func (s *SparkApi) discoverDeviceDetails(device *Device) (e error) {
	// Lookup the Variable and Function names for the given Device.
	// Call lookupDeviceVariable for each variable.

	// Look up device details.
	requestUrl := DEVICES_URL + "/" + device.Id
	response, err := s.urlToResponseWithTokenRefresh(requestUrl)
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
		Id        string
		Variables map[string]interface{}
		Functions []string
	}

	err = json.Unmarshal(bodyText, &parsedResponse)
	if err != nil {
		return fmt.Errorf("Can't unmarshel device details: %s\n%s", err, string(bodyText))
	}

	// There are a wide variety of error responses, but none seem
	// to include an Id field.
	if parsedResponse.Id != device.Id {
		return fmt.Errorf("Error Response on Lookup: %s", string(bodyText))
	}

	// Save the details we looked up.
	device.Variables = map[string]interface{}{}
	for variable := range parsedResponse.Variables {
		device.Variables[variable] = nil

		err = s.lookupDeviceVariable(device, variable)
		if err != nil {
			return err
		}
	}

	device.Functions = parsedResponse.Functions

	return nil
}

func (s *SparkApi) lookupDeviceVariable(device *Device, variable string) (e error) {
	// Lookup and fill in the current value for a given variable on the Device.

	// Look up device details.
	requestUrl := DEVICES_URL + "/" + device.Id + "/" + variable
	response, err := s.urlToResponseWithTokenRefresh(requestUrl)
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
		Name   string
		Result interface{}
	}

	err = json.Unmarshal(bodyText, &parsedResponse)
	if err != nil {
		return err
	}

	// There are a wide variety of error responses, but none seem
	// to include an Id field.
	if parsedResponse.Name != variable {
		return fmt.Errorf("Error Response on Lookup: %s", string(bodyText))
	}

	// Save the value we looked up.
	device.Variables[variable] = parsedResponse.Result
	return nil
}
