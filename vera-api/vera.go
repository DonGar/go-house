package veraapi

import (
	"encoding/json"
	"fmt"
	"github.com/DonGar/go-house/stoppable"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type VeraApiInterface interface {
	Updates() <-chan []Device
	Stop()
}

type ResponseError struct {
	Status     string
	StatusCode int
	BodyText   string
}

func (r ResponseError) Error() string {
	return "Request got code: " + r.Status + "\nBody:\n" + r.BodyText
}

type VeraApi struct {
	stoppable.Base

	// API Configuration.
	hostname string

	// Track current known devices.
	devices []Device

	// Publish to external listeners our current known devices.
	// Will have nil values, if nobody is listening.
	deviceUpdates chan []Device

	// Internal request channels.
	refreshDevices chan bool
}

func NewVeraApi(hostname string) *VeraApi {
	a := &VeraApi{
		stoppable.NewBase(),
		hostname,
		[]Device{},
		nil,
		make(chan bool),
	}

	// Start our background thread.
	go a.handler()

	return a
}

func (a *VeraApi) Updates() <-chan []Device {
	if a.deviceUpdates == nil {
		a.deviceUpdates = make(chan []Device)
		a.refreshDevices <- true
	}

	return a.deviceUpdates
}

func (a *VeraApi) handler() {
	var err error

	// We don't need an exact timing, just... occasionally.
	refreshTimer := time.NewTicker(24 * time.Hour)

	// Start our event handling loop.
	for {
		select {
		case <-a.StopChan:
			refreshTimer.Stop()
			a.StopChan <- true
			return

		case <-refreshTimer.C:
			// Refresh our device list, at least once a day.
			go func() {
				if a.deviceUpdates != nil {
					a.refreshDevices <- true
				}
			}()

		case <-a.refreshDevices:
			log.Println("Vera: Requesting devices from:", a.hostname)

			a.devices, err = a.discoverDevices()
			if err != nil {
				log.Println("Vera: discoverDevices failed: ", err.Error())
			}
			a.deviceUpdates <- a.devices
		}
	}
}

func (a *VeraApi) discoverDevices() ([]Device, error) {
	// Example: http://vera:3480/data_request?id=sdata
	requestUrl := fmt.Sprintf("http://%s:3480/data_request?id=sdata", a.hostname)

	// Lookup the list of devices, and discoverDeviceDetails on each.

	// Do the device lookup.
	bodyText, err := a.urlToBytes(requestUrl)
	if err != nil {
		return nil, err
	}

	return a.parseVeraData(bodyText)
}

func (a *VeraApi) parseVeraData(bodyText []byte) ([]Device, error) {

	type section struct {
		Name string
		Id   int
	}
	type room struct {
		Name    string
		Id      int
		Section int
	}
	type scene struct {
		Name   string
		Id     int
		Room   int
		Active int
	}
	type device struct {
		Name        string
		Id          int
		Category    int
		Subcategory int
		Room        int

		State int

		// Optional Category specific status variables.
		Armed        string
		Batterylevel string
		Status       string
		Temperature  string
		Light        string
		Humidity     string
		Tripped      string
		Armedtripped string
		Lasttrip     string
		Level        string
		Locked       string
	}
	type category struct {
		Name string
		Id   int
	}

	// Parse the response.
	var parsedResponse struct {
		Sections   []section
		Rooms      []room
		Scenes     []scene
		Devices    []device
		Categories []category
	}

	err := json.Unmarshal(bodyText, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("Can't unmarshel device: %s\n%s", err, string(bodyText))
	}

	sectionMap := map[int]section{}
	for _, v := range parsedResponse.Sections {
		sectionMap[v.Id] = v
	}

	roomMap := map[int]room{}
	for _, v := range parsedResponse.Rooms {
		roomMap[v.Id] = v
	}

	sceneMap := map[int]scene{}
	for _, v := range parsedResponse.Scenes {
		sceneMap[v.Id] = v
	}

	deviceMap := map[int]device{}
	for _, v := range parsedResponse.Devices {
		deviceMap[v.Id] = v
	}

	categoryMap := map[int]category{}
	for _, v := range parsedResponse.Categories {
		categoryMap[v.Id] = v
	}

	devices := make([]Device, len(parsedResponse.Devices))

	// Fill out device details.
	for i, r := range parsedResponse.Devices {
		devices[i].Id = r.Id
		devices[i].Name = r.Name
		devices[i].Category = "foo"
		// devices[i].Category = r.Category
		// devices[i].Subcategory = r.Subcategory
		// devices[i].Room = r.Room
		devices[i].State = r.State

		// Look up more details.
		// err = a.discoverDeviceDetails(&devices[i])
		// if err != nil {
		// 	return nil, err
		// }
	}

	return devices, nil
}

func (a *VeraApi) urlToBytes(requestUrl string) ([]byte, error) {
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
