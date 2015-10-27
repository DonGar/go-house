package veraapi

import (
	"fmt"
	"github.com/DonGar/go-house/http-client"
	"github.com/DonGar/go-house/stoppable"
	"io"
	"io/ioutil"
	"log"
	"time"
)

type VeraApiInterface interface {
	Updates() <-chan []Device
	Stop()
}

type VeraApi struct {
	stoppable.Base

	// Helper for HTTP requests.
	httpclient.HttpClientInterface

	// API Configuration.
	hostname string

	// Values received during previous load.
	parseResult

	// Publish to external listeners our current known devices.
	// Will have nil values, if nobody is listening.
	deviceUpdates chan []Device

	// Internal request channels.
	refreshTimer *time.Timer
}

func NewVeraApi(hostname string) *VeraApi {
	// Use this to create real VeraApi instances.
	return NewVeraApiWithHttp(hostname, &httpclient.HttpClient{})
}

func NewVeraApiWithHttp(hostname string, httpClient httpclient.HttpClientInterface) (a *VeraApi) {
	// Mostly used to create instances with FakeHttpClients.
	a = &VeraApi{
		stoppable.NewBase(),
		httpClient,
		hostname,
		*newParseResult(),
		make(chan []Device, 1),
		time.NewTimer(0 * time.Second),
	}

	// Start our background thread.
	go a.handler()

	return a
}

func (a *VeraApi) Updates() <-chan []Device {
	return a.deviceUpdates
}

func (a *VeraApi) handler() {
	// Values for the blocking read we do. Kept here, so we can close on shutdown.
	var refreshReaderCloser io.ReadCloser
	var err error

	// Start our event handling loop.
	for {
		select {
		case <-a.StopChan:
			a.refreshTimer.Stop()
			if refreshReaderCloser != nil {
				refreshReaderCloser.Close()
			}
			a.StopChan <- true
			return

		case <-a.refreshTimer.C:
			// Ensure we only have a single refresh cycle in progress.
			log.Println("Vera: Requesting devices from:", a.hostname)
			refreshReaderCloser, err = httpclient.UrlToReadCloser(a, a.deviceUrl())
			if err != nil {
				a.handleRefreshError("open", err)
				continue
			}

			// We opened an device connection, listen for the blocking response.
			go func() {
				bodyText, err := ioutil.ReadAll(refreshReaderCloser)
				if err != nil {
					a.handleRefreshError("read", err)
					return
				}

				// log.Println("Vera parsing:", string(bodyText))
				result, err := parseVeraData(bodyText)
				if err != nil {
					a.handleRefreshError("parse", err)
					return
				}

				// Save off valid result.
				a.parseResult = *result

				a.sendDevicesNonBlocking()
				a.refreshTimer.Reset(2 * time.Second)
			}()
		}
	}
}

func (a *VeraApi) handleRefreshError(step string, err error) {
	log.Printf("refreshDevices %s failed: %s", step, err.Error())
	a.refreshTimer.Reset(30 * time.Second)
}

func (a *VeraApi) deviceUrl() string {
	// Example: http://vera:3480/data_request?id=sdata
	requestUrl := fmt.Sprintf("http://%s:3480/data_request?id=sdata", a.hostname)

	if a.loadtime != 0 || a.dataversion != 0 {
		requestUrl += fmt.Sprintf("&loadtime=%d&dataversion=%d&timeout=60&minimumdelay=1000",
			a.loadtime, a.dataversion)
	}

	return requestUrl
}

func (a *VeraApi) sendDevicesNonBlocking() {
	devices := make([]Device, 0, len(a.devices))
	for _, value := range a.devices {
		devices = append(devices, value)
	}

	log.Printf("Sending devices update with %d devices.", len(devices))

	// Make sure the send channel has room to send.
	select {
	case <-a.deviceUpdates:
	default:
	}

	// Now send.
	a.deviceUpdates <- devices
}
