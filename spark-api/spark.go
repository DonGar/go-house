package sparkapi

import (
	"bufio"
	"github.com/DonGar/go-house/stoppable"
	"log"
	"net/http"
	"strings"
	"time"
)

var SPARK_IO_URL string = "https://api.spark.io/"

type SparkApiInterface interface {
	CallFunction(device, function, argument string) (int, error)
	Updates() (<-chan []Device, <-chan Event)
	Stop()
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

	// Internally trigger a refresh of our known devices.
	listenEvents   chan bool
	refreshDevices chan bool
}

func NewSparkApi(username, password string) *SparkApi {
	s := &SparkApi{
		stoppable.NewBase(),
		username, password, "",
		[]Device{},
		nil,
		nil,
		make(chan bool),
		make(chan bool),
	}

	// Start our background thread.
	go s.handler()

	return s
}

func (s *SparkApi) CallFunction(device, function, argument string) (int, error) {
	return -1, nil
}

func (s *SparkApi) Updates() (<-chan []Device, <-chan Event) {
	if s.deviceUpdates == nil {
		s.deviceUpdates = make(chan []Device)
		s.events = make(chan Event)
		// Start listening for events, which also kicks off device refresh.
		go func() { s.listenEvents <- true }()
	}

	return s.deviceUpdates, s.events
}

func (s *SparkApi) sendDevicesUpdate(devices []Device) {

	if devices == nil {
		s.deviceUpdates <- nil
		return
	}

	// Make a copy of our Devices structure.
	updateValue := make([]Device, len(devices))
	for i := range s.devices {
		updateValue[i] = devices[i].Copy()
	}

	// Send out our Devices copy.
	s.deviceUpdates <- updateValue
}

func (s *SparkApi) readEvents(reader *bufio.Reader) error {
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
			s.refreshDevices <- true
		}

		s.events <- *event
	}
}

func (s *SparkApi) reconnectAfterDelay() {
	// Try to reconnect the even listener, after a delay.
	time.Sleep(60 * time.Second)
	log.Println("Requesting new Spark API connection.")
	s.listenEvents <- true
}

func (s *SparkApi) handler() {
	var eventResponse *http.Response
	var err error

	// We don't need an exact timing, just... occasionally.
	refreshTimer := time.NewTicker(24 * time.Hour)

	// Start our event handling loop.
	for {
		select {
		case <-s.StopChan:
			refreshTimer.Stop()
			if eventResponse != nil {
				eventResponse.Body.Close()
			}

			s.StopChan <- true
			return

		case <-refreshTimer.C:
			// Refresh our device list, at least once a day.
			go func() {
				if s.deviceUpdates != nil {
					s.refreshDevices <- true
				}
			}()

		case <-s.refreshDevices:
			s.devices, err = s.discoverDevices()
			if err != nil {
				log.Println("discoverDevices failed: ", err.Error())
			}
			s.sendDevicesUpdate(s.devices)

		case <-s.listenEvents:
			var reader *bufio.Reader

			eventResponse, reader, err = s.openEventConnection()
			if err != nil {
				log.Println("openEventConnection failed: ", err.Error())
				go s.reconnectAfterDelay()
				continue
			}

			// We opened an event connection, listen to it.
			go func() {
				s.refreshDevices <- true

				// readEvents never returns without an error.
				err = s.readEvents(reader)
				log.Println("readEvents failed: ", err.Error())

				// Always reconnect.
				s.reconnectAfterDelay()
			}()
		}
	}
}
