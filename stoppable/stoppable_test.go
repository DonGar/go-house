package stoppable

import (
	"gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (suite *MySuite) TestStoppableInterface(c *check.C) {
	// Make sure Base conforms to the Stoppable interface.
	var _ Stoppable = &Base{}
}

//
// Compose Base, WITHOUT defining a Handler function.
//
type MinimalStoppable struct {
	Base
}

func NewMinimalStoppable() *MinimalStoppable {
	result := &MinimalStoppable{NewBase()}
	go result.Handler()
	return result
}

func (suite *MySuite) TestStoppableMinimalStartStop(c *check.C) {
	var b Stoppable = NewMinimalStoppable()
	b.Stop()
}

//
// Create a complex type that composes our base type.
//

type TestNormalStoppable struct {
	Base
	foo int
}

func NewTestNormalStoppable() *TestNormalStoppable {
	result := &TestNormalStoppable{NewBase(), 5}
	go result.Handler()
	return result
}

func (b *TestNormalStoppable) Handler() {
	for {
		select {
		//
		// Additional cases related to other channels to watch go here.
		//

		case <-b.StopChan:
			b.foo = 0
			b.StopChan <- true
			return
		}
	}
}

func (suite *MySuite) TestStoppableNormalStartStop(c *check.C) {
	obj := NewTestNormalStoppable()
	var b Stoppable = obj

	// Verify that our custom stop routine has the expected effects.
	c.Check(obj.foo, check.Equals, 5)
	b.Stop()
	c.Check(obj.foo, check.Equals, 0)
}
