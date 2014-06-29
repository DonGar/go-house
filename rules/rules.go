package rules

import (
	// "github.com/cpucycle/astrotime"
	"github.com/DonGar/go-house/status"
)

// An interface all rules are expected to implement.
type rule interface {
	Stop() error
	Revision() int
}

// A constructor interface all rules are expected to implement.
type newRule func(base base) (r rule, e error)

// An interface for firing actions that should be provided to all rules.
type actionHelper func(action *status.Status)

// The base type all rules should compose with.
type base struct {
	status       *status.Status
	actionHelper actionHelper
	name         string         // Name of this rule.
	revision     int            // Status revision of the rule definition.
	body         *status.Status // Substatus of the rule definition.
}

func newBaseRule(base base) (rule, error) {
	return &base, nil
}

func (b *base) Revision() int {
	return b.revision
}

func (b *base) Stop() error {
	return nil
}

// Called when the rule decides to fire it's action.
func (b *base) fire() {
	// If the rule has no action, we'll look up a nil.
	action, _, _ := b.body.GetSubStatus("status://action")

	// Perform the action.
	b.actionHelper(action)
}
