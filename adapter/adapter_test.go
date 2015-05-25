package adapter

import (
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/status"
	"github.com/DonGar/go-house/stoppable"
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
      	"config": "./testdata",
        "adapters": {
          "base": {
            "TestBase": {
            }
          },
          "file": {
            "TestFile": {
            },
            "TestFileSpecified": {
              "filename": "TestFile.json"
            }
          },
          "spark": {
            "TestSpark": {
            	"username": "foo",
            	"password": "bar"
            }
          },
          "web": {
            "TestWeb": {
            }
          }
        }
      }
    }`),
		0)
	c.Assert(e, check.IsNil)

	return s
}

func setupTestAdapter(c *check.C, configUrl string, adapterUrl string) (s *status.Status, mgr *Manager, b base) {
	s = setupTestStatus(c)

	config, _, e := s.GetSubStatus(configUrl)
	c.Assert(e, check.IsNil)

	b, e = newBase(s, config, adapterUrl)
	c.Assert(e, check.IsNil)

	// We need just enough of a manager for our tests.
	mgr = &Manager{actionsMgr: actions.NewManager(), webUrls: map[string]adapter{}}

	return
}

func (suite *MySuite) TestBaseStop(c *check.C) {
	s := setupTestStatus(c)

	base := base{stoppable.NewBase(), s, &status.Status{}, "status://TestBase"}
	go base.Handler()

	base.Stop()

	// Make sure status://TestBase is gone.
	v, r, e := s.Get(base.adapterUrl)
	c.Check(v, check.Equals, nil)
	c.Check(r, check.Equals, 2)
	c.Check(e, check.IsNil)
}

func checkAdaptorContents(c *check.C, adaptor *base, jsonValue string) {
	// Let the background routine catchup.
	time.Sleep(time.Microsecond)

	v := adaptor.status.PrettyDump(adaptor.adapterUrl)
	c.Check(string(v), check.DeepEquals, jsonValue)
}
