package particleapi

import (
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestDeviceCopy(c *check.C) {
	device := Device{
		"id",
		"name",
		"lastheard",
		true,
		map[string]interface{}{"foo": "bar"},
		[]string{"func", "func2"},
	}

	deviceCopy := device.Copy()

	// The copy should be a new instance with same contents.
	c.Check(&device, check.Not(check.Equals), &deviceCopy)
	c.Check(device, check.DeepEquals, deviceCopy)
}
