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
	status := Status{value: statusMap{}}

	// Verify not creating children.
	statuses, e := status.urlPathToStatuses("status://", false)
	c.Check(statuses, check.DeepEquals, []*Status{&status})
	c.Check(e, check.IsNil)

	// Verify creating children.
	statuses, e = status.urlPathToStatuses("status://", true)
	c.Check(statuses, check.DeepEquals, []*Status{&status})
	c.Check(e, check.IsNil)

	statuses, e = status.urlPathToStatuses("status://foo/bar", true)
	c.Check(statuses[0], check.Equals, &status)
	c.Check(statuses[1], check.Equals, statuses[0].value.(statusMap)["foo"])
	// c.Check(statuses[2], check.Equals, statuses[1].value.(statusMap)["bar"])
	c.Check(e, check.IsNil)

}

func (s *MySuite) TestGetSet(c *check.C) {
	status := Status{}

	// Get from empty status.
	value, e := status.Get("status://")
	c.Check(value, check.IsNil)
	c.Check(e, check.IsNil)

	value, e = status.Get("status://foo/bar")
	c.Check(value, check.IsNil)
	c.Check(e, check.NotNil)
	r, e := status.Revision("status://")
	c.Check(r, check.Equals, 0)

	// Set value to empty status
	e = status.Set("status://", 5)
	c.Check(e, check.IsNil)
	r, e = status.Revision("status://")
	c.Check(r, check.Equals, 1)

	value, e = status.Get("status://")
	c.Check(value, check.Equals, 5)
	c.Check(e, check.IsNil)
	r, e = status.Revision("status://")
	c.Check(r, check.Equals, 1)

	// Clear all contents.
	e = status.Set("status://", nil)
	c.Check(e, check.IsNil)
	r, e = status.Revision("status://")
	c.Check(r, check.Equals, 2)

	value, e = status.Get("status://")
	c.Check(value, check.Equals, nil)
	c.Check(e, check.IsNil)
	r, e = status.Revision("status://")
	c.Check(r, check.Equals, 2)

	// Create a sub-path.
	e = status.Set("status://foo", 5)
	c.Check(e, check.IsNil)
	r, e = status.Revision("status://")
	c.Check(r, check.Equals, 3)

	r, e = status.Revision("status://foo")
	c.Check(r, check.Equals, 3)

	value, e = status.Get("status://foo")
	c.Check(value, check.Equals, 5)
	c.Check(e, check.IsNil)

	// Clear all contents.
	e = status.Set("status://", nil)
	c.Check(e, check.IsNil)
	r, e = status.Revision("status://")
	c.Check(r, check.Equals, 4)

	value, e = status.Get("status://")
	c.Check(value, check.Equals, nil)
	c.Check(e, check.IsNil)

	// Create a complex tree.
	e = status.Set("status://sub1/sub2/int", 5)
	c.Check(e, check.IsNil)
	e = status.Set("status://sub1/sub2/float", 2.5)
	c.Check(e, check.IsNil)
	e = status.Set("status://sub1/string", "string value")
	c.Check(e, check.IsNil)
	e = status.Set("status://sub1/sub2/array", []interface{}{1, "foo", 2.5})
	c.Check(e, check.IsNil)
	e = status.Set("status://sub1/sub2/nested", map[string]interface{}{
		"subnested": map[string]interface{}{}})
	c.Check(e, check.IsNil)

	r, e = status.Revision("status://")
	c.Check(r, check.Equals, 9)
	r, e = status.Revision("status://sub1/sub2/nested")
	c.Check(r, check.Equals, 9)
	r, e = status.Revision("status://sub1/sub2/nested/subnested")
	c.Check(r, check.Equals, 9)
	r, e = status.Revision("status://sub1/string")
	c.Check(r, check.Equals, 7)

	// Verify it, one value at a time.
	value, e = status.Get("status://sub1/sub2/int")
	c.Check(e, check.IsNil)
	c.Check(value, check.Equals, 5)
	value, e = status.Get("status://sub1/sub2/float")
	c.Check(e, check.IsNil)
	c.Check(value, check.Equals, 2.5)
	value, e = status.Get("status://sub1/string")
	c.Check(e, check.IsNil)
	c.Check(value, check.Equals, "string value")
	value, e = status.Get("status://sub1/sub2/array")
	c.Check(e, check.IsNil)
	c.Check(value, check.DeepEquals, []interface{}{1, "foo", 2.5})

	value, e = status.Get("status://float")
	c.Check(e, check.NotNil)

	// Verify the whole tree.
	value, e = status.Get("status://")
	c.Check(e, check.IsNil)
	c.Check(value,
		check.DeepEquals,
		map[string]interface{}{
			"sub1": map[string]interface{}{
				"sub2": map[string]interface{}{
					"int":   5,
					"float": 2.5,
					"array": []interface{}{1, "foo", 2.5},
					"nested": map[string]interface{}{
						"subnested": map[string]interface{}{}},
				},
				"string": "string value"}})
}

