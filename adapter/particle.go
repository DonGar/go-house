package adapter

import (
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/particle-api"
	"github.com/DonGar/go-house/status"
	"log"
	"path/filepath"
	"strings"
)

type particleAdapter struct {
	base
	particleapi.ParticleApiInterface
	actionName  string
	actionsMgr  *actions.Manager
	targetWatch <-chan status.UrlMatches
}

func newParticleAdapter(m *Manager, b base) (a adapter, e error) {
	//
	// Look up config values.
	//
	username, _, e := b.config.GetString("status://username")
	if e != nil {
		return nil, e
	}

	password, _, e := b.config.GetString("status://password")
	if e != nil {
		return nil, e
	}

	watch, e := b.status.WatchForUpdate(b.adapterUrl + "/core/*/*")
	if e != nil {
		return nil, e
	}

	// Create an start adapter.
	sa := &particleAdapter{
		b,
		particleapi.NewParticleApi(username, password),
		filepath.Base(b.adapterUrl) + ".function",
		m.actionsMgr,
		watch,
	}

	go sa.Handler()

	return sa, nil
}

func (a *particleAdapter) Handler() {
	err := a.actionsMgr.RegisterAction(a.actionName, a.functionAction)
	if err != nil {
		panic(err)
	}

	// Create the root for the core devices we are about to discover.
	err = a.status.SetJson(a.adapterUrl+"/core", []byte(`{}`), status.UNCHECKED_REVISION)
	if err != nil {
		panic(err)
	}

	deviceUpdates, events := a.ParticleApiInterface.Updates()

	for {
		select {
		case devices := <-deviceUpdates:
			log.Printf("Particle: Got devices. %+v\n", devices)
			a.updateDeviceList(devices)

		case event := <-events:
			log.Printf("Particle: Got event. %+v\n", event)
			a.updateFromEvent(event)

		case matches := <-a.targetWatch:
			// Don't log, since this often fires when there is no action to take.
			a.checkForTargetToFire(matches)

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

func (a *particleAdapter) Stop() {
	a.ParticleApiInterface.Stop()
	a.status.ReleaseWatch(a.targetWatch)
	a.base.Stop()
}

func (a particleAdapter) updateFromEvent(event particleapi.Event) {

	device_url := a.findDeviceUrl(event.CoreId)
	if device_url == "" {
		// If we have no associated device (yet), ignore the event.
		log.Println("Particle: Received event from unknown device: ", event.CoreId)
		return
	}

	if strings.HasPrefix(event.Name, "spark/") ||
		strings.HasPrefix(event.Name, "particle/") {
		log.Printf("Particle: Ignoring system event %s from %s.\n", event.Name, event.CoreId)
		return
	}

	//
	// Publish detailed event information.
	//
	event_url := device_url + "/details/events/" + event.Name

	// Event data may be in JSON.
	data_url := event_url + "/data"
	a.status.SetJsonOrString(data_url, event.Data, status.UNCHECKED_REVISION)
	a.status.Set(event_url+"/published", event.Published_at, status.UNCHECKED_REVISION)

	//
	// Publish event as device property.
	//
	property_url := device_url + "/" + event.Name
	a.status.SetJsonOrString(property_url, event.Data, status.UNCHECKED_REVISION)
}

func (a *particleAdapter) checkForTargetToFire(matches status.UrlMatches) {
	for target_url, raw_value := range matches {
		// If the target was updated to 'nil', we can ignore it.
		if raw_value.Value == nil {
			continue
		}

		// Parse URL to find device and target names.

		inside_adapter := target_url[len(a.adapterUrl+"/core/"):]

		// Valid inside_adapter is expected to be of the form: <device>/<property_target>
		device_end := strings.Index(inside_adapter, "/")
		device := inside_adapter[:device_end]
		target := inside_adapter[device_end+1:]

		if !strings.HasSuffix(target, "_target") {
			continue
		}

		argument, revision, err := a.status.GetStringOrJson(target_url)
		if revision != raw_value.Revision || err != nil {
			continue
		}

		// We ignore results, but they are logged.
		a.ParticleApiInterface.CallFunctionAsync(device, target, string(argument))

		// Clear the target value. Again, ignore error. The most likely cause
		// is that someone else updated the target again, which doesn't bother us.
		a.status.Set(target_url, nil, raw_value.Revision)
	}
}

func (a particleAdapter) findDeviceUrl(id string) string {

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
		search_id, ok := raw_search_id.Value.(string)
		if !ok {
			continue
		}

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

func (a particleAdapter) updateDeviceList(devices []particleapi.Device) {
	core_url := a.adapterUrl + "/core"

	// Remove any old cores that don't exist any more.
	oldNames, _, err := a.status.GetChildNames(core_url)
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
		a.updateDevice(d)
	}
}

func (a particleAdapter) updateDevice(device particleapi.Device) {
	// Add/update devices that exist.
	safeName := status.EscapeUriElement(device.Name)

	device_details_url := a.adapterUrl + "/core" + "/" + safeName + "/details"

	wasConnected := a.status.GetBoolWithDefault(device_details_url+"/connected", false)

	funcNames := make([]interface{}, len(device.Functions))
	for i, name := range device.Functions {
		funcNames[i] = name
	}

	if device.Connected && !wasConnected {
		// Refresh is called for devices that just connected, and after server
		// restart to allow current events to be resent.
		a.callRefreshIfPresent(device)
	}

	deviceDetails := map[string]interface{}{
		"id":         device.Id,
		"last_heard": device.LastHeard,
		"connected":  device.Connected,
		"variables":  device.Variables,
		"functions":  funcNames,
	}

	// If the device existed, and we had events for it, preserve them.
	events, _, _ := a.status.Get(device_details_url + "/events")
	if events != nil {
		deviceDetails["events"] = events
	}

	err := a.status.Set(device_details_url, deviceDetails, status.UNCHECKED_REVISION)
	if err != nil {
		panic(err)
	}

	a.createEmptyTargets(device)
}

func (a particleAdapter) callRefreshIfPresent(device particleapi.Device) {
	for _, name := range device.Functions {
		if name == "refresh" {
			// Request refresh in background, to avoid slowing devices update.
			a.ParticleApiInterface.CallFunctionAsync(device.Name, "refresh", "")
		}
	}
}

func (a particleAdapter) createEmptyTargets(device particleapi.Device) (e error) {
	device_url := a.adapterUrl + "/core" + "/" + device.Name

	for _, name := range device.Functions {
		if strings.HasSuffix(name, "_target") {
			target_url := device_url + "/" + name
			// Ignore failures. They are expected if target existed.
			_ = a.status.Set(target_url, nil, status.NONEXISTENT)
		}
	}
	return nil
}

func (a particleAdapter) functionAction(s *status.Status, action *status.Status) (e error) {
	device, _, err := action.GetString("status://device")
	if err != nil {
		return err
	}

	function, _, err := action.GetString("status://function")
	if err != nil {
		return err
	}

	argument, _, err := action.GetString("status://argument")
	if err != nil {
		return err
	}

	// Log detailed results.
	_, err = a.ParticleApiInterface.CallFunction(device, function, argument)
	return err
}
