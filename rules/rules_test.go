package rules

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

// This creates standard status/options objects used by most Adapters tests.
func setupTestStatusOptions(c *check.C) (s *status.Status, e error) {
	s = &status.Status{}
	e = s.SetJson("status://",
		[]byte(`
    {
      "server": {
        "adapters": {
        }
      }
    }`),
		0)
	c.Assert(e, check.IsNil)

	return s, nil
}
