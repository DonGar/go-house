package adapter

import (
	"github.com/DonGar/go-house/spark-api"
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

func (suite *MySuite) TestSparkAdapterStartStop(c *check.C) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/spark/TestSpark", "status://TestSpark")

	// Create a spark adapter.
	a, e := newSparkAdapter(mgr, b)
	c.Assert(e, check.IsNil)

	// Make sure empty adaptor contents are created correctly.
	checkAdaptorContents(c, &b, `{"core":{}}`)

	a.Stop()

	// Make sure no status contents are left over.
	checkAdaptorContents(c, &b, `null`)
}

//
// Create a mockSparkApi to help test the adaptor.
//

type mockFunctionCall struct {
	device, function, argument string
}

// Conforms to SparkApiInterface
type mockSparkApi struct {
	devices      chan []sparkapi.Device
	events       chan sparkapi.Event
	actionResult error
	actionArgs   mockFunctionCall
}

func newMockSparkApi() *mockSparkApi {
	return &mockSparkApi{
		make(chan []sparkapi.Device),
		make(chan sparkapi.Event),
		nil,
		mockFunctionCall{},
	}
}

func (s *mockSparkApi) CallFunction(device, function, argument string) (int, error) {
	s.actionArgs = mockFunctionCall{device, function, argument}
	return 0, s.actionResult
}

func (m *mockSparkApi) Updates() (<-chan []sparkapi.Device, <-chan sparkapi.Event) {
	return m.devices, m.events
}

func (m mockSparkApi) Stop() {
}

//
// Helper to setup an adaptor that uses the mock api.
//

func setupSparkAdaptorMockApi(m *Manager, b base) (mock *mockSparkApi, a *sparkAdapter) {

	mockApi := newMockSparkApi()

	// Create an start adapter.
	sa := &sparkAdapter{b, mockApi, "mock_action", m.actionsMgr}
	go sa.Handler()
	return mockApi, sa
}

//
// Tests that run against the Mock API.
//

var deviceA sparkapi.Device = sparkapi.Device{
	"aaa",
	"a",
	"date_time",
	true,
	map[string]interface{}{},
	[]string{},
}

var deviceB sparkapi.Device = sparkapi.Device{
	"bbb",
	"b",
	"date_time",
	false,
	map[string]interface{}{"var1": "val1", "var2": 2},
	[]string{"func_a", "func_b"},
}

func (suite *MySuite) TestSparkAdapterStartStopMock(c *check.C) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/spark/TestSpark", "status://TestSpark")

	// Create a spark adapter.
	mock, adaptor := setupSparkAdaptorMockApi(mgr, b)

	checkAdaptorContents(c, &b, `{"core":{}}`)

	mock.devices <- []sparkapi.Device{deviceA, deviceB}

	checkAdaptorContents(c, &b,
		`{"core":{`+
			`"a":{"connected":true,"functions":[],"id":"aaa","last_heard":"date_time","variables":{}},`+
			`"b":{"connected":false,"functions":["func_a","func_b"],"id":"bbb","last_heard":"date_time","variables":{"var1":"val1","var2":2}}`+
			`}}`)

	adaptor.Stop()

	checkAdaptorContents(c, &b, `null`)
}

func (suite *MySuite) TestSparkAdapterRefreshCalled(c *check.C) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/spark/TestSpark", "status://TestSpark")

	// Create a spark adapter.
	mock, adaptor := setupSparkAdaptorMockApi(mgr, b)

	// Send to devices without a refresh method.
	mock.devices <- []sparkapi.Device{deviceA, deviceB}
	time.Sleep(time.Microsecond)

	// Check that no refresh method was called.
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{})

	// Define an online device with a refresh method.
	deviceRefresh := sparkapi.Device{
		"ccc",
		"c",
		"date_time",
		true,
		map[string]interface{}{"var1": "val1", "var2": 2},
		[]string{"func_a", "refresh"},
	}

	// Send a device with a refresh method.
	mock.devices <- []sparkapi.Device{deviceA, deviceB, deviceRefresh}
	time.Sleep(time.Microsecond)

	// Check that refresh method was called.
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{"c", "refresh", ""})

	// Clear the mock, and send the same devices with connected status unchanged.
	mock.actionArgs = mockFunctionCall{}
	mock.devices <- []sparkapi.Device{deviceA, deviceB, deviceRefresh}
	time.Sleep(time.Microsecond)

	// Check that no refresh method was called.
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{})

	// Take the refresh device offline.
	deviceRefresh.Connected = false
	mock.devices <- []sparkapi.Device{deviceA, deviceB, deviceRefresh}
	time.Sleep(time.Microsecond)

	// Check that no refresh method was called.
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{})

	// Bring the device back online.
	deviceRefresh.Connected = true
	mock.devices <- []sparkapi.Device{deviceA, deviceB, deviceRefresh}
	time.Sleep(time.Microsecond)

	// Check that refresh method was called.
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{"c", "refresh", ""})

	adaptor.Stop()
}

func (suite *MySuite) TestSparkAdapterAction(c *check.C) {

	s, mgr, b := setupTestAdapter(c,
		"status://server/adapters/spark/TestSpark", "status://TestSpark")

	// Create a spark adapter.
	mock, adaptor := setupSparkAdaptorMockApi(mgr, b)

	mock.devices <- []sparkapi.Device{deviceA, deviceB}

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

func (suite *MySuite) TestSparkAdapterEventHandling(c *check.C) {
	_, mgr, b := setupTestAdapter(c,
		"status://server/adapters/spark/TestSpark", "status://TestSpark")

	// Create a spark adapter.
	mock, adaptor := setupSparkAdaptorMockApi(mgr, b)

	checkAdaptorContents(c, &b, `{"core":{}}`)

	// Create a device.
	mock.devices <- []sparkapi.Device{deviceA}

	// Send a valid event for device.
	mock.events <- sparkapi.Event{"standard", "value", "p_date", "aaa"}

	// Send an event for an unknown device (to be ignored).
	mock.events <- sparkapi.Event{"standard", "value", "p_date", "bogus_core_id"}

	checkAdaptorContents(c, &b,
		`{"core":{`+
			`"a":{"connected":true,"events":{"standard":{"data":"value","published":"p_date"}},"functions":[],"id":"aaa","last_heard":"date_time","variables":{}}`+
			`}}`)

	// Update device, verify event is still there.
	mock.devices <- []sparkapi.Device{deviceA}

	checkAdaptorContents(c, &b,
		`{"core":{`+
			`"a":{"connected":true,"events":{"standard":{"data":"value","published":"p_date"}},"functions":[],"id":"aaa","last_heard":"date_time","variables":{}}`+
			`}}`)

	// Update an event value.
	mock.events <- sparkapi.Event{"standard", "updated", "p_date", "aaa"}

	checkAdaptorContents(c, &b,
		`{"core":{`+
			`"a":{"connected":true,"events":{"standard":{"data":"updated","published":"p_date"}},"functions":[],"id":"aaa","last_heard":"date_time","variables":{}}`+
			`}}`)

	// Send a system event to make sure it's ignored.
	mock.events <- sparkapi.Event{"spark/status", "online", "p_date", "aaa"}

	checkAdaptorContents(c, &b,
		`{"core":{`+
			`"a":{"connected":true,"events":{"standard":{"data":"updated","published":"p_date"}},"functions":[],"id":"aaa","last_heard":"date_time","variables":{}}`+
			`}}`)

	adaptor.Stop()

	checkAdaptorContents(c, &b, `null`)
}
