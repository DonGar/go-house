package adapter

import (
	"github.com/DonGar/go-house/status"
	"github.com/DonGar/go-house/vera-api"
	"log"
	"strings"
)

type veraAdapter struct {
	base
	veraapi.VeraApiInterface
	targetWatch <-chan status.UrlMatches
	devices     []veraapi.Device
}

func newVeraAdapter(m *Manager, b base) (sa adapter, err error) {
	// Normal version of the constructor.

	//
	// Look up config values.
	//
	hostname, _, e := b.config.GetString("status://hostname")
	if e != nil {
		return nil, e
	}

	return newVeraAdapterDetailed(m, b, veraapi.NewVeraApi(hostname))
}

func newVeraAdapterDetailed(m *Manager, b base, vera_api veraapi.VeraApiInterface) (sa *veraAdapter, err error) {
	// This version of the constructor gives test code more control.

	watch, err := b.status.WatchForUpdate(b.adapterUrl + "/*/*/*")
	if err != nil {
		return nil, err
	}

	// Create an start adapter.
	sa = &veraAdapter{
		b,
		vera_api,
		watch,
		[]veraapi.Device{},
	}

	go sa.Handler()

	return sa, nil
}

func (a *veraAdapter) Handler() {
	// Create the root for the devices we are about to discover.
	err := a.status.SetJson(a.adapterUrl, []byte(`{}`), status.UNCHECKED_REVISION)
	if err != nil {
		panic(err)
	}

	deviceUpdates := a.VeraApiInterface.Updates()

	for {
		select {
		case devices := <-deviceUpdates:
			log.Printf("Vera: Got devices. %+v\n", devices)
			a.updateDeviceList(devices)

		case matches := <-a.targetWatch:
			// case matches := <-a.targetWatch:
			// Don't log, since this often fires when there is no action to take.
			a.checkForTargetToFire(matches)
			continue

		case <-a.StopChan:
			a.StopChan <- true
			return
		}
	}
}

func (a *veraAdapter) Stop() {
	a.VeraApiInterface.Stop()
	a.status.ReleaseWatch(a.targetWatch)
	a.base.Stop()
}

func (a *veraAdapter) checkForTargetToFire(matches status.UrlMatches) {
	for target_url, raw_value := range matches {

		// If the updated value isn't a control target, ignore it.
		if !strings.HasSuffix(target_url, "_target") {
			continue
		}

		// If the target was updated to 'nil', we can ignore it.
		if raw_value.Value == nil {
			continue
		}

		// XXX Do Something Here.

		// Clear the target value. Again, ignore error. The most likely cause is
		// that someone else updated the target again, which doesn't bother us.
		a.status.Set(target_url, nil, raw_value.Revision)
	}
}

func (a veraAdapter) findDeviceUrl(id int) string {

	// Find the id's of all devices.
	devices_url := a.adapterUrl + "/*/*/id"

	// Find all device URLs, the look for the one we want.
	matches, err := a.status.GetMatchingUrls(devices_url)
	if err != nil {
		// Not possible/hard to handle
		panic(err)
	}

	// Find the device with our requested id.
	for search_url, raw_search_id := range matches {
		search_id, ok := raw_search_id.Value.(int)
		if !ok {
			continue
		}

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

func (a *veraAdapter) updateDeviceList(devices []veraapi.Device) {
OldNames:
	for _, old := range a.devices {
		for _, d := range devices {
			if d.Id == old.Id {
				continue OldNames // This name still exists, leave it.
			}
		}

		old_dev_url := a.findDeviceUrl(old.Id)
		if old_dev_url != "" {
			err := a.status.Remove(old_dev_url, status.UNCHECKED_REVISION)
			if err != nil {
				panic(err)
			}
		}
	}

	// Add/update devices that exist.
	for _, d := range devices {
		a.updateDevice(d)
	}

	a.devices = devices
}

func (a veraAdapter) updateDevice(device veraapi.Device) {
	if device.Category == "" {
		device.Category = "Generic"
	}

	// Add/update devices that exist.
	device_url := a.adapterUrl + "/" + device.Category + "/" + device.Name

	device_values := map[string]interface{}{
		"id":   device.Id,
		"name": device.Name,
	}

	for name, value := range device.Values {
		device_values[name] = value
	}

	err := a.status.Set(device_url, device_values, status.UNCHECKED_REVISION)
	if err != nil {
		panic(err)
	}
}
