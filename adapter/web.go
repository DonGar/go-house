package adapter

import (
	"github.com/DonGar/go-house/status"
)

type webAdapter struct {
	base
	adapterMgr *Manager
}

func newWebAdapter(m *Manager, base base) (a adapter, e error) {

	e = base.status.SetJson(base.adapterUrl, []byte(`{}`), status.UNCHECKED_REVISION)
	if e != nil {
		return nil, e
	}

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
