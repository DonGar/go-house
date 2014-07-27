package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
)

func setupWatchCondition(c *check.C, url string) (*status.Status, *watchCondition) {
	s := &status.Status{}

	body := &status.Status{}
	e := body.Set("status://watch", url, 0)
	c.Assert(e, check.IsNil)

	cond, e := newWatchCondition(s, body)
	c.Assert(e, check.IsNil)

	return s, cond
}

func setupTriggerWatchCondition(c *check.C, url string, trigger interface{}) (*status.Status, *watchCondition) {
	s := &status.Status{}

	body := &status.Status{}
	e := body.Set("status://watch", url, 0)
	c.Assert(e, check.IsNil)
	e = body.Set("status://trigger", trigger, 1)
	c.Assert(e, check.IsNil)

	cond, e := newWatchCondition(s, body)
	c.Assert(e, check.IsNil)

	return s, cond
}

func (suite *MySuite) TestWatchStartStop(c *check.C) {
	_, cond := setupWatchCondition(c, "status://foo")
	c.Check(cond.hasTrigger, check.Equals, false)
	cond.Stop()
}

func (suite *MySuite) TestTriggerWatchStartStop(c *check.C) {
	_, cond := setupTriggerWatchCondition(c, "status://foo", "foo")
	c.Check(cond.hasTrigger, check.Equals, true)
	cond.Stop()
}

func (suite *MySuite) TestWatchNoUrl(c *check.C) {
	s := &status.Status{}
	body := &status.Status{}

	cond, e := newWatchCondition(s, body)
	c.Assert(e, check.NotNil)
	c.Assert(cond, check.IsNil)
}

func (suite *MySuite) TestWatchBadUrl(c *check.C) {
	s := &status.Status{}

	body := &status.Status{}
	e := body.Set("status://watch", "Bad Url", 0)
	c.Assert(e, check.IsNil)

	cond, e := newWatchCondition(s, body)
	c.Assert(e, check.NotNil)
	c.Assert(cond, check.IsNil)
}

func (suite *MySuite) TestWatchWithUpdates(c *check.C) {
	s, cond := setupWatchCondition(c, "status://foo")

	validateChannelEmpty(c, cond)

	s.Set("status://foo", 1, status.UNCHECKED_REVISION)

	validateChannelRead(c, cond, true)
	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	s.Set("status://foo", 2, status.UNCHECKED_REVISION)

	validateChannelRead(c, cond, true)
	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	cond.Stop()
}

func helperTestTrigger(c *check.C, trigger interface{}, badValues ...interface{}) {
	url := "status://foo"

	s, cond := setupTriggerWatchCondition(c, url, trigger)

	validateChannelEmpty(c, cond)
	validateChannelEmpty(c, cond)

	for _, badValue := range badValues {
		s.Set(url, badValue, status.UNCHECKED_REVISION)
		validateChannelEmpty(c, cond)
	}

	s.Set(url, trigger, status.UNCHECKED_REVISION)
	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	s.Set(url, badValues[0], status.UNCHECKED_REVISION)
	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	for _, badValue := range badValues {
		s.Set(url, badValue, status.UNCHECKED_REVISION)
		validateChannelEmpty(c, cond)
	}

	s.Set(url, trigger, status.UNCHECKED_REVISION)
	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	s.Set(url, trigger, status.UNCHECKED_REVISION)
	validateChannelEmpty(c, cond)

	cond.Stop()
}

func (suite *MySuite) TestWatchWithTriggersInitialTrue(c *check.C) {
	url := "status://foo"
	trigger := true

	s := &status.Status{}
	e := s.Set(url, trigger, 0)
	c.Assert(e, check.IsNil)

	body := &status.Status{}
	e = body.Set("status://watch", url, 0)
	c.Assert(e, check.IsNil)

	e = body.Set("status://trigger", trigger, 1)
	c.Assert(e, check.IsNil)

	cond, e := newWatchCondition(s, body)
	c.Assert(e, check.IsNil)

	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	s.Set(url, false, status.UNCHECKED_REVISION)
	validateChannelRead(c, cond, false)
	validateChannelEmpty(c, cond)

	cond.Stop()
}

func (suite *MySuite) TestWatchWithTriggersInitialFalse(c *check.C) {
	url := "status://foo"
	trigger := true

	s := &status.Status{}
	e := s.Set(url, false, 0)
	c.Assert(e, check.IsNil)

	body := &status.Status{}
	e = body.Set("status://watch", url, 0)
	c.Assert(e, check.IsNil)

	e = body.Set("status://trigger", trigger, 1)
	c.Assert(e, check.IsNil)

	cond, e := newWatchCondition(s, body)
	c.Assert(e, check.IsNil)

	validateChannelEmpty(c, cond)

	s.Set(url, true, status.UNCHECKED_REVISION)
	validateChannelRead(c, cond, true)
	validateChannelEmpty(c, cond)

	cond.Stop()
}

func (suite *MySuite) TestWatchWithTriggers(c *check.C) {
	helperTestTrigger(c, true, false)
	helperTestTrigger(c, 1, 0, 2, 3)
	helperTestTrigger(c, nil, 0, 2, 3, "foo")
	helperTestTrigger(c, "foo", 0, 2, 3, "bar", map[string]interface{}{})
	helperTestTrigger(c, []interface{}{1, 2, "foo"},
		nil, []interface{}{}, []interface{}{1, 2, "bar"}, []interface{}{1, 2, "foo", "extra"})
	helperTestTrigger(c, map[string]interface{}{"foo": "bar"},
		nil, []interface{}{}, map[string]interface{}{"foo": "foo"},
		map[string]interface{}{"foo": "bar", "bar": "foo"})
}

func (suite *MySuite) TestWatchWithPresetValue(c *check.C) {
	s := &status.Status{}
	e := s.Set("status://foo", 0, 0)
	c.Assert(e, check.IsNil)

	body := &status.Status{}
	e = body.Set("status://watch", "status://foo", 0)
	c.Assert(e, check.IsNil)

	cond, e := newWatchCondition(s, body)
	c.Assert(e, check.IsNil)

	// Make sure we don't set true, if the value was set before we start.
	validateChannelEmpty(c, cond)

	cond.Stop()
}
