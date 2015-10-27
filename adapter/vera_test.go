package adapter

import (
	// "github.com/DonGar/go-house/status"
	"github.com/DonGar/go-house/vera-api"
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestVeraAdapterStartStop(c *check.C) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/vera/TestVera", "status://TestVera")

	// Create a vera adapter.
	a, e := newVeraAdapter(mgr, b)
	c.Assert(e, check.IsNil)

	// Make sure empty adaptor contents are created correctly.
	checkAdaptorContents(c, &b, `{}`)

	a.Stop()

	// Make sure no status contents are left over.
	checkAdaptorContents(c, &b, `null`)
}

//
// Create a mockVeraApi to help test the adaptor.
//

// Conforms to VeraApiInterface
type mockVeraApi struct {
	devices      chan []veraapi.Device
	actionResult error
}

func newMockVeraApi() *mockVeraApi {
	return &mockVeraApi{
		make(chan []veraapi.Device),
		nil,
	}
}

func (m *mockVeraApi) Updates() <-chan []veraapi.Device {
	return m.devices
}

func (m mockVeraApi) Stop() {
}

//
// Helper to setup an adaptor that uses the mock api.
//

func setupVeraAdaptorMockApi(m *Manager, b base) (*mockVeraApi, *veraAdapter) {

	mockApi := newMockVeraApi()

	// This must be kept in sync with newVeraAdapter, especially the URL used.
	watch, e := b.status.WatchForUpdate(b.adapterUrl + "/*/*/*")
	if e != nil {
		panic(e)
	}

	// Create an start adapter.
	sa := &veraAdapter{b, mockApi, watch, []veraapi.Device{}}
	go sa.Handler()

	return mockApi, sa
}

//
// Tests that run against the Mock API.
//

func (suite *MySuite) TestVeraAdapterStartStopMock(c *check.C) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/vera/TestVera", "status://TestVera")

	// Create a vera adapter.
	mock, adaptor := setupVeraAdaptorMockApi(mgr, b)

	checkAdaptorContents(c, &b, `{}`)

	deviceA := veraapi.Device{}
	deviceA.Id = 1
	deviceA.Name = "aaa"
	deviceA.Category = "foo"

	deviceB := veraapi.Device{}
	deviceB.Id = 2
	deviceB.Name = "bbb"
	deviceB.Category = "bar"

	mock.devices <- []veraapi.Device{deviceA, deviceB}

	checkAdaptorContents(c, &b, `{
	    "bar": {
	        "bbb": {
	            "category": "bar",
	            "id": 2,
	            "name": "bbb",
	            "state": 0
	        }
	    },
	    "foo": {
	        "aaa": {
	            "category": "foo",
	            "id": 1,
	            "name": "aaa",
	            "state": 0
	        }
	    }
	}`)

	adaptor.Stop()

	checkAdaptorContents(c, &b, `null`)
}
