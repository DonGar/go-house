package sparkapi

import (
	"bufio"
	"fmt"
	"github.com/DonGar/go-house/stoppable"
	"log"
	"net/http"
	"strings"
	"time"
)

var SPARK_IO_URL string = "https://api.particle.io/"

type SparkApiInterface interface {
	CallFunction(device, function, argument string) (int, error)
	CallFunctionAsync(device, function, argument string)
	Updates() (<-chan []Device, <-chan Event)
	Stop()
}

type funcResponse struct {
	result int
	err    error
}

type SparkApi struct {
	stoppable.Base

	// Our connection information.
	username string
	password string
	token    string

	// Track current known devices.
	devices []Device

	// Publish to external listeners our current known devices.
	// Will have nil values, if nobody is listening.
	deviceUpdates chan []Device
	events        chan Event
	funcCall      chan [3]string
	funcResponse  chan funcResponse

	// Internally trigger a refresh of our known devices.
	listenEvents   chan bool
	refreshDevices chan bool
}

func NewSparkApi(username, password string) *SparkApi {
	a := &SparkApi{
		stoppable.NewBase(),
		username, password, "",
		[]Device{},
		nil,
		nil,
		make(chan [3]string, 10),
		make(chan funcResponse),
		make(chan bool),
		make(chan bool),
	}

	// Start our background thread.
	go a.handler()

	return a
}

func (a *SparkApi) CallFunction(device, function, argument string) (result int, err error) {
	// This is a blocking call, but processed in handler routine.
	a.funcCall <- [3]string{device, function, argument}
	return a.waitForFunctionResult(device, function, argument)
}

func (a *SparkApi) CallFunctionAsync(device, function, argument string) {
	// Queue request synchronously, then log results async.
	// The sync'd queuing ensures in-order processing.
	a.funcCall <- [3]string{device, function, argument}
	go a.waitForFunctionResult(device, function, argument)
}

func (a *SparkApi) waitForFunctionResult(device, function, argument string) (result int, err error) {
	fullResult := <-a.funcResponse

	if fullResult.err == nil {
		log.Printf("Spark %s.%s(%s) result: %d\n", device, function, argument, fullResult.result)
	} else {
		log.Printf("Spark %s.%s(%s) error: %s\n", device, function, argument, fullResult.err)
	}

	return fullResult.result, fullResult.err
}

func (a *SparkApi) Updates() (<-chan []Device, <-chan Event) {
	if a.deviceUpdates == nil {
		a.deviceUpdates = make(chan []Device)
		a.events = make(chan Event)
		// Start listening for events, which also kicks off device refresh.
		go func() { a.listenEvents <- true }()
	}

	return a.deviceUpdates, a.events
}

func (a *SparkApi) sendDevicesUpdate(devices []Device) {

	if devices == nil {
		a.deviceUpdates <- nil
		return
	}

	// Make a copy of our Devices structure.
	updateValue := make([]Device, len(devices))
	for i := range a.devices {
		updateValue[i] = devices[i].Copy()
	}

	// Send out our Devices copy.
	a.deviceUpdates <- updateValue
}

func (a *SparkApi) readEvents(reader *bufio.Reader) error {
	for {
		event, err := parseEvent(reader)
		if err != nil {
			return err
		}

		if event == nil {
			continue
		}

		// If it's an update to a Core status, fully refresh.
		if strings.HasPrefix(event.Name, "spark/") {
			a.refreshDevices <- true
		}

		a.events <- *event
	}
}

func (a *SparkApi) reconnectAfterDelay() {
	// Try to reconnect the even listener, after a delay.
	time.Sleep(60 * time.Second)
	log.Println("Requesting new Spark API connection.")
	a.listenEvents <- true
}

func (a *SparkApi) handler() {
	var eventResponse *http.Response
	var err error

	// We don't need an exact timing, just... occasionally.
	refreshTimer := time.NewTicker(24 * time.Hour)

	// Start our event handling loop.
	for {
		select {
		case <-a.StopChan:
			refreshTimer.Stop()
			if eventResponse != nil {
				eventResponse.Body.Close()
			}

			a.StopChan <- true
			return

		case args := <-a.funcCall:
			device, function, argument := args[0], args[1], args[2]
			d := a.findDevice(device)

			if d != nil {
				result, err := a.callFunction(d, function, argument)
				a.funcResponse <- funcResponse{result, err}
			} else {
				err = fmt.Errorf("Can't find device '%s' to invoke %s\n", device, function)
				a.funcResponse <- funcResponse{-1, err}
			}

		case <-refreshTimer.C:
			// Refresh our device list, at least once a day.
			go func() {
				if a.deviceUpdates != nil {
					a.refreshDevices <- true
				}
			}()

		case <-a.refreshDevices:
			a.devices, err = a.discoverDevices()
			if err != nil {
				log.Println("discoverDevices failed: ", err.Error())
			}
			a.sendDevicesUpdate(a.devices)

		case <-a.listenEvents:
			var reader *bufio.Reader

			eventResponse, reader, err = a.openEventConnection()
			if err != nil {
				log.Println("openEventConnection failed: ", err.Error())
				go a.reconnectAfterDelay()
				continue
			}

			// We opened an event connection, listen to it.
			go func() {
				a.refreshDevices <- true

				// readEvents never returns without an error.
				err = a.readEvents(reader)
				log.Println("readEvents failed: ", err.Error())

				// Always reconnect.
				a.reconnectAfterDelay()
			}()
		}
	}
}
