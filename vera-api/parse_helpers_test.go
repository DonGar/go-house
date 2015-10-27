package veraapi

import (
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestParseString(c *check.C) {
	testSuccess := func(value string) {
		result, err := parseString(value)
		c.Check(err, check.IsNil)
		c.Check(result, check.Equals, value)
	}

	testFailure := func(value interface{}) {
		result, err := parseString(value)
		c.Check(err, check.NotNil)
		c.Check(result, check.Equals, "")
	}

	testSuccess("I'm a happy string!")

	testFailure(nil)
	testFailure("")
	testFailure(true)
	testFailure(3.1415)
	testFailure([]int{1, 2, 3})
}

func (suite *MySuite) TestParseBool(c *check.C) {
	testSuccess := func(value interface{}, expected bool) {
		result, err := parseBool(value)
		c.Check(err, check.IsNil)
		c.Check(result, check.Equals, expected)
	}

	testFailure := func(value interface{}) {
		result, err := parseBool(value)
		c.Check(err, check.NotNil)
		c.Check(result, check.Equals, false)
	}

	testSuccess(true, true)
	testSuccess("true", true)
	testSuccess(float64(1), true)

	testSuccess(false, false)
	testSuccess("false", false)
	testSuccess(float64(0), false)

	testFailure(nil)
	testFailure("")
	testFailure("I'm a sad string!")
	testFailure(float64(0.1))
	testFailure(float64(5))
	testFailure(float64(-1))
	testFailure(float64(3.1415))
	testFailure([]int{1, 2, 3})
}

func (suite *MySuite) TestParseInt(c *check.C) {
	testSuccess := func(value interface{}, expected int) {
		result, err := parseInt(value)
		c.Check(err, check.IsNil)
		c.Check(result, check.Equals, expected)
	}

	testFailure := func(value interface{}) {
		result, err := parseInt(value)
		c.Check(err, check.NotNil)
		c.Check(result, check.Equals, 0)
	}

	testSuccess(float64(0), 0)
	testSuccess(float64(42), 42)
	testSuccess(float64(8221425010.0), 8221425010)
	testSuccess("0", 0)
	testSuccess("42", 42)
	testSuccess("-3769421", -3769421)

	testFailure(nil)
	testFailure("")
	testFailure(true)
	testFailure(false)
	testFailure("I'm a sad string!")
	testFailure([]int{1, 2, 3})
}

func (suite *MySuite) TestParseFloat(c *check.C) {
	testSuccess := func(value interface{}, expected float64) {
		result, err := parseFloat(value)
		c.Check(err, check.IsNil)
		c.Check(result, check.Equals, expected)
	}

	testFailure := func(value interface{}) {
		result, err := parseFloat(value)
		c.Check(err, check.NotNil)
		c.Check(result, check.Equals, 0.0)
	}

	testSuccess(float64(0), 0)
	testSuccess(float64(42), 42)
	testSuccess(float64(8221425010.0), 8221425010)
	testSuccess("0", 0)
	testSuccess("42", 42)
	testSuccess("-3769421.533395696", -3769421.533395696)

	testFailure(nil)
	testFailure("")
	testFailure(true)
	testFailure(false)
	testFailure("I'm a sad string!")
	testFailure([]int{1, 2, 3})
}

func (suite *MySuite) TestParseInsertRawString(c *check.C) {
	values := ValuesMap{}

	// Success
	err := insertRawString(values, "happy", "I'm a happy string!")
	c.Check(err, check.IsNil)

	// Ignored
	err = insertRawString(values, "nil", nil)
	c.Check(err, check.IsNil)
	err = insertRawString(values, "empty", "")
	c.Check(err, check.IsNil)

	// Failure
	err = insertRawString(values, "true", true)
	c.Check(err, check.NotNil)
	err = insertRawString(values, "float", 3.1415)
	c.Check(err, check.NotNil)
	err = insertRawString(values, "array", []int{1, 2, 3})
	c.Check(err, check.NotNil)

	c.Check(values, check.DeepEquals, ValuesMap{
		"happy": "I'm a happy string!",
	})
}

func (suite *MySuite) TestParseInsertRawBool(c *check.C) {
	values := ValuesMap{}

	// Success
	err := insertRawBool(values, "true", true)
	c.Check(err, check.IsNil)
	err = insertRawBool(values, "trueStr", "true")
	c.Check(err, check.IsNil)
	err = insertRawBool(values, "trueFlt", float64(1))
	c.Check(err, check.IsNil)
	err = insertRawBool(values, "false", false)
	c.Check(err, check.IsNil)

	// Ignored.
	err = insertRawBool(values, "nil", nil)
	c.Check(err, check.IsNil)
	err = insertRawBool(values, "empty", "")
	c.Check(err, check.IsNil)

	// Failure
	err = insertRawBool(values, "sad", "I'm a sad string!")
	c.Check(err, check.NotNil)
	err = insertRawBool(values, "negative", float64(-1))
	c.Check(err, check.NotNil)

	c.Check(values, check.DeepEquals, ValuesMap{
		"true":    true,
		"trueStr": true,
		"trueFlt": true,
		"false":   false,
	})
}

func (suite *MySuite) TestParseInsertRawInt(c *check.C) {
	values := ValuesMap{}

	// Success
	err := insertRawInt(values, "zero", float64(0))
	c.Check(err, check.IsNil)
	err = insertRawInt(values, "42", "42")
	c.Check(err, check.IsNil)

	// Ignored.
	err = insertRawInt(values, "nil", nil)
	c.Check(err, check.IsNil)
	err = insertRawInt(values, "empty", "")
	c.Check(err, check.IsNil)

	// Failure
	err = insertRawInt(values, "true", true)
	c.Check(err, check.NotNil)
	err = insertRawInt(values, "sad", "I'm a sad string!")
	c.Check(err, check.NotNil)

	c.Check(values, check.DeepEquals, ValuesMap{
		"zero": 0,
		"42":   42,
	})
}

func (suite *MySuite) TestParseInsertRawFloat(c *check.C) {
	values := ValuesMap{}

	// Success
	err := insertRawFloat(values, "zero", float64(0))
	c.Check(err, check.IsNil)
	err = insertRawFloat(values, "negative", "-37694.5333")
	c.Check(err, check.IsNil)

	// Ignored.
	err = insertRawFloat(values, "nil", nil)
	c.Check(err, check.IsNil)
	err = insertRawFloat(values, "empty", "")
	c.Check(err, check.IsNil)

	// Failure
	err = insertRawFloat(values, "true", true)
	c.Check(err, check.NotNil)
	err = insertRawFloat(values, "sad", "I'm a sad string!")
	c.Check(err, check.NotNil)

	c.Check(values, check.DeepEquals, ValuesMap{
		"zero":     0.0,
		"negative": -37694.5333,
	})
}
