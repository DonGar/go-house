package status

import (
	"gopkg.in/check.v1"
	"sort"
)

var URL string = "status://foo"
var BAD_URL string = "/foo"

func (s *MySuite) TestJsonIdentity(c *check.C) {
	testStatus := Status{}

	verifySetGet := func(url, stringJson string) {
		rawJson := []byte(stringJson)

		var e error
		var after []byte

		e = testStatus.SetJson(url, rawJson, UNCHECKED_REVISION)
		c.Check(e, check.IsNil)

		after, _, e = testStatus.GetJson(url)
		c.Check(e, check.IsNil)
		c.Check(string(after), check.DeepEquals, string(rawJson))
	}

	verifySetGet(URL, `null`)
	verifySetGet(URL, `"foo"`)
	verifySetGet(URL, `true`)
	verifySetGet(URL, `false`)
	verifySetGet(URL, `{"foo":"bar","int":3}`)
	verifySetGet(URL, `[1,2,3,"foo"]`)
	verifySetGet(URL, `{}`)
	verifySetGet(URL, `{"sub":{"subsub1":1,"subsub2":2}}`)
	verifySetGet(URL, `{"sub":{"subsub":{"foo":"bar"}}}`)
	verifySetGet(URL, `{"array":[1,2,3,"foo"],"float":1.1,"int":1,"nil":null,"string":"foobar","sub":{"subsub1":1,"subsub2":{}}}`)
}

func (s *MySuite) TestSetGetJson(c *check.C) {
	testStatus := Status{}

	// Good
	e := testStatus.SetJson(URL, []byte(`{"foo":"bar","int":3}`), 0)
	c.Check(e, check.IsNil)

	// Bad Url
	e = testStatus.SetJson(BAD_URL, []byte(`{"foo":"bar","int":3}`), UNCHECKED_REVISION)
	c.Check(e, check.NotNil)

	// Bad Json
	e = testStatus.SetJson(URL, []byte(`{"foo":"bar","int":3,}`), UNCHECKED_REVISION)
	c.Check(e, check.NotNil)

	// Bad Revision
	e = testStatus.SetJson(URL, []byte(`{"foo":"bar","int":3}`), 23)
	c.Check(e, check.NotNil)

	// Good
	value, revision, e := testStatus.GetJson(URL)
	c.Check(string(value), check.Equals, `{"foo":"bar","int":3}`)
	c.Check(revision, check.Equals, 1)
	c.Check(e, check.IsNil)

	// Bad URL
	value, revision, e = testStatus.GetJson(BAD_URL)
	c.Check(value, check.IsNil)
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)
}

func (s *MySuite) TestPrettyDump(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":"bar","int":3}`), 0)
	c.Check(e, check.IsNil)

	value := testStatus.PrettyDump(URL)
	c.Check(value, check.Equals, `{
    "foo": "bar",
    "int": 3
}`)
}

func (s *MySuite) TestGetSubStatus(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":"bar","int":3}`), 0)
	c.Check(e, check.IsNil)

	// Good
	value, revision, e := testStatus.GetSubStatus(URL)
	c.Check(revision, check.Equals, 1)
	c.Check(e, check.IsNil)

	subvalue, revision, e := value.GetJson("status://")
	c.Check(string(subvalue), check.Equals, `{"foo":"bar","int":3}`)
	c.Check(revision, check.Equals, 1)
	c.Check(e, check.IsNil)

	// Bad URL
	value, revision, e = testStatus.GetSubStatus(BAD_URL)
	c.Check(value, check.IsNil)
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)
}

func (s *MySuite) TestGetChildNames(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":"bar","int":3}`), 0)
	c.Check(e, check.IsNil)

	// Good
	value, revision, e := testStatus.GetChildNames(URL)
	sort.Strings(value)
	c.Check(value, check.DeepEquals, []string{"foo", "int"})
	c.Check(revision, check.Equals, 1)
	c.Check(e, check.IsNil)

	// Bad URL
	value, revision, e = testStatus.GetChildNames(BAD_URL)
	c.Check(value, check.IsNil)
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)

	// Non-map URL.
	value, revision, e = testStatus.GetChildNames("status://foo/foo")
	c.Check(value, check.IsNil)
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)
}

func (s *MySuite) TestGetString(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":"bar","int":3}`), 0)
	c.Check(e, check.IsNil)

	// Good
	value, revision, e := testStatus.GetString("status://foo/foo")
	c.Check(value, check.Equals, "bar")
	c.Check(revision, check.Equals, 1)
	c.Check(e, check.IsNil)

	// Bad URL
	value, revision, e = testStatus.GetString(BAD_URL)
	c.Check(value, check.Equals, "")
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)

	// Non-string URL.
	value, revision, e = testStatus.GetString("status://foo/int")
	c.Check(value, check.Equals, "")
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)
}

func (s *MySuite) TestGetStringWithDefault(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":"bar","int":3}`), 0)
	c.Check(e, check.IsNil)

	// Good
	value := testStatus.GetStringWithDefault("status://foo/foo", "def")
	c.Check(value, check.Equals, "bar")

	// Bad URL
	value = testStatus.GetStringWithDefault(BAD_URL, "def")
	c.Check(value, check.Equals, "def")

	// Non-string URL.
	value = testStatus.GetStringWithDefault("status://foo/int", "def")
	c.Check(value, check.Equals, "def")
}

