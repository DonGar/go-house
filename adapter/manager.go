package adapter

import (
	"fmt"
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"log"
)

type Manager struct {
	status     *status.Status
	actionsMgr *actions.Manager
	adapters   map[string]adapter // Map options.ADAPTERS to Adapter.
	webUrls    map[string]adapter // These are updated directly by WebAdapter.
}

// Map type name to factory method.
var adapterFactories = map[string]newAdapter{
	"base":     newBaseAdapter,
	"file":     newFileAdapter,
	"particle": newParticleAdapter,
	"vera":     newVeraAdapter,
	"web":      newWebAdapter,
}

func NewManager(status *status.Status, actionsMgr *actions.Manager) (mgr *Manager, e error) {
	// Create the new manager.
	mgr = &Manager{status, actionsMgr, map[string]adapter{}, map[string]adapter{}}
	err := mgr.createAdapters()
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

func (m *Manager) createAdapters() error {
	// Look for adapter configs.
	adapterTypes, _, e := m.status.GetChildNames(options.ADAPTERS)
	if e != nil {
		// If there are no adapters configured, just don't set any up.
		adapterTypes = []string{}
	}

	// Loop through the types of adapters.
	for _, adapterType := range adapterTypes {
		adapterNames, _, e := m.status.GetChildNames(options.ADAPTERS + "/" + adapterType)
		if e != nil {
			return e
		}

		// Loop through the adapters of a given type.
		for _, name := range adapterNames {
			newAdapter, e := m.createAdaptor(adapterType, name)
			if e != nil {
				return e
			}

			m.adapters[name] = newAdapter
		}
	}

	return nil
}

func (m *Manager) createAdaptor(adapterType, name string) (adapter, error) {
	adapterConfigUrl := options.ADAPTERS + "/" + adapterType + "/" + name
	adapterUrl := "status://" + name

	factory, ok := adapterFactories[adapterType]
	if !ok {
		return nil, fmt.Errorf("Adapter: Unknown type: %s.", adapterType)
	}

	adapterConfig, _, e := m.status.GetSubStatus(adapterConfigUrl)
	if e != nil {
		return nil, e
	}

	log.Printf("Create Adapter: %s", name)
	b, e := newBase(m.status, adapterConfig, adapterUrl)
	if e != nil {
		return nil, e
	}

	return factory(m, b)
}

func (m *Manager) Stop() {
	for name, adapter := range m.adapters {
		log.Printf("Stop Adapter: %s", name)
		adapter.Stop()
		delete(m.adapters, name)
	}
}

func (m *Manager) WebAdapterStatusUrls() (result []string) {
	result = []string{}

	for k := range m.webUrls {
		result = append(result, k)
	}

	return result
}
