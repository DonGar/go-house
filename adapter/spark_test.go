package adapter

import (
	"errors"
	"github.com/DonGar/go-house/spark-api"
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

func (s *mockSparkApi) CallFunction(device, function, argument string) error {
	s.actionArgs = mockFunctionCall{device, function, argument}
	return s.actionResult
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
			`"a":{"actions":{},"connected":true,"functions":[],"id":"aaa","last_heard":"date_time","variables":{}},`+
			`"b":{"actions":{"func_a":{"action":"mock_action","argument":"","device":"b","function":"func_a"},"func_b":{"action":"mock_action","argument":"","device":"b","function":"func_b"}},"connected":false,"functions":["func_a","func_b"],"id":"bbb","last_heard":"date_time","variables":{"var1":"val1","var2":2}}`+
			`}}`)

	adaptor.Stop()

	checkAdaptorContents(c, &b, `null`)
}

func (suite *MySuite) TestSparkAdapterAction(c *check.C) {
	s, mgr, b := setupTestAdapter(c,
		"status://server/adapters/spark/TestSpark", "status://TestSpark")

	// Create a spark adapter.
	mock, adaptor := setupSparkAdaptorMockApi(mgr, b)

	mock.devices <- []sparkapi.Device{deviceA, deviceB}

	// Let the background routine catchup.
	time.Sleep(time.Microsecond)

	// Fetch generated action sub statuses for these functions.
	action_a, _, err := s.GetSubStatus(adaptor.adapterUrl + "/core/b/actions/func_a")
	c.Assert(err, check.IsNil)

	action_b, _, err := s.GetSubStatus(adaptor.adapterUrl + "/core/b/actions/func_b")
	c.Assert(err, check.IsNil)

	// Check that a normal calls work for both functions.
	err = adaptor.actionsMgr.FireAction(s, action_a)
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{"b", "func_a", ""})
	c.Check(err, check.IsNil)

	err = adaptor.actionsMgr.FireAction(s, action_b)
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{"b", "func_b", ""})
	c.Check(err, check.IsNil)

	// Check that a failure is returns.
	mock.actionResult = errors.New("Mock Error")
	err = adaptor.actionsMgr.FireAction(s, action_a)
	c.Check(mock.actionArgs, check.DeepEquals, mockFunctionCall{"b", "func_a", ""})
	c.Check(err, check.Equals, mock.actionResult)

	adaptor.Stop()
}
