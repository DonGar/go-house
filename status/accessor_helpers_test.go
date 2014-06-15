package status

import (
	"gopkg.in/check.v1"
)

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

// TODO: Add tests for other accessors.
