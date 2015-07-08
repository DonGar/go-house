package adapter

import (
	"github.com/DonGar/go-house/particle-api"
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

func (suite *MySuite) TestParticleAdapterStartStop(c *check.C) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/particle/TestParticle", "status://TestParticle")

	// Create a particle adapter.
	a, e := newParticleAdapter(mgr, b)
	c.Assert(e, check.IsNil)

	// Make sure empty adaptor contents are created correctly.
	checkAdaptorContents(c, &b, `{
    "core": {}
}`)

	a.Stop()

	// Make sure no status contents are left over.
	checkAdaptorContents(c, &b, `null`)
}

//
// Create a mockParticleApi to help test the adaptor.
//

type mockFunctionCall struct {
	device, function, argument string
}

// Conforms to ParticleApiInterface
type mockParticleApi struct {
	devices      chan []particleapi.Device
	events       chan particleapi.Event
	actionResult error
	actionArgs   mockFunctionCall
}

func newMockParticleApi() *mockParticleApi {
	return &mockParticleApi{
		make(chan []particleapi.Device),
		make(chan particleapi.Event),
		nil,
		mockFunctionCall{},
	}
}

func (s *mockParticleApi) CallFunction(device, function, argument string) (int, error) {
	s.actionArgs = mockFunctionCall{device, function, argument}
	return 0, s.actionResult
}

func (s *mockParticleApi) CallFunctionAsync(device, function, argument string) {
	s.actionArgs = mockFunctionCall{device, function, argument}
}

func (m *mockParticleApi) Updates() (<-chan []particleapi.Device, <-chan particleapi.Event) {
	return m.devices, m.events
}

func (m mockParticleApi) Stop() {
}

//
// Helper to setup an adaptor that uses the mock api.
//

func setupParticleAdaptorMockApi(m *Manager, b base) (*mockParticleApi, *particleAdapter) {

	mockApi := newMockParticleApi()

	// This must be kept in sync with newParticleAdapter, especially the URL used.
	watch, e := b.status.WatchForUpdate(b.adapterUrl + "/core/*/*")
	if e != nil {
		panic(e)
	}

	// Create an start adapter.
	sa := &particleAdapter{b, mockApi, "mock_action", m.actionsMgr, watch}
	go sa.Handler()
	return mockApi, sa
}

//
// Tests that run against the Mock API.
//

var deviceA particleapi.Device = particleapi.Device{
	"aaa",
	"a",
	"date_time",
	true,
	map[string]interface{}{},
	[]string{},
}

var deviceFuncs particleapi.Device = particleapi.Device{
	"bbb",
	"b",
	"date_time",
	false,
	map[string]interface{}{"var1": "val1", "var2": 2},
	[]string{"func_a", "prop_target"},
}

func (suite *MySuite) TestParticleAdapterStartStopMock(c *check.C) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/particle/TestParticle", "status://TestParticle")

	// Create a particle adapter.
	mock, adaptor := setupParticleAdaptorMockApi(mgr, b)

	checkAdaptorContents(c, &b, `{
    "core": {}
}`)

	mock.devices <- []particleapi.Device{deviceA, deviceFuncs}

	checkAdaptorContents(c, &b, `{
    "core": {
        "a": {
            "details": {
                "connected": true,
                "functions": [],
                "id": "aaa",
                "last_heard": "date_time",
                "variables": {}
            }
        },
        "b": {
            "details": {
                "connected": false,
                "functions": [
                    "func_a",
                    "prop_target"
                ],
                "id": "bbb",
                "last_heard": "date_time",
                "variables": {
                    "var1": "val1",
                    "var2": 2
                }
            },
            "prop_target": null
        }
    }
}`)

	adaptor.Stop()

	checkAdaptorContents(c, &b, `null`)
}

