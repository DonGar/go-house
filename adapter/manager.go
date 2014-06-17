package adapter

import (
	"fmt"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
)

type AdapterManager struct {
	status   *status.Status
	adapters map[string]Adapter // Map configUrl to Adapter.
	webUrls  map[string]Adapter // These are updated directly by WebAdapter.
}

var configUrl = "status://server/adapters"

// Map type name to factory method.
var adapterFactories = map[string]NewAdapter{
	"base": NewBaseAdapter,
	"file": NewFileAdapter,
	"web":  NewWebAdapter,
}

func NewAdapterManager(options *options.Options, status *status.Status) (adapterMgr *AdapterManager, e error) {

	// Create the new manager.
	adapterMgr = &AdapterManager{
		status:   status,
		adapters: map[string]Adapter{},
		webUrls:  map[string]Adapter{},
	}

	// Look for adapter configs.
	childNames, e := status.GetChildNames(configUrl)
	if e != nil {
		return nil, e
	}

	for _, name := range childNames {
		adapterConfigUrl := configUrl + "/" + name
		adapterUrl := "status://" + name

		adapterType, e := status.GetString(adapterConfigUrl + "/type")
		if e != nil {
			return nil, e
		}

		factory, ok := adapterFactories[adapterType]
		if !ok {
			return nil, fmt.Errorf("Adapter: Unknown type: %s.", adapterType)
		}

		newAdapter, e := factory(adapterMgr, base{
			options:    options,
			status:     status,
			configUrl:  adapterConfigUrl,
			adapterUrl: adapterUrl,
		})
		if e != nil {
			return nil, e
		}

		adapterMgr.adapters[name] = newAdapter
	}

	return adapterMgr, nil
}

func (m *AdapterManager) Stop() (e error) {
	for k, v := range m.adapters {
		e = v.Stop()
		if e != nil {
			return e
		}
		delete(m.adapters, k)
	}
	return nil
}

func (m *AdapterManager) WebAdapterStatusUrls() (result []string) {
	result = []string{}

	for k := range m.webUrls {
		result = append(result, k)
	}

	return result
}
