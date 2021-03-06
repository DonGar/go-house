package status

import (
	"encoding/json"
	"gopkg.in/check.v1"
)

// Compare two UrlMatches structures since DeepEquals can't.
func compareUrlMatches(c *check.C, left, right UrlMatches) {
	c.Assert((left == nil), check.Equals, (right == nil))
	c.Assert(len(left), check.Equals, len(right))

	for k, lValue := range left {
		rValue, rPresent := right[k]
		c.Assert(rPresent, check.Equals, true)
		c.Assert(lValue.Revision, check.Equals, rValue.Revision)

		lJson, e := json.Marshal(lValue.Value)
		if e != nil {
			panic(e)
		}

		rJson, e := json.Marshal(rValue.Value)
		if e != nil {
			panic(e)
		}

		c.Assert(string(lJson), check.DeepEquals, string(rJson))
	}
}

// Validate that a received notification matchers expectations.
func checkPending(c *check.C, wc <-chan UrlMatches, expected UrlMatches) {
	var received UrlMatches

	select {
	case received = <-wc:
		compareUrlMatches(c, received, expected)
	default:
		c.Fatal("Nothing received")
	}
}

// Validate that no notification is pending.
func checkNotPending(c *check.C, wc <-chan UrlMatches) {
	var received UrlMatches

	select {
	case received = <-wc:
		c.Fatal("Unexpected received: ", received)
	default:
	}
}

// Check that we can correctly create watches.
func (s *MySuite) TestWatchForUpdateSetup(c *check.C) {
	status := Status{}

	var e error

	// Make sure we can't watch a bad URL.
	_, e = status.WatchForUpdate("foo")
	c.Check(e, check.NotNil)

	_, e = status.WatchForUpdate("status://sub/*//int")
	c.Check(e, check.NotNil)

	// Make sure we can watch good urls.
	root, e := status.WatchForUpdate("status://")
	c.Check(e, check.IsNil)

	wild, e := status.WatchForUpdate("status://sub/*/int")
	c.Check(e, check.IsNil)

	dup, e := status.WatchForUpdate("status://sub/*/int") // Same URL twice.
	c.Check(e, check.IsNil)

	c.Check(len(status.watchers), check.Equals, 3)

	status.ReleaseWatch(root)
	status.ReleaseWatch(wild)
	status.ReleaseWatch(dup)

	c.Check(len(status.watchers), check.Equals, 0)
}

// Create a root level watch, and verify correct behavior.
func (s *MySuite) TestWatchForUpdateRoot(c *check.C) {
	status := Status{}

	watch, e := status.WatchForUpdate("status://")
	c.Check(e, check.IsNil)

	// Validate internal lastSeen structure on watcher.
	c.Check(
		status.watchers[0].lastSeen,
		check.DeepEquals,
		map[string]int{"status://": 0})

	// We expect to receive the initial match right away.
	checkPending(c, watch,
		UrlMatches{"status://": UrlMatch{}})

	// Then have nothing pending, until we make a change.
	checkNotPending(c, watch)

	e = status.SetJson("status://foo", []byte("1"), UNCHECKED_REVISION)
	c.Check(e, check.IsNil)

	checkPending(c, watch,
		UrlMatches{"status://": UrlMatch{
			Revision: 1, Value: map[string]interface{}{"foo": 1}}})

	// Verify the internal state of the watcher.
	c.Check(
		status.watchers[0].lastSeen,
		check.DeepEquals,
		map[string]int{"status://": 1})

	// Make sure no notifications after we read in changes.
	checkNotPending(c, watch)
}

// Create a deeper watch with wildcards, and verify correct behavior.
func (s *MySuite) TestWatchForUpdateWildcards(c *check.C) {
	status := Status{}

	watch, e := status.WatchForUpdate("status://*/int")
	c.Check(e, check.IsNil)

	// We expect to receive the initial match right away.
	checkPending(c, watch, UrlMatches{})

	// Then have nothing pending, until we make a change.
	checkNotPending(c, watch)

	e = status.SetJson("status://foo", []byte("1"), UNCHECKED_REVISION)
	c.Check(e, check.IsNil)

	checkNotPending(c, watch)

	// Create a matching value, and make sure we are notified.
	e = status.SetJson("status://sub/int", []byte("1"), UNCHECKED_REVISION)
	c.Check(e, check.IsNil)

	checkPending(c, watch,
		UrlMatches{"status://sub/int": UrlMatch{
			Revision: 2, Value: 1}})
	checkNotPending(c, watch)

	// Create a different matching value, and make sure we are notified.
	e = status.SetJson("status://sub2/int", []byte("2"), UNCHECKED_REVISION)
	c.Check(e, check.IsNil)

	checkPending(c, watch,
		UrlMatches{
			"status://sub/int":  UrlMatch{Revision: 2, Value: 1},
			"status://sub2/int": UrlMatch{Revision: 3, Value: 2},
		})
	checkNotPending(c, watch)

	// Update a matching value, and make sure we are notified.
	e = status.SetJson("status://sub/int", []byte("3"), UNCHECKED_REVISION)
	c.Check(e, check.IsNil)

	checkPending(c, watch,
		UrlMatches{
			"status://sub/int":  UrlMatch{Revision: 4, Value: 3},
			"status://sub2/int": UrlMatch{Revision: 3, Value: 2},
		})
	checkNotPending(c, watch)

	// Set un unrelated value, make sure we are not notified.
	e = status.SetJson("status://foo", []byte("3"), UNCHECKED_REVISION)
	c.Check(e, check.IsNil)

	c.Check(
		status.watchers[0].lastSeen,
		check.DeepEquals,
		map[string]int{"status://sub2/int": 3, "status://sub/int": 4})

	checkNotPending(c, watch)

	// Remove an existing match, and make sure we are notified.
	e = status.SetJson("status://sub2", []byte("null"), UNCHECKED_REVISION)
	c.Check(e, check.IsNil)

	checkPending(c, watch,
		UrlMatches{
			"status://sub/int": UrlMatch{Revision: 4, Value: 3},
		})
	checkNotPending(c, watch)
}

// Create a root level watch, and verify correct behavior.
func (s *MySuite) TestNoReadDeadlock(c *check.C) {
	status := Status{}

	_, e := status.WatchForUpdate("status://")
	c.Check(e, check.IsNil)

	// By writing twice, we try to exaust the channel buffer.
	e = status.SetJson("status://foo", []byte("1"), UNCHECKED_REVISION)
	c.Check(e, check.IsNil)
	e = status.SetJson("status://foo", []byte("2"), UNCHECKED_REVISION)
	c.Check(e, check.IsNil)
}
