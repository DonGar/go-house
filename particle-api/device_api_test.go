package particleapi

import (
	"gopkg.in/check.v1"
)

// I don't (yet ) have a good way to mock out the web
// service, and so can't meaningfully test most behavior.

func (suite *MySuite) TestFindDevice(c *check.C) {

	dev_a := Device{Name: "dev_a"}
	dev_b := Device{Name: "dev_b"}

	sa := ParticleApi{}
	sa.devices = []Device{dev_a, dev_b}

	result := sa.findDevice("nonexistant")
	c.Check(result, check.IsNil)

	result = sa.findDevice("dev_a")
	c.Check(*result, check.DeepEquals, dev_a)

	result = sa.findDevice("dev_b")
	c.Check(*result, check.DeepEquals, dev_b)
}
