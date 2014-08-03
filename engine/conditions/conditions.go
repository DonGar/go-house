package conditions

import (
	"fmt"
	"github.com/DonGar/go-house/status"
)

type Condition interface {
	Result() <-chan bool
	Stop()
}

// A constructor interface all rules are expected to implement.
func NewCondition(
	s *status.Status,
	body *status.Status) (c Condition, e error) {

	conditionValue, _, e := body.Get("status://")
	if e != nil {
		return nil, e
	}

	switch typedValue := conditionValue.(type) {
	case string:
		// A simple string is a redirect status URL, like an action.
		redirectBody, _, e := s.GetSubStatus(typedValue)
		if e != nil {
			return nil, fmt.Errorf("Condition: url invalid: %s: %s", typedValue, e.Error())
		}
		return NewCondition(s, redirectBody)

	case []interface{}:
		// A array is syntactic sugur for an 'and' condition.
		andBody := &status.Status{}
		e = andBody.Set("status://test", "and", status.UNCHECKED_REVISION)
		if e != nil {
			panic(e)
		}
		e = andBody.Set("status://conditions", conditionValue, status.UNCHECKED_REVISION)
		if e != nil {
			panic(e)
		}
		return NewCondition(s, andBody)

	case map[string]interface{}:
		// We received a dictionary, this is (hopefully) a registered action.
		conditionName, e := body.GetString("status://test")
		if e != nil {
			return nil, fmt.Errorf("Condition: No condition specified: %s", conditionName)
		}

		switch conditionName {
		case "base":
			// This type only exists for basic testing.
			return newBaseCondition(s), nil
		case "after":
			return newAfterCondition(s, body)
		case "and":
			return newAndCondition(s, body)
		case "daily":
			return newDailyCondition(s, body)
		case "periodic":
			return newPeriodicCondition(s, body)
		case "watch":
			return newWatchCondition(s, body)
		default:
			return nil, fmt.Errorf("Condition: No known type: %s", conditionName)
		}

	default:
		return nil, fmt.Errorf("Condition: Doin't understand %#v", conditionValue)
	}
}

// The base type all rules should compose with.
type base struct {
	status *status.Status
	result chan bool // This will be sent on all condition result transitions.
	stop   chan bool
}

func (b *base) Result() <-chan bool {
	return b.result
}

func (b *base) Stop() {
	b.stop <- true
	<-b.stop
}

// This only exists for testing, it should not be used by classes that compose
// base.
func newBaseCondition(s *status.Status) Condition {
	c := &base{s, make(chan bool), make(chan bool)}
	go c.handleEvents()

	return c
}

func (c *base) handleEvents() {
	for {
		select {
		case <-c.stop:
			c.stop <- true
			return
		}
	}
}
