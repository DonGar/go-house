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

func (a sparkAdapter) updateFromEvent(event sparkapi.Event) {

	device_url := a.findDeviceUrl(event.CoreId)
	if device_url == "" {
		// If we have no associated device (yet), ignore the event.
		log.Println("Spark: Received event from unknown device: ", event.CoreId)
		return
	}

	event_url := device_url + "/events/" + event.Name

	// Event data may be in JSON.
	data_url := event_url + "/data"
	err := a.status.SetJson(data_url, []byte(event.Data), status.UNCHECKED_REVISION)
	if err != nil {
		// Not valid JSON, treat as a simple string.
		err = a.status.Set(data_url, event.Data, status.UNCHECKED_REVISION)
		if err != nil {
			// Not possible/hard to handle
			panic(err)
		}
	}
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
		device_url := core_url + "/" + d.Name

		funcNames := make([]interface{}, len(d.Functions))
		for i, name := range d.Functions {
			funcNames[i] = name
		}

		actionsContents := map[string]interface{}{}
		for _, name := range d.Functions {
			actionContents := map[string]interface{}{
				"action":   a.actionName,
				"device":   d.Name,
				"function": name,
				"argument": "",
			}
			actionsContents[name] = actionContents
		}

		coreContents := map[string]interface{}{
			"id":         d.Id,
			"last_heard": d.LastHeard,
			"connected":  d.Connected,
			"variables":  d.Variables,
			"functions":  funcNames,
			"actions":    actionsContents,
		}

		err = a.status.Set(device_url, coreContents, status.UNCHECKED_REVISION)
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

	return a.SparkApiInterface.CallFunction(device, function, argument)
}
