package veraapi

import (
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestParseResult(c *check.C) {
	empty := &parseResult{}
	result := newParseResult()

	c.Check(result, check.DeepEquals, result)
	c.Check(result, check.Not(check.DeepEquals), empty)
}

func (suite *MySuite) TestParseResultCopyEmpty(c *check.C) {
	empty := &parseResult{}

	emptyCopy := empty.copy()
	c.Check(emptyCopy, check.DeepEquals, newParseResult())

}

func testCopyHelper(c *check.C, src *parseResult) {
	dest := src.copy()

	c.Check(src, check.Not(check.Equals), dest)
	c.Check(src, check.DeepEquals, dest)
}

func testCopyHelperFromFile(c *check.C, filename string) {
	bodyText := readFile(c, filename)
	result, err := parseVeraData(bodyText, nil)
	c.Assert(err, check.IsNil)

	testCopyHelper(c, result)
}

func (suite *MySuite) TestParseResultCopy(c *check.C) {
	testCopyHelper(c, newParseResult())

	testCopyHelperFromFile(c, MINIMAL_JSON)
	testCopyHelperFromFile(c, SIMPLE_JSON)
	testCopyHelperFromFile(c, FULL_JSON)
	// testCopyHelperFromFile(c, PARTIAL_JSON)
}
