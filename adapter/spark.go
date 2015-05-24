package adapter

import (
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/spark-api"
	"github.com/DonGar/go-house/status"
	"log"
	"path/filepath"
	"strings"
)

type sparkAdapter struct {
	base
	sparkapi.SparkApiInterface
	actionName string
	actionsMgr *actions.Manager
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

	// Create an start adapter.
	sa := &sparkAdapter{
		b,
		sparkapi.NewSparkApi(username, password),
		filepath.Base(b.adapterUrl) + ".function",
		m.actionsMgr,
	}

	go sa.Handler()

	return sa, nil
}

func (a *sparkAdapter) Handler() {
	err := a.actionsMgr.RegisterAction(a.actionName, a.functionAction)
	if err != nil {
		panic(err)
	}

	// Create the root for the core devices we are about to discover.
	err = a.status.SetJson(a.adapterUrl+"/core", []byte(`{}`), status.UNCHECKED_REVISION)
	if err != nil {
		panic(err)
	}

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
			err := a.actionsMgr.UnRegisterAction(a.actionName)
			if err != nil {
				panic(err)
			}

			a.StopChan <- true
			return
		}
	}
}

func (a *sparkAdapter) Stop() {
	a.SparkApiInterface.Stop()
	a.base.Stop()
}

func (a sparkAdapter) setJsonOrString(url, data string) {
	// Save the given value into status. Try to handle the value as Json,
	// but store as a string otherwise.
	err := a.status.SetJson(url, []byte(data), status.UNCHECKED_REVISION)
	if err != nil {
		// Not valid JSON, treat as a simple string.
		err = a.status.Set(url, data, status.UNCHECKED_REVISION)
		if err != nil {
			// Not possible/hard to handle
			panic(err)
		}
	}
}

func (a sparkAdapter) updateFromEvent(event sparkapi.Event) {

	device_url := a.findDeviceUrl(event.CoreId)
	if device_url == "" {
		// If we have no associated device (yet), ignore the event.
		log.Println("Spark: Received event from unknown device: ", event.CoreId)
		return
	}

	if strings.HasPrefix(event.Name, "spark") {
		log.Printf("Spark: Ignoring system event %s from %s.\n", event.Name, event.CoreId)
		return
	}

	//
	// Publish detailed event information.
	//
	event_url := device_url + "/details/events/" + event.Name

	// Event data may be in JSON.
	data_url := event_url + "/data"
	a.setJsonOrString(data_url, event.Data)
	a.status.Set(event_url+"/published", event.Published_at, status.UNCHECKED_REVISION)

	//
	// Publish event as device property.
	//
	property_url := device_url + "/" + event.Name
	a.setJsonOrString(property_url, event.Data)
}

func (a sparkAdapter) findDeviceUrl(id string) string {

	// Find the id's of all devices.
	devices_url := a.adapterUrl + "/core/*/details/id"

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
			search_url = search_url[:last_break]

			// Strip off /details from the end of the URL.
			last_break = strings.LastIndex(search_url, "/")
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
		panic(err)
	}

OldNames:
	for _, old := range oldNames {
		for _, d := range devices {
			if d.Name == old {
				continue OldNames // This name still exists, leave it.
			}
		}

		err = a.status.Remove(core_url+"/"+old, status.UNCHECKED_REVISION)
		if err != nil {
			panic(err)
		}
	}

	// Add/update devices that exist.
	for _, d := range devices {
		device_details_url := core_url + "/" + d.Name + "/details"

		wasConnected := a.status.GetBoolWithDefault(device_details_url+"/connected", false)

		funcNames := make([]interface{}, len(d.Functions))
		for i, name := range d.Functions {
			funcNames[i] = name
		}

		if d.Connected && !wasConnected {
			for _, name := range d.Functions {
				// Refresh is called for devices that just connected, and after server
				// restart to allow current events to be resent.
				if name == "refresh" {
					// Request refresh in background, to avoid slowing devices update.
					go a.callFunction(d.Name, "refresh", "")
				}
			}
		}

		coreDetails := map[string]interface{}{
			"id":         d.Id,
			"last_heard": d.LastHeard,
			"connected":  d.Connected,
			"variables":  d.Variables,
			"functions":  funcNames,
		}

		// If the device existed, and we had events for it, preserve them.
		events, _, _ := a.status.Get(device_details_url + "/events")
		if events != nil {
			coreDetails["events"] = events
		}

		err = a.status.Set(device_details_url, coreDetails, status.UNCHECKED_REVISION)
		if err != nil {
			panic(err)
		}
	}
}

func (a sparkAdapter) functionAction(s *status.Status, action *status.Status) (e error) {
	device, err := action.GetString("status://device")
	if err != nil {
		return err
	}

	function, err := action.GetString("status://function")
	if err != nil {
		return err
	}

	argument, err := action.GetString("status://argument")
	if err != nil {
		return err
	}

	// Log detailed results.
	return a.callFunction(device, function, argument)
}

func (a sparkAdapter) callFunction(device, function, argument string) (e error) {
	result, err := a.SparkApiInterface.CallFunction(device, function, argument)
	if err == nil {
		log.Printf("Spark %s.%s(%s) result: %d\n", device, function, argument, result)
	} else {
		log.Printf("Spark %s.%s(%s) error: %s\n", device, function, argument, err)
	}

	return err
}
