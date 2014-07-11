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

	watchUrl, e := body.GetString("status://")
	if e == nil {
		_ = watchUrl
		return nil, e
		// watchBody := &status.Status{}
		// _ = watchUrl
		// // Blah, blah, fill in values, TODO
		// return NewCondition(s, watchBody)
	}

	// We received a dictionary, this is (hopefully) a registered action.
	conditionName, e := body.GetString("status://test")
	if e != nil {
		return nil, fmt.Errorf("Condition: No condition specified: %s", conditionName)
	}

	switch conditionName {
	case "base":
		// This type only exists for basic testing.
		return newBaseCondition(), nil
	case "daily":
		return newDailyCondition(s, body)
	case "periodic":
		return newPeriodicCondition(body)
	default:
		return nil, fmt.Errorf("Condition: No known type: %s", conditionName)
	}
}

// The base type all rules should compose with.
type base struct {
	result chan bool // This will be sent on all condition result transitions.
}

func (b *base) Result() <-chan bool {
	return b.result
}

func (b *base) Stop() {
}

func newBaseCondition() Condition {
	return &base{make(chan bool)}
}
