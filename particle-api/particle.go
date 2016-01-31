package particleapi

import (
	"bufio"
	"fmt"
	"github.com/DonGar/go-house/http-client"
	"github.com/DonGar/go-house/stoppable"
	"io"
	"log"
	"strings"
	"time"
)

var PARTICLE_IO_URL string = "https://api.particle.io/"

type ParticleApiInterface interface {
	CallFunction(device, function, argument string) (int, error)
	CallFunctionAsync(device, function, argument string)
	Updates() (<-chan []Device, <-chan Event)
	Stop()
}

type funcResponse struct {
	result int
	err    error
}

type ParticleApi struct {
	stoppable.Base

	// Our connection information.
	username string
	password string
	token    string

	// Track current known devices.
	devices []Device

	// Helper that allows web requests to be mocked out.
	hc httpclient.HttpClientInterface

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

func NewParticleApi(username, password string) *ParticleApi {
	a := &ParticleApi{
		stoppable.NewBase(),
		username, password, "",
		[]Device{},
		&httpclient.HttpClient{},
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

func (a *ParticleApi) CallFunction(device, function, argument string) (result int, err error) {
	// This is a blocking call, but processed in handler routine.
	a.funcCall <- [3]string{device, function, argument}
	return a.waitForFunctionResult(device, function, argument)
}

func (a *ParticleApi) CallFunctionAsync(device, function, argument string) {
	// Queue request synchronously, then log results async.
	// The sync'd queuing ensures in-order processing.
	a.funcCall <- [3]string{device, function, argument}
	go a.waitForFunctionResult(device, function, argument)
}

func (a *ParticleApi) waitForFunctionResult(device, function, argument string) (result int, err error) {
	fullResult := <-a.funcResponse

	if fullResult.err == nil {
		log.Printf("Particle %s.%s(%s) result: %d\n", device, function, argument, fullResult.result)
	} else {
		log.Printf("Particle %s.%s(%s) error: %s\n", device, function, argument, fullResult.err)
	}

	return fullResult.result, fullResult.err
}

func (a *ParticleApi) Updates() (<-chan []Device, <-chan Event) {
	if a.deviceUpdates == nil {
		a.deviceUpdates = make(chan []Device)
		a.events = make(chan Event)
		// Start listening for events, which also kicks off device refresh.
		go func() { a.listenEvents <- true }()
	}

	return a.deviceUpdates, a.events
}

func (a *ParticleApi) sendDevicesUpdate(devices []Device) {

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

func (a *ParticleApi) readEvents(reader *bufio.Reader) error {
	for {
		event, err := parseEvent(reader)
		if err != nil {
			return err
		}

		if event == nil {
			continue
		}

		// If it's an update to a Core status, fully refresh.
		if strings.HasPrefix(event.Name, "spark/") ||
			strings.HasPrefix(event.Name, "particle/") {
			a.refreshDevices <- true
		}

		a.events <- *event
	}
}

func (a *ParticleApi) reconnectAfterDelay() {
	// Try to reconnect the even listener, after a delay.
	time.Sleep(60 * time.Second)
	log.Println("Requesting new Particle API connection.")
	a.listenEvents <- true
}

func (a *ParticleApi) handler() {
	var eventReaderCloser io.ReadCloser
	var err error

	// We don't need an exact timing, just... occasionally.
	refreshTimer := time.NewTicker(24 * time.Hour)

	// Start our event handling loop.
	for {
		select {
		case <-a.StopChan:
			refreshTimer.Stop()
			if eventReaderCloser != nil {
				eventReaderCloser.Close()
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
			var bufferedEventReader *bufio.Reader

			eventReaderCloser, bufferedEventReader, err = a.openEventConnection()
			if err != nil {
				log.Println("openEventConnection failed: ", err.Error())
				go a.reconnectAfterDelay()
				continue
			}

			// We opened an event connection, listen to it.
			go func() {
				a.refreshDevices <- true

				// readEvents never returns without an error.
				err = a.readEvents(bufferedEventReader)
				log.Println("readEvents failed: ", err.Error())

				// Always reconnect.
				a.reconnectAfterDelay()
			}()
		}
	}
}
