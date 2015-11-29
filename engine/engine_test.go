package engine

import (
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/status"
	"gopkg.in/check.v1"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func bogus_action(am actions.Manager, s *status.Status, action *status.Status) (e error) {
	return nil
}

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
    					"test": "false"
    				},
    				"on": null
					},
    			"RuleTwo": {
    				"condition": {
    					"test": "false"
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

func (suite *MySuite) TestEngineStartStopEmpty(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s := &status.Status{}
	a := actions.NewManager()

	engine, e := NewEngine(s, a)
	c.Assert(e, check.IsNil)

	// Stop it.
	engine.Stop()
}

func (suite *MySuite) TestEngineStartStopPopulated(c *check.C) {
	// Setup a couple of base adapters and verify their contents.
	s := setupTestStatus(c)
	a := actions.NewManager()

	engine, e := NewEngine(s, a)
	c.Assert(e, check.IsNil)

	// We give the watcher a little time to finish initializing.
	time.Sleep(100 * time.Millisecond)
	c.Check(len(engine.rules.active), check.Equals, 2)

	// Stop it.
	engine.Stop()

	c.Check(len(engine.rules.active), check.Equals, 0)
}
