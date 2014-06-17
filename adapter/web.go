package adapter

import (
	"github.com/DonGar/go-house/status"
)

type WebAdapter struct {
	base
	adapterMgr *AdapterManager
}

func NewWebAdapter(m *AdapterManager, base base) (a Adapter, e error) {

	e = base.status.SetJson(base.adapterUrl, []byte(`{}`), status.UNCHECKED_REVISION)
	if e != nil {
		return nil, e
	}

	a = &WebAdapter{base, m}

	// Remember this adapter in the web URLs section.
	m.webUrls[base.adapterUrl] = a

	return a, nil
}

// Remove this adapter from the web URLs section, the default Stop.
func (a *WebAdapter) Stop() (e error) {
	delete(a.adapterMgr.webUrls, a.base.adapterUrl)
	return a.base.Stop()
}
