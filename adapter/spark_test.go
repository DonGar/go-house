package adapter

import (
	"github.com/DonGar/go-house/spark-api"
	"gopkg.in/check.v1"
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

// Conforms to SparkApiInterface
type mockSparkApi struct {
	devices chan []sparkapi.Device
	events  chan sparkapi.Event
}

func newMockSparkApi() *mockSparkApi {
	return &mockSparkApi{make(chan []sparkapi.Device), make(chan sparkapi.Event)}
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
	sa := &sparkAdapter{b, mockApi}
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
		`{"core":{"a":{"connected":true,"functions":[],"id":"aaa","last_heard":"date_time","variables":{}},`+
			`"b":{"connected":false,"functions":["func_a","func_b"],"id":"bbb","last_heard":"date_time","variables":{"var1":"val1","var2":2}}}}`)

	adaptor.Stop()

	checkAdaptorContents(c, &b, `null`)
}