func (suite *MySuite) TestParticleAdapterRefreshCalled(c *check.C) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/particle/TestParticle", "status://TestParticle")

	// Create a particle adapter.
	mock, adaptor := setupParticleAdaptorMockApi(mgr, b)

	// Send to devices without a refresh method.
	mock.devices <- []particleapi.Device{deviceA, deviceFuncs}
	time.Sleep(time.Microsecond)

	// Check that no refresh method was called.
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{})

	// Define an online device with a refresh method.
	deviceRefresh := particleapi.Device{
		"ccc",
		"c",
		"date_time",
		true,
		map[string]interface{}{"var1": "val1", "var2": 2},
		[]string{"func_a", "refresh"},
	}

	// Send a device with a refresh method.
	mock.devices <- []particleapi.Device{deviceA, deviceFuncs, deviceRefresh}
	time.Sleep(time.Microsecond)

	// Check that refresh method was called.
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{"c", "refresh", ""})

	// Clear the mock, and send the same devices with connected status unchanged.
	mock.actionArgs = mockFunctionCall{}
	mock.devices <- []particleapi.Device{deviceA, deviceFuncs, deviceRefresh}
	time.Sleep(time.Microsecond)

	// Check that no refresh method was called.
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{})

	// Take the refresh device offline.
	deviceRefresh.Connected = false
	mock.devices <- []particleapi.Device{deviceA, deviceFuncs, deviceRefresh}
	time.Sleep(time.Microsecond)

	// Check that no refresh method was called.
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{})

	// Bring the device back online.
	deviceRefresh.Connected = true
	mock.devices <- []particleapi.Device{deviceA, deviceFuncs, deviceRefresh}
	time.Sleep(time.Microsecond)

	// Check that refresh method was called.
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{"c", "refresh", ""})

	adaptor.Stop()
}

func (suite *MySuite) TestParticleAdapterAction(c *check.C) {

	s, mgr, b := setupTestAdapter(c,
		"status://server/adapters/particle/TestParticle", "status://TestParticle")

	// Create a particle adapter.
	mock, adaptor := setupParticleAdaptorMockApi(mgr, b)

	mock.devices <- []particleapi.Device{deviceA, deviceFuncs}

	verifyFailure := func(actionContents map[string]interface{}) {
		// Create action definition.
		action := &status.Status{}
		err := action.Set("status://", actionContents, status.UNCHECKED_REVISION)
		c.Assert(err, check.IsNil)

		// Fire Action
		err = adaptor.actionsMgr.FireAction(s, action)

		c.Check(err, check.NotNil)
	}

	verifySuccess := func(device, function, argument string) {
		// Create action definition.
		action := &status.Status{}
		actionContents := map[string]interface{}{
			"action":   "mock_action",
			"device":   device,
			"function": function,
			"argument": argument,
		}
		err := action.Set("status://", actionContents, status.UNCHECKED_REVISION)
		c.Assert(err, check.IsNil)

		// Fire Action
		err = adaptor.actionsMgr.FireAction(s, action)
		c.Check(err, check.IsNil)
		c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{device, function, argument})
	}

	// Verify assorted failure modes.

	// Bogus action name.
	verifyFailure(map[string]interface{}{
		"action":   "bogus",
		"device":   "dev",
		"function": "func",
		"argument": "arg",
	})

	// No device.
	verifyFailure(map[string]interface{}{
		"action":   "mock_action",
		"function": "func",
		"argument": "arg",
	})

	// No function.
	verifyFailure(map[string]interface{}{
		"action":   "mock_action",
		"device":   "dev",
		"argument": "arg",
	})

	// No argument.
	verifyFailure(map[string]interface{}{
		"action":   "mock_action",
		"device":   "dev",
		"function": "func",
	})

	// Test success cases.
	verifySuccess("dev", "func", "")
	verifySuccess("dev", "func", "arg")
}

