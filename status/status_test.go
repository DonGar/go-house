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

func (suite *MySuite) TestUrlPathToNodes(c *check.C) {
	status := Status{node: node{value: statusMap{}}}

	// Verify not creating children.
	nodes, e := status.urlPathToNodes("status://", false)
	c.Check(nodes, check.DeepEquals, []*node{&status.node})
	c.Check(e, check.IsNil)

	// Verify creating children.
	nodes, e = status.urlPathToNodes("status://", true)
	c.Check(nodes, check.DeepEquals, []*node{&status.node})
	c.Check(e, check.IsNil)

	nodes, e = status.urlPathToNodes("status://foo/bar", true)
	c.Check(nodes[0], check.Equals, &status.node)
	c.Check(nodes[1], check.Equals, nodes[0].value.(statusMap)["foo"])
	c.Check(nodes[2], check.Equals, nodes[1].value.(statusMap)["bar"])
	c.Check(e, check.IsNil)
}

func (s *MySuite) TestGetSet(c *check.C) {
	status := Status{}

	// Get from empty status.
	v, r, e := status.Get("status://")
	c.Check(v, check.IsNil)
	c.Check(r, check.Equals, 0)
	c.Check(e, check.IsNil)

	v, r, e = status.Get("status://foo/bar")
	c.Check(v, check.IsNil)
	c.Check(e, check.NotNil)
	c.Check(r, check.Equals, 0)

	// Set value to empty status
	e = status.Set("status://", 5, 0)
	c.Check(e, check.IsNil)
	v, r, e = status.Get("status://")
	c.Check(v, check.Equals, 5)
	c.Check(r, check.Equals, 1)
	c.Check(e, check.IsNil)

	// Clear all contents.
	e = status.Set("status://", nil, 1)
	c.Check(e, check.IsNil)

	v, r, e = status.Get("status://")
	c.Check(v, check.Equals, nil)
	c.Check(r, check.Equals, 2)
	c.Check(e, check.IsNil)

	// Create a sub-path.
	e = status.Set("status://foo", 5, 2)
	c.Check(e, check.IsNil)

	v, r, e = status.Get("status://")
	c.Check(r, check.Equals, 3)
	c.Check(e, check.IsNil)

	v, r, e = status.Get("status://foo")
	c.Check(v, check.Equals, 5)
	c.Check(r, check.Equals, 3)
	c.Check(e, check.IsNil)

	// Clear all contents.
	e = status.Set("status://", nil, 3)
	c.Check(e, check.IsNil)

	v, r, e = status.Get("status://")
	c.Check(v, check.Equals, nil)
	c.Check(r, check.Equals, 4)
	c.Check(e, check.IsNil)

	// Create a complex tree.
	e = status.Set("status://sub1/sub2/int", 5, 4)
	c.Check(e, check.IsNil)
	e = status.Set("status://sub1/sub2/float", 2.5, 5)
	c.Check(e, check.IsNil)
	e = status.Set("status://sub1/string", "string value", 6)
	c.Check(e, check.IsNil)
	e = status.Set("status://sub1/sub2/array", []interface{}{1, "foo", 2.5}, 6)
	c.Check(e, check.IsNil)
	e = status.Set("status://sub1/sub2/nested",
		map[string]interface{}{"subnested": map[string]interface{}{}},
		8)
	c.Check(e, check.IsNil)

	// Set with a UNCHECKED_REVISION revision (doesn't match, still works)
	e = status.Set("status://unchecked_rev", "value", UNCHECKED_REVISION)
	c.Check(e, check.IsNil)

	v, r, e = status.Get("status://")
	c.Check(r, check.Equals, 10)
	v, r, e = status.Get("status://sub1/sub2/nested")
	c.Check(r, check.Equals, 9)
	v, r, e = status.Get("status://sub1/sub2/nested/subnested")
	c.Check(r, check.Equals, 9)
	v, r, e = status.Get("status://sub1/string")
	c.Check(v, check.Equals, "string value")
	c.Check(r, check.Equals, 7)

	v, r, e = status.Get("status://sub1/sub2/int")
	c.Check(v, check.Equals, 5)
	c.Check(r, check.Equals, 5)

	v, r, e = status.Get("status://sub1/sub2/float")
	c.Check(v, check.Equals, 2.5)
	c.Check(r, check.Equals, 6)

	v, r, e = status.Get("status://sub1/sub2/array")
	c.Check(v, check.DeepEquals, []interface{}{1, "foo", 2.5})
	c.Check(r, check.Equals, 8)

	v, r, e = status.Get("status://float")
	c.Check(e, check.NotNil)

	// Verify the whole tree.
	v, r, e = status.Get("status://")
	c.Check(e, check.IsNil)
	c.Check(r, check.Equals, 10)
	c.Check(
		v,
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
				"string": "string value"},
			"unchecked_rev": "value"})
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