func (s *MySuite) TestGetBool(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":true,"bar":false,"int":3}`), 0)
	c.Check(e, check.IsNil)

	// Good
	value, revision, e := testStatus.GetBool("status://foo/foo")
	c.Check(value, check.Equals, true)
	c.Check(revision, check.Equals, 1)
	c.Check(e, check.IsNil)

	value, revision, e = testStatus.GetBool("status://foo/bar")
	c.Check(value, check.Equals, false)
	c.Check(revision, check.Equals, 1)
	c.Check(e, check.IsNil)

	// Bad URL
	value, revision, e = testStatus.GetBool(BAD_URL)
	c.Check(value, check.Equals, false)
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)

	// Non-bool URL.
	value, revision, e = testStatus.GetBool("status://foo/int")
	c.Check(value, check.Equals, false)
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)
}

func (s *MySuite) TestGetBoolWithDefault(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":true,"int":3}`), 0)
	c.Check(e, check.IsNil)

	// Good
	value := testStatus.GetBoolWithDefault("status://foo/foo", false)
	c.Check(value, check.Equals, true)

	// Bad URL
	value = testStatus.GetBoolWithDefault(BAD_URL, false)
	c.Check(value, check.Equals, false)

	// Non-bool URL.
	value = testStatus.GetBoolWithDefault("status://foo/int", false)
	c.Check(value, check.Equals, false)
}

func (s *MySuite) TestGetInt(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":"bar","int":3,"neg":-25}`), 0)
	c.Check(e, check.IsNil)

	// Good
	value, revision, e := testStatus.GetInt("status://foo/int")
	c.Check(value, check.Equals, 3)
	c.Check(revision, check.Equals, 1)
	c.Check(e, check.IsNil)

	value, revision, e = testStatus.GetInt("status://foo/neg")
	c.Check(value, check.Equals, -25)
	c.Check(revision, check.Equals, 1)
	c.Check(e, check.IsNil)

	// Bad URL
	value, revision, e = testStatus.GetInt(BAD_URL)
	c.Check(value, check.Equals, 0)
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)

	// Non-int URL.
	value, revision, e = testStatus.GetInt("status://foo/foo")
	c.Check(value, check.Equals, 0)
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)
}

func (s *MySuite) TestGetIntWithDefault(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":"bar","int":3}`), 0)
	c.Check(e, check.IsNil)

	// Good
	value := testStatus.GetIntWithDefault("status://foo/int", 2)
	c.Check(value, check.Equals, 3)

	// Bad URL
	value = testStatus.GetIntWithDefault(BAD_URL, 2)
	c.Check(value, check.Equals, 2)

	// Non-int URL.
	value = testStatus.GetIntWithDefault("status://foo/foo", 2)
	c.Check(value, check.Equals, 2)
}

func (s *MySuite) TestGetFloat(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":"bar","float":3.0,"neg":-3.1415}`), 0)
	c.Check(e, check.IsNil)

	// Good
	value, revision, e := testStatus.GetFloat("status://foo/float")
	c.Check(value, check.Equals, 3.0)
	c.Check(revision, check.Equals, 1)
	c.Check(e, check.IsNil)

	value, revision, e = testStatus.GetFloat("status://foo/neg")
	c.Check(value, check.Equals, -3.1415)
	c.Check(revision, check.Equals, 1)
	c.Check(e, check.IsNil)

	// Bad URL
	value, revision, e = testStatus.GetFloat(BAD_URL)
	c.Check(value, check.Equals, 0.0)
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)

	// Non-int URL.
	value, revision, e = testStatus.GetFloat("status://foo/foo")
	c.Check(value, check.Equals, 0.0)
	c.Check(revision, check.Equals, 0)
	c.Check(e, check.NotNil)
}

func (s *MySuite) TestGetFloatWithDefault(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":"bar","float":3.0,"neg":-3.1415}`), 0)
	c.Check(e, check.IsNil)

	// Good
	value := testStatus.GetFloatWithDefault("status://foo/float", 2)
	c.Check(value, check.Equals, 3.0)

	// Bad URL
	value = testStatus.GetFloatWithDefault(BAD_URL, 2)
	c.Check(value, check.Equals, 2.0)

	// Non-int URL.
	value = testStatus.GetFloatWithDefault("status://foo/foo", 2)
	c.Check(value, check.Equals, 2.0)
}

func (s *MySuite) TestGetStrings(c *check.C) {
	testStatus := Status{}
	e := testStatus.SetJson(URL, []byte(`{"foo":"bar","other":"o","int":3}`), 0)
	c.Check(e, check.IsNil)

	// Good
	value, e := testStatus.GetStrings([]string{"status://foo/foo", "status://foo/other"})
	c.Check(value, check.DeepEquals, []string{"bar", "o"})
	c.Check(e, check.IsNil)

	value, e = testStatus.GetStrings([]string{"status://foo/foo"})
	c.Check(value, check.DeepEquals, []string{"bar"})
	c.Check(e, check.IsNil)

	value, e = testStatus.GetStrings([]string{})
	c.Check(value, check.DeepEquals, []string{})
	c.Check(e, check.IsNil)

	// Bad URL
	value, e = testStatus.GetStrings([]string{"status://foo/foo", BAD_URL})
	c.Check(value, check.IsNil)
	c.Check(e, check.NotNil)

	// Non-int URL.
	value, e = testStatus.GetStrings([]string{"status://foo/int"})
	c.Check(value, check.IsNil)
	c.Check(e, check.NotNil)
}
