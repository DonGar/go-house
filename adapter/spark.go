package adapter

import (
	"github.com/DonGar/go-house/spark-api"
	"github.com/DonGar/go-house/status"
	"log"
	"strings"
)

type sparkAdapter struct {
	base
	sparkapi.SparkApiInterface
}

func newSparkAdapter(m *Manager, b base) (a adapter, e error) {
	//
	// Look up config values.
	//
	username, e := b.config.GetString("status://username")
	if e != nil {
		return nil, e
	}

	password, e := b.config.GetString("status://password")
	if e != nil {
		return nil, e
	}

	token, e := b.config.GetString("status://token")
	if e != nil {
		token = ""
	}

	// Create an start adapter.
	sa := &sparkAdapter{
		b,
		sparkapi.NewSparkApi(username, password, token),
	}

	// Create the root for the core devices we are about to discover.
	e = sa.status.SetJson(sa.adapterUrl+"/core", []byte(`{}`), status.UNCHECKED_REVISION)

	go sa.Handler()

	return sa, nil
}

func (a *sparkAdapter) Handler() {

	deviceUpdates, events := a.SparkApiInterface.Updates()

	for {
		select {
		case devices := <-deviceUpdates:
			log.Printf("Spark: Got devices. %+v\n", devices)
			a.updateDeviceList(devices)

		case event := <-events:
			log.Printf("Spark: Got event. %+v\n", event)
			a.updateFromEvent(event)

		case <-a.StopChan:
			a.StopChan <- true
			return
		}
	}
}

func (a *sparkAdapter) Stop() {
	a.SparkApiInterface.Stop()
	a.base.Stop()
}

func (a sparkAdapter) updateFromEvent(event sparkapi.Event) {

	device_url := a.findDeviceUrl(event.CoreId)
	if device_url == "" {
		// If we have no associated device (yet), ignore the event.
		log.Println("Spark: Received event from unknown device: ", event.CoreId)
		return
	}

	event_url := device_url + "/events/" + event.Name

	a.status.Set(event_url+"/data", event.Data, status.UNCHECKED_REVISION)
	a.status.Set(event_url+"/published", event.Published_at, status.UNCHECKED_REVISION)
}

func (a sparkAdapter) findDeviceUrl(id string) string {

	// Find the id's of all devices.
	devices_url := a.adapterUrl + "/core/*/id"

	// Find all device URLs, the look for the one we want.
	matches, err := a.status.GetMatchingUrls(devices_url)
	if err != nil {
		// Not possible/hard to handle
		panic(err)
	}

	// Find the device with our requested id.
	for search_url, raw_search_id := range matches {
		search_id := raw_search_id.Value.(string)

		// If we found it.
		if search_id == id {
			// Strip off /id from the end of the URL.
			last_break := strings.LastIndex(search_url, "/")
			return search_url[:last_break]
		}
	}

	// No device, no url.
	return ""
}

func (a sparkAdapter) updateDeviceList(devices []sparkapi.Device) {
	core_url := a.adapterUrl + "/core"

	// Remove any old cores that don't exist any more.
	oldNames, err := a.status.GetChildNames(core_url)
	if err != nil {
		return
	}

OldNames:
	for _, old := range oldNames {
		for _, d := range devices {
			if d.Name == old {
				continue OldNames // This name still exists, leave it.
			}
		}

		a.status.Remove(core_url+"/"+old, status.UNCHECKED_REVISION)
	}

	// Add/update devices that exist.
	for _, d := range devices {
		device_url := core_url + "/" + d.Name
		a.status.Set(device_url+"/id", d.Id, status.UNCHECKED_REVISION)
		a.status.Set(device_url+"/last_heard", d.LastHeard, status.UNCHECKED_REVISION)
		a.status.Set(device_url+"/connected", d.Connected, status.UNCHECKED_REVISION)
		a.status.Set(device_url+"/variables", d.Variables, status.UNCHECKED_REVISION)
		a.status.Set(device_url+"/functions", d.Functions, status.UNCHECKED_REVISION)
	}
}
