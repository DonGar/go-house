package engine

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"time"
)

type mockWatchedItem struct {
}

func newWatched(url string, body *status.Status) (stoppable, error) {
	return &mockWatchedItem{}, nil
}

func (m *mockWatchedItem) Stop() {
}

func (suite *MySuite) TestWatcherStartStopEmpty(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s := &status.Status{}

	factoryAssert := func(url string, body *status.Status) (stoppable, error) {
		c.Error("Unexpected factory call: ", url)
		return nil, nil
	}

	// Create the watcher.
	watcher := newWatcher(s, "status://*", factoryAssert)

	// We give the watcher a little time for a delayed update.
	time.Sleep(1 * time.Millisecond)

	// Stop it.
	watcher.Stop()

	// We verify that there are no active.
	c.Check(len(watcher.active), check.Equals, 0)
}

func (suite *MySuite) TestWatcherAddRule(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s := &status.Status{}

	watcher := newWatcher(s, "status://active/*", newWatched)

	// We give the watcher a little time for a delayed update.
	time.Sleep(1 * time.Millisecond)
	c.Check(len(watcher.active), check.Equals, 0)

	// Force add a new rule.
	s.Set("status://active/Test", "active value", status.UNCHECKED_REVISION)
	time.Sleep(1 * time.Millisecond)
	c.Check(len(watcher.active), check.Equals, 1)

	// Stop it.
	watcher.Stop()

	// We verify that there are no rules.
	c.Check(len(watcher.active), check.Equals, 0)
}

func (suite *MySuite) TestMgrStartEditStop(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s := setupTestStatus(c)

	watcher := newWatcher(s, "status://*/rule/*", newWatched)

	// We give the watcher a little time for a delayed update.
	time.Sleep(1 * time.Millisecond)
	c.Check(len(watcher.active), check.Equals, 2)

	e := s.Set("status://testAdapter/rule/RuleThree", "Booga", 1)
	c.Assert(e, check.IsNil)

	// We give the watcher a little time to finish initializing.
	time.Sleep(1 * time.Millisecond)
	c.Check(len(watcher.active), check.Equals, 3)

	e = s.Remove("status://testAdapter/rule/RuleTwo", 2)
	c.Assert(e, check.IsNil)

	// We give the watcher a little time for a delayed update.
	time.Sleep(1 * time.Millisecond)
	c.Check(len(watcher.active), check.Equals, 2)

	// Stop it.
	watcher.Stop()

	// We verify that there are no rules.
	c.Check(len(watcher.active), check.Equals, 0)
}