func (suite *MySuite) TestParticleAdapterEventHandling(c *check.C) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/particle/TestParticle", "status://TestParticle")

	// Create a particle adapter.
	mock, adaptor := setupParticleAdaptorMockApi(mgr, b)

	checkAdaptorContents(c, &b, `{
    "core": {}
}`)

	adaptor_no_event := `{
    "core": {
        "a": {
            "details": {
                "connected": true,
                "functions": [],
                "id": "aaa",
                "last_heard": "date_time",
                "variables": {}
            }
        }
    }
}`

	adaptor_event_1 := `{
    "core": {
        "a": {
            "details": {
                "connected": true,
                "events": {
                    "standard": {
                        "data": "value",
                        "published": "p_date"
                    }
                },
                "functions": [],
                "id": "aaa",
                "last_heard": "date_time",
                "variables": {}
            },
            "standard": "value"
        }
    }
}`

	adaptor_event_2 := `{
    "core": {
        "a": {
            "details": {
                "connected": true,
                "events": {
                    "standard": {
                        "data": "updated",
                        "published": "p_date"
                    }
                },
                "functions": [],
                "id": "aaa",
                "last_heard": "date_time",
                "variables": {}
            },
            "standard": "updated"
        }
    }
}`

	adaptor_event_json := `{
    "core": {
        "a": {
            "details": {
                "connected": true,
                "events": {
                    "standard": {
                        "data": [
                            1,
                            "1",
                            3.1
                        ],
                        "published": "p_date"
                    }
                },
                "functions": [],
                "id": "aaa",
                "last_heard": "date_time",
                "variables": {}
            },
            "standard": [
                1,
                "1",
                3.1
            ]
        }
    }
}`

	// Create a device.
	mock.devices <- []particleapi.Device{deviceA}
	checkAdaptorContents(c, &b, adaptor_no_event)

	// Send an event for an unknown device (to be ignored).
	mock.events <- particleapi.Event{"standard", "value", "p_date", "bogus_core_id"}
	checkAdaptorContents(c, &b, adaptor_no_event)

	// Send a valid event for device.
	mock.events <- particleapi.Event{"standard", "value", "p_date", "aaa"}
	checkAdaptorContents(c, &b, adaptor_event_1)

	// Update device, verify event is still there.
	mock.devices <- []particleapi.Device{deviceA}
	checkAdaptorContents(c, &b, adaptor_event_1)

	// Update an event value.
	mock.events <- particleapi.Event{"standard", "updated", "p_date", "aaa"}
	checkAdaptorContents(c, &b, adaptor_event_2)

	// Send a system event to make sure it's ignored.
	mock.events <- particleapi.Event{"spark/status", "online", "p_date", "aaa"}
	checkAdaptorContents(c, &b, adaptor_event_2)

	// Send a system event to make sure it's ignored.
	mock.events <- particleapi.Event{"particle/status", "online", "p_date", "aaa"}
	checkAdaptorContents(c, &b, adaptor_event_2)

	// Update an event value.
	mock.events <- particleapi.Event{"standard", "[1, \"1\", 3.1]", "p_date", "aaa"}
	checkAdaptorContents(c, &b, adaptor_event_json)

	adaptor.Stop()
	checkAdaptorContents(c, &b, `null`)
}

func (suite *MySuite) TestParticleAdapterTargetHandling(c *check.C) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/particle/TestParticle", "status://TestParticle")

	// Create a particle adapter.
	mock, adaptor := setupParticleAdaptorMockApi(mgr, b)
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{})

	// Send to devices with a target method.
	mock.devices <- []particleapi.Device{deviceFuncs}
	time.Sleep(time.Microsecond)
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{})

	target_url := adaptor.adapterUrl + "/core/b/prop_target"
	property_url := adaptor.adapterUrl + "/core/b/prop"
	bad_target_url := adaptor.adapterUrl + "/core/b/bogus/prop_target"

	// Set target to nil. Should have no effect.
	mock.actionArgs = mockFunctionCall{}
	err := adaptor.status.Set(target_url, nil, status.UNCHECKED_REVISION)
	c.Assert(err, check.IsNil)
	time.Sleep(time.Microsecond)
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{})

	// Set target to value. Should invoke function.
	mock.actionArgs = mockFunctionCall{}
	err = adaptor.status.Set(target_url, "foo", status.UNCHECKED_REVISION)
	c.Assert(err, check.IsNil)
	time.Sleep(time.Microsecond)

	value, _, err := adaptor.status.Get(target_url)
	c.Check(value, check.IsNil)
	c.Assert(err, check.IsNil)
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{"b", "prop_target", "foo"})

	// Set target to value. Should invoke function.
	mock.actionArgs = mockFunctionCall{}
	err = adaptor.status.Set(target_url, "bar", status.UNCHECKED_REVISION)
	c.Assert(err, check.IsNil)
	time.Sleep(time.Microsecond)

	value, _, err = adaptor.status.Get(target_url)
	c.Check(value, check.IsNil)
	c.Assert(err, check.IsNil)
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{"b", "prop_target", "bar"})

	// Set property to value. Should have no effect.
	mock.actionArgs = mockFunctionCall{}
	err = adaptor.status.Set(property_url, "prop_val", status.UNCHECKED_REVISION)
	c.Assert(err, check.IsNil)
	time.Sleep(time.Microsecond)
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{})

	// Set invalid thing that looks like target to value. Should have no effect.
	mock.actionArgs = mockFunctionCall{}
	err = adaptor.status.Set(bad_target_url, "prop_val", status.UNCHECKED_REVISION)
	c.Assert(err, check.IsNil)
	time.Sleep(time.Microsecond)
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{})

	// Set target to Json value.
	mock.actionArgs = mockFunctionCall{}
	err = adaptor.status.SetJson(target_url, []byte("0"), status.UNCHECKED_REVISION)
	c.Assert(err, check.IsNil)
	time.Sleep(time.Microsecond)

	value, _, err = adaptor.status.Get(target_url)
	c.Check(value, check.IsNil)
	c.Assert(err, check.IsNil)
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{"b", "prop_target", "0"})

	adaptor.Stop()
	checkAdaptorContents(c, &b, `null`)
}
