package adapter

import ()

type webAdapter struct {
	base
	adapterMgr *Manager
}

func newWebAdapter(m *Manager, base base) (a adapter, e error) {
	result := &webAdapter{base, m}

	// Remember this adapter in the web URLs section.
	m.webUrls[base.adapterUrl] = result

	go result.Handler()

	return result, nil
}

// Remove this adapter from the web URLs section, the default Stop.
func (a *webAdapter) Stop() {
	delete(a.adapterMgr.webUrls, a.base.adapterUrl)
	a.base.Stop()
}
