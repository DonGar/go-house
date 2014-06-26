package rules

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
	o = &options.Options{
		ConfigDir: "./testdata",
	}

	s = &status.Status{}
	e = s.SetJson("status://",
		[]byte(`
    {
      "server": {
        "adapters": {
        }
      },
      "testAdapter": {
      	"rules": {
      		"periodic": {
      			"TestIntervalRule": {
							"target": "target",
							"interval": "1s"
						}
      		}
      	}
      }
    }`),
		0)
	c.Assert(e, check.IsNil)

	return o, s, nil
}
