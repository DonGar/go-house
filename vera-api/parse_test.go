package veraapi

import (
	"gopkg.in/check.v1"
	"io/ioutil"
)

const (
	MINIMAL_JSON = "./testdata/minimal.json"
	SIMPLE_JSON  = "./testdata/simple.json"
	FULL_JSON    = "./testdata/full.json"
	PARTIAL_JSON = "./testdata/partial.json"
)

func readFile(c *check.C, filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	c.Assert(err, check.IsNil)
	return data
}

func (suite *MySuite) TestParseErrorEmpty(c *check.C) {
	bodyText := []byte("{}")
	result, err := parseVeraData(bodyText, nil)

	c.Check(err, check.NotNil)
	c.Check(result, check.IsNil)
}

func (suite *MySuite) TestParseErrorInvalid(c *check.C) {
	bodyText := []byte("{")
	result, err := parseVeraData(bodyText, nil)

	c.Check(err, check.NotNil)
	c.Check(result, check.IsNil)
}

func (suite *MySuite) TestParseEmpty(c *check.C) {
	bodyText := readFile(c, MINIMAL_JSON)
	result, err := parseVeraData(bodyText, nil)

	c.Check(err, check.IsNil)
	c.Assert(result, check.NotNil)

	c.Check(result.full, check.Equals, true)
	c.Check(result.loadtime, check.Equals, 1446822100)
	c.Check(result.dataversion, check.Equals, 822142501)
	c.Check(result.devices, check.HasLen, 0)
}

func (suite *MySuite) TestParseSimple(c *check.C) {
	bodyText := readFile(c, SIMPLE_JSON)
	result, err := parseVeraData(bodyText, nil)

	c.Check(err, check.IsNil)
	c.Assert(result, check.NotNil)

	c.Check(result.full, check.Equals, true)
	c.Check(result.loadtime, check.Equals, 1446822100)
	c.Check(result.dataversion, check.Equals, 822142501)
	c.Check(result.devices, check.HasLen, 2)

	c.Check(result.devices, check.DeepEquals,
		deviceMap{
			3: Device{
				Id:          3,
				Name:        "Office Light",
				Category:    "On/Off Switch",
				Subcategory: "",
				Room:        "2 Office",
				Values: ValuesMap{
					"status": false,
				},
			},
			4: Device{
				Id:          4,
				Name:        "Fully Populated",
				Category:    "On/Off Switch",
				Subcategory: "",
				Room:        "2 Office",
				Values: ValuesMap{
					"status":       true,
					"level":        65,
					"armed":        true,
					"lasttrip":     "1235763",
					"tripped":      false,
					"humidity":     23,
					"batterylevel": 23,
					"armedtripped": false,
					"temperature":  76.2,
					"light":        78,
					"locked":       false,
					"watts":        1.75,
				},
			},
		})
}

func (suite *MySuite) TestParseFull(c *check.C) {
	bodyText := readFile(c, FULL_JSON)
	result, err := parseVeraData(bodyText, nil)

	c.Check(err, check.IsNil)
	c.Assert(result, check.NotNil)

	c.Check(result.full, check.Equals, true)
	c.Check(result.loadtime, check.Equals, 1455972866)
	c.Check(result.dataversion, check.Equals, 958637939)
	c.Check(result.devices, check.HasLen, 50)

	// Validate a few devices from the list, not everything.
	c.Check(result.devices[22], check.DeepEquals, Device{
		Id:          22,
		Name:        "Bedroom Bath Light",
		Category:    "On/Off Switch",
		Subcategory: "",
		Room:        "2 Bedroom",
		Values: ValuesMap{
			"status": false,
		},
	})

	c.Check(result.devices[33], check.DeepEquals, Device{
		Id:          33,
		Name:        "Kitchen Spots",
		Category:    "On/Off Switch",
		Subcategory: "",
		Room:        "1 Kitchen",
		Values: ValuesMap{
			"status": false,
		},
	})

	c.Check(result.devices[117], check.DeepEquals, Device{
		Id:          117,
		Name:        "CPAP",
		Category:    "On/Off Switch",
		Subcategory: "",
		Room:        "2 Bedroom",
		Values: ValuesMap{
			"status": true,
			"watts":  1.653,
		},
	})
}

func (suite *MySuite) TestParseFullPrevious(c *check.C) {

	bodyText := readFile(c, MINIMAL_JSON)

	// We can parse a full result, with a previous of nil.
	resultNil, err := parseVeraData(bodyText, nil)
	c.Assert(err, check.IsNil)

	// We can parse a full result, with a previous of empty.
	resultNew, err := parseVeraData(bodyText, newParseResult())
	c.Assert(err, check.IsNil)

	// We can parse a full result, with a previous of ourselves.
	resultMinimal, err := parseVeraData(bodyText, resultNil)
	c.Assert(err, check.IsNil)

	bodyFull := readFile(c, FULL_JSON)
	resultFull, err := parseVeraData(bodyFull, nil)
	c.Assert(err, check.IsNil)

	// We can parse a full result, with a previous of different full.
	resultFull, err = parseVeraData(bodyText, resultFull)
	c.Assert(err, check.IsNil)

	// But we always get the same result.
	c.Check(resultNil, check.DeepEquals, resultNew)
	c.Check(resultNil, check.DeepEquals, resultMinimal)
	c.Check(resultNil, check.DeepEquals, resultFull)
}

func (suite *MySuite) TestParsePartialNoPrevious(c *check.C) {
	bodyText := readFile(c, PARTIAL_JSON)

	// Parse partial with previous of nil, and fail.
	result, err := parseVeraData(bodyText, nil)
	c.Check(result, check.IsNil)
	c.Check(err, check.NotNil)

	// Parse partial with previous of empty (!full), and fail.
	result, err = parseVeraData(bodyText, newParseResult())
	c.Check(result, check.IsNil)
	c.Check(err, check.NotNil)
}

func (suite *MySuite) TestParsePartialPrevious(c *check.C) {
	// Parse a full result that the partial goes on top of.
	resultFull, err := parseVeraData(readFile(c, FULL_JSON), nil)
	c.Assert(resultFull, check.NotNil)
	c.Assert(err, check.IsNil)

	// Make a duplicate of the full result.
	resultFullCopy := resultFull.copy()

	// Make sure we can parse on top of the full.
	result, err := parseVeraData(readFile(c, PARTIAL_JSON), resultFullCopy)
	c.Check(err, check.IsNil)
	c.Check(result, check.NotNil)

	// Make sure the full wasn't modified.
	c.Check(resultFull, check.DeepEquals, resultFullCopy)

	// Make sure our partial result != to full it's based on.
	c.Check(resultFull, check.Not(check.DeepEquals), result)
}
