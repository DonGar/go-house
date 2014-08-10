package adapter

import (
	"github.com/DonGar/go-house/status"
	"github.com/DonGar/go-house/stoppable"
)

// Create a standard type all adapters must conform too.
type adapter stoppable.Stoppable

// All Adapters must implement a factory with this signature.
type newAdapter func(m *Manager, base base) (a adapter, e error)

// All Adapters may compose this type for convenience.
type base struct {
	stoppable.Base
	status     *status.Status
	config     *status.Status
	adapterUrl string
}

// This is really only present for testing purposes.
func newBaseAdapter(m *Manager, base base) (a adapter, e error) {
	e = base.status.SetJson(base.adapterUrl, []byte(`{}`), status.UNCHECKED_REVISION)
	if e != nil {
		return nil, e
	}

	go base.Handler()
	return &base, nil
}

// This creates a default Stop method for adapters.
func (b *base) Stop() {
	b.Base.Stop()
	b.status.Set(b.adapterUrl, nil, status.UNCHECKED_REVISION)
}