func (s *MySuite) TestJsonIdentity(c *check.C) {
	testStatus := Status{}

	verifySetGet := func(url, stringJson string) {
		rawJson := []byte(stringJson)

		var e error
		var after []byte

		e = testStatus.SetJson(url, rawJson, -1)
		c.Check(e, check.IsNil)

		after, _, e = testStatus.GetJson(url)
		c.Check(e, check.IsNil)
		c.Check(string(after), check.DeepEquals, string(rawJson))
	}

	verifySetGet("status://foo", `null`)
	verifySetGet("status://foo", `"foo"`)
	verifySetGet("status://foo", `true`)
	verifySetGet("status://foo", `false`)
	verifySetGet("status://foo", `{"foo":"bar","int":3}`)
	verifySetGet("status://foo", `[1,2,3,"foo"]`)
	verifySetGet("status://foo", `{}`)
	verifySetGet("status://foo", `{"sub":{"subsub1":1,"subsub2":2}}`)
	verifySetGet("status://foo", `{"sub":{"subsub":{"foo":"bar"}}}`)
	verifySetGet("status://foo", `{"array":[1,2,3,"foo"],"float":1.1,"int":1,"nil":null,"string":"foobar","sub":{"subsub1":1,"subsub2":{}}}`)
}

func (s *MySuite) TestGetMatchingUrls(c *check.C) {
	status := Status{}

	var found UrlMatches
	var e error

	nested_value := map[string]interface{}{
		"subnested": map[string]interface{}{}}

	tree_value := map[string]interface{}{
		"sub1": map[string]interface{}{
			"sub2": map[string]interface{}{
				"int":    5,
				"float":  2.5,
				"array":  []interface{}{1, "foo", 2.5},
				"nested": nested_value,
			},
			"string": "string value"}}

	e = status.Set(
		"status://",
		tree_value,
		0)
	c.Check(e, check.IsNil)

	// Test bad URL.
	found, e = status.getMatchingUrls("")
	c.Check(e, check.NotNil)

	// Test base url.
	found, e = status.getMatchingUrls("status://")
	c.Check(found, check.DeepEquals, UrlMatches{"status://": {1, tree_value}})
	c.Check(e, check.IsNil)

	// Test non-existent url.
	found, e = status.getMatchingUrls("status://bogus")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		UrlMatches{})

	// Test non-existent wildcard.
	found, e = status.getMatchingUrls("status://bogus/*")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		UrlMatches{})

	// Test exact matching url to map.
	found, e = status.getMatchingUrls("status://sub1")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		UrlMatches{"status://sub1": {revision: 1, value: tree_value["sub1"]}})

	// Test exact matching url to value.
	found, e = status.getMatchingUrls("status://sub1/string")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		UrlMatches{"status://sub1/string": {revision: 1, value: "string value"}})

	// Test wildcard matching url to single value.
	found, e = status.getMatchingUrls("status://sub1/*/int")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		UrlMatches{"status://sub1/sub2/int": {revision: 1, value: 5}})

	// Test wildcard matching url to multiple values.
	found, e = status.getMatchingUrls("status://sub1/*/*")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		UrlMatches{
			"status://sub1/sub2/nested": {revision: 1, value: nested_value},
			"status://sub1/sub2/array":  {revision: 1, value: []interface{}{1, "foo", 2.5}},
			"status://sub1/sub2/float":  {revision: 1, value: 2.5},
			"status://sub1/sub2/int":    {revision: 1, value: 5},
		})

	// Test wildcard url to value children.
	found, e = status.getMatchingUrls("status://sub1/sub2/array/*")
	c.Check(e, check.IsNil)
	c.Check(
		found,
		check.DeepEquals,
		UrlMatches{})

}
