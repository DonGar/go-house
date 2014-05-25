package status

import (
	"gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (suite *MySuite) TestUrlParsing(c *check.C) {
	var path []string
	var e error

	// Test some bad URLs
	path, e = parseUrl("")
	c.Check(path, check.IsNil)
	c.Check(e, check.ErrorMatches, ".*")

	path, e = parseUrl("http://www.google.com/")
	c.Check(path, check.IsNil)
	c.Check(e, check.ErrorMatches, ".*")

	path, e = parseUrl("status")
	c.Check(path, check.IsNil)
	c.Check(e, check.ErrorMatches, ".*")

	path, e = parseUrl("gs://foo/bar")
	c.Check(path, check.IsNil)
	c.Check(e, check.ErrorMatches, ".*")

	path, e = parseUrl("status://foo//bar")
	c.Check(path, check.IsNil)
	c.Check(e, check.ErrorMatches, ".*")

	// Test some good URLs
	path, e = parseUrl("status://")
	c.Check(path, check.DeepEquals, []string{})
	c.Check(e, check.IsNil)

	path, e = parseUrl("status://foo")
	c.Check(path, check.DeepEquals, []string{"foo"})
	c.Check(e, check.IsNil)

	path, e = parseUrl("status://foo/bar/4")
	c.Check(path, check.DeepEquals, []string{"foo", "bar", "4"})
	c.Check(e, check.IsNil)

	path, e = parseUrl("status://foo/bar/")
	c.Check(path, check.DeepEquals, []string{"foo", "bar"})
	c.Check(e, check.IsNil)
}

func (suite *MySuite) TestUrlPathToStatuses(c *check.C) {
	status := Status{}

	// Verify the empty case.
	statuses, e := status.urlPathToStatuses("status://", false)
	c.Check(statuses, check.DeepEquals, []*Status{&status})
	c.Check(e, check.IsNil)

	statuses, e = status.urlPathToStatuses("status://", true)
	c.Check(statuses, check.DeepEquals, []*Status{&status})
	c.Check(e, check.IsNil)

}

func (s *MySuite) TestIdentity(c *check.C) {

	verifyIdentity := func(value interface{}, intermediate *Status) {
		var s *Status
		var e error

		s, e = valueToStatus(22, value)
		c.Check(e, check.IsNil)
		c.Check(s.revision, check.Equals, 22)

		c.Check(s, check.DeepEquals, intermediate)

		var after interface{}
		after, e = s.toValue()
		c.Check(e, check.IsNil)

		c.Check(after, check.DeepEquals, value)
	}

	verifyIdentity(nil, &Status{22, nil})
	verifyIdentity("foo", &Status{22, "foo"})
	verifyIdentity(true, &Status{22, true})
	verifyIdentity(false, &Status{22, false})

	verifyIdentity(
		map[string]interface{}{
			"UTC": 0,
			"EST": -5,
			"CST": -6,
			"MST": -7,
			"PST": -8,
		},
		&Status{22, statusMap{
			"UTC": &Status{22, 0},
			"EST": &Status{22, -5},
			"CST": &Status{22, -6},
			"MST": &Status{22, -7},
			"PST": &Status{22, -8}}})

	verifyIdentity(
		[]interface{}{1, 2, 3, "foo"},
		&Status{22, []statusValue{1, 2, 3, "foo"}},
	)

	verifyIdentity(
		map[string]interface{}{},
		&Status{22, statusMap{}})

	verifyIdentity(
		map[string]interface{}{
			"sub": map[string]interface{}{"subsub1": 1, "subsub2": 2}},
		&Status{22, statusMap{
			"sub": &Status{22, statusMap{
				"subsub1": &Status{22, 1},
				"subsub2": &Status{22, 2}}}}})

	verifyIdentity(
		map[string]interface{}{
			"sub": map[string]interface{}{"subsub": map[string]interface{}{"foo": "bar"}},
		},
		&Status{22, statusMap{
			"sub": &Status{22, statusMap{
				"subsub": &Status{22, statusMap{
					"foo": &Status{22, "bar"}}}}}}})

	verifyIdentity(
		map[string]interface{}{
			"int":    1,
			"string": "foobar",
			"float":  1.0,
			"array":  []interface{}{1, 2, 3, "foo"},
			"nil":    nil,
			"sub":    map[string]interface{}{"subsub1": 1, "subsub2": map[string]interface{}{}},
		},
		&Status{22, statusMap{
			"int":    &Status{22, 1},
			"string": &Status{22, "foobar"},
			"float":  &Status{22, 1.0},
			"array":  &Status{22, []statusValue{1, 2, 3, "foo"}},
			"nil":    &Status{22, nil},
			"sub": &Status{22, statusMap{
				"subsub1": &Status{22, 1},
				"subsub2": &Status{22, statusMap{}}}},
		}})

}