func (s *MySuite) TestIdentity(c *check.C) {

	verifyIdentity := func(value interface{}) {
		var e error
		var s statusValue
		var after interface{}

		s, e = valueToStatusValue(value, 22)
		c.Check(e, check.IsNil)

		after, e = statusValueToValue(s)
		c.Check(e, check.IsNil)

		c.Check(after, check.DeepEquals, value)
	}

	verifyIdentity(nil)
	verifyIdentity("foo")
	verifyIdentity(true)
	verifyIdentity(false)

	verifyIdentity(
		map[string]interface{}{
			"UTC": 0,
			"EST": -5,
			"CST": -6,
			"MST": -7,
			"PST": -8,
		})

	verifyIdentity([]interface{}{1, 2, 3, "foo"})

	verifyIdentity(map[string]interface{}{})

	verifyIdentity(
		map[string]interface{}{
			"sub": map[string]interface{}{"subsub1": 1, "subsub2": 2}})

	verifyIdentity(
		map[string]interface{}{
			"sub": map[string]interface{}{"subsub": map[string]interface{}{"foo": "bar"}},
		})

	verifyIdentity(
		map[string]interface{}{
			"int":    1,
			"string": "foobar",
			"float":  1.0,
			"array":  []interface{}{1, 2, 3, "foo"},
			"nil":    nil,
			"sub":    map[string]interface{}{"subsub1": 1, "subsub2": map[string]interface{}{}},
		})
}

func (s *MySuite) TestGetMatchingUrls(c *check.C) {
	status := Status{}

	e := status.Set(
		"status://",
		map[string]interface{}{
			"sub1": map[string]interface{}{
				"sub2": map[string]interface{}{
					"int":   5,
					"float": 2.5,
					"array": []interface{}{1, "foo", 2.5},
					"nested": map[string]interface{}{
						"subnested": map[string]interface{}{}},
				},
				"string": "string value"}})
	c.Check(e, check.IsNil)

	var found []string

	found, e = status.GetMatchingUrls("")
	c.Check(e, check.NotNil)

	found, e = status.GetMatchingUrls("status://")
	c.Check(e, check.IsNil)
	c.Check(found, check.DeepEquals, []string{"status://"})

	found, e = status.GetMatchingUrls("status://bogus")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		[]string{})

	found, e = status.GetMatchingUrls("status://bogus/*")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		[]string{})

	found, e = status.GetMatchingUrls("status://sub1")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		[]string{"status://sub1"})

	found, e = status.GetMatchingUrls("status://sub1/string")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		[]string{"status://sub1/string"})

	found, e = status.GetMatchingUrls("status://sub1/*/int")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		[]string{"status://sub1/sub2/int"})

	found, e = status.GetMatchingUrls("status://sub1/*/*")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		[]string{
			"status://sub1/sub2/nested", "status://sub1/sub2/array",
			"status://sub1/sub2/float", "status://sub1/sub2/int"})

	found, e = status.GetMatchingUrls("status://sub1/sub2/array/*")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		[]string{})

}
