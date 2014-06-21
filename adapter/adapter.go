package adapter

import (
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
)

// All Adapters must confirm to this interface.
type adapter interface {
	Stop() (e error)
}

// All Adapters must implement a factory with this signature.
type newAdapter func(m *AdapterManager, base base) (a adapter, e error)

// All Adapters may compose this type for convenience.
type base struct {
	options    *options.Options
	status     *status.Status
	config     *status.Status
	adapterUrl string
}

// This is really only present for testing purposes.
func newBaseAdapter(m *AdapterManager, base base) (a adapter, e error) {
	e = base.status.SetJson(base.adapterUrl, []byte(`{}`), status.UNCHECKED_REVISION)
	if e != nil {
		return nil, e
	}
	return &base, nil
}

// This creates a default Stop method for adapters.
func (ab *base) Stop() (e error) {
	// TODO: Remove fully, when that's supported.
	return ab.status.Set(ab.adapterUrl, nil, status.UNCHECKED_REVISION)
}
