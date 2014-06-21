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

func NewAdapterManager(options *options.Options, status *status.Status) (mgr *AdapterManager, e error) {

	// Create the new manager.
	mgr = &AdapterManager{
		status:   status,
		adapters: map[string]Adapter{},
		webUrls:  map[string]Adapter{},
	}

	// Look for adapter configs.
	adapterTypes, e := status.GetChildNames(configUrl)
	if e != nil {
		// If there are no adapters configured, just don't set any up.
		adapterTypes = []string{}
	}

	// Loop through the types of adapters.
	for _, adapterType := range adapterTypes {

		factory, ok := adapterFactories[adapterType]
		if !ok {
			return nil, fmt.Errorf("Adapter: Unknown type: %s.", adapterType)
		}

		adapterNames, e := status.GetChildNames(configUrl + "/" + adapterType)
		if e != nil {
			return nil, e
		}

		// Loop through the adapters of a given type.
		for _, name := range adapterNames {
			adapterConfigUrl := configUrl + "/" + adapterType + "/" + name
			adapterUrl := "status://" + name

			adapterConfig, _, e := status.GetSubStatus(adapterConfigUrl)
			if e != nil {
				return nil, e
			}

			newAdapter, e := factory(mgr, base{
				options:    options,
				status:     status,
				config:     adapterConfig,
				adapterUrl: adapterUrl,
			})
			if e != nil {
				return nil, e
			}

			mgr.adapters[name] = newAdapter
		}
	}

	return mgr, nil
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
