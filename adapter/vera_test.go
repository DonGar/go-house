package adapter

import (
	// "github.com/DonGar/go-house/status"
	"github.com/DonGar/go-house/vera-api"
	"gopkg.in/check.v1"
)

// func (suite *MySuite) TestVeraAdapterStartStop(c *check.C) {
// 	_, mgr, b := setupTestAdapter(c,
// 		"status://server/adapters/vera/TestVera", "status://TestVera")

// 	// Create a vera adapter.
// 	a, e := newVeraAdapterDetailed(mgr, b, fhc)
// 	c.Assert(e, check.IsNil)

// 	// Make sure empty adaptor contents are created correctly.
// 	checkAdaptorContents(c, &&b, `{}`)

// 	a.Stop()

// 	// Make sure no status contents are left over.
// 	checkAdaptorContents(c, &&b, `null`)
// }

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

func setupVeraAdaptorMockApi(c *check.C) (*mockVeraApi, *veraAdapter) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/vera/TestVera", "status://TestVera")

	mockApi := newMockVeraApi()

	a, err := newVeraAdapterDetailed(mgr, b, mockApi)
	c.Assert(err, check.IsNil)

	return mockApi, a
}

//
// Tests that run against the Mock API.
//

func (suite *MySuite) TestVeraAdapterStartStopMock(c *check.C) {
	// Create a vera adapter.
	_, adaptor := setupVeraAdaptorMockApi(c)
	checkAdaptorContents(c, &adaptor.base, `{}`)
	adaptor.Stop()
	checkAdaptorContents(c, &adaptor.base, `null`)
}

func (suite *MySuite) TestVeraAdapterCreateDevices(c *check.C) {
	deviceA := veraapi.Device{Id: 1, Name: "aaa", Category: "foo"}
	deviceB := veraapi.Device{Id: 2, Name: "bbb", Category: "bar"}

	// Create a vera adapter.
	mock, adaptor := setupVeraAdaptorMockApi(c)
	checkAdaptorContents(c, &adaptor.base, `{}`)

	mock.devices <- []veraapi.Device{deviceA, deviceB}

	checkAdaptorContents(c, &adaptor.base, `{
      "bar": {
          "bbb": {
              "id": 2,
              "name": "bbb"
          }
      },
      "foo": {
          "aaa": {
              "id": 1,
              "name": "aaa"
          }
      }
  }`)

	adaptor.Stop()

	checkAdaptorContents(c, &adaptor.base, `null`)
}
