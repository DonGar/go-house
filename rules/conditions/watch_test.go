package conditions

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
)

func setupWatchCondition(c *check.C, url string) (*status.Status, *watchCondition) {
	s := &status.Status{}

	body := &status.Status{}
	e := body.Set("status://", map[string]interface{}{"watch": url}, 0)
	c.Assert(e, check.IsNil)

	cond, e := newWatchCondition(s, body)
	c.Assert(e, check.IsNil)

	return s, cond
}

func (suite *MySuite) TestWatchStartStop(c *check.C) {
	_, cond := setupWatchCondition(c, "status://foo")
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
	e := body.Set("status://", map[string]interface{}{"watch": "Bad Url"}, 0)
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
