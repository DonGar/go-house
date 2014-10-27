package engine

import (
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

// This creates standard status/options objects used by most Adapters tests.
func setupTestStatus(c *check.C) (s *status.Status) {
	s = &status.Status{}
	e := s.SetJson("status://",
		[]byte(`
    {
      "server": {
        "adapters": {
        }
      },
      "testAdapter": {
      	"rule": {
    			"RuleOne": {
    				"condition": {
    					"test": "base"
    				},
    				"on": null
					},
    			"RuleTwo": {
    				"condition": {
    					"test": "base"
    				},
    				"off": null
					}
      	},
      	"property": {
      	}
      }
    }`),
		0)
	c.Assert(e, check.IsNil)

	return s
}

func (suite *MySuite) TestMgrStartStopEmpty(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s := &status.Status{}

	engine, e := NewEngine(s)
	c.Assert(e, check.IsNil)

	// Stop it.
	engine.Stop()
}

func (suite *MySuite) TestMgrStartStopPopulated(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s := setupTestStatus(c)

	engine, e := NewEngine(s)
	c.Assert(e, check.IsNil)

	// We give the watcher a little time to finish initializing.
	time.Sleep(1 * time.Millisecond)
	c.Check(len(engine.rules.active), check.Equals, 2)

	// Stop it.
	engine.Stop()

	c.Check(len(engine.rules.active), check.Equals, 0)
}
