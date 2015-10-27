package veraapi

import (
	"gopkg.in/check.v1"
)

func (suite *MySuite) TestDeviceCopy(c *check.C) {
	_ = Device{
		// General device description.
		5,
		"Name",
		"Category",
		"Subcategory",
		"Room",

		// Connectivity State
		-1,

		// Values produced by device.
		"Armed",
		"Batterylevel",
		"Status",
		"Temperature",
		"Light",
		"Humidity",
		"Tripped",
		"Armedtripped",
		"Lasttrip",
		"Level",
		"Locked",
	}
}
