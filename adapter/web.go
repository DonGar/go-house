package adapter

import (
	"github.com/DonGar/go-house/status"
)

type webAdapter struct {
	base
	adapterMgr *AdapterManager
}

func newWebAdapter(m *AdapterManager, base base) (a adapter, e error) {

	e = base.status.SetJson(base.adapterUrl, []byte(`{}`), status.UNCHECKED_REVISION)
	if e != nil {
		return nil, e
	}

	a = &webAdapter{base, m}

	// Remember this adapter in the web URLs section.
	m.webUrls[base.adapterUrl] = a

	return a, nil
}

// Remove this adapter from the web URLs section, the default Stop.
func (a *webAdapter) Stop() (e error) {
	delete(a.adapterMgr.webUrls, a.base.adapterUrl)
	return a.base.Stop()
}
