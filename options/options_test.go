package options

import (
	"gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (suite *MySuite) TestFindOptions(c *check.C) {
	var options Options
	var e error

	options, e = FindOptions()
	c.Check(e, check.IsNil)
	c.Check(
		options.ConfigDir,
		check.Matches,
		".*github.com/DonGar/go-house/options/_test")

}
