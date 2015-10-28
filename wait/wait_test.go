package wait

import (
	"gopkg.in/check.v1"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (suite *MySuite) TestSuccess(c *check.C) {
	// Test that we exit right away if it passes the first time.
	Wait(time.Second, func() bool { return true })
}
