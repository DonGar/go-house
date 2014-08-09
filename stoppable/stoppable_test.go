package stoppable

import (
	"gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

//
// Create a simple type that composes our base type.
//
type TestStoppable struct {
	Base
}

func NewTestStoppable() *TestStoppable {
	result := &TestStoppable{NewBase()}
	go result.Handler()
	return result
}

//
// Create a complex type that composes our base type.
//

type TestComplexStoppable struct {
	Base
	foo int
}

func NewTestComplexStoppable() *TestComplexStoppable {
	result := &TestComplexStoppable{NewBase(), 5}
	go result.Handler()
	return result
}

func (b *TestComplexStoppable) Handler() {
	for {
		select {
		case <-b.StopChan:
			b.foo = 0
			b.StopChan <- true
			return
		}
	}
}

//
// Test Cases
//

func (suite *MySuite) TestBaseStartStop(c *check.C) {
	base := NewBase()
	var b Stoppable = &base

	// base does not start it's background thread, so we do.
	go base.Handler()

	b.Stop()
}

func (suite *MySuite) TestComposedStartStop(c *check.C) {
	var b Stoppable = NewTestStoppable()

	b.Stop()
}

func (suite *MySuite) TestComplexComposedStartStop(c *check.C) {
	obj := NewTestComplexStoppable()
	var b Stoppable = obj

	// Verify that our custom stop routine has the expected effects.
	c.Check(obj.foo, check.Equals, 5)
	b.Stop()
	c.Check(obj.foo, check.Equals, 0)
}
