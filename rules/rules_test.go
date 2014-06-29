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

type mockActionHelper struct {
	fireCount int
}

func (m *mockActionHelper) helper(action *status.Status) {
	m.fireCount += 1
}
