package adapter

import (
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
)

// All Adapters must confirm to this interface.
type Adapter interface {
	Stop() (e error)
}

// All Adapters must implement a factory with this signature.
type NewAdapter func(base base) (a Adapter, e error)

// All Adapters may compose this type for convenience.
type base struct {
	options    *options.Options
	status     *status.Status
	configUrl  string
	adapterUrl string
}

// This is really only present for testing purposes.
func NewBaseAdapter(base base) (a Adapter, e error) {
	base.status.SetJson(base.adapterUrl, []byte(`{}`), status.UNCHECKED_REVISION)
	return &base, nil
}

// This creates a default Stop method for adapters.
func (ab *base) Stop() (e error) {
	// TODO: Remove fully, when that's supported.
	return ab.status.Set(ab.adapterUrl, nil, status.UNCHECKED_REVISION)
}