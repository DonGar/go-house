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
	result, err := parseVeraData(bodyText)

	c.Check(err, check.NotNil)
	c.Check(result, check.IsNil)
}

func (suite *MySuite) TestParseErrorInvalid(c *check.C) {
	bodyText := []byte("{")
	result, err := parseVeraData(bodyText)

	c.Check(err, check.NotNil)
	c.Check(result, check.IsNil)
}

func (suite *MySuite) TestParseEmpty(c *check.C) {
	bodyText := readFile(c, MINIMAL_JSON)
	result, err := parseVeraData(bodyText)

	c.Check(err, check.IsNil)
	c.Assert(result, check.NotNil)

	c.Check(result.full, check.Equals, true)
	c.Check(result.loadtime, check.Equals, 1446822100)
	c.Check(result.dataversion, check.Equals, 822142501)
	c.Check(result.devices, check.HasLen, 0)
}

func (suite *MySuite) TestParseSimple(c *check.C) {
	bodyText := readFile(c, SIMPLE_JSON)
	result, err := parseVeraData(bodyText)

	c.Check(err, check.IsNil)
	c.Assert(result, check.NotNil)

	c.Check(result.full, check.Equals, true)
	c.Check(result.loadtime, check.Equals, 1446822100)
	c.Check(result.dataversion, check.Equals, 822142501)
	c.Check(result.devices, check.HasLen, 2)

	c.Check(result.devices, check.DeepEquals,
		map[int]Device{
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
	result, err := parseVeraData(bodyText)

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

func (suite *MySuite) TestParsePartial(c *check.C) {
	bodyText := readFile(c, PARTIAL_JSON)
	_, err := parseVeraData(bodyText)

	// For now, ensure we always fail with partial updates.
	c.Check(err, check.NotNil)

	// c.Assert(result, check.NotNil)

	// c.Check(result.full, check.Equals, true)
	// c.Check(result.loadtime, check.Equals, 1455972866)
	// c.Check(result.dataversion, check.Equals, 958637941)
	// c.Check(result.devices, check.HasLen, 54)
}
