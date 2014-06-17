package adapter

import (
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

// This creates standard status/options objects used by most Adapters tests.
func setupTestStatusOptions(c *check.C) (o *options.Options, s *status.Status, e error) {
	s = &status.Status{}
	e = s.SetJson("status://",
		[]byte(`
    {
      "server": {
        "adapters": {
          "TestBase": {
            "type": "base"
          },
          "TestFile": {
            "type": "file"
          },
          "TestFileSpecified": {
            "type": "file",
            "filename": "TestFile.json"
          },
          "TestWeb": {
            "type": "web"
          }
        }
      }
    }`),
		0)
	c.Assert(e, check.IsNil)

	o = &options.Options{
		ConfigDir: "./testdata",
	}

	return o, s, nil
}

func (suite *MySuite) TestBaseStop(c *check.C) {
	o, s, e := setupTestStatusOptions(c)

	base := base{
		status:     s,
		options:    o,
		configUrl:  "status://server/adapters/TestBase",
		adapterUrl: "status://TestBase",
	}

	e = base.Stop()
	c.Check(e, check.IsNil)

	// Make sure status://TestBase is gone.
	v, r, e := s.Get(base.adapterUrl)
	c.Check(v, check.Equals, nil)
	c.Check(r, check.Equals, 2)
	c.Check(e, check.IsNil)
}
